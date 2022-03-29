package message

import (
	"fmt"
	"strconv"

	"github.com/Mrs4s/MiraiGo/client/pb/msg"
)

//go:generate go run generate.go

type TextElement struct {
	Content string
}

type VoiceElement struct {
	Name string
	Md5  []byte
	Size int32
	Url  string

	// --- sending ---
	Data []byte
}

type GroupVoiceElement struct {
	Data []byte
	Ptt  *msg.Ptt
}

type PrivateVoiceElement = GroupVoiceElement

type FaceElement struct {
	Index int32
	Name  string
}

type AtElement struct {
	Target  int64
	Display string
	SubType AtType
}

type GroupFileElement struct {
	Name  string
	Size  int64
	Path  string
	Busid int32
}

type ReplyElement struct {
	ReplySeq int32
	Sender   int64
	GroupID  int64 // 私聊回复群聊时
	Time     int32
	Elements []IMessageElement

	// original []*msg.Elem
}

type ShortVideoElement struct {
	Name      string
	Uuid      []byte
	Size      int32
	ThumbSize int32
	Md5       []byte
	ThumbMd5  []byte
	Url       string
	Guild     bool
}

type ServiceElement struct {
	Id      int32
	Content string
	ResId   string
	SubType string
}

type LightAppElement struct {
	Content string
}

type RedBagElement struct {
	MsgType RedBagMessageType
	Title   string
}

// MusicShareElement 音乐分享卡片
//
// 请使用 SendGroupMusicShare 或者 SendFriendMusicShare 发送
type MusicShareElement struct {
	MusicType  int    // 音乐类型,请使用 QQMusic 等常量
	Title      string // 标题(歌曲名)
	Brief      string
	Summary    string // 简介(歌手名)
	Url        string // 点击跳转链接
	PictureUrl string // 显示图片链接
	MusicUrl   string // 音乐播放链接
}

type AnimatedSticker struct {
	ID   int32
	Name string
}

type (
	RedBagMessageType int
	AtType            int
)

// /com/tencent/mobileqq/data/MessageForQQWalletMsg.java
const (
	RedBagSimple             RedBagMessageType = 2
	RedBagLucky              RedBagMessageType = 3
	RedBagSimpleTheme        RedBagMessageType = 4
	RedBagLuckyTheme         RedBagMessageType = 5
	RedBagWord               RedBagMessageType = 6
	RedBagSimpleSpecify      RedBagMessageType = 7
	RedBagLuckySpecify       RedBagMessageType = 8
	RedBagSimpleSpecifyOver3 RedBagMessageType = 11
	RedBagLuckySpecifyOver3  RedBagMessageType = 12
	RedBagVoice              RedBagMessageType = 13
	RedBagLook               RedBagMessageType = 14 // ?
	RedBagVoiceC2C           RedBagMessageType = 15
	RedBagH5                 RedBagMessageType = 17
	RedBagKSong              RedBagMessageType = 18
	RedBagEmoji              RedBagMessageType = 19
	RedBagDraw               RedBagMessageType = 22
	RedBagH5Common           RedBagMessageType = 20
	RedBagWordChain          RedBagMessageType = 24
	RedBagKeyword            RedBagMessageType = 25 // ?
	RedBagDrawMultiModel     RedBagMessageType = 26 // ??

	AtTypeGroupMember  = 0 // At群成员
	AtTypeGuildMember  = 1 // At频道成员
	AtTypeGuildChannel = 2 // At频道
)

func NewText(s string) *TextElement {
	return &TextElement{Content: s}
}

func NewFace(index int32) *FaceElement {
	name := faceMap[int(index)]
	if name == "" {
		name = "未知表情"
	}
	return &FaceElement{
		Index: index,
		Name:  name,
	}
}

func NewAt(target int64, display ...string) *AtElement {
	dis := "@" + strconv.FormatInt(target, 10)
	if target == 0 {
		dis = "@全体成员"
	}
	if len(display) != 0 {
		dis = display[0]
	}
	return &AtElement{
		Target:  target,
		Display: dis,
	}
}

func AtAll() *AtElement {
	return NewAt(0)
}

func NewReply(m *GroupMessage) *ReplyElement {
	return &ReplyElement{
		ReplySeq: m.Id,
		Sender:   m.Sender.Uin,
		Time:     m.Time,
		// original: m.OriginalElements,
		Elements: m.Elements,
	}
}

func NewPrivateReply(m *PrivateMessage) *ReplyElement {
	return &ReplyElement{
		ReplySeq: m.Id,
		Sender:   m.Sender.Uin,
		Time:     m.Time,
		Elements: m.Elements,
	}
}

func NewUrlShare(url, title, content, image string) *ServiceElement {
	template := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?><msg templateID="12345" action="web" brief="[分享] %s" serviceID="1" url="%s"><item layout="2"><picture cover="%v"/><title>%v</title><summary>%v</summary></item><source/></msg>`,
		title, url, image, title, content)
	/*
		template := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8' standalone='yes'?><msg templateID="123" url="%s" serviceID="33" action="web" actionData="" brief="【链接】%s" flag="8"><item layout="2"><picture cover="%s"/><title>%s</title><summary>%s</summary></item></msg>`,
			url, url, image, title, content,
		)
	*/
	return &ServiceElement{
		Id:      1,
		Content: template,
		ResId:   url,
		SubType: "UrlShare",
	}
}

func NewRichXml(template string, resID int64) *ServiceElement {
	if resID == 0 {
		resID = 60 // 默认值60
	}
	return &ServiceElement{
		Id:      int32(resID),
		Content: template,
		SubType: "xml",
	}
}

func NewRichJson(template string) *ServiceElement {
	return &ServiceElement{
		Id:      1,
		Content: template,
		SubType: "json",
	}
}

func NewLightApp(content string) *LightAppElement {
	return &LightAppElement{Content: content}
}

func (e *TextElement) Type() ElementType {
	return Text
}

func (e *FaceElement) Type() ElementType {
	return Face
}

func (e *AtElement) Type() ElementType {
	return At
}

func (e *ServiceElement) Type() ElementType {
	return Service
}

func (e *ReplyElement) Type() ElementType {
	return Reply
}

func (e *GroupFileElement) Type() ElementType {
	return File
}

func (e *GroupVoiceElement) Type() ElementType {
	return Voice
}

func (e *VoiceElement) Type() ElementType {
	return Voice
}

func (e *ShortVideoElement) Type() ElementType {
	return Video
}

func (e *LightAppElement) Type() ElementType {
	return LightApp
}

// Type implement message.IMessageElement
func (e *MusicShareElement) Type() ElementType {
	return LightApp
}

func (e *RedBagElement) Type() ElementType {
	return RedBag
}

func (e *AnimatedSticker) Type() ElementType {
	return Face
}
