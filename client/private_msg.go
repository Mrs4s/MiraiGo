package client

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
)

func (c *QQClient) SendPrivateMessage(target int64, m *message.SendingMessage) *message.PrivateMessage {
	mr := int32(rand.Uint32())
	var seq int32
	t := time.Now().Unix()
	imgCount := 0
	frag := true
	for _, e := range m.Elements {
		switch e.Type() {
		case message.Image:
			imgCount++
		case message.Reply:
			frag = false
		}
	}
	msgLen := message.EstimateLength(m.Elements)
	if msgLen > message.MaxMessageSize || imgCount > 50 {
		return nil
	}
	if frag && (msgLen > 300 || imgCount > 2) && target == c.Uin {
		div := int32(rand.Uint32())
		fragmented := m.ToFragmented()
		for i, elems := range fragmented {
			fseq := c.nextFriendSeq()
			if i == 0 {
				seq = fseq
			}
			_, pkt := c.buildFriendSendingPacket(target, fseq, mr, int32(len(fragmented)), int32(i), div, t, elems)
			_ = c.sendPacket(pkt)
		}
	} else {
		seq = c.nextFriendSeq()
		if target != c.Uin {
			_, pkt := c.buildFriendSendingPacket(target, seq, mr, 1, 0, 0, t, m.Elements)
			_ = c.sendPacket(pkt)
		}
	}
	c.stat.MessageSent.Add(1)
	ret := &message.PrivateMessage{
		Id:         seq,
		InternalId: mr,
		Self:       c.Uin,
		Target:     target,
		Time:       int32(t),
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Elements: m.Elements,
	}
	go c.SelfPrivateMessageEvent.dispatch(c, ret)
	return ret
}

func (c *QQClient) SendGroupTempMessage(groupCode, target int64, m *message.SendingMessage) *message.TempMessage {
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
			Self:      c.Uin,
			Sender:    pm.Sender,
			Elements:  m.Elements,
		}
	}
	mr := int32(rand.Uint32())
	seq := c.nextFriendSeq()
	t := time.Now().Unix()
	_, pkt := c.buildGroupTempSendingPacket(group.Uin, target, seq, mr, t, m)
	_ = c.sendPacket(pkt)
	c.stat.MessageSent.Add(1)
	return &message.TempMessage{
		Id:        seq,
		GroupCode: group.Code,
		GroupName: group.Name,
		Self:      c.Uin,
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Elements: m.Elements,
	}
}

func (c *QQClient) sendWPATempMessage(target int64, sig []byte, m *message.SendingMessage) *message.TempMessage {
	mr := int32(rand.Uint32())
	seq := c.nextFriendSeq()
	t := time.Now().Unix()
	_, pkt := c.buildWPATempSendingPacket(target, sig, seq, mr, t, m)
	_ = c.sendPacket(pkt)
	c.stat.MessageSent.Add(1)
	return &message.TempMessage{
		Id:   seq,
		Self: c.Uin,
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Elements: m.Elements,
	}
}

func (s *TempSessionInfo) SendMessage(m *message.SendingMessage) (*message.TempMessage, error) {
	switch s.Source {
	case GroupSource:
		return s.client.SendGroupTempMessage(s.GroupCode, s.Sender, m), nil
	case ConsultingSource:
		return s.client.sendWPATempMessage(s.Sender, s.sig, m), nil
	default:
		return nil, errors.New("unsupported message source")
	}
}

/* this function is unused
func (c *QQClient) buildGetOneDayRoamMsgRequest(target, lastMsgTime, random int64, count uint32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.PbGetOneDayRoamMsgReq{
		PeerUin:     proto.Uint64(uint64(target)),
		LastMsgTime: proto.Uint64(uint64(lastMsgTime)),
		Random:      proto.Uint64(uint64(random)),
		ReadCnt:     &count,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbGetOneDayRoamMsg", 1, c.SessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}
*/

// MessageSvc.PbSendMsg
func (c *QQClient) buildFriendSendingPacket(target int64, msgSeq, r, pkgNum, pkgIndex, pkgDiv int32, time int64, m []message.IMessageElement) (uint16, []byte) {
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
		MsgSeq:     &msgSeq,
		MsgRand:    &r,
		SyncCookie: syncCookie(time),
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("MessageSvc.PbSendMsg", payload)
}

// MessageSvc.PbSendMsg
func (c *QQClient) buildGroupTempSendingPacket(groupUin, target int64, msgSeq, r int32, time int64, m *message.SendingMessage) (uint16, []byte) {
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
		MsgSeq:     &msgSeq,
		MsgRand:    &r,
		SyncCookie: syncCookie(time),
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("MessageSvc.PbSendMsg", payload)
}

func (c *QQClient) buildWPATempSendingPacket(uin int64, sig []byte, msgSeq, r int32, time int64, m *message.SendingMessage) (uint16, []byte) {
	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{WpaTmp: &msg.WPATmp{
			ToUin: proto.Uint64(uint64(uin)),
			Sig:   sig,
		}},
		ContentHead: &msg.ContentHead{PkgNum: proto.Int32(1)},
		MsgBody: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m.Elements, false),
			},
		},
		MsgSeq:     &msgSeq,
		MsgRand:    &r,
		SyncCookie: syncCookie(time),
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("MessageSvc.PbSendMsg", payload)
}

func syncCookie(time int64) []byte {
	cookie, _ := proto.Marshal(&msg.SyncCookie{
		Time:   &time,
		Ran1:   proto.Int64(rand.Int63()),
		Ran2:   proto.Int64(rand.Int63()),
		Const1: &syncConst1,
		Const2: &syncConst2,
		Const3: proto.Int64(0x1d),
	})
	return cookie
}
