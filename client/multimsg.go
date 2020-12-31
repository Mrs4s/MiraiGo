package client

import (
	"fmt"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/longmsg"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/multimsg"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func init() {
	decoders["MultiMsg.ApplyUp"] = decodeMultiApplyUpResponse
	decoders["MultiMsg.ApplyDown"] = decodeMultiApplyDownResponse
}

// MultiMsg.ApplyUp
func (c *QQClient) buildMultiApplyUpPacket(data, hash []byte, buType int32, groupUin int64) (uint16, []byte) {
	seq := c.nextSeq()
	req := &multimsg.MultiReqBody{
		Subcmd:       1,
		TermType:     5,
		PlatformType: 9,
		NetType:      3,
		BuildVer:     "8.2.0.1296",
		MultimsgApplyupReq: []*multimsg.MultiMsgApplyUpReq{
			{
				DstUin:  groupUin,
				MsgSize: int64(len(data)),
				MsgMd5:  hash,
				MsgType: 3,
			},
		},
		BuType: buType,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MultiMsg.ApplyUp", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MultiMsg.ApplyUp
func decodeMultiApplyUpResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	body := multimsg.MultiRspBody{}
	if err := proto.Unmarshal(payload, &body); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(body.MultimsgApplyupRsp) == 0 {
		return nil, errors.New("rsp is empty")
	}
	rsp := body.MultimsgApplyupRsp[0]
	switch rsp.Result {
	case 0:
		return rsp, nil
	case 193:
		return nil, errors.New("too large")
	}
	return nil, errors.Errorf("unexpected multimsg apply up response: %d", rsp.Result)
}

// MultiMsg.ApplyDown
func (c *QQClient) buildMultiApplyDownPacket(resId string) (uint16, []byte) {
	seq := c.nextSeq()
	req := &multimsg.MultiReqBody{
		Subcmd:       2,
		TermType:     5,
		PlatformType: 9,
		NetType:      3,
		BuildVer:     "8.2.0.1296",
		MultimsgApplydownReq: []*multimsg.MultiMsgApplyDownReq{
			{
				MsgResid: []byte(resId),
				MsgType:  3,
			},
		},
		BuType:         2,
		ReqChannelType: 2,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MultiMsg.ApplyDown", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MultiMsg.ApplyDown
func decodeMultiApplyDownResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	body := multimsg.MultiRspBody{}
	if err := proto.Unmarshal(payload, &body); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(body.MultimsgApplydownRsp) == 0 {
		return nil, errors.New("not found")
	}
	rsp := body.MultimsgApplydownRsp[0]
	prefix := func() string {
		if rsp.MsgExternInfo != nil && rsp.MsgExternInfo.ChannelType == 2 {
			return "https://ssl.htdata.qq.com"
		}
		return fmt.Sprintf("http://%s:%d", binary.UInt32ToIPV4Address(uint32(rsp.Uint32DownIp[0])), body.MultimsgApplydownRsp[0].Uint32DownPort[0])
	}()
	b, err := utils.HttpGetBytes(fmt.Sprintf("%s%s", prefix, string(rsp.ThumbDownPara)), "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to download by multi apply down")
	}
	if b[0] != 40 {
		return nil, errors.New("unexpected body data")
	}
	tea := binary.NewTeaCipher(body.MultimsgApplydownRsp[0].MsgKey)
	r := binary.NewReader(b[1:])
	i1 := r.ReadInt32()
	i2 := r.ReadInt32()
	if i1 > 0 {
		r.ReadBytes(int(i1)) // im msg head
	}
	data := tea.Decrypt(r.ReadBytes(int(i2)))
	lb := longmsg.LongRspBody{}
	if err = proto.Unmarshal(data, &lb); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	uc := binary.GZipUncompress(lb.MsgDownRsp[0].MsgContent)
	mt := msg.PbMultiMsgTransmit{}
	if err = proto.Unmarshal(uc, &mt); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return &mt, nil
}
