package client

import (
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/longmsg"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/multimsg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

func init() {
	decoders["MultiMsg.ApplyUp"] = decodeMultiApplyUpResponse
	decoders["MultiMsg.ApplyDown"] = decodeMultiApplyDownResponse
}

// MultiMsg.ApplyUp
func (c *QQClient) buildMultiApplyUpPacket(data, hash []byte, buType int32, groupUin int64) (uint16, []byte) {
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
	return c.uniPacket("MultiMsg.ApplyUp", payload)
}

// MultiMsg.ApplyUp
func decodeMultiApplyUpResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
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
func (c *QQClient) buildMultiApplyDownPacket(resID string) (uint16, []byte) {
	req := &multimsg.MultiReqBody{
		Subcmd:       2,
		TermType:     5,
		PlatformType: 9,
		NetType:      3,
		BuildVer:     "8.2.0.1296",
		MultimsgApplydownReq: []*multimsg.MultiMsgApplyDownReq{
			{
				MsgResid: []byte(resID),
				MsgType:  3,
			},
		},
		BuType:         2,
		ReqChannelType: 2,
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("MultiMsg.ApplyDown", payload)
}

// MultiMsg.ApplyDown
func decodeMultiApplyDownResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
	body := multimsg.MultiRspBody{}
	if err := proto.Unmarshal(payload, &body); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(body.MultimsgApplydownRsp) == 0 {
		return nil, errors.New("not found")
	}
	rsp := body.MultimsgApplydownRsp[0]

	var prefix string
	if rsp.MsgExternInfo != nil && rsp.MsgExternInfo.ChannelType == 2 {
		prefix = "https://ssl.htdata.qq.com"
	} else {
		prefix = fmt.Sprintf("http://%s:%d", binary.UInt32ToIPV4Address(uint32(rsp.Uint32DownIp[0])), body.MultimsgApplydownRsp[0].Uint32DownPort[0])
	}
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
	msgContent := lb.MsgDownRsp[0].MsgContent
	if msgContent == nil {
		return nil, errors.New("message content is empty")
	}
	uc := binary.GZipUncompress(msgContent)
	mt := msg.PbMultiMsgTransmit{}
	if err = proto.Unmarshal(uc, &mt); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return &mt, nil
}

type forwardMsgLinker struct {
	items map[string]*msg.PbMultiMsgItem
}

func (l *forwardMsgLinker) link(name string) *message.ForwardMessage {
	item := l.items[name]
	if item == nil {
		return nil
	}
	nodes := make([]*message.ForwardNode, 0, len(item.GetBuffer().GetMsg()))
	for _, m := range item.GetBuffer().GetMsg() {
		name := m.Head.GetFromNick()
		if m.Head.GetMsgType() == 82 && m.Head.GroupInfo != nil {
			name = m.Head.GroupInfo.GetGroupCard()
		}

		msgElems := message.ParseMessageElems(m.Body.RichText.Elems)
		for i, elem := range msgElems {
			if forward, ok := elem.(*message.ForwardElement); ok {
				if forward.FileName != "" {
					msgElems[i] = l.link(forward.FileName) // 递归处理嵌套转发
				}
			}
		}

		nodes = append(nodes, &message.ForwardNode{
			SenderId:   m.Head.GetFromUin(),
			SenderName: name,
			Time:       m.Head.GetMsgTime(),
			Message:    msgElems,
		})
	}
	return &message.ForwardMessage{Nodes: nodes}
}

func (c *QQClient) GetForwardMessage(resID string) *message.ForwardMessage {
	m := c.DownloadForwardMessage(resID)
	if m == nil {
		return nil
	}
	linker := forwardMsgLinker{
		items: make(map[string]*msg.PbMultiMsgItem),
	}
	for _, item := range m.Items {
		linker.items[item.GetFileName()] = item
	}
	return linker.link(m.FileName)
}

func (c *QQClient) DownloadForwardMessage(resId string) *message.ForwardElement {
	i, err := c.sendAndWait(c.buildMultiApplyDownPacket(resId))
	if err != nil {
		return nil
	}
	multiMsg := i.(*msg.PbMultiMsgTransmit)
	if multiMsg.GetPbItemList() == nil {
		return nil
	}
	var pv string
	for i := 0; i < int(math.Min(4, float64(len(multiMsg.GetMsg())))); i++ {
		m := multiMsg.Msg[i]
		pv += fmt.Sprintf(`<title size="26" color="#777777">%s: %s</title>`,
			func() string {
				if m.Head.GetMsgType() == 82 && m.Head.GroupInfo != nil {
					return m.Head.GroupInfo.GetGroupCard()
				}
				return m.Head.GetFromNick()
			}(),
			message.ToReadableString(
				message.ParseMessageElems(multiMsg.Msg[i].GetBody().GetRichText().Elems),
			),
		)
	}
	return genForwardTemplate(
		resId, pv, "群聊的聊天记录", "[聊天记录]", "聊天记录",
		fmt.Sprintf("查看 %d 条转发消息", len(multiMsg.GetMsg())),
		time.Now().UnixNano(),
		multiMsg.GetPbItemList(),
	)
}
