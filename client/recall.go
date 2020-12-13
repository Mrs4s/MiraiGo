package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// 撤回相关处理逻辑

func (c *QQClient) RecallGroupMessage(groupCode int64, msgId, msgInternalId int32) error {
	if m, _ := c.GetGroupMessages(groupCode, int64(msgId), int64(msgId)); len(m) > 0 {
		content := m[0].OriginalObject.Content
		if content.GetPkgNum() > 1 {
			if m, err := c.GetGroupMessages(groupCode, int64(msgId-content.GetPkgIndex()-1), int64(msgId+(content.GetPkgNum()-content.GetPkgIndex()+1))); err == nil {
				if flag, _ := c.internalGroupRecall(groupCode, msgInternalId, m); flag {
					return nil
				}
			}
		}
	}
	_, err := c.sendAndWait(c.buildGroupRecallPacket(groupCode, msgId, msgInternalId))
	return err
}

func (c *QQClient) internalGroupRecall(groupCode int64, msgInternalId int32, m []*message.GroupMessage) (flag bool, err error) {
	for _, item := range m {
		if item.InternalId == msgInternalId {
			flag = true
			if _, err := c.sendAndWait(c.buildGroupRecallPacket(groupCode, item.Id, item.InternalId)); err != nil {
				return false, err
			}
		}
	}
	return flag, nil
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
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(rsp.C2CWithDraw) > 0 {
		if rsp.C2CWithDraw[0].GetErrMsg() != "" && rsp.C2CWithDraw[0].GetErrMsg() != "Success" {
			return nil, errors.Errorf("recall error: %v msg: %v", rsp.C2CWithDraw[0].GetResult(), rsp.C2CWithDraw[0].GetErrMsg())
		}
	}
	if len(rsp.GroupWithDraw) > 0 {
		if rsp.GroupWithDraw[0].GetErrMsg() != "" && rsp.GroupWithDraw[0].GetErrMsg() != "Success" {
			return nil, errors.Errorf("recall error: %v msg: %v", rsp.GroupWithDraw[0].GetResult(), rsp.GroupWithDraw[0].GetErrMsg())
		}
	}
	return nil, nil
}
