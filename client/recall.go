package client

import (
	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
)

// 撤回相关处理逻辑

func (g *Group) recallGroupMessage(id *messageIdentifier) error {
	groupCode := g.i.Code
	msgSeq := id.Sequence
	if m, _ := g.c.GetGroupMessages(groupCode, msgSeq, msgSeq); len(m) > 0 {
		content := m[0].OriginalObject.Content
		if content.PkgNum.Unwrap() > 1 {
			if m, err := g.c.GetGroupMessages(
				groupCode,
				msgSeq-MessageSequence(content.PkgIndex.Unwrap()-1),
				msgSeq+MessageSequence(content.PkgNum.Unwrap()-content.PkgIndex.Unwrap()+1),
			); err == nil {
				if flag, _ := g.internalGroupRecall(id, m); flag {
					return nil
				}
			}
		}
	}
	_, err := g.c.sendAndWait(g.c.buildGroupRecallPacket(groupCode, int32(id.Sequence), int32(id.Internal)))
	return err
}

func (g *Group) internalGroupRecall(id *messageIdentifier, m []*message.GroupMessage) (flag bool, err error) {
	groupCode := g.i.Code
	for _, item := range m {
		if int64(item.InternalId) == id.Internal {
			flag = true
			if _, err := g.c.sendAndWait(g.c.buildGroupRecallPacket(groupCode, item.Id, item.InternalId)); err != nil {
				return false, err
			}
		}
	}
	return flag, nil
}

func (c *QQClient) RecallPrivateMessage(uin, ts int64, msgSeq MessageSequence, msgInternalId MessageInternal) error {
	_, err := c.sendAndWait(c.buildPrivateRecallPacket(uin, ts, int32(msgSeq), int32(msgInternalId)))
	return err
}

// PbMessageSvc.PbMsgWithDraw
func (c *QQClient) buildGroupRecallPacket(groupCode int64, msgSeq, msgRan int32) (uint16, []byte) {
	req := &msg.MsgWithDrawReq{
		GroupWithDraw: []*msg.GroupMsgWithDrawReq{
			{
				SubCmd:    proto.Int32(1),
				GroupCode: proto.Some(groupCode),
				MsgList: []*msg.GroupMsgInfo{
					{
						MsgSeq:    proto.Some(msgSeq),
						MsgRandom: proto.Some(msgRan),
						MsgType:   proto.Int32(0),
					},
				},
				UserDef: []byte{0x08, 0x00},
			},
		},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("PbMessageSvc.PbMsgWithDraw", payload)
}

func (c *QQClient) buildPrivateRecallPacket(uin, ts int64, msgSeq, random int32) (uint16, []byte) {
	req := &msg.MsgWithDrawReq{C2CWithDraw: []*msg.C2CMsgWithDrawReq{
		{
			MsgInfo: []*msg.C2CMsgInfo{
				{
					FromUin:   proto.Some(c.Uin),
					ToUin:     proto.Some(uin),
					MsgTime:   proto.Some(ts),
					MsgUid:    proto.Int64(0x0100_0000_0000_0000 | (int64(random) & 0xFFFFFFFF)),
					MsgSeq:    proto.Some(msgSeq),
					MsgRandom: proto.Some(random),
					RoutingHead: &msg.RoutingHead{
						C2C: &msg.C2C{
							ToUin: proto.Some(uin),
						},
					},
				},
			},
			LongMessageFlag: proto.Int32(0),
			Reserved:        []byte{0x08, 0x00},
			SubCmd:          proto.Int32(1),
		},
	}}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("PbMessageSvc.PbMsgWithDraw", payload)
}

func decodeMsgWithDrawResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := msg.MsgWithDrawResp{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(rsp.C2CWithDraw) > 0 {
		if rsp.C2CWithDraw[0].ErrMsg.Unwrap() != "" && rsp.C2CWithDraw[0].ErrMsg.Unwrap() != "Success" {
			return nil, errors.Errorf("recall error: %v msg: %v", rsp.C2CWithDraw[0].Result.Unwrap(), rsp.C2CWithDraw[0].ErrMsg.Unwrap())
		}
	}
	if len(rsp.GroupWithDraw) > 0 {
		if rsp.GroupWithDraw[0].ErrMsg.Unwrap() != "" && rsp.GroupWithDraw[0].ErrMsg.Unwrap() != "Success" {
			return nil, errors.Errorf("recall error: %v msg: %v", rsp.GroupWithDraw[0].Result.Unwrap(), rsp.GroupWithDraw[0].ErrMsg.Unwrap())
		}
	}
	return nil, nil
}
