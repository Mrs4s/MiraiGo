package message

import (
	"crypto/md5"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/gjson"
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
		Id             int32
		InternalId     int32
		GroupCode      int64
		GroupName      string
		Sender         *Sender
		Time           int32
		Elements       []IMessageElement
		OriginalObject *msg.Message
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

	RichMessage struct {
		Title      string
		Summary    string
		Brief      string
		Url        string
		PictureUrl string
		MusicUrl   string
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

	IRichMessageElement interface {
		Pack() []*msg.Elem
	}

	ElementType int

	GroupGift int
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

	HoldingYourHand GroupGift = 280
	CuteCat         GroupGift = 281
	MysteryMask     GroupGift = 284
	SweetWink       GroupGift = 285
	ImBusy          GroupGift = 286
	HappyCola       GroupGift = 289
	LuckyBracelet   GroupGift = 290
	Cappuccino      GroupGift = 299
	CatWatch        GroupGift = 302
	FleeceGloves    GroupGift = 307
	RainbowCandy    GroupGift = 308
	LoveMask        GroupGift = 312
	Stronger        GroupGift = 313
	LoveMicrophone  GroupGift = 367
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
		case *FriendFlashImgElement:
			// NOTE: ignore other components
			return "[Image (flash):" + e.Filename + "]"
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
		case *GroupFlashImgElement:
			// NOTE: ignore other components
			return "[Image (flash):" + e.Filename + "]"
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
			for _, text := range utils.ChunkString(o.Content, 80) {
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
					SenderUin: &reply.Sender,
					Time:      &reply.Time,
					Flag:      proto.Int32(1),
					Elems:     ToSrcProtoElems(reply.Elements),
					RichMsg:   []byte{},
					PbReserve: []byte{},
					SrcMsg:    []byte{},
					TroopName: []byte{},
				},
			})
		}
	}
	for _, elem := range elems {
		if e, ok := elem.(IRichMessageElement); ok {
			r = append(r, e.Pack()...)
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
							LongTextFlag:  proto.Int32(1),
							LongTextResid: &e.ResId,
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
					Str: proto.String("[图片]"),
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
					Time:     elem.SrcMsg.GetTime(),
					Sender:   elem.SrcMsg.GetSenderUin(),
					Elements: ParseMessageElems(elem.SrcMsg.Elems),
				}
				res = append(res, r)
			}
			continue
		}
		if elem.TransElemInfo != nil {
			if elem.TransElemInfo.GetElemType() == 24 { // QFile
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
			var content []byte
			if elem.LightApp.Data[0] == 0 {
				content = elem.LightApp.Data[1:]
			}
			if elem.LightApp.Data[0] == 1 {
				content = binary.ZlibUncompress(elem.LightApp.Data[1:])
			}
			if len(content) > 0 && len(content) < 1024*1024*1024 { // 解析出错 or 非法内容
				// TODO: 解析具体的APP
				return append(res, &LightAppElement{Content: string(content)})
			}
		}
		if elem.VideoFile != nil {
			return append(res, &ShortVideoElement{
				Name: string(elem.VideoFile.FileName),
				Uuid: elem.VideoFile.FileUuid,
				Size: elem.VideoFile.GetFileSize(),
				Md5:  elem.VideoFile.FileMd5,
			})
		}
		if elem.Text != nil {
			if len(elem.Text.Attr6Buf) == 0 {
				res = append(res, NewText(func() string {
					// 这么处理应该没问题
					if strings.Contains(elem.Text.GetStr(), "\r") && !strings.Contains(elem.Text.GetStr(), "\r\n") {
						return strings.ReplaceAll(elem.Text.GetStr(), "\r", "\r\n")
					}
					return elem.Text.GetStr()
				}()))
			} else {
				att6 := binary.NewReader(elem.Text.Attr6Buf)
				att6.ReadBytes(7)
				target := int64(uint32(att6.ReadInt32()))
				res = append(res, NewAt(target, elem.Text.GetStr()))
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
				if elem.RichMsg.GetServiceId() == 35 {
					reg := regexp.MustCompile(`m_resid="(\w+?.*?)"`)
					sub := reg.FindAllStringSubmatch(content, -1)
					if len(sub) > 0 && len(sub[0]) > 1 {
						res = append(res, &ForwardElement{ResId: reg.FindAllStringSubmatch(content, -1)[0][1]})
						continue
					}
				}
				if elem.RichMsg.GetServiceId() == 33 {
					continue // 前面一个 elem 已经解析到链接
				}
				if isOk := strings.Contains(content, "<?xml"); isOk {
					res = append(res, NewRichXml(content, int64(elem.RichMsg.GetServiceId())))
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
			if len(elem.CustomFace.Md5) == 0 {
				continue
			}
			res = append(res, &ImageElement{
				Filename: elem.CustomFace.GetFilePath(),
				Size:     elem.CustomFace.GetSize(),
				Width:    elem.CustomFace.GetWidth(),
				Height:   elem.CustomFace.GetHeight(),
				Url: func() string {
					if elem.CustomFace.GetOrigUrl() == "" {
						return "http://gchat.qpic.cn/gchatpic_new/0/0-0-" + strings.ReplaceAll(binary.CalculateImageResourceId(elem.CustomFace.Md5)[1:37], "-", "") + "/0?term=2"
					}
					return "http://gchat.qpic.cn" + elem.CustomFace.GetOrigUrl()
				}(),
				Md5: elem.CustomFace.Md5,
			})
		}
		if elem.NotOnlineImage != nil {
			var img string
			if elem.NotOnlineImage.GetOrigUrl() != "" {
				img = "http://c2cpicdw.qpic.cn" + elem.NotOnlineImage.GetOrigUrl()
			} else {
				img = "http://c2cpicdw.qpic.cn/offpic_new/0/" + elem.NotOnlineImage.GetResId() + "/0?term=2"
			}
			res = append(res, &ImageElement{
				Filename: elem.NotOnlineImage.GetFilePath(),
				Size:     elem.NotOnlineImage.GetFileLen(),
				Url:      img,
				Md5:      elem.NotOnlineImage.PicMd5,
			})
		}
		if elem.QQWalletMsg != nil && elem.QQWalletMsg.AioBody != nil {
			msgType := elem.QQWalletMsg.AioBody.GetMsgType()
			if msgType == 2 || msgType == 3 || msgType == 6 {
				return []IMessageElement{
					&RedBagElement{
						MsgType: RedBagMessageType(msgType),
						Title:   elem.QQWalletMsg.AioBody.Receiver.GetTitle(),
					},
				}
			}
		}
		if elem.Face != nil {
			res = append(res, NewFace(elem.Face.GetIndex()))
		}
		if elem.CommonElem != nil {
			switch elem.CommonElem.GetServiceType() {
			case 3:
				flash := &msg.MsgElemInfoServtype3{}
				_ = proto.Unmarshal(elem.CommonElem.PbElem, flash)
				if flash.FlashTroopPic != nil {
					res = append(res, &GroupFlashImgElement{
						ImageElement{
							Filename: flash.FlashTroopPic.GetFilePath(),
							Size:     flash.FlashTroopPic.GetSize(),
							Width:    flash.FlashTroopPic.GetWidth(),
							Height:   flash.FlashTroopPic.GetHeight(),
							Md5:      flash.FlashTroopPic.Md5,
						},
					})
					return res
				}
				if flash.FlashC2CPic != nil {
					res = append(res, &GroupFlashImgElement{
						ImageElement{
							Filename: flash.FlashC2CPic.GetFilePath(),
							Size:     flash.FlashC2CPic.GetFileLen(),
							Md5:      flash.FlashC2CPic.PicMd5,
						},
					})
					return res
				}
			case 33:
				newSysFaceMsg := &msg.MsgElemInfoServtype33{}
				_ = proto.Unmarshal(elem.CommonElem.PbElem, newSysFaceMsg)
				res = append(res, NewFace(int32(newSysFaceMsg.GetIndex())))
			}
		}
	}
	return res
}

func (forMsg *ForwardMessage) CalculateValidationData(seq, random int32, groupCode int64) ([]byte, []byte) {
	var msgs []*msg.Message
	for _, node := range forMsg.Nodes {
		msgs = append(msgs, &msg.Message{
			Head: &msg.MessageHead{
				FromUin: &node.SenderId,
				MsgSeq:  &seq,
				MsgTime: &node.Time,
				MsgUid:  proto.Int64(0x01000000000000000 | (int64(random) & 0xFFFFFFFF)),
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
	buf, _ := proto.Marshal(&msg.PbMultiMsgNew{Msg: msgs})
	trans := &msg.PbMultiMsgTransmit{Msg: msgs, PbItemList: []*msg.PbMultiMsgItem{
		{
			FileName: proto.String("MultiMsg"),
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
		// NOTE: flash pic is singular
		// To be clarified
		// case *GroupFlashImgElement:
		// 	return "[闪照]"
		// case *FriendFlashImgElement:
		// 	return "[闪照]"
		case *AtElement:
			r += e.Display
		}
	}
	return
}
