package message

import (
	"bytes"
	"crypto/md5"

	"google.golang.org/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/utils"
)

// *----- Definitions -----* //

// ForwardMessage 添加 Node 请用 AddNode 方法
type ForwardMessage struct {
	Nodes []*ForwardNode
	items []*msg.PbMultiMsgItem
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

func (e *ForwardElement) Pack() []*msg.Elem {
	rich := &msg.Elem{
		RichMsg: &msg.RichMsg{
			Template1: append([]byte{1}, binary.ZlibCompress(utils.S2B(e.Content))...),
			ServiceId: proto.Int32(35),
			MsgResId:  []byte{},
		},
	}
	txt := &msg.Elem{
		Text: &msg.Text{
			Str: proto.String("你的QQ暂不支持查看[转发多条消息]，请期待后续版本。"),
		},
	}
	return []*msg.Elem{rich, txt}
}

func NewForwardMessage() *ForwardMessage {
	return &ForwardMessage{}
}

// AddNode adds a node to the forward message. return for method chaining.
func (f *ForwardMessage) AddNode(node *ForwardNode) *ForwardMessage {
	f.Nodes = append(f.Nodes, node)
	for _, item := range node.Message {
		if item.Type() != Forward { // quick path
			continue
		}
		if forward, ok := item.(*ForwardElement); ok {
			f.items = append(f.items, forward.Items...)
		}
	}
	return f
}

// Length return the length of Nodes.
func (f *ForwardMessage) Length() int { return len(f.Nodes) }

func (f *ForwardMessage) Brief() string {
	var brief bytes.Buffer
	for _, n := range f.Nodes {
		brief.WriteString(ToReadableString(n.Message))
		if brief.Len() >= 27 {
			break
		}
	}
	return brief.String()
}

func (f *ForwardMessage) Preview() string {
	var pv bytes.Buffer
	for i, node := range f.Nodes {
		if i >= 4 {
			break
		}
		pv.WriteString(`<title size="26" color="#777777">`)
		pv.WriteString(utils.XmlEscape(node.SenderName))
		pv.WriteString(": ")
		pv.WriteString(utils.XmlEscape(ToReadableString(node.Message)))
		pv.WriteString("</title>")
	}
	return pv.String()
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
	trans.PbItemList = append(trans.PbItemList, f.items...)
	b, _ := proto.Marshal(trans)
	data := binary.GZipCompress(b)
	hash := md5.Sum(data)
	return data, hash[:], trans.PbItemList
}

func (f *ForwardMessage) packForwardMsg(seq int32, random int32, groupCode int64) []*msg.Message {
	ml := make([]*msg.Message, 0, len(f.Nodes))
	for _, node := range f.Nodes {
		ml = append(ml, &msg.Message{
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
	return ml
}
