package client

import (
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"google.golang.org/protobuf/proto"
)

// 撤回相关处理逻辑

func (c *QQClient) RecallGroupMessage(groupCode int64, msgId, msgInternalId int32) error {
	_, err := c.sendAndWait(c.buildGroupRecallPacket(groupCode, msgId, msgInternalId))
	return err
}

func (c *QQClient) RecallPrivateMessage(uin, ts int64, msgId, msgInternalId int32) error {
	_, err := c.sendAndWait(c.buildPrivateRecallPacket(uin, ts, msgId, msgInternalId))
	return err
}

// PbMessageSvc.PbMsgWithDraw
func (c *QQClient) buildGroupRecallPacket(groupCode int64, msgSeq, msgRan int32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.MsgWithDrawReq{
		GroupWithDraw: []*msg.GroupMsgWithDrawReq{
			{
				SubCmd:    proto.Int32(1),
				GroupCode: &groupCode,
				MsgList: []*msg.GroupMsgInfo{
					{
						MsgSeq:    &msgSeq,
						MsgRandom: &msgRan,
						MsgType:   proto.Int32(0),
					},
				},
				UserDef: []byte{0x08, 0x00},
			},
		},
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "PbMessageSvc.PbMsgWithDraw", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildPrivateRecallPacket(uin, ts int64, msgSeq, random int32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.MsgWithDrawReq{C2CWithDraw: []*msg.C2CMsgWithDrawReq{
		{
			MsgInfo: []*msg.C2CMsgInfo{
				{
					FromUin:   &c.Uin,
					ToUin:     &uin,
					MsgTime:   &ts,
					MsgUid:    proto.Int64(int64(random)),
					MsgSeq:    &msgSeq,
					MsgRandom: &random,
				},
			},
			Reserved: []byte{0x08, 0x00},
			SubCmd:   proto.Int32(1),
		},
	}}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "PbMessageSvc.PbMsgWithDraw", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func decodeMsgWithDrawResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.MsgWithDrawResp{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if len(rsp.C2CWithDraw) > 0 {
		if rsp.C2CWithDraw[0].GetResult() != 0 {
			return nil, errors.New(fmt.Sprintf("recall error: %v msg: %v", rsp.C2CWithDraw[0].GetResult(), rsp.C2CWithDraw[0].GetErrMsg()))
		}
	}
	if len(rsp.GroupWithDraw) > 0 {
		if rsp.GroupWithDraw[0].GetResult() != 0 {
			return nil, errors.New(fmt.Sprintf("recall error: %v msg: %v", rsp.GroupWithDraw[0].GetResult(), rsp.GroupWithDraw[0].GetErrMsg()))
		}
	}
	return nil, nil
}
