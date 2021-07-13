package message

import (
	"crypto/md5"

	"google.golang.org/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
)

// *----- Definitions -----* //

type ForwardMessage struct {
	Nodes []*ForwardNode
}

type ForwardNode struct {
	SenderId   int64
	SenderName string
	Time       int32
	Message    []IMessageElement
}

type ForwardElement struct {
	FileName string
	Content  string
	ResId    string
	Items    []*msg.PbMultiMsgItem
}

// *----- Implementations -----* //

// Type impl IMessageElement
func (e *ForwardElement) Type() ElementType {
	return Forward
}

func (e *ForwardElement) Pack() (r []*msg.Elem) {
	r = []*msg.Elem{}
	r = append(r, &msg.Elem{
		RichMsg: &msg.RichMsg{
			Template1: append([]byte{1}, binary.ZlibCompress([]byte(e.Content))...),
			ServiceId: proto.Int32(35),
			MsgResId:  []byte{},
		},
	})
	r = append(r, &msg.Elem{
		Text: &msg.Text{
			Str: proto.String("你的QQ暂不支持查看[转发多条消息]，请期待后续版本。"),
		},
	})
	return
}

// Type impl IMessageElement
func (f *ForwardMessage) Type() ElementType {
	return Forward
}

func (f *ForwardMessage) CalculateValidationData(seq, random int32, groupCode int64) ([]byte, []byte) {
	msgs := f.packForwardMsg(seq, random, groupCode)
	trans := &msg.PbMultiMsgTransmit{Msg: msgs, PbItemList: []*msg.PbMultiMsgItem{
		{
			FileName: proto.String("MultiMsg"),
			Buffer:   &msg.PbMultiMsgNew{Msg: msgs},
		},
	}}
	b, _ := proto.Marshal(trans)
	data := binary.GZipCompress(b)
	hash := md5.Sum(data)
	return data, hash[:]
}

// CalculateValidationDataForward 屎代码
func (f *ForwardMessage) CalculateValidationDataForward(seq, random int32, groupCode int64) ([]byte, []byte, []*msg.PbMultiMsgItem) {
	msgs := f.packForwardMsg(seq, random, groupCode)
	trans := &msg.PbMultiMsgTransmit{Msg: msgs, PbItemList: []*msg.PbMultiMsgItem{
		{
			FileName: proto.String("MultiMsg"),
			Buffer:   &msg.PbMultiMsgNew{Msg: msgs},
		},
	}}
	for _, node := range f.Nodes {
		for _, message := range node.Message {
			if forwardElement, ok := message.(*ForwardElement); ok {
				trans.PbItemList = append(trans.PbItemList, forwardElement.Items...)
			}
		}
	}
	b, _ := proto.Marshal(trans)
	data := binary.GZipCompress(b)
	hash := md5.Sum(data)
	return data, hash[:], trans.PbItemList
}

func (f *ForwardMessage) packForwardMsg(seq int32, random int32, groupCode int64) []*msg.Message {
	msgs := make([]*msg.Message, 0, len(f.Nodes))
	for _, node := range f.Nodes {
		msgs = append(msgs, &msg.Message{
			Head: &msg.MessageHead{
				FromUin: &node.SenderId,
				MsgSeq:  &seq,
				MsgTime: &node.Time,
				MsgUid:  proto.Int64(0x0100_0000_0000_0000 | (int64(random) & 0xFFFFFFFF)),
				MutiltransHead: &msg.MutilTransHead{
					MsgId: proto.Int32(1),
				},
				MsgType: proto.Int32(82),
				GroupInfo: &msg.GroupInfo{
					GroupCode: &groupCode,
					GroupRank: []byte{},
					GroupName: []byte{},
					GroupCard: &node.SenderName,
				},
			},
			Body: &msg.MessageBody{
				RichText: &msg.RichText{
					Elems: ToProtoElems(node.Message, false),
				},
			},
		})
	}
	return msgs
}
