package message

import (
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"strconv"
)

type PrivateMessage struct {
	Id       int32
	Sender   *Sender
	Elements []IMessageElement
}

type GroupMessage struct {
	Id        int32
	GroupUin  int64
	GroupName string
	Sender    *Sender
	Elements  []IMessageElement
}

type SendingMessage struct {
	Elements []IMessageElement
}

type Sender struct {
	Uin      int64
	Nickname string
	CardName string
	IsFriend bool
}

type IMessageElement interface {
	Type() ElementType
}

type ElementType int

const (
	Text ElementType = iota
	Image
	Face
	At
	Reply
)

func (s *Sender) IsAnonymous() bool {
	return s.Uin == 80000000
}

func NewSendingMessage() *SendingMessage {
	return &SendingMessage{}
}

func (msg *PrivateMessage) ToString() (res string) {
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *TextElement:
			res += e.Content
		case *ImageElement:
			res += "[Image:" + e.Filename + "]"
		case *FaceElement:
			res += "[" + e.Name + "]"
		case *AtElement:
			res += e.Display
		}
	}
	return
}

func (msg *GroupMessage) ToString() (res string) {
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *TextElement:
			res += e.Content
		case *ImageElement:
			res += "[Image:" + e.Filename + "]"
		case *FaceElement:
			res += "[" + e.Name + "]"
		case *GroupImageElement:
			res += "[Image: " + e.ImageId + "]"
		case *AtElement:
			res += e.Display
		case *ReplyElement:
			res += "[Reply:" + strconv.FormatInt(int64(e.ReplySeq), 10) + "]"
		}
	}
	return
}

func (msg *SendingMessage) Append(e IMessageElement) *SendingMessage {
	msg.Elements = append(msg.Elements, e)
	return msg
}

func ToProtoElems(elems []IMessageElement) (r []*msg.Elem) {
	for _, elem := range elems {
		switch e := elem.(type) {
		case *TextElement:
			r = append(r, &msg.Elem{
				Text: &msg.Text{
					Str: e.Content,
				},
			})
		case *FaceElement:
			r = append(r, &msg.Elem{
				Face: &msg.Face{
					Index: e.Index,
					Old:   binary.ToBytes(int16(0x1445 - 4 + e.Index)),
					Buf:   []byte{0x00, 0x01, 0x00, 0x04, 0x52, 0xCC, 0xF5, 0xD0},
				},
			})
		case *AtElement:
			r = append(r, &msg.Elem{
				Text: &msg.Text{
					Str: e.Display,
					Attr6Buf: binary.NewWriterF(func(w *binary.Writer) {
						w.WriteUInt16(1)
						w.WriteUInt16(0)
						w.WriteUInt16(uint16(len(e.Display)))
						w.WriteByte(func() byte {
							if e.Target == 0 {
								return 1
							}
							return 0
						}())
						w.WriteUInt32(uint32(e.Target))
						w.WriteUInt16(0)
					}),
				},
			})
			r = append(r, &msg.Elem{Text: &msg.Text{Str: " "}})
		case *GroupImageElement:
			r = append(r, &msg.Elem{
				CustomFace: &msg.CustomFace{
					FilePath: e.ImageId,
					Md5:      e.Md5[:],
					Flag:     make([]byte, 4),
					OldData: []byte{0x15, 0x36, 0x20, 0x39, 0x32, 0x6B, 0x41, 0x31, 0x00, 0x38, 0x37, 0x32, 0x66, 0x30, 0x36, 0x36, 0x30, 0x33, 0x61, 0x65, 0x31, 0x30, 0x33, 0x62, 0x37, 0x20, 0x20, 0x20, 0x20, 0x20,
						0x20, 0x35, 0x30, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x7B, 0x30, 0x31, 0x45, 0x39, 0x34, 0x35, 0x31, 0x42, 0x2D, 0x37, 0x30, 0x45, 0x44,
						0x2D, 0x45, 0x41, 0x45, 0x33, 0x2D, 0x42, 0x33, 0x37, 0x43, 0x2D, 0x31, 0x30, 0x31, 0x46, 0x31, 0x45, 0x45, 0x42, 0x46, 0x35, 0x42, 0x35, 0x7D, 0x2E, 0x70, 0x6E, 0x67, 0x41},
				},
			})
		}
	}
	return
}
