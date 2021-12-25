package client

//go:generate go run github.com/Mrs4s/MiraiGo/internal/generator/c2c_switcher

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

type (
	TempSessionInfo struct {
		Source    TempSessionSource
		GroupCode int64
		Sender    int64

		sig    []byte
		client *QQClient
	}
	TempSessionSource int
)

const (
	GroupSource         TempSessionSource = iota // 来自群聊
	ConsultingSource                             // 来自QQ咨询
	SearchSource                                 // 来自查找
	MovieSource                                  // 来自QQ电影
	HotChatSource                                // 来自热聊
	SystemMessageSource                          // 来自验证消息
	MultiChatSource                              // 来自多人聊天
	DateSource                                   // 来自约会
	AddressBookSource                            // 来自通讯录
)

func (c *QQClient) c2cMessageSyncProcessor(rsp *msg.GetMessageResponse, resp *network.Response) {
	c.sig.SyncCookie = rsp.SyncCookie
	c.sig.PubAccountCookie = rsp.PubAccountCookie
	// c.msgCtrlBuf = rsp.MsgCtrlBuf
	if rsp.UinPairMsgs == nil {
		return
	}
	var delItems []*pb.MessageItem
	for _, pairMsg := range rsp.UinPairMsgs {
		for _, pMsg := range pairMsg.Messages {
			// delete message
			delItem := &pb.MessageItem{
				FromUin: pMsg.Head.GetFromUin(),
				ToUin:   pMsg.Head.GetToUin(),
				MsgType: pMsg.Head.GetMsgType(),
				MsgSeq:  pMsg.Head.GetMsgSeq(),
				MsgUid:  pMsg.Head.GetMsgUid(),
			}
			delItems = append(delItems, delItem)
			if pMsg.Head.GetToUin() != c.Uin {
				continue
			}
			if (int64(pairMsg.GetLastReadTime()) & 4294967295) > int64(pMsg.Head.GetMsgTime()) {
				continue
			}
			c.commMsgProcessor(pMsg, resp)
		}
	}
	if delItems != nil {
		_, _ = c.call(c.buildDeleteMessageRequestPacket(delItems))
	}
	if rsp.GetSyncFlag() != msg.SyncFlag_STOP {
		c.Debug("continue sync with flag: %v", rsp.SyncFlag)
		req := c.buildGetMessageRequest(rsp.GetSyncFlag(), time.Now().Unix())
		req.Params = resp.Params
		_, _ = c.callAndDecode(req, decodeMessageSvcPacket)
	}
}

func (c *QQClient) commMsgProcessor(pMsg *msg.Message, resp *network.Response) {
	strKey := fmt.Sprintf("%d%d%d%d", pMsg.Head.GetFromUin(), pMsg.Head.GetToUin(), pMsg.Head.GetMsgSeq(), pMsg.Head.GetMsgUid())
	if _, ok := c.msgSvcCache.GetAndUpdate(strKey, time.Hour); ok {
		c.Debug("c2c msg %v already exists in cache. skip.", pMsg.Head.GetMsgUid())
		return
	}
	c.msgSvcCache.Add(strKey, "", time.Hour)
	if c.lastC2CMsgTime > int64(pMsg.Head.GetMsgTime()) && (c.lastC2CMsgTime-int64(pMsg.Head.GetMsgTime())) > 60*10 {
		c.Debug("c2c msg filtered by time. lastMsgTime: %v  msgTime: %v", c.lastC2CMsgTime, pMsg.Head.GetMsgTime())
		return
	}
	c.lastC2CMsgTime = int64(pMsg.Head.GetMsgTime())
	if resp.Params.Bool("init") {
		return
	}
	if decoder, _ := peekC2CDecoder(pMsg.Head.GetMsgType()); decoder != nil {
		decoder(c, pMsg, resp)
	} else {
		c.Debug("unknown msg type on c2c processor: %v - %v", pMsg.Head.GetMsgType(), pMsg.Head.GetC2CCmd())
	}
}

func privateMessageDecoder(c *QQClient, pMsg *msg.Message, _ *network.Response) {
	switch pMsg.Head.GetC2CCmd() {
	case 11, 175: // friend msg
		if pMsg.Head.GetFromUin() == c.Uin {
			for {
				frdSeq := c.friendSeq.Load()
				if frdSeq < pMsg.Head.GetMsgSeq() {
					if c.friendSeq.CAS(frdSeq, pMsg.Head.GetMsgSeq()) {
						break
					}
				} else {
					break
				}
			}
		}
		if pMsg.Body.RichText == nil || pMsg.Body.RichText.Elems == nil {
			return
		}
		if pMsg.Head.GetFromUin() == c.Uin {
			c.dispatchPrivateMessageSelf(c.parsePrivateMessage(pMsg))
			return
		}
		c.dispatchPrivateMessage(c.parsePrivateMessage(pMsg))
	default:
		c.Debug("unknown c2c cmd on private msg decoder: %v", pMsg.Head.GetC2CCmd())
	}
}

func privatePttDecoder(c *QQClient, pMsg *msg.Message, _ *network.Response) {
	if pMsg.Body == nil || pMsg.Body.RichText == nil || pMsg.Body.RichText.Ptt == nil {
		return
	}
	if len(pMsg.Body.RichText.Ptt.Reserve) != 0 {
		// m := binary.NewReader(pMsg.Body.RichText.Ptt.Reserve[1:]).ReadTlvMap(1)
		// T3 -> timestamp T8 -> voiceType T9 -> voiceLength T10 -> PbReserveStruct
	}
	c.dispatchPrivateMessage(c.parsePrivateMessage(pMsg))
}

func tempSessionDecoder(c *QQClient, pMsg *msg.Message, _ *network.Response) {
	if pMsg.Head.C2CTmpMsgHead == nil || pMsg.Body == nil {
		return
	}
	if (pMsg.Head.GetMsgType() == 529 && pMsg.Head.GetC2CCmd() == 6) || pMsg.Body.RichText != nil {
		genTempSessionInfo := func() *TempSessionInfo {
			if pMsg.Head.C2CTmpMsgHead.GetServiceType() == 0 {
				group := c.FindGroup(pMsg.Head.C2CTmpMsgHead.GetGroupCode())
				if group == nil {
					return nil
				}
				return &TempSessionInfo{
					Source:    GroupSource,
					GroupCode: group.Code,
					Sender:    pMsg.Head.GetFromUin(),
					client:    c,
				}
			}
			info := &TempSessionInfo{
				Source: 0,
				Sender: pMsg.Head.GetFromUin(),
				sig:    pMsg.Head.C2CTmpMsgHead.GetSig(),
				client: c,
			}

			switch pMsg.Head.C2CTmpMsgHead.GetServiceType() {
			case 1:
				info.Source = MultiChatSource
			case 130:
				info.Source = AddressBookSource
			case 132:
				info.Source = HotChatSource
			case 134:
				info.Source = SystemMessageSource
			case 201:
				info.Source = ConsultingSource
			default:
				return nil
			}
			return info
		}
		session := genTempSessionInfo()
		if session == nil {
			return
		}
		/*
			group := c.FindGroup(pMsg.Head.C2CTmpMsgHead.GetGroupCode())
			if group == nil {
				return
			}
		*/
		if pMsg.Head.GetFromUin() == c.Uin {
			return
		}
		c.dispatchTempMessage(&TempMessageEvent{
			Message: c.parseTempMessage(pMsg),
			Session: session,
		})
	}
}

func troopAddMemberBroadcastDecoder(c *QQClient, pMsg *msg.Message, resp *network.Response) {
	groupJoinLock.Lock()
	defer groupJoinLock.Unlock()
	group := c.FindGroupByUin(pMsg.Head.GetFromUin())
	if pMsg.Head.GetAuthUin() == c.Uin {
		if group == nil && c.ReloadGroupList() == nil {
			c.dispatchJoinGroupEvent(c.FindGroupByUin(pMsg.Head.GetFromUin()))
		}
	} else {
		if group != nil && group.FindMember(pMsg.Head.GetAuthUin()) == nil {
			mem, err := c.GetMemberInfo(group.Code, pMsg.Head.GetAuthUin())
			if err != nil {
				c.Debug("error to fetch new member info: %v", err)
				return
			}
			group.Update(func(info *GroupInfo) {
				info.Members = append(info.Members, mem)
				info.sort()
			})
			c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
				Group:  group,
				Member: mem,
			})
		}
	}
}

func systemMessageDecoder(c *QQClient, _ *msg.Message, _ *network.Response) {
	_, _ = c.call(c.buildSystemMsgNewFriendRequest())
}

func troopSystemMessageDecoder(c *QQClient, pMsg *msg.Message, info *network.Response) {
	if !info.Params.Bool("used_reg_proxy") && pMsg.Head.GetMsgType() != 85 && pMsg.Head.GetMsgType() != 36 {
		c.exceptAndDispatchGroupSysMsg()
	}
	if len(pMsg.Body.GetMsgContent()) == 0 {
		return
	}
	reader := binary.NewReader(pMsg.GetBody().GetMsgContent())
	groupCode := uint32(reader.ReadInt32())
	if info := c.FindGroup(int64(groupCode)); info != nil && pMsg.Head.GetGroupName() != "" && info.Name != pMsg.Head.GetGroupName() {
		c.Debug("group %v name updated. %v -> %v", groupCode, info.Name, pMsg.Head.GetGroupName())
		info.Name = pMsg.Head.GetGroupName()
	}
}

func msgType0x211Decoder(c *QQClient, pMsg *msg.Message, info *network.Response) {
	if pMsg.Head.GetC2CCmd() == 6 || pMsg.Head.C2CTmpMsgHead != nil {
		tempSessionDecoder(c, pMsg, info)
	}
	sub4 := msg.SubMsgType0X4Body{}
	if err := proto.Unmarshal(pMsg.Body.MsgContent, &sub4); err != nil {
		err = errors.Wrap(err, "unmarshal sub msg 0x4 error")
		c.Error("unmarshal sub msg 0x4 error: %v", err)
		return
	}
	if sub4.NotOnlineFile != nil && sub4.NotOnlineFile.GetSubcmd() == 1 { // subcmd: 1 -> sendPacket, 2-> recv
		rsp, err := c.callAndDecode(c.buildOfflineFileDownloadRequestPacket(sub4.NotOnlineFile.FileUuid),
			decodeOfflineFileDownloadResponse) // offline_file.go
		if err != nil {
			return
		}
		c.dispatchOfflineFileEvent(&OfflineFileEvent{
			FileName:    string(sub4.NotOnlineFile.FileName),
			FileSize:    sub4.NotOnlineFile.GetFileSize(),
			Sender:      pMsg.Head.GetFromUin(),
			DownloadUrl: rsp.(string),
		})
	}
}
