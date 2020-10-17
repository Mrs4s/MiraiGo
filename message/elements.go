package message

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"strconv"
	"strings"
)

type TextElement struct {
	Content string `json:"content"`
}

type ImageElement struct {
	Filename string `json:"filename"`
	Size     int32  `json:"size"`
	Width    int32  `json:"width"`
	Height   int32  `json:"height"`
	Url      string `json:"url"`
	Md5      []byte `json:"md5"`
	Data     []byte `json:"data"`
}

type GroupImageElement struct {
	ImageId string `json:"image_id"`
	FileId  int64  `json:"file_id"`
	Size    int32  `json:"size"`
	Width   int32  `json:"width"`
	Height  int32  `json:"height"`
	Md5     []byte `json:"md5"`
	Url     string `json:"url"`
}

type VoiceElement struct {
	Name string `json:"name"`
	Md5  []byte `json:"md5"`
	Size int32  `json:"size"`
	Url  string `json:"url"`

	// --- sending ---
	Data []byte `json:"data"`
}

type GroupVoiceElement struct {
	Data []byte
	Ptt  *msg.Ptt
}

type PrivateVoiceElement struct {
	Data []byte
	Ptt  *msg.Ptt
}

type FriendImageElement struct {
	ImageId string `json:"image_id"`
	Md5     []byte `json:"md5"`
	Url     string `json:"url"`
}

type FaceElement struct {
	Index int32  `json:"index"`
	Name  string `json:"name"`
}

type AtElement struct {
	Target  int64  `json:"target"`
	Display string `json:"display"`
}

type GroupFileElement struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Path  string `json:"path"`
	Busid int32  `json:"busid"`
}

type ReplyElement struct {
	ReplySeq int32 `json:"reply_seq"`
	Sender   int64 `json:"sender"`
	Time     int32 `json:"time"`
	Elements []IMessageElement

	//original []*msg.Elem
}

type ShortVideoElement struct {
	Name string `json:"name"`
	Uuid []byte `json:"uuid"`
	Size int32  `json:"size"`
	Md5  []byte `json:"md5"`
	Url  string `json:"url"`
}

type ServiceElement struct {
	Id      int32  `json:"id"`
	Content string `json:"content"`
	ResId   string `json:"res_id"`
	SubType string `json:"sub_type"`
}

type ForwardElement struct {
	ResId string `json:"res_id"`
}

type LightAppElement struct {
	Content string `json:"content"`
}

type RedBagElement struct {
	MsgType RedBagMessageType `json:"msg_type"`
	Title   string            `json:"title"`
}

type GroupFlashPicElement struct {
	GroupImageElement
}

type GroupShowPicElement struct {
	GroupImageElement
	EffectId int32 `json:"effect_id"`
}

type FriendFlashPicElement struct {
	FriendImageElement
}

type RedBagMessageType int

const (
	Simple RedBagMessageType = 2
	Lucky  RedBagMessageType = 3
	World  RedBagMessageType = 6
)

func NewText(s string) *TextElement {
	return &TextElement{Content: s}
}

func NewImage(data []byte) *ImageElement {
	return &ImageElement{
		Data: data,
	}
}

func NewGroupImage(id string, md5 []byte, fid int64, size, width, height int32) *GroupImageElement {
	return &GroupImageElement{
		ImageId: id,
		FileId:  fid,
		Md5:     md5,
		Size:    size,
		Width:   width,
		Height:  height,
		Url:     "http://gchat.qpic.cn/gchatpic_new/1/0-0-" + strings.ReplaceAll(binary.CalculateImageResourceId(md5)[1:37], "-", "") + "/0?term=2",
	}
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
		//original: m.OriginalElements,
		Elements: m.Elements,
	}
}

func NewUrlShare(url, title, content, image string) *ServiceElement {
	template := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8' standalone='yes'?><msg templateID="123" url="%s" serviceID="33" action="web" actionData="" brief="【链接】%s" flag="8"><item layout="2"><picture cover="%s"/><title>%s</title><summary>%s</summary></item></msg>`,
		url, url, image, title, content,
	)
	return &ServiceElement{
		Id:      33,
		Content: template,
		ResId:   url,
		SubType: "UrlShare",
	}
}
func NewRichXml(template string, ResId int64) *ServiceElement {
	if ResId == 0 {
		ResId = 60 //默认值60
	}
	return &ServiceElement{
		Id:      int32(ResId),
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

func (e *ImageElement) Type() ElementType {
	return Image
}

func (e *FaceElement) Type() ElementType {
	return Face
}

func (e *GroupImageElement) Type() ElementType {
	return Image
}

func (e *FriendImageElement) Type() ElementType {
	return Image
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

func (e *ForwardElement) Type() ElementType {
	return Forward
}

func (e *GroupFileElement) Type() ElementType {
	return File
}

func (e *GroupVoiceElement) Type() ElementType {
	return Voice
}

func (e *PrivateVoiceElement) Type() ElementType {
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

func (e *RedBagElement) Type() ElementType {
	return RedBag
}

var faceMap = map[int]string{
	14:  "微笑",
	1:   "撇嘴",
	2:   "色",
	3:   "发呆",
	4:   "得意",
	5:   "流泪",
	6:   "害羞",
	7:   "闭嘴",
	8:   "睡",
	9:   "大哭",
	10:  "尴尬",
	11:  "发怒",
	12:  "调皮",
	13:  "呲牙",
	0:   "惊讶",
	15:  "难过",
	16:  "酷",
	96:  "冷汗",
	18:  "抓狂",
	19:  "吐",
	20:  "偷笑",
	21:  "可爱",
	22:  "白眼",
	23:  "傲慢",
	24:  "饥饿",
	25:  "困",
	26:  "惊恐",
	27:  "流汗",
	28:  "憨笑",
	29:  "大兵",
	30:  "奋斗",
	31:  "咒骂",
	32:  "疑问",
	33:  "嘘",
	34:  "晕",
	35:  "折磨",
	36:  "衰",
	37:  "骷髅",
	38:  "敲打",
	39:  "再见",
	97:  "擦汗",
	98:  "抠鼻",
	99:  "鼓掌",
	100: "糗大了",
	101: "坏笑",
	102: "左哼哼",
	103: "右哼哼",
	104: "哈欠",
	105: "鄙视",
	106: "委屈",
	107: "快哭了",
	108: "阴险",
	109: "亲亲",
	110: "吓",
	111: "可怜",
	172: "眨眼睛",
	182: "笑哭",
	179: "doge",
	173: "泪奔",
	174: "无奈",
	212: "托腮",
	175: "卖萌",
	178: "斜眼笑",
	177: "喷血",
	180: "惊喜",
	181: "骚扰",
	176: "小纠结",
	183: "我最美",
	112: "菜刀",
	89:  "西瓜",
	113: "啤酒",
	114: "篮球",
	115: "乒乓",
	171: "茶",
	60:  "咖啡",
	61:  "饭",
	46:  "猪头",
	63:  "玫瑰",
	64:  "凋谢",
	116: "示爱",
	66:  "爱心",
	67:  "心碎",
	53:  "蛋糕",
	54:  "闪电",
	55:  "炸弹",
	56:  "刀",
	57:  "足球",
	117: "瓢虫",
	59:  "便便",
	75:  "月亮",
	74:  "太阳",
	69:  "礼物",
	49:  "拥抱",
	76:  "强",
	77:  "弱",
	78:  "握手",
	79:  "胜利",
	118: "抱拳",
	119: "勾引",
	120: "拳头",
	121: "差劲",
	122: "爱你",
	123: "NO",
	124: "OK",
	42:  "爱情",
	85:  "飞吻",
	43:  "跳跳",
	41:  "发抖",
	86:  "怄火",
	125: "转圈",
	126: "磕头",
	127: "回头",
	128: "跳绳",
	129: "挥手",
	130: "激动",
	131: "街舞",
	132: "献吻",
	133: "左太极",
	134: "右太极",
	136: "双喜",
	137: "鞭炮",
	138: "灯笼",
	140: "K歌",
	144: "喝彩",
	145: "祈祷",
	146: "爆筋",
	147: "棒棒糖",
	148: "喝奶",
	151: "飞机",
	158: "钞票",
	168: "药",
	169: "手枪",
	188: "蛋",
	192: "红包",
	184: "河蟹",
	185: "羊驼",
	190: "菊花",
	187: "幽灵",
	193: "大笑",
	194: "不开心",
	197: "冷漠",
	198: "呃",
	199: "好棒",
	200: "拜托",
	201: "点赞",
	202: "无聊",
	203: "托脸",
	204: "吃",
	205: "送花",
	206: "害怕",
	207: "花痴",
	208: "小样儿",
	210: "飙泪",
	211: "我不看",
}
