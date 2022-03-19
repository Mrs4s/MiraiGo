package message

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

type (
	PrivateMessage struct {
		Id         int32
		InternalId int32
		Self       int64
		Target     int64
		Time       int32
		Sender     *Sender
		Elements   []IMessageElement
	}

	TempMessage struct {
		Id        int32
		GroupCode int64
		GroupName string
		Self      int64
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
		// OriginalElements []*msg.Elem
	}

	SendingMessage struct {
		Elements []IMessageElement
	}

	Sender struct {
		Uin           int64
		Nickname      string
		CardName      string
		AnonymousInfo *AnonymousInfo
		IsFriend      bool
	}

	AnonymousInfo struct {
		AnonymousId   string
		AnonymousNick string
	}

	IMessageElement interface {
		Type() ElementType
	}

	IRichMessageElement interface {
		Pack() []*msg.Elem
	}

	ElementType int
)

// MusicType values.
const (
	QQMusic    = iota // QQ音乐
	CloudMusic        // 网易云音乐
	MiguMusic         // 咪咕音乐
	KugouMusic        // 酷狗音乐
	KuwoMusic         // 酷我音乐
)

//go:generate stringer -type ElementType -linecomment
const (
	Text     ElementType = iota // 文本
	Image                       // 图片
	Face                        // 表情
	At                          // 艾特
	Reply                       // 回复
	Service                     // 服务
	Forward                     // 转发
	File                        // 文件
	Voice                       // 语音
	Video                       // 视频
	LightApp                    // 轻应用
	RedBag                      // 红包
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
		case *FaceElement:
			res += "[" + e.Name + "]"
		case *MarketFaceElement:
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
			for _, text := range utils.ChunkString(o.Content, 80) {
				fragmented = append(fragmented, []IMessageElement{NewText(text)})
			}
		default:
			fragmented = append(fragmented, []IMessageElement{o})
		}
	}
	return fragmented
}

// 单条消息发送的大小限制（预估）
const MaxMessageSize = 5000

func EstimateLength(elems []IMessageElement) int {
	sum := 0
	for _, elem := range elems {
		switch e := elem.(type) {
		case *TextElement:
			sum += len(e.Content)
		case *AtElement:
			sum += len(e.Display)
		case *ReplyElement:
			sum += 444 + EstimateLength(e.Elements)
		case *GroupImageElement, *FriendImageElement:
			sum += 100
		default:
			sum += len(ToReadableString([]IMessageElement{elem}))
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
			if len(elems) > 1 {
				if elems[0].Type() == Image || elems[1].Type() == Image {
					r = append(r, &msg.Elem{Text: &msg.Text{Str: proto.String(" ")}})
				}
			}
		}
	}
	for _, elem := range elems {
		if e, ok := elem.(*ShortVideoElement); ok {
			r = e.Pack()
			break
		}
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
				// d, _ := hex.DecodeString("08097800C80100F00100F80100900200C80200980300A00320B00300C00300D00300E803008A04020803900480808010B80400C00400")
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

var photoTextElem IMessageElement = NewText("[图片]")

func ToSrcProtoElems(elems []IMessageElement) []*msg.Elem {
	elems2 := make([]IMessageElement, len(elems))
	copy(elems2, elems)
	for i, elem := range elems2 {
		if elem.Type() == Image {
			elems2[i] = photoTextElem
		}
	}
	return ToProtoElems(elems2, false)
}

func ParseMessageElems(elems []*msg.Elem) []IMessageElement {
	var res []IMessageElement
	for _, elem := range elems {
		if elem.SrcMsg != nil && len(elem.SrcMsg.OrigSeqs) != 0 {
			r := &ReplyElement{
				ReplySeq: elem.SrcMsg.OrigSeqs[0],
				Time:     elem.SrcMsg.GetTime(),
				Sender:   elem.SrcMsg.GetSenderUin(),
				GroupID:  elem.SrcMsg.GetToUin(),
				Elements: ParseMessageElems(elem.SrcMsg.Elems),
			}
			res = append(res, r)
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
			return []IMessageElement{
				&ShortVideoElement{
					Name:      string(elem.VideoFile.FileName),
					Uuid:      elem.VideoFile.FileUuid,
					Size:      elem.VideoFile.GetFileSize(),
					ThumbSize: elem.VideoFile.GetThumbFileSize(),
					Md5:       elem.VideoFile.FileMd5,
					ThumbMd5:  elem.VideoFile.ThumbFileMd5,
				},
			}
		}
		if elem.Text != nil {
			switch {
			case len(elem.Text.Attr6Buf) > 0:
				att6 := binary.NewReader(elem.Text.Attr6Buf)
				att6.ReadBytes(7)
				target := int64(uint32(att6.ReadInt32()))
				at := NewAt(target, elem.Text.GetStr())
				at.SubType = AtTypeGroupMember
				res = append(res, at)
			case len(elem.Text.PbReserve) > 0:
				resv := new(msg.TextResvAttr)
				_ = proto.Unmarshal(elem.Text.PbReserve, resv)
				if resv.GetAtType() == 2 {
					at := NewAt(int64(resv.GetAtMemberTinyid()), elem.Text.GetStr())
					at.SubType = AtTypeGuildMember
					res = append(res, at)
					break
				}
				if resv.GetAtType() == 4 {
					at := NewAt(int64(resv.AtChannelInfo.GetChannelId()), elem.Text.GetStr())
					at.SubType = AtTypeGuildChannel
					res = append(res, at)
					break
				}
				fallthrough
			default:
				res = append(res, NewText(func() string {
					// 这么处理应该没问题
					if strings.Contains(elem.Text.GetStr(), "\r") && !strings.Contains(elem.Text.GetStr(), "\r\n") {
						return strings.ReplaceAll(elem.Text.GetStr(), "\r", "\r\n")
					}
					return elem.Text.GetStr()
				}()))
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
					elem := forwardMsgFromXML(content)
					if elem != nil {
						res = append(res, elem)
						continue
					}
				}
				if elem.RichMsg.GetServiceId() == 33 {
					continue // 前面一个 elem 已经解析到链接
				}
				if isOk := strings.Contains(content, "<?xml"); isOk {
					res = append(res, NewRichXml(content, int64(elem.RichMsg.GetServiceId())))
					continue
				} else if json.Valid(utils.S2B(content)) {
					res = append(res, NewRichJson(content))
					continue
				}
				res = append(res, NewText(content))
			}
		}
		if elem.CustomFace != nil {
			if len(elem.CustomFace.Md5) == 0 {
				continue
			}
			var url string
			if elem.CustomFace.GetOrigUrl() == "" {
				url = "https://gchat.qpic.cn/gchatpic_new/0/0-0-" + strings.ReplaceAll(binary.CalculateImageResourceId(elem.CustomFace.Md5)[1:37], "-", "") + "/0?term=2"
			} else {
				url = "https://gchat.qpic.cn" + elem.CustomFace.GetOrigUrl()
			}
			if strings.Contains(elem.CustomFace.GetOrigUrl(), "qmeet") {
				res = append(res, &GuildImageElement{
					FileId:   int64(elem.CustomFace.GetFileId()),
					FilePath: elem.CustomFace.GetFilePath(),
					Size:     elem.CustomFace.GetSize(),
					Width:    elem.CustomFace.GetWidth(),
					Height:   elem.CustomFace.GetHeight(),
					Url:      url,
					Md5:      elem.CustomFace.Md5,
				})
				continue
			}
			res = append(res, &GroupImageElement{
				FileId:  int64(elem.CustomFace.GetFileId()),
				ImageId: elem.CustomFace.GetFilePath(),
				Size:    elem.CustomFace.GetSize(),
				Width:   elem.CustomFace.GetWidth(),
				Height:  elem.CustomFace.GetHeight(),
				Url:     url,
				ImageBizType: func() ImageBizType {
					if len(elem.CustomFace.PbReserve) == 0 {
						return UnknownBizType
					}
					attr := new(msg.ResvAttr)
					if proto.Unmarshal(elem.CustomFace.PbReserve, attr) != nil {
						return UnknownBizType
					}
					return ImageBizType(attr.GetImageBizType())
				}(),
				Md5: elem.CustomFace.Md5,
			})
		}
		if elem.MarketFace != nil {
			face := &MarketFaceElement{
				Name:       utils.B2S(elem.MarketFace.FaceName),
				FaceId:     elem.MarketFace.FaceId,
				TabId:      int32(elem.MarketFace.GetTabId()),
				ItemType:   int32(elem.MarketFace.GetItemType()),
				SubType:    int32(elem.MarketFace.GetSubType()),
				MediaType:  int32(elem.MarketFace.GetMediaType()),
				EncryptKey: elem.MarketFace.Key,
				MagicValue: utils.B2S(elem.MarketFace.Mobileparam),
			}
			if face.Name == "[骰子]" || face.Name == "[随机骰子]" {
				return []IMessageElement{
					&DiceElement{
						MarketFaceElement: face,
						Value: func() int32 {
							v := strings.SplitN(face.MagicValue, "=", 2)[1]
							t, _ := strconv.ParseInt(v, 10, 32)
							return int32(t) + 1
						}(),
					},
				}
			}
			if face.Name == "[猜拳]" {
				v := strings.SplitN(face.MagicValue, "=", 2)[1]
				t, _ := strconv.ParseInt(v, 10, 32)
				return []IMessageElement{
					&FingerGuessingElement{
						MarketFaceElement: face,
						Value:             int32(t),
						Name:              fingerGuessingName[int32(t)],
					},
				}
			}
			return []IMessageElement{face}
		}
		if elem.NotOnlineImage != nil {
			var img string
			if elem.NotOnlineImage.GetOrigUrl() != "" {
				img = "https://c2cpicdw.qpic.cn" + elem.NotOnlineImage.GetOrigUrl()
			} else {
				img = "https://c2cpicdw.qpic.cn/offpic_new/0"
				downloadPath := elem.NotOnlineImage.GetResId()
				if elem.NotOnlineImage.GetDownloadPath() != "" {
					downloadPath = elem.NotOnlineImage.GetDownloadPath()
				}
				if !strings.HasPrefix(downloadPath, "/") {
					img += "/"
				}
				img += downloadPath + "/0?term=3"
			}
			res = append(res, &FriendImageElement{
				ImageId: elem.NotOnlineImage.GetFilePath(),
				Size:    elem.NotOnlineImage.GetFileLen(),
				Url:     img,
				Md5:     elem.NotOnlineImage.PicMd5,
			})
		}
		if elem.QQWalletMsg != nil && elem.QQWalletMsg.AioBody != nil {
			// /com/tencent/mobileqq/data/MessageForQQWalletMsg.java#L366
			msgType := elem.QQWalletMsg.AioBody.GetMsgType()
			if msgType <= 1000 && elem.QQWalletMsg.AioBody.RedType != nil {
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
					res = append(res, &GroupImageElement{
						FileId:  int64(flash.FlashTroopPic.GetFileId()),
						ImageId: flash.FlashTroopPic.GetFilePath(),
						Size:    flash.FlashTroopPic.GetSize(),
						Width:   flash.FlashTroopPic.GetWidth(),
						Height:  flash.FlashTroopPic.GetHeight(),
						Md5:     flash.FlashTroopPic.Md5,
						Flash:   true,
					})
					return res
				}
				if flash.FlashC2CPic != nil {
					res = append(res, &FriendImageElement{
						ImageId: flash.FlashC2CPic.GetFilePath(),
						Size:    flash.FlashC2CPic.GetFileLen(),
						Md5:     flash.FlashC2CPic.PicMd5,
						Flash:   true,
					})
					return res
				}
			case 33:
				newSysFaceMsg := &msg.MsgElemInfoServtype33{}
				_ = proto.Unmarshal(elem.CommonElem.PbElem, newSysFaceMsg)
				res = append(res, NewFace(int32(newSysFaceMsg.GetIndex())))
			case 37:
				animatedStickerMsg := &msg.MsgElemInfoServtype37{}
				_ = proto.Unmarshal(elem.CommonElem.PbElem, animatedStickerMsg)
				sticker := &AnimatedSticker{
					ID:   int32(animatedStickerMsg.GetQsid()),
					Name: strings.TrimPrefix(string(animatedStickerMsg.Text), "/"),
				}
				return []IMessageElement{sticker} // sticker 永远为单独消息
			}
		}
	}
	return res
}

func ToReadableString(m []IMessageElement) string {
	sb := new(strings.Builder)
	for _, elem := range m {
		switch e := elem.(type) {
		case *TextElement:
			sb.WriteString(e.Content)
		case *GroupImageElement, *FriendImageElement:
			sb.WriteString("[图片]")
		case *FaceElement:
			sb.WriteByte('/')
			sb.WriteString(e.Name)
		case *ForwardElement:
			sb.WriteString("[聊天记录]")
		// NOTE: flash pic is singular
		// To be clarified
		// case *GroupFlashImgElement:
		// 	return "[闪照]"
		// case *FriendFlashImgElement:
		// 	return "[闪照]"
		case *AtElement:
			sb.WriteString(e.Display)
		}
	}
	return sb.String()
}

func FaceNameById(id int) string {
	if name, ok := faceMap[id]; ok {
		return name
	}
	return "未知表情"
}

// SplitLongMessage 将过长的消息分割为若干个适合发送的消息
func SplitLongMessage(sendingMessage *SendingMessage) []*SendingMessage {
	// 合并连续文本消息
	sendingMessage = mergeContinuousTextMessages(sendingMessage)

	// 分割过长元素
	sendingMessage = splitElements(sendingMessage)

	// 将元素分为多组，确保各组不超过单条消息的上限
	splitMessages := splitMessages(sendingMessage)

	return splitMessages
}

// mergeContinuousTextMessages 预先将所有连续的文本消息合并为到一起，方便后续统一切割
func mergeContinuousTextMessages(sendingMessage *SendingMessage) *SendingMessage {
	// 检查下是否有连续的文本消息，若没有，则可以直接返回
	lastIsText := false
	hasContinuousText := false
	for _, message := range sendingMessage.Elements {
		if message.Type() == Text {
			if lastIsText {
				// 有连续的文本消息，需要进行处理
				hasContinuousText = true
				break
			}

			// 遇到文本元素先存放起来，方便将连续的文本元素合并
			lastIsText = true
			continue
		} else {
			lastIsText = false
		}
	}
	if !hasContinuousText {
		return sendingMessage
	}

	// 存在连续的文本消息，需要进行合并处理
	textBuffer := strings.Builder{}
	lastIsText = false
	totalMessageCount := 0
	for _, message := range sendingMessage.Elements {
		if msgVal, ok := message.(*TextElement); ok {
			// 遇到文本元素先存放起来，方便将连续的文本元素合并
			textBuffer.WriteString(msgVal.Content)
			lastIsText = true
			continue
		}

		// 如果之前的是文本元素（可能是多个合并起来的），则在这里将其实际放入消息中
		if lastIsText {
			sendingMessage.Elements[totalMessageCount] = NewText(textBuffer.String())
			totalMessageCount += 1
			textBuffer.Reset()
		}
		lastIsText = false

		// 非文本元素则直接处理
		sendingMessage.Elements[totalMessageCount] = message
		totalMessageCount += 1
	}
	// 处理最后几个元素是文本的情况
	if textBuffer.Len() != 0 {
		sendingMessage.Elements[totalMessageCount] = NewText(textBuffer.String())
		totalMessageCount += 1
		textBuffer.Reset()
	}
	sendingMessage.Elements = sendingMessage.Elements[:totalMessageCount]

	return sendingMessage
}

// splitElements 将原有消息的各个元素先尝试处理，如过长的文本消息按需分割为多个元素
func splitElements(sendingMessage *SendingMessage) *SendingMessage {
	// 检查下是否存在需要文本消息，若不存在，则直接返回
	needSplit := false
	for _, message := range sendingMessage.Elements {
		if msgVal, ok := message.(*TextElement); ok {
			if textNeedSplit(msgVal.Content) {
				needSplit = true
				break
			}
		}
	}
	if !needSplit {
		return sendingMessage
	}

	// 开始尝试切割
	messageParts := NewSendingMessage()

	for _, message := range sendingMessage.Elements {
		switch msgVal := message.(type) {
		case *TextElement:
			messageParts.Elements = append(messageParts.Elements, splitPlainMessage(msgVal.Content)...)
		default:
			messageParts.Append(message)
		}
	}

	return messageParts
}

// splitMessages 根据大小分为多个消息进行发送
func splitMessages(sendingMessage *SendingMessage) []*SendingMessage {
	var splitMessages []*SendingMessage

	messagePart := NewSendingMessage()
	msgSize := 0
	for _, part := range sendingMessage.Elements {
		estimateSize := EstimateLength([]IMessageElement{part})
		// 若当前分消息加上新的元素后大小会超限，且已经有元素（确保不会无限循环），则开始切分为新的一个元素
		if msgSize+estimateSize > MaxMessageSize && len(messagePart.Elements) > 0 {
			splitMessages = append(splitMessages, messagePart)

			messagePart = NewSendingMessage()
			msgSize = 0
		}

		// 加上新的元素
		messagePart.Append(part)
		msgSize += estimateSize
	}
	// 将最后一个分片加上
	if len(messagePart.Elements) != 0 {
		splitMessages = append(splitMessages, messagePart)
	}

	return splitMessages
}

func splitPlainMessage(content string) []IMessageElement {
	if !textNeedSplit(content) {
		return []IMessageElement{NewText(content)}
	}

	splittedMessage := make([]IMessageElement, 0, (len(content)+MaxMessageSize-1)/MaxMessageSize)

	last := 0
	for runeIndex, runeValue := range content {
		// 如果加上新的这个字符后，会超出大小，则从这个字符前分一次片
		if runeIndex+len(string(runeValue))-last > MaxMessageSize {
			splittedMessage = append(splittedMessage, NewText(content[last:runeIndex]))
			last = runeIndex
		}
	}
	if last != len(content) {
		splittedMessage = append(splittedMessage, NewText(content[last:]))
	}

	return splittedMessage
}

func textNeedSplit(content string) bool {
	return len(content) > MaxMessageSize
}
