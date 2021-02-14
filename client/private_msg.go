package client

import (
	"math/rand"
	"time"

	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/golang/protobuf/proto"
)

func (c *QQClient) SendPrivateMessage(target int64, m *message.SendingMessage) *message.PrivateMessage {
	mr := int32(rand.Uint32())
	seq := c.nextFriendSeq()
	t := time.Now().Unix()
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	msgLen := message.EstimateLength(m.Elements, 703)
	if msgLen > 5000 || imgCount > 50 {
		return nil
	}
	if msgLen > 300 || imgCount > 2 {
		div := int32(rand.Uint32())
		fragmented := m.ToFragmented()
		for i, elems := range fragmented {
			_, pkt := c.buildFriendSendingPacket(target, c.nextFriendSeq(), mr, int32(len(fragmented)), int32(i), div, t, elems)
			_ = c.send(pkt)
		}
	} else {
		_, pkt := c.buildFriendSendingPacket(target, seq, mr, 1, 0, 0, t, m.Elements)
		_ = c.send(pkt)
	}
	c.stat.MessageSent++
	return &message.PrivateMessage{
		Id:         seq,
		InternalId: mr,
		Target:     target,
		Time:       int32(t),
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Elements: m.Elements,
	}
}

func (c *QQClient) SendTempMessage(groupCode, target int64, m *message.SendingMessage) *message.TempMessage {
	group := c.FindGroup(groupCode)
	if group == nil {
		return nil
	}
	if c.FindFriend(target) != nil {
		pm := c.SendPrivateMessage(target, m)
		return &message.TempMessage{
			Id:        pm.Id,
			GroupCode: group.Code,
			GroupName: group.Name,
			Sender:    pm.Sender,
			Elements:  m.Elements,
		}
	}
	mr := int32(rand.Uint32())
	seq := c.nextFriendSeq()
	t := time.Now().Unix()
	_, pkt := c.buildTempSendingPacket(group.Uin, target, seq, mr, t, m)
	_ = c.send(pkt)
	c.stat.MessageSent++
	return &message.TempMessage{
		Id:        seq,
		GroupCode: group.Code,
		GroupName: group.Name,
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Elements: m.Elements,
	}
}

func (c *QQClient) buildGetOneDayRoamMsgRequest(target, lastMsgTime, random int64, count uint32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.PbGetOneDayRoamMsgReq{
		PeerUin:     proto.Uint64(uint64(target)),
		LastMsgTime: proto.Uint64(uint64(lastMsgTime)),
		Random:      proto.Uint64(uint64(random)),
		ReadCnt:     &count,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbGetOneDayRoamMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MessageSvc.PbSendMsg
func (c *QQClient) buildFriendSendingPacket(target int64, msgSeq, r, pkgNum, pkgIndex, pkgDiv int32, time int64, m []message.IMessageElement) (uint16, []byte) {
	seq := c.nextSeq()
	var ptt *msg.Ptt
	if len(m) > 0 {
		if p, ok := m[0].(*message.PrivateVoiceElement); ok {
			ptt = p.Ptt
			m = []message.IMessageElement{}
		}
	}
	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{C2C: &msg.C2C{ToUin: &target}},
		ContentHead: &msg.ContentHead{PkgNum: &pkgNum, PkgIndex: &pkgIndex, DivSeq: &pkgDiv},
		MsgBody: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m, false),
				Ptt:   ptt,
			},
		},
		MsgSeq:  &msgSeq,
		MsgRand: &r,
		SyncCookie: func() []byte {
			cookie := &msg.SyncCookie{
				Time:   &time,
				Ran1:   proto.Int64(rand.Int63()),
				Ran2:   proto.Int64(rand.Int63()),
				Const1: &syncConst1,
				Const2: &syncConst2,
				Const3: proto.Int64(0x1d),
			}
			b, _ := proto.Marshal(cookie)
			return b
		}(),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbSendMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MessageSvc.PbSendMsg
func (c *QQClient) buildTempSendingPacket(groupUin, target int64, msgSeq, r int32, time int64, m *message.SendingMessage) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{GrpTmp: &msg.GrpTmp{
			GroupUin: &groupUin,
			ToUin:    &target,
		}},
		ContentHead: &msg.ContentHead{PkgNum: proto.Int32(1)},
		MsgBody: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m.Elements, false),
			},
		},
		MsgSeq:  &msgSeq,
		MsgRand: &r,
		SyncCookie: func() []byte {
			cookie := &msg.SyncCookie{
				Time:   &time,
				Ran1:   proto.Int64(rand.Int63()),
				Ran2:   proto.Int64(rand.Int63()),
				Const1: &syncConst1,
				Const2: &syncConst2,
				Const3: proto.Int64(0x1d),
			}
			b, _ := proto.Marshal(cookie)
			return b
		}(),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbSendMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}
