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

func (c *QQClient) c2cMessageSyncProcessor(rsp *msg.GetMessageResponse, info *network.IncomingPacketInfo) {
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
				FromUin: pMsg.Head.FromUin.Unwrap(),
				ToUin:   pMsg.Head.ToUin.Unwrap(),
				MsgType: pMsg.Head.MsgType.Unwrap(),
				MsgSeq:  pMsg.Head.MsgSeq.Unwrap(),
				MsgUid:  pMsg.Head.MsgUid.Unwrap(),
			}
			delItems = append(delItems, delItem)
			if pMsg.Head.ToUin.Unwrap() != c.Uin {
				continue
			}
			if (int64(pairMsg.LastReadTime.Unwrap()) & 4294967295) > int64(pMsg.Head.MsgTime.Unwrap()) {
				continue
			}
			c.commMsgProcessor(pMsg, info)
		}
	}
	if delItems != nil {
		_, _ = c.sendAndWait(c.buildDeleteMessageRequestPacket(delItems))
	}
	if rsp.SyncFlag.Unwrap() != msg.SyncFlag_STOP {
		c.debug("continue sync with flag: %v", rsp.SyncFlag)
		seq, pkt := c.buildGetMessageRequestPacket(rsp.SyncFlag.Unwrap(), time.Now().Unix())
		_, _ = c.sendAndWait(seq, pkt, info.Params)
	}
}

func (c *QQClient) commMsgProcessor(pMsg *msg.Message, info *network.IncomingPacketInfo) {
	strKey := fmt.Sprintf("%d%d%d%d", pMsg.Head.FromUin.Unwrap(), pMsg.Head.ToUin.Unwrap(), pMsg.Head.MsgSeq.Unwrap(), pMsg.Head.MsgUid.Unwrap())
	if _, ok := c.msgSvcCache.GetAndUpdate(strKey, time.Hour); ok {
		c.debug("c2c msg %v already exists in cache. skip.", pMsg.Head.MsgUid.Unwrap())
		return
	}
	c.msgSvcCache.Add(strKey, unit{}, time.Hour)
	if c.lastC2CMsgTime > int64(pMsg.Head.MsgTime.Unwrap()) && (c.lastC2CMsgTime-int64(pMsg.Head.MsgTime.Unwrap())) > 60*10 {
		c.debug("c2c msg filtered by time. lastMsgTime: %v  msgTime: %v", c.lastC2CMsgTime, pMsg.Head.MsgTime.Unwrap())
		return
	}
	c.lastC2CMsgTime = int64(pMsg.Head.MsgTime.Unwrap())
	if info.Params.Bool("init") {
		return
	}
	if decoder, _ := peekC2CDecoder(pMsg.Head.MsgType.Unwrap()); decoder != nil {
		decoder(c, pMsg, info)
	} else {
		c.debug("unknown msg type on c2c processor: %v - %v", pMsg.Head.MsgType.Unwrap(), pMsg.Head.C2CCmd.Unwrap())
	}
}

func privateMessageDecoder(c *QQClient, pMsg *msg.Message, _ *network.IncomingPacketInfo) {
	switch pMsg.Head.C2CCmd.Unwrap() {
	case 11, 175: // friend msg
		if pMsg.Head.FromUin.Unwrap() == c.Uin {
			for {
				frdSeq := c.friendSeq.Load()
				if frdSeq < pMsg.Head.MsgSeq.Unwrap() {
					if c.friendSeq.CompareAndSwap(frdSeq, pMsg.Head.MsgSeq.Unwrap()) {
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

		// handle fragmented message
		if pMsg.Content != nil && pMsg.Content.PkgNum.Unwrap() > 1 {
			seq := pMsg.Content.DivSeq.Unwrap()
			builder := c.messageBuilder(seq)
			builder.append(pMsg)
			if builder.len() < pMsg.Content.PkgNum.Unwrap() {
				// continue to receive other fragments
				return
			}
			c.msgBuilders.Delete(seq)
			pMsg = builder.build()
		}

		if pMsg.Head.FromUin.Unwrap() == c.Uin {
			c.SelfPrivateMessageEvent.dispatch(c, c.parsePrivateMessage(pMsg))
			return
		}
		c.PrivateMessageEvent.dispatch(c, c.parsePrivateMessage(pMsg))
	default:
		c.debug("unknown c2c cmd on private msg decoder: %v", pMsg.Head.C2CCmd.Unwrap())
	}
}

func privatePttDecoder(c *QQClient, pMsg *msg.Message, _ *network.IncomingPacketInfo) {
	if pMsg.Body == nil || pMsg.Body.RichText == nil || pMsg.Body.RichText.Ptt == nil {
		return
	}
	if len(pMsg.Body.RichText.Ptt.Reserve) != 0 {
		// m := binary.NewReader(pMsg.Body.RichText.Ptt.Reserve[1:]).ReadTlvMap(1)
		// T3 -> timestamp T8 -> voiceType T9 -> voiceLength T10 -> PbReserveStruct
	}
	c.PrivateMessageEvent.dispatch(c, c.parsePrivateMessage(pMsg))
}

func tempSessionDecoder(c *QQClient, pMsg *msg.Message, _ *network.IncomingPacketInfo) {
	if pMsg.Head.C2CTmpMsgHead == nil || pMsg.Body == nil {
		return
	}
	if (pMsg.Head.MsgType.Unwrap() == 529 && pMsg.Head.C2CCmd.Unwrap() == 6) || pMsg.Body.RichText != nil {
		genTempSessionInfo := func() *TempSessionInfo {
			if pMsg.Head.C2CTmpMsgHead.ServiceType.Unwrap() == 0 {
				group := c.FindGroup(pMsg.Head.C2CTmpMsgHead.GroupCode.Unwrap())
				if group == nil {
					return nil
				}
				return &TempSessionInfo{
					Source:    GroupSource,
					GroupCode: group.Code,
					Sender:    pMsg.Head.FromUin.Unwrap(),
					client:    c,
				}
			}
			info := &TempSessionInfo{
				Source: 0,
				Sender: pMsg.Head.FromUin.Unwrap(),
				sig:    pMsg.Head.C2CTmpMsgHead.Sig,
				client: c,
			}

			switch pMsg.Head.C2CTmpMsgHead.ServiceType.Unwrap() {
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
			group := c.FindGroup(pMsg.Head.C2CTmpMsgHead.GroupCode.Unwrap())
			if group == nil {
				return
			}
		*/
		if pMsg.Head.FromUin.Unwrap() == c.Uin {
			return
		}
		c.TempMessageEvent.dispatch(c, &TempMessageEvent{
			Message: c.parseTempMessage(pMsg),
			Session: session,
		})
	}
}

func troopAddMemberBroadcastDecoder(c *QQClient, pMsg *msg.Message, _ *network.IncomingPacketInfo) {
	groupJoinLock.Lock()
	defer groupJoinLock.Unlock()
	group := c.FindGroupByUin(pMsg.Head.FromUin.Unwrap())
	if pMsg.Head.AuthUin.Unwrap() == c.Uin {
		if group == nil && c.ReloadGroupList() == nil {
			c.GroupJoinEvent.dispatch(c, c.FindGroupByUin(pMsg.Head.FromUin.Unwrap()))
		}
	} else {
		if group != nil && group.FindMember(pMsg.Head.AuthUin.Unwrap()) == nil {
			mem, err := c.GetMemberInfo(group.Code, pMsg.Head.AuthUin.Unwrap())
			if err != nil {
				c.debug("error to fetch new member info: %v", err)
				return
			}
			group.Update(func(info *GroupInfo) {
				info.Members = append(info.Members, mem)
				info.sort()
			})
			c.GroupMemberJoinEvent.dispatch(c, &MemberJoinGroupEvent{
				Group:  group,
				Member: mem,
			})
		}
	}
}

func systemMessageDecoder(c *QQClient, _ *msg.Message, _ *network.IncomingPacketInfo) {
	_, pkt := c.buildSystemMsgNewFriendPacket()
	_ = c.sendPacket(pkt)
}

func troopSystemMessageDecoder(c *QQClient, pMsg *msg.Message, info *network.IncomingPacketInfo) {
	if !info.Params.Bool("used_reg_proxy") && pMsg.Head.MsgType.Unwrap() != 85 && pMsg.Head.MsgType.Unwrap() != 36 {
		c.exceptAndDispatchGroupSysMsg()
	}
	if len(pMsg.Body.MsgContent) == 0 {
		return
	}
	reader := binary.NewReader(pMsg.Body.MsgContent)
	groupCode := uint32(reader.ReadInt32())
	if info := c.FindGroup(int64(groupCode)); info != nil && pMsg.Head.GroupName.Unwrap() != "" && info.Name != pMsg.Head.GroupName.Unwrap() {
		c.debug("group %v name updated. %v -> %v", groupCode, info.Name, pMsg.Head.GroupName.Unwrap())
		info.Name = pMsg.Head.GroupName.Unwrap()
	}
}

func msgType0x211Decoder(c *QQClient, pMsg *msg.Message, info *network.IncomingPacketInfo) {
	if pMsg.Head.C2CCmd.Unwrap() == 6 || pMsg.Head.C2CTmpMsgHead != nil {
		tempSessionDecoder(c, pMsg, info)
	}
	sub4 := msg.SubMsgType0X4Body{}
	if err := proto.Unmarshal(pMsg.Body.MsgContent, &sub4); err != nil {
		err = errors.Wrap(err, "unmarshal sub msg 0x4 error")
		c.error("unmarshal sub msg 0x4 error: %v", err)
		return
	}
	if sub4.NotOnlineFile != nil && sub4.NotOnlineFile.Subcmd.Unwrap() == 1 { // subcmd: 1 -> sendPacket, 2-> recv
		rsp, err := c.sendAndWait(c.buildOfflineFileDownloadRequestPacket(sub4.NotOnlineFile.FileUuid)) // offline_file.go
		if err != nil {
			return
		}
		c.OfflineFileEvent.dispatch(c, &OfflineFileEvent{
			FileName:    string(sub4.NotOnlineFile.FileName),
			FileSize:    sub4.NotOnlineFile.FileSize.Unwrap(),
			Sender:      pMsg.Head.FromUin.Unwrap(),
			DownloadUrl: rsp.(string),
		})
	}
}
