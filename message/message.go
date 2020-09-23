package message

import (
	"crypto/md5"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/gjson"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type (
	PrivateMessage struct {
		Id         int32
		InternalId int32
		Target     int64
		Time       int32
		Sender     *Sender
		Elements   []IMessageElement
	}

	TempMessage struct {
		Id        int32
		GroupCode int64
		GroupName string
		Sender    *Sender
		Elements  []IMessageElement
	}

	GroupMessage struct {
		Id         int32
		InternalId int32
		GroupCode  int64
		GroupName  string
		Sender     *Sender
		Time       int32
		Elements   []IMessageElement
		//OriginalElements []*msg.Elem
	}

	SendingMessage struct {
		Elements []IMessageElement
	}

	ForwardMessage struct {
		Nodes []*ForwardNode
	}

	ForwardNode struct {
		SenderId   int64
		SenderName string
		Time       int32
		Message    []IMessageElement
	}

	Sender struct {
		Uin      int64
		Nickname string
		CardName string
		IsFriend bool
	}

	IMessageElement interface {
		Type() ElementType
	}

	ElementType int
)

const (
	Text ElementType = iota
	Image
	Face
	At
	Reply
	Service
	Forward
	File
	Voice
	Video
	LightApp
	RedBag
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

func (msg *TempMessage) ToString() (res string) {
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
		case *RedBagElement:
			res += "[RedBag:" + e.Title + "]"
		case *ReplyElement:
			res += "[Reply:" + strconv.FormatInt(int64(e.ReplySeq), 10) + "]"
		}
	}
	return
}

func (msg *SendingMessage) Append(e IMessageElement) *SendingMessage {
	v := reflect.ValueOf(e)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		msg.Elements = append(msg.Elements, e)
	}
	return msg
}

func (msg *SendingMessage) Any(filter func(e IMessageElement) bool) bool {
	for _, e := range msg.Elements {
		if filter(e) {
			return true
		}
	}
	return false
}

func (msg *SendingMessage) FirstOrNil(filter func(e IMessageElement) bool) IMessageElement {
	for _, e := range msg.Elements {
		if filter(e) {
			return e
		}
	}
	return nil
}

func (msg *SendingMessage) Count(filter func(e IMessageElement) bool) (c int) {
	for _, e := range msg.Elements {
		if filter(e) {
			c++
		}
	}
	return
}

func (msg *SendingMessage) ToFragmented() [][]IMessageElement {
	var fragmented [][]IMessageElement
	for _, elem := range msg.Elements {
		switch o := elem.(type) {
		case *TextElement:
			for _, text := range utils.ChunkString(o.Content, 220) {
				fragmented = append(fragmented, []IMessageElement{NewText(text)})
			}
		default:
			fragmented = append(fragmented, []IMessageElement{o})
		}
	}
	return fragmented
}

func EstimateLength(elems []IMessageElement, limit int) int {
	sum := 0
	for _, elem := range elems {
		if sum >= limit {
			break
		}
		left := int(math.Max(float64(limit-sum), 0))
		switch e := elem.(type) {
		case *TextElement:
			sum += utils.ChineseLength(e.Content, left)
		case *AtElement:
			sum += utils.ChineseLength(e.Display, left)
		case *ReplyElement:
			sum += 444 + EstimateLength(e.Elements, left)
		case *ImageElement, *GroupImageElement, *FriendImageElement:
			sum += 260
		default:
			sum += utils.ChineseLength(ToReadableString([]IMessageElement{elem}), left)
		}
	}
	return sum
}

func (s *Sender) DisplayName() string {
	if s.CardName == "" {
		return s.Nickname
	}
	return s.CardName
}

func ToProtoElems(elems []IMessageElement, generalFlags bool) (r []*msg.Elem) {
	if len(elems) == 0 {
		return nil
	}
	for _, elem := range elems {
		if reply, ok := elem.(*ReplyElement); ok {
			r = append(r, &msg.Elem{
				SrcMsg: &msg.SourceMsg{
					OrigSeqs:  []int32{reply.ReplySeq},
					SenderUin: reply.Sender,
					Time:      reply.Time,
					Flag:      1,
					Elems:     ToSrcProtoElems(reply.Elements),
					RichMsg:   []byte{},
					PbReserve: []byte{},
					SrcMsg:    []byte{},
					TroopName: []byte{},
				},
			})
		}
	}
	imgOld := []byte{0x15, 0x36, 0x20, 0x39, 0x32, 0x6B, 0x41, 0x31, 0x00, 0x38, 0x37, 0x32, 0x66, 0x30, 0x36, 0x36, 0x30, 0x33, 0x61, 0x65, 0x31, 0x30, 0x33, 0x62, 0x37, 0x20, 0x20, 0x20, 0x20, 0x20,
		0x20, 0x35, 0x30, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x7B, 0x30, 0x31, 0x45, 0x39, 0x34, 0x35, 0x31, 0x42, 0x2D, 0x37, 0x30, 0x45, 0x44,
		0x2D, 0x45, 0x41, 0x45, 0x33, 0x2D, 0x42, 0x33, 0x37, 0x43, 0x2D, 0x31, 0x30, 0x31, 0x46, 0x31, 0x45, 0x45, 0x42, 0x46, 0x35, 0x42, 0x35, 0x7D, 0x2E, 0x70, 0x6E, 0x67, 0x41}
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
						w.WriteUInt16(uint16(len([]rune(e.Display))))
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
		case *ImageElement:
			r = append(r, &msg.Elem{
				CustomFace: &msg.CustomFace{
					FilePath: e.Filename,
					Md5:      e.Md5,
					Size:     e.Size,
					Flag:     make([]byte, 4),
					OldData:  imgOld,
				},
			})
		case *GroupImageElement:
			r = append(r, &msg.Elem{
				CustomFace: &msg.CustomFace{
					FileType: 66,
					Useful:   1,
					Origin:   1,
					FileId:   int32(e.FileId),
					FilePath: e.ImageId,
					Size:     e.Size,
					Md5:      e.Md5[:],
					Flag:     make([]byte, 4),
					//OldData:  imgOld,
				},
			})
		case *FriendImageElement:
			r = append(r, &msg.Elem{
				NotOnlineImage: &msg.NotOnlineImage{
					FilePath:     e.ImageId,
					ResId:        e.ImageId,
					OldPicMd5:    false,
					PicMd5:       e.Md5,
					DownloadPath: e.ImageId,
					Original:     1,
					PbReserve:    []byte{0x78, 0x02},
				},
			})
		case *ServiceElement:
			if e.Id == 35 {
				r = append(r, &msg.Elem{
					RichMsg: &msg.RichMsg{
						Template1: append([]byte{1}, binary.ZlibCompress([]byte(e.Content))...),
						ServiceId: e.Id,
						MsgResId:  []byte{},
					},
				})
				r = append(r, &msg.Elem{
					Text: &msg.Text{
						Str: "你的QQ暂不支持查看[转发多条消息]，请期待后续版本。",
					},
				})
				continue
			}
			if e.Id == 33 {
				r = append(r, &msg.Elem{
					Text: &msg.Text{Str: e.ResId},
				})
				r = append(r, &msg.Elem{
					RichMsg: &msg.RichMsg{
						Template1: append([]byte{1}, binary.ZlibCompress([]byte(e.Content))...),
						ServiceId: e.Id,
						MsgResId:  []byte{},
					},
				})
				continue
			}
			r = append(r, &msg.Elem{
				RichMsg: &msg.RichMsg{
					Template1: append([]byte{1}, binary.ZlibCompress([]byte(e.Content))...),
					ServiceId: e.Id,
				},
			})
		case *LightAppElement:
			r = append(r, &msg.Elem{
				LightApp: &msg.LightAppElem{
					Data:     append([]byte{1}, binary.ZlibCompress([]byte(e.Content))...),
					MsgResid: []byte{1},
				},
			})
		}
	}
	if generalFlags {
	L:
		for _, elem := range elems {
			switch e := elem.(type) {
			case *ServiceElement:
				if e.SubType == "Long" {
					r = append(r, &msg.Elem{
						GeneralFlags: &msg.GeneralFlags{
							LongTextFlag:  1,
							LongTextResid: e.ResId,
							PbReserve:     []byte{0x78, 0x00, 0xF8, 0x01, 0x00, 0xC8, 0x02, 0x00},
						},
					})
					break L
				}
				//d, _ := hex.DecodeString("08097800C80100F00100F80100900200C80200980300A00320B00300C00300D00300E803008A04020803900480808010B80400C00400")
				r = append(r, &msg.Elem{
					GeneralFlags: &msg.GeneralFlags{
						PbReserve: []byte{
							0x08, 0x09, 0x78, 0x00, 0xC8, 0x01, 0x00, 0xF0, 0x01, 0x00, 0xF8, 0x01, 0x00, 0x90, 0x02, 0x00,
							0xC8, 0x02, 0x00, 0x98, 0x03, 0x00, 0xA0, 0x03, 0x20, 0xB0, 0x03, 0x00, 0xC0, 0x03, 0x00, 0xD0,
							0x03, 0x00, 0xE8, 0x03, 0x00, 0x8A, 0x04, 0x02, 0x08, 0x03, 0x90, 0x04, 0x80, 0x80, 0x80, 0x10,
							0xB8, 0x04, 0x00, 0xC0, 0x04, 0x00,
						},
					},
				})
				break L
			}
		}
	}
	return
}

func ToSrcProtoElems(elems []IMessageElement) (r []*msg.Elem) {
	for _, elem := range elems {
		switch e := elem.(type) {
		case *ImageElement, *GroupImageElement, *FriendImageElement:
			r = append(r, &msg.Elem{
				Text: &msg.Text{
					Str: "[图片]",
				},
			})
		default:
			r = append(r, ToProtoElems([]IMessageElement{e}, false)...)
		}
	}
	return
}

func ParseMessageElems(elems []*msg.Elem) []IMessageElement {
	var res []IMessageElement
	for _, elem := range elems {
		if elem.SrcMsg != nil {
			if len(elem.SrcMsg.OrigSeqs) != 0 {
				r := &ReplyElement{
					ReplySeq: elem.SrcMsg.OrigSeqs[0],
					Time:     elem.SrcMsg.Time,
					Sender:   elem.SrcMsg.SenderUin,
					Elements: ParseMessageElems(elem.SrcMsg.Elems),
				}
				res = append(res, r)
			}
			continue
		}
		if elem.TransElemInfo != nil {
			if elem.TransElemInfo.ElemType == 24 { // QFile
				i3 := len(elem.TransElemInfo.ElemValue)
				r := binary.NewReader(elem.TransElemInfo.ElemValue)
				if i3 > 3 {
					if r.ReadByte() == 1 {
						pb := r.ReadBytes(int(r.ReadUInt16()))
						objMsg := msg.ObjMsg{}
						if err := proto.Unmarshal(pb, &objMsg); err == nil && len(objMsg.MsgContentInfo) > 0 {
							info := objMsg.MsgContentInfo[0]
							res = append(res, &GroupFileElement{
								Name:  info.MsgFile.FileName,
								Size:  info.MsgFile.FileSize,
								Path:  string(info.MsgFile.FilePath),
								Busid: info.MsgFile.BusId,
							})
						}
					}
				}
			}
		}
		if elem.LightApp != nil && len(elem.LightApp.Data) > 1 {
			var content string
			if elem.LightApp.Data[0] == 0 {
				content = string(elem.LightApp.Data[1:])
			}
			if elem.LightApp.Data[0] == 1 {
				content = string(binary.ZlibUncompress(elem.LightApp.Data[1:]))
			}
			if content != "" {
				// TODO: 解析具体的APP
				return append(res, &LightAppElement{Content: content})
			}
		}
		if elem.VideoFile != nil {
			return append(res, &ShortVideoElement{
				Name: string(elem.VideoFile.FileName),
				Uuid: elem.VideoFile.FileUuid,
				Size: elem.VideoFile.FileSize,
				Md5:  elem.VideoFile.FileMd5,
			})
		}
		if elem.Text != nil {
			if len(elem.Text.Attr6Buf) == 0 {
				res = append(res, NewText(func() string {
					// 这么处理应该没问题
					if strings.Contains(elem.Text.Str, "\r") && !strings.Contains(elem.Text.Str, "\r\n") {
						return strings.ReplaceAll(elem.Text.Str, "\r", "\r\n")
					}
					return elem.Text.Str
				}()))
			} else {
				att6 := binary.NewReader(elem.Text.Attr6Buf)
				att6.ReadBytes(7)
				target := int64(uint32(att6.ReadInt32()))
				res = append(res, NewAt(target, elem.Text.Str))
			}
		}
		if elem.RichMsg != nil {
			var content string
			if elem.RichMsg.Template1[0] == 0 {
				content = string(elem.RichMsg.Template1[1:])
			}
			if elem.RichMsg.Template1[0] == 1 {
				content = string(binary.ZlibUncompress(elem.RichMsg.Template1[1:]))
			}
			if content != "" {
				if elem.RichMsg.ServiceId == 35 {
					reg := regexp.MustCompile(`m_resid="(\w+?.*?)"`)
					res = append(res, &ForwardElement{ResId: reg.FindAllStringSubmatch(content, -1)[0][1]})
					continue
				}
				if elem.RichMsg.ServiceId == 33 {
					continue // 前面一个 elem 已经解析到链接
				}
				if isOk := strings.Contains(content, "<?xml"); isOk {
					res = append(res, NewRichXml(content, int64(elem.RichMsg.ServiceId)))
					continue
				} else {
					if gjson.Valid(content) {
						res = append(res, NewRichJson(content))
						continue
					}
				}
				res = append(res, NewText(content))
			}
		}
		if elem.CustomFace != nil {
			res = append(res, &ImageElement{
				Filename: elem.CustomFace.FilePath,
				Size:     elem.CustomFace.Size,
				Width:    elem.CustomFace.Width,
				Height:   elem.CustomFace.Height,
				Url: func() string {
					if elem.CustomFace.OrigUrl == "" {
						return "http://gchat.qpic.cn/gchatpic_new/0/0-0-" + strings.ReplaceAll(binary.CalculateImageResourceId(elem.CustomFace.Md5)[1:37], "-", "") + "/0?term=2"
					}
					return "http://gchat.qpic.cn" + elem.CustomFace.OrigUrl
				}(),
				Md5: elem.CustomFace.Md5,
			})
		}
		if elem.NotOnlineImage != nil {
			var img string
			if elem.NotOnlineImage.OrigUrl != "" {
				img = "http://c2cpicdw.qpic.cn" + elem.NotOnlineImage.OrigUrl
			} else {
				img = "http://c2cpicdw.qpic.cn/offpic_new/0/" + elem.NotOnlineImage.ResId + "/0?term=2"
			}
			res = append(res, &ImageElement{
				Filename: elem.NotOnlineImage.FilePath,
				Size:     elem.NotOnlineImage.FileLen,
				Url:      img,
				Md5:      elem.NotOnlineImage.PicMd5,
			})
		}
		if elem.QQWalletMsg != nil && elem.QQWalletMsg.AioBody != nil {
			msgType := elem.QQWalletMsg.AioBody.MsgType
			if msgType == 2 || msgType == 3 || msgType == 6 {
				return []IMessageElement{
					&RedBagElement{
						MsgType: RedBagMessageType(msgType),
						Title:   elem.QQWalletMsg.AioBody.Receiver.Title,
					},
				}
			}
		}
		if elem.Face != nil {
			res = append(res, NewFace(elem.Face.Index))
		}
	}
	return res
}

func (forMsg *ForwardMessage) CalculateValidationData(seq, random int32, groupCode int64) ([]byte, []byte) {
	var msgs []*msg.Message
	for _, node := range forMsg.Nodes {
		msgs = append(msgs, &msg.Message{
			Head: &msg.MessageHead{
				FromUin: node.SenderId,
				MsgSeq:  seq,
				MsgTime: node.Time,
				MsgUid:  0x01000000000000000 | (int64(random) & 0xFFFFFFFF),
				MutiltransHead: &msg.MutilTransHead{
					MsgId: 1,
				},
				MsgType: 82,
				GroupInfo: &msg.GroupInfo{
					GroupCode: groupCode,
					GroupRank: []byte{},
					GroupName: []byte{},
					GroupCard: node.SenderName,
				},
			},
			Body: &msg.MessageBody{
				RichText: &msg.RichText{
					Elems: ToProtoElems(node.Message, false),
				},
			},
		})
	}
	buf, _ := proto.Marshal(&msg.PbMultiMsgNew{Msg: msgs})
	trans := &msg.PbMultiMsgTransmit{Msg: msgs, PbItemList: []*msg.PbMultiMsgItem{
		{
			FileName: "MultiMsg",
			Buffer:   buf,
		},
	}}
	b, _ := proto.Marshal(trans)
	data := binary.GZipCompress(b)
	hash := md5.Sum(data)
	return data, hash[:]
}

func ToReadableString(m []IMessageElement) (r string) {
	for _, elem := range m {
		switch e := elem.(type) {
		case *TextElement:
			r += e.Content
		case *ImageElement:
			r += "[图片]"
		case *FaceElement:
			r += "/" + e.Name
		case *GroupImageElement:
			r += "[图片]"
		case *AtElement:
			r += e.Display
		}
	}
	return
}
