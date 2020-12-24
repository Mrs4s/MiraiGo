package client

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client/pb/longmsg"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/multimsg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"math"
	"math/rand"
	"time"
)

func init() {
	decoders["OnlinePush.PbPushGroupMsg"] = decodeGroupMessagePacket
	decoders["MessageSvc.PbSendMsg"] = decodeMsgSendResponse
	decoders["MessageSvc.PbGetGroupMsg"] = decodeGetGroupMsgResponse
	decoders["OidbSvc.0x8a7_0"] = decodeAtAllRemainResponse
}

// SendGroupMessage 发送群消息
func (c *QQClient) SendGroupMessage(groupCode int64, m *message.SendingMessage, f ...bool) *message.GroupMessage {
	useFram := false
	if len(f) > 0 {
		useFram = f[0]
	}
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	if useFram {
		if m.Any(func(e message.IMessageElement) bool { return e.Type() == message.Reply }) {
			useFram = false
		}
	}
	msgLen := message.EstimateLength(m.Elements, 703)
	if msgLen > 5000 || imgCount > 50 {
		return nil
	}
	if (msgLen > 200 || imgCount > 1) && !useFram {
		ret := c.sendGroupLongOrForwardMessage(groupCode, true, &message.ForwardMessage{Nodes: []*message.ForwardNode{
			{
				SenderId:   c.Uin,
				SenderName: c.Nickname,
				Time:       int32(time.Now().Unix()),
				Message:    m.Elements,
			},
		}})
		return ret
	}
	return c.sendGroupMessage(groupCode, false, m)
}

// SendGroupForwardMessage 发送群合并转发消息
func (c *QQClient) SendGroupForwardMessage(groupCode int64, m *message.ForwardMessage) *message.GroupMessage {
	return c.sendGroupLongOrForwardMessage(groupCode, false, m)
}

// GetGroupMessages 从服务器获取历史信息
func (c *QQClient) GetGroupMessages(groupCode, beginSeq, endSeq int64) ([]*message.GroupMessage, error) {
	i, err := c.sendAndWait(c.buildGetGroupMsgRequest(groupCode, beginSeq, endSeq))
	if err != nil {
		return nil, err
	}
	return i.([]*message.GroupMessage), nil
}

func (c *QQClient) GetAtAllRemain(groupCode int64) (*AtAllRemainInfo, error) {
	i, err := c.sendAndWait(c.buildAtAllRemainRequestPacket(groupCode))
	if err != nil {
		return nil, err
	}
	return i.(*AtAllRemainInfo), nil
}

func (c *QQClient) sendGroupMessage(groupCode int64, forward bool, m *message.SendingMessage) *message.GroupMessage {
	eid := utils.RandomString(6)
	mr := int32(rand.Uint32())
	ch := make(chan int32)
	c.onGroupMessageReceipt(eid, func(c *QQClient, e *groupMessageReceiptEvent) {
		if e.Rand == mr && !utils.IsChanClosed(ch) {
			ch <- e.Seq
		}
	})
	defer c.onGroupMessageReceipt(eid)
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	msgLen := message.EstimateLength(m.Elements, 703)
	if (msgLen > 100 || imgCount > 1) && !forward && !m.Any(func(e message.IMessageElement) bool {
		_, ok := e.(*message.GroupVoiceElement)
		_, ok2 := e.(*message.ServiceElement)
		_, ok3 := e.(*message.ReplyElement)
		return ok || ok2 || ok3
	}) {
		div := int32(rand.Uint32())
		fragmented := m.ToFragmented()
		for i, elems := range fragmented {
			_, pkt := c.buildGroupSendingPacket(groupCode, mr, int32(len(fragmented)), int32(i), div, forward, elems)
			_ = c.send(pkt)
		}
	} else {
		_, pkt := c.buildGroupSendingPacket(groupCode, mr, 1, 0, 0, forward, m.Elements)
		_ = c.send(pkt)
	}
	var mid int32
	ret := &message.GroupMessage{
		Id:         -1,
		InternalId: mr,
		GroupCode:  groupCode,
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Time:     int32(time.Now().Unix()),
		Elements: m.Elements,
	}
	select {
	case mid = <-ch:
		ret.Id = mid
		return ret
	case <-time.After(time.Second * 5):
		if g, err := c.GetGroupInfo(groupCode); err == nil {
			if history, err := c.GetGroupMessages(groupCode, g.lastMsgSeq-10, g.lastMsgSeq+1); err == nil {
				for _, m := range history {
					if m.InternalId == mr {
						return m
					}
				}
			}
		}
		return ret
	}
}

func (c *QQClient) sendGroupLongOrForwardMessage(groupCode int64, isLong bool, m *message.ForwardMessage) *message.GroupMessage {
	if len(m.Nodes) >= 200 {
		return nil
	}
	ts := time.Now().Unix()
	seq := c.nextGroupSeq()
	data, hash := m.CalculateValidationData(seq, rand.Int31(), groupCode)
	i, err := c.sendAndWait(c.buildMultiApplyUpPacket(data, hash, func() int32 {
		if isLong {
			return 1
		} else {
			return 2
		}
	}(), utils.ToGroupUin(groupCode)))
	if err != nil {
		return nil
	}
	rsp := i.(*multimsg.MultiMsgApplyUpRsp)
	body, _ := proto.Marshal(&longmsg.LongReqBody{
		Subcmd:       1,
		TermType:     5,
		PlatformType: 9,
		MsgUpReq: []*longmsg.LongMsgUpReq{
			{
				MsgType:    3,
				DstUin:     utils.ToGroupUin(groupCode),
				MsgContent: data,
				StoreType:  2,
				MsgUkey:    rsp.MsgUkey,
			},
		},
	})
	for i, ip := range rsp.Uint32UpIp {
		err := c.highwayUpload(uint32(ip), int(rsp.Uint32UpPort[i]), rsp.MsgSig, body, 27)
		if err == nil {
			if !isLong {
				var pv string
				for i := 0; i < int(math.Min(4, float64(len(m.Nodes)))); i++ {
					pv += fmt.Sprintf(`<title size="26" color="#777777">%s: %s</title>`, m.Nodes[i].SenderName, message.ToReadableString(m.Nodes[i].Message))
				}
				return c.sendGroupMessage(groupCode, true, genForwardTemplate(rsp.MsgResid, pv, "群聊的聊天记录", "[聊天记录]", "聊天记录", fmt.Sprintf("查看 %d 条转发消息", len(m.Nodes)), ts))
			}
			bri := func() string {
				var r string
				for _, n := range m.Nodes {
					r += message.ToReadableString(n.Message)
					if len(r) >= 27 {
						break
					}
				}
				return r
			}()
			return c.sendGroupMessage(groupCode, false, genLongTemplate(rsp.MsgResid, bri, ts))
		}
	}
	return nil
}

// MessageSvc.PbSendMsg
func (c *QQClient) buildGroupSendingPacket(groupCode int64, r, pkgNum, pkgIndex, pkgDiv int32, forward bool, m []message.IMessageElement) (uint16, []byte) {
	seq := c.nextSeq()
	var ptt *message.GroupVoiceElement
	if len(m) > 0 {
		if p, ok := m[0].(*message.GroupVoiceElement); ok {
			ptt = p
			m = []message.IMessageElement{}
		}
	}
	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{Grp: &msg.Grp{GroupCode: &groupCode}},
		ContentHead: &msg.ContentHead{PkgNum: &pkgNum, PkgIndex: &pkgIndex, DivSeq: &pkgDiv},
		MsgBody: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m, true),
				Ptt: func() *msg.Ptt {
					if ptt != nil {
						return ptt.Ptt
					}
					return nil
				}(),
			},
		},
		MsgSeq:     proto.Int32(c.nextGroupSeq()),
		MsgRand:    &r,
		SyncCookie: EmptyBytes,
		MsgVia:     proto.Int32(1),
		MsgCtrl: func() *msg.MsgCtrl {
			if forward {
				return &msg.MsgCtrl{MsgFlag: proto.Int32(4)}
			}
			return nil
		}(),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbSendMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildGetGroupMsgRequest(groupCode, beginSeq, endSeq int64) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.GetGroupMsgReq{
		GroupCode:   proto.Uint64(uint64(groupCode)),
		BeginSeq:    proto.Uint64(uint64(beginSeq)),
		EndSeq:      proto.Uint64(uint64(endSeq)),
		PublicGroup: proto.Bool(false),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbGetGroupMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildAtAllRemainRequestPacket(groupCode int64) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D8A7ReqBody{
		SubCmd:                    proto.Uint32(1),
		LimitIntervalTypeForUin:   proto.Uint32(2),
		LimitIntervalTypeForGroup: proto.Uint32(1),
		Uin:                       proto.Uint64(uint64(c.Uin)),
		GroupCode:                 proto.Uint64(uint64(groupCode)),
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:    2215,
		Bodybuffer: b,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x8a7_0", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OnlinePush.PbPushGroupMsg
func decodeGroupMessagePacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkt := msg.PushMessagePacket{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if pkt.Message.Head.GetFromUin() == c.Uin {
		c.dispatchGroupMessageReceiptEvent(&groupMessageReceiptEvent{
			Rand: pkt.Message.Body.RichText.Attr.GetRandom(),
			Seq:  pkt.Message.Head.GetMsgSeq(),
			Msg:  c.parseGroupMessage(pkt.Message),
		})
		return nil, nil
	}
	if pkt.Message.Content != nil && pkt.Message.Content.GetPkgNum() > 1 {
		var builder *groupMessageBuilder // TODO: 支持多SEQ
		i, ok := c.groupMsgBuilders.Load(pkt.Message.Content.DivSeq)
		if !ok {
			builder = &groupMessageBuilder{}
			c.groupMsgBuilders.Store(pkt.Message.Content.DivSeq, builder)
		} else {
			builder = i.(*groupMessageBuilder)
		}
		builder.MessageSlices = append(builder.MessageSlices, pkt.Message)
		if int32(len(builder.MessageSlices)) >= pkt.Message.Content.GetPkgNum() {
			c.groupMsgBuilders.Delete(pkt.Message.Content.DivSeq)
			c.dispatchGroupMessage(c.parseGroupMessage(builder.build()))
		}
		return nil, nil
	}
	c.dispatchGroupMessage(c.parseGroupMessage(pkt.Message))
	return nil, nil
}

func decodeMsgSendResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.SendMessageResponse{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.GetResult() != 0 {
		c.Error("send msg error: %v %v", rsp.GetResult(), rsp.GetErrMsg())
	}
	return nil, nil
}

func decodeGetGroupMsgResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.GetGroupMsgResp{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.GetResult() != 0 {
		c.Error("get msg error: %v %v", rsp.GetResult(), rsp.GetErrmsg())
		return nil, errors.Errorf("get msg error: %v msg: %v", rsp.GetResult(), rsp.GetErrmsg())
	}
	var ret []*message.GroupMessage
	for _, m := range rsp.Msg {
		if m.Head.FromUin == nil {
			continue
		}
		ret = append(ret, c.parseGroupMessage(m))
	}
	return ret, nil
}

func decodeAtAllRemainResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D8A7RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return &AtAllRemainInfo{
		CanAtAll:                 rsp.GetCanAtAll(),
		RemainAtAllCountForGroup: rsp.GetRemainAtAllCountForGroup(),
		RemainAtAllCountForUin:   rsp.GetRemainAtAllCountForUin(),
	}, nil
}
