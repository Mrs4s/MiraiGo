package message

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
)

type TextElement struct {
	Content string
}

type ImageElement struct {
	Filename string
	Size     int32
	Width    int32
	Height   int32
	Url      string
	Md5      []byte
	Data     []byte
}

type GroupImageElement struct {
	ImageId   string
	FileId    int64
	ImageType int32
	Size      int32
	Width     int32
	Height    int32
	Md5       []byte
	Url       string
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

type PrivateVoiceElement struct {
	Data []byte
	Ptt  *msg.Ptt
}

type FriendImageElement struct {
	ImageId string
	Md5     []byte
	Url     string
}

type FaceElement struct {
	Index      int32
	NewSysFace bool
	Name       string
}

type AtElement struct {
	Target  int64
	Display string
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
	Time     int32
	Elements []IMessageElement

	//original []*msg.Elem
}

type ShortVideoElement struct {
	Name      string
	Uuid      []byte
	Size      int32
	ThumbSize int32
	Md5       []byte
	ThumbMd5  []byte
	Url       string
}

type ServiceElement struct {
	Id      int32
	Content string
	ResId   string
	SubType string
}

type ForwardElement struct {
	FileName string
	Content  string
	ResId    string
	Items    []*msg.PbMultiMsgItem
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

// TODO: 总之就是非常傻逼

type GroupFlashImgElement struct {
	ImageElement
}

type GroupFlashPicElement struct {
	GroupImageElement
}

type GroupShowPicElement struct {
	GroupImageElement
	EffectId int32
}

type FriendFlashImgElement struct {
	ImageElement
}

type FriendFlashPicElement struct {
	FriendImageElement
}

type RedBagMessageType int

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
)

func NewText(s string) *TextElement {
	return &TextElement{Content: s}
}

func NewImage(data []byte) *ImageElement {
	return &ImageElement{
		Data: data,
	}
}

func NewGroupImage(id string, md5 []byte, fid int64, size, width, height, imageType int32) *GroupImageElement {
	return &GroupImageElement{
		ImageId:   id,
		FileId:    fid,
		Md5:       md5,
		Size:      size,
		ImageType: imageType,
		Width:     width,
		Height:    height,
		Url:       "http://gchat.qpic.cn/gchatpic_new/1/0-0-" + strings.ReplaceAll(binary.CalculateImageResourceId(md5)[1:37], "-", "") + "/0?term=2",
	}
}

func NewFace(index int32) *FaceElement {
	name := faceMap[int(index)]
	if name == "" {
		name = newSysFaceMap[int(index)]
		if name != "" {
			return &FaceElement{
				Index:      index,
				NewSysFace: true,
				Name:       name,
			}
		}
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

func (e *GroupFlashImgElement) Type() ElementType {
	return Image
}

func (e *FriendFlashImgElement) Type() ElementType {
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

// Type implement message.IMessageElement
func (e *MusicShareElement) Type() ElementType {
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
	305: "右亲亲",
	109: "左亲亲",
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

	// newSysFaceMap

	192: "红包",
	137: "嗨皮牛耶",
	138: "灯笼",
	136: "双喜",

	// newSysFaceMap

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

	145: "祈祷",
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
	140: "K歌",
	144: "喝彩",
	146: "爆筋",
	147: "棒棒糖",
	148: "喝奶",
	151: "飞机",
	158: "钞票",
	168: "药",
	169: "手枪",
	188: "蛋",
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

	// newSysFaceMap
}

var newSysFaceMap = map[int]string{

	245: "加油必胜",
	246: "加油抱抱",
	247: "口罩护体",
	260: "搬砖中",
	261: "忙到飞起",
	262: "脑阔疼",
	263: "沧桑",
	264: "捂脸",
	265: "辣眼睛",
	266: "哦呦",
	267: "头秃",
	268: "问号脸",
	269: "暗中观察",
	270: "emm",
	271: "吃瓜",
	272: "呵呵哒",
	277: "汪汪",
	307: "牛转钱坤",
	306: "牛气冲天",
	281: "无眼笑",
	282: "敬礼",
	283: "狂笑",
	284: "面无表情",
	285: "摸鱼",
	293: "摸锦鲤",
	286: "魔鬼笑",
	287: "哦",
	288: "请",
	289: "睁眼",
	294: "期待",
	295: "拿到红包",
	296: "真好",
	297: "拜谢",
	298: "元宝",
	299: "牛啊",
	300: "胖三斤",
	301: "好闪",
	303: "右拜年",
	302: "左拜年",
	304: "红包包",

	273: "我酸了",
	274: "太南了",

	308: "求红包",
	309: "谢红包",
	310: "新年烟花",
	290: "敲开心",
	291: "震惊",
	292: "让我康康",
	278: "汗",
	279: "打脸",
	280: "击掌",
	242: "头撞击",
	243: "甩头",
	244: "扔狗",
	215: "糊脸",
	237: "偷看",
	226: "拍桌",
	214: "啵啵",
	217: "扯一扯",
	240: "喷脸",
	216: "拍头",
	218: "舔一舔",
	229: "干杯",
	238: "扇脸",
	219: "蹭一蹭",
	225: "撩一撩",
	231: "哼",
	233: "掐一掐",
	221: "顶呱呱",
	222: "抱抱",
	239: "原谅",
	232: "佛系",
	220: "拽炸天",
	235: "颤抖",
	241: "生日快乐",
	230: "嘲讽",
	224: "开枪",
	236: "啃头",
	228: "恭喜",
	234: "惊呆",
	223: "暴击",
	227: "拍手",

	276: "辣椒酱", // 疑似删除
}
