package client

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"

	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

var privateMsgDecoders = map[int32]func(*QQClient, *msg.Message, *incomingPacketInfo){
	9: privateMessageDecoder, 10: privateMessageDecoder, 31: privateMessageDecoder,
	79: privateMessageDecoder, 97: privateMessageDecoder, 120: privateMessageDecoder,
	132: privateMessageDecoder, 133: privateMessageDecoder, 166: privateMessageDecoder,
	167: privateMessageDecoder, 140: tempSessionDecoder, 141: tempSessionDecoder,
	208: privatePttDecoder,
}

var troopSystemMsgDecoders = map[int32]func(*QQClient, *msg.Message, *incomingPacketInfo){
	35: troopSystemMessageDecoder, 36: troopSystemMessageDecoder, 37: troopSystemMessageDecoder,
	45: troopSystemMessageDecoder, 46: troopSystemMessageDecoder, 84: troopSystemMessageDecoder,
	85: troopSystemMessageDecoder, 86: troopSystemMessageDecoder, 87: troopSystemMessageDecoder,
}

var sysMsgDecoders = map[int32]func(*QQClient, *msg.Message, *incomingPacketInfo){
	187: systemMessageDecoder, 188: systemMessageDecoder, 189: systemMessageDecoder,
	190: systemMessageDecoder, 191: systemMessageDecoder,
}

var otherDecoders = map[int32]func(*QQClient, *msg.Message, *incomingPacketInfo){
	33: troopAddMemberBroadcastDecoder, 529: msgType0x211Decoder,
}

var c2cDecoders = map[int32]func(*QQClient, *msg.Message, *incomingPacketInfo){}

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
	GroupSource         TempSessionSource = 0 // 来自群聊
	ConsultingSource    TempSessionSource = 1 // 来自QQ咨询
	SearchSource        TempSessionSource = 2 // 来自查找
	MovieSource         TempSessionSource = 3 // 来自QQ电影
	HotChatSource       TempSessionSource = 4 // 来自热聊
	SystemMessageSource TempSessionSource = 6 // 来自验证消息
	MultiChatSource     TempSessionSource = 7 // 来自多人聊天
	DateSource          TempSessionSource = 8 // 来自约会
	AddressBookSource   TempSessionSource = 9 // 来自通讯录
)

func init() {
	merge := func(m map[int32]func(*QQClient, *msg.Message, *incomingPacketInfo)) {
		for k, v := range m {
			c2cDecoders[k] = v
		}
	}
	merge(privateMsgDecoders)
	merge(troopSystemMsgDecoders)
	merge(sysMsgDecoders)
	merge(otherDecoders)
}

func (c *QQClient) c2cMessageSyncProcessor(rsp *msg.GetMessageResponse, info *incomingPacketInfo) {
	c.syncCookie = rsp.SyncCookie
	c.pubAccountCookie = rsp.PubAccountCookie
	c.msgCtrlBuf = rsp.MsgCtrlBuf
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
			c.commMsgProcessor(pMsg, info)
		}
	}
	if delItems != nil {
		_, _ = c.sendAndWait(c.buildDeleteMessageRequestPacket(delItems))
	}
	if rsp.GetSyncFlag() != msg.SyncFlag_STOP {
		c.Debug("continue sync with flag: %v", rsp.SyncFlag.String())
		seq, pkt := c.buildGetMessageRequestPacket(rsp.GetSyncFlag(), time.Now().Unix())
		_, _ = c.sendAndWait(seq, pkt, info.Params)
	}
}

func (c *QQClient) commMsgProcessor(pMsg *msg.Message, info *incomingPacketInfo) {
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
	if info.Params.bool("init") {
		return
	}
	if decoder, ok := c2cDecoders[pMsg.Head.GetMsgType()]; ok {
		decoder(c, pMsg, info)
	} else {
		c.Debug("unknown msg type on c2c processor: %v - %v", pMsg.Head.GetMsgType(), pMsg.Head.GetC2CCmd())
	}
}

func privateMessageDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
	switch pMsg.Head.GetC2CCmd() {
	case 11, 175: // friend msg
		if pMsg.Head.GetFromUin() == c.Uin {
			for {
				frdSeq := atomic.LoadInt32(&c.friendSeq)
				if frdSeq < pMsg.Head.GetMsgSeq() {
					if atomic.CompareAndSwapInt32(&c.friendSeq, frdSeq, pMsg.Head.GetMsgSeq()) {
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

func privatePttDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
	if pMsg.Body == nil || pMsg.Body.RichText == nil || pMsg.Body.RichText.Ptt == nil {
		return
	}
	if len(pMsg.Body.RichText.Ptt.Reserve) != 0 {
		// m := binary.NewReader(pMsg.Body.RichText.Ptt.Reserve[1:]).ReadTlvMap(1)
		// T3 -> timestamp T8 -> voiceType T9 -> voiceLength T10 -> PbReserveStruct
	}
	c.dispatchPrivateMessage(c.parsePrivateMessage(pMsg))
}

func tempSessionDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
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

func troopAddMemberBroadcastDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
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

func systemMessageDecoder(c *QQClient, _ *msg.Message, _ *incomingPacketInfo) {
	_, pkt := c.buildSystemMsgNewFriendPacket()
	_ = c.send(pkt)
}

func troopSystemMessageDecoder(c *QQClient, pMsg *msg.Message, info *incomingPacketInfo) {
	if !info.Params.bool("used_reg_proxy") && pMsg.Head.GetMsgType() != 85 && pMsg.Head.GetMsgType() != 36 {
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

func msgType0x211Decoder(c *QQClient, pMsg *msg.Message, info *incomingPacketInfo) {
	if pMsg.Head.GetC2CCmd() == 6 || pMsg.Head.C2CTmpMsgHead != nil {
		tempSessionDecoder(c, pMsg, info)
	}
	sub4 := msg.SubMsgType0X4Body{}
	if err := proto.Unmarshal(pMsg.Body.MsgContent, &sub4); err != nil {
		err = errors.Wrap(err, "unmarshal sub msg 0x4 error")
		c.Error("unmarshal sub msg 0x4 error: %v", err)
		return
	}
	if sub4.NotOnlineFile != nil && sub4.NotOnlineFile.GetSubcmd() == 1 { // subcmd: 1 -> send, 2-> recv
		rsp, err := c.sendAndWait(c.buildOfflineFileDownloadRequestPacket(sub4.NotOnlineFile.FileUuid)) // offline_file.go
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
