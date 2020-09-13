package client

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/Mrs4s/MiraiGo/binary"
	devinfo "github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"sort"
)

type DeviceInfo struct {
	Display     []byte
	Product     []byte
	Device      []byte
	Board       []byte
	Brand       []byte
	Model       []byte
	Bootloader  []byte
	FingerPrint []byte
	BootId      []byte
	ProcVersion []byte
	BaseBand    []byte
	SimInfo     []byte
	OSType      []byte
	MacAddress  []byte
	IpAddress   []byte
	WifiBSSID   []byte
	WifiSSID    []byte
	IMSIMd5     []byte
	IMEI        string
	AndroidId   []byte
	APN         []byte
	Guid        []byte
	TgtgtKey    []byte
	Protocol    ClientProtocol
	Version     *Version
}

type Version struct {
	Incremental []byte
	Release     []byte
	CodeName    []byte
	Sdk         uint32
}

type DeviceInfoFile struct {
	Display     string `json:"display"`
	Product     string `json:"product"`
	Device      string `json:"device"`
	Board       string `json:"board"`
	Model       string `json:"model"`
	FingerPrint string `json:"finger_print"`
	BootId      string `json:"boot_id"`
	ProcVersion string `json:"proc_version"`
	Protocol    int    `json:"protocol"` // 0: Pad 1: Phone 2: Watch
	IMEI        string `json:"imei"`
}

type groupMessageBuilder struct {
	MessageSeq    int32
	MessageCount  int32
	MessageSlices []*msg.Message
}

// default
var SystemDeviceInfo = &DeviceInfo{
	Display:     []byte("MIRAI.123456.001"),
	Product:     []byte("mirai"),
	Device:      []byte("mirai"),
	Board:       []byte("mirai"),
	Brand:       []byte("mamoe"),
	Model:       []byte("mirai"),
	Bootloader:  []byte("unknown"),
	FingerPrint: []byte("mamoe/mirai/mirai:10/MIRAI.200122.001/1234567:user/release-keys"),
	BootId:      []byte("cb886ae2-00b6-4d68-a230-787f111d12c7"),
	ProcVersion: []byte("Linux version 3.0.31-cb886ae2 (android-build@xxx.xxx.xxx.xxx.com)"),
	BaseBand:    []byte{},
	SimInfo:     []byte("T-Mobile"),
	OSType:      []byte("android"),
	MacAddress:  []byte("00:50:56:C0:00:08"),
	IpAddress:   []byte{10, 0, 1, 3}, // 10.0.1.3
	WifiBSSID:   []byte("00:50:56:C0:00:08"),
	WifiSSID:    []byte("<unknown ssid>"),
	IMEI:        "468356291846738",
	AndroidId:   []byte("MIRAI.123456.001"),
	APN:         []byte("wifi"),
	Protocol:    AndroidPad,
	Version: &Version{
		Incremental: []byte("5891938"),
		Release:     []byte("10"),
		CodeName:    []byte("REL"),
		Sdk:         29,
	},
}

var EmptyBytes = []byte{}
var NumberRange = "0123456789"

func init() {
	r := make([]byte, 16)
	rand.Read(r)
	t := md5.Sum(r)
	SystemDeviceInfo.IMSIMd5 = t[:]
	SystemDeviceInfo.GenNewGuid()
	SystemDeviceInfo.GenNewTgtgtKey()
}

func GenRandomDevice() {
	r := make([]byte, 16)
	rand.Read(r)
	SystemDeviceInfo.Display = []byte("MIRAI." + utils.RandomStringRange(6, NumberRange) + ".001")
	SystemDeviceInfo.FingerPrint = []byte("mamoe/mirai/mirai:10/MIRAI.200122.001/" + utils.RandomStringRange(7, NumberRange) + ":user/release-keys")
	SystemDeviceInfo.BootId = []byte(binary.GenUUID(r))
	SystemDeviceInfo.ProcVersion = []byte("Linux version 3.0.31-" + utils.RandomString(8) + " (android-build@xxx.xxx.xxx.xxx.com)")
	rand.Read(r)
	t := md5.Sum(r)
	SystemDeviceInfo.IMSIMd5 = t[:]
	SystemDeviceInfo.IMEI = utils.RandomStringRange(15, NumberRange)
	SystemDeviceInfo.AndroidId = SystemDeviceInfo.Display
	SystemDeviceInfo.GenNewGuid()
	SystemDeviceInfo.GenNewTgtgtKey()
}

func (info *DeviceInfo) ToJson() []byte {
	f := &DeviceInfoFile{
		Display:     string(info.Display),
		Product:     string(info.Product),
		Device:      string(info.Device),
		Board:       string(info.Board),
		Model:       string(info.Model),
		FingerPrint: string(info.FingerPrint),
		BootId:      string(info.BootId),
		ProcVersion: string(info.ProcVersion),
		IMEI:        info.IMEI,
		Protocol: func() int {
			switch info.Protocol {
			case AndroidPad:
				return 0
			case AndroidPhone:
				return 1
			case AndroidWatch:
				return 2
			}
			return 0
		}(),
	}
	d, _ := json.Marshal(f)
	return d
}

func (info *DeviceInfo) ReadJson(d []byte) error {
	var f DeviceInfoFile
	if err := json.Unmarshal(d, &f); err != nil {
		return err
	}
	info.Display = []byte(f.Display)
	if f.Product != "" {
		info.Product = []byte(f.Product)
		info.Device = []byte(f.Device)
		info.Board = []byte(f.Board)
		info.Model = []byte(f.Model)
	}
	info.FingerPrint = []byte(f.FingerPrint)
	info.BootId = []byte(f.BootId)
	info.ProcVersion = []byte(f.ProcVersion)
	info.IMEI = f.IMEI
	info.AndroidId = SystemDeviceInfo.Display
	switch f.Protocol {
	case 1:
		info.Protocol = AndroidPhone
	case 2:
		info.Protocol = AndroidWatch
	default:
		info.Protocol = AndroidPad
	}
	SystemDeviceInfo.GenNewGuid()
	SystemDeviceInfo.GenNewTgtgtKey()
	return nil
}

func (info *DeviceInfo) GenNewGuid() {
	t := md5.Sum(append(info.AndroidId, info.MacAddress...))
	info.Guid = t[:]
}

func (info *DeviceInfo) GenNewTgtgtKey() {
	r := make([]byte, 16)
	rand.Read(r)
	t := md5.Sum(append(r, info.Guid...))
	info.TgtgtKey = t[:]
}

func (info *DeviceInfo) GenDeviceInfoData() []byte {
	m := &devinfo.DeviceInfo{
		Bootloader:   string(info.Bootloader),
		ProcVersion:  string(info.ProcVersion),
		Codename:     string(info.Version.CodeName),
		Incremental:  string(info.Version.Incremental),
		Fingerprint:  string(info.FingerPrint),
		BootId:       string(info.BootId),
		AndroidId:    string(info.AndroidId),
		BaseBand:     string(info.BaseBand),
		InnerVersion: string(info.Version.Incremental),
	}
	data, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	return data
}

func (c *QQClient) parsePrivateMessage(msg *msg.Message) *message.PrivateMessage {
	friend := c.FindFriend(msg.Head.FromUin)
	var sender *message.Sender
	if friend == nil {
		sender = &message.Sender{
			Uin:      msg.Head.FromUin,
			Nickname: msg.Head.FromNick,
			IsFriend: false,
		}
	} else {
		sender = &message.Sender{
			Uin:      friend.Uin,
			Nickname: friend.Nickname,
		}
	}
	ret := &message.PrivateMessage{
		Id:       msg.Head.MsgSeq,
		Target:   c.Uin,
		Time:     msg.Head.MsgTime,
		Sender:   sender,
		Elements: message.ParseMessageElems(msg.Body.RichText.Elems),
	}
	if msg.Body.RichText.Attr != nil {
		ret.InternalId = msg.Body.RichText.Attr.Random
	}
	return ret
}

func (c *QQClient) parseTempMessage(msg *msg.Message) *message.TempMessage {
	group := c.FindGroupByUin(msg.Head.C2CTmpMsgHead.GroupUin)
	mem := group.FindMember(msg.Head.FromUin)
	sender := &message.Sender{
		Uin:      msg.Head.FromUin,
		Nickname: "Unknown",
		IsFriend: false,
	}
	if mem != nil {
		sender.Nickname = mem.Nickname
		sender.CardName = mem.CardName
	}
	return &message.TempMessage{
		Id:        msg.Head.MsgSeq,
		GroupCode: group.Code,
		GroupName: group.Name,
		Sender:    sender,
		Elements:  message.ParseMessageElems(msg.Body.RichText.Elems),
	}
}

func (c *QQClient) parseGroupMessage(m *msg.Message) *message.GroupMessage {
	group := c.FindGroup(m.Head.GroupInfo.GroupCode)
	if group == nil {
		return nil
	}
	var anonInfo *msg.AnonymousGroupMessage
	for _, e := range m.Body.RichText.Elems {
		if e.AnonGroupMsg != nil {
			anonInfo = e.AnonGroupMsg
		}
	}
	var sender *message.Sender
	if anonInfo != nil {
		sender = &message.Sender{
			Uin:      80000000,
			Nickname: string(anonInfo.AnonNick),
			IsFriend: false,
		}
	} else {
		mem := group.FindMember(m.Head.FromUin)
		if mem == nil {
			return nil
		}
		sender = &message.Sender{
			Uin:      mem.Uin,
			Nickname: mem.Nickname,
			CardName: mem.CardName,
			IsFriend: c.FindFriend(mem.Uin) != nil,
		}
	}
	var g *message.GroupMessage
	g = &message.GroupMessage{
		Id:        m.Head.MsgSeq,
		GroupCode: group.Code,
		GroupName: string(m.Head.GroupInfo.GroupName),
		Sender:    sender,
		Time:      m.Head.MsgTime,
		Elements:  message.ParseMessageElems(m.Body.RichText.Elems),
	}
	// pre parse
	for _, elem := range m.Body.RichText.Elems {
		// 为什么小程序会同时通过RichText和long text发送
		if elem.LightApp != nil {
			break
		}
		// is rich long msg
		if elem.GeneralFlags != nil && elem.GeneralFlags.LongTextResid != "" {
			if f := c.GetForwardMessage(elem.GeneralFlags.LongTextResid); f != nil && len(f.Nodes) == 1 {
				g = &message.GroupMessage{
					Id:        m.Head.MsgSeq,
					GroupCode: group.Code,
					GroupName: string(m.Head.GroupInfo.GroupName),
					Sender:    sender,
					Time:      m.Head.MsgTime,
					Elements:  f.Nodes[0].Message,
				}
			}
		}
	}
	if m.Body.RichText.Ptt != nil {
		g.Elements = []message.IMessageElement{
			&message.VoiceElement{
				Name: m.Body.RichText.Ptt.FileName,
				Md5:  m.Body.RichText.Ptt.FileMd5,
				Size: m.Body.RichText.Ptt.FileSize,
				Url:  "http://grouptalk.c2c.qq.com" + string(m.Body.RichText.Ptt.DownPara),
			},
		}
	}
	if m.Body.RichText.Attr != nil {
		g.InternalId = m.Body.RichText.Attr.Random
	}
	return g
}

func (b *groupMessageBuilder) build() *msg.Message {
	sort.Slice(b.MessageSlices, func(i, j int) bool {
		return b.MessageSlices[i].Content.PkgIndex < b.MessageSlices[i].Content.PkgIndex
	})
	base := b.MessageSlices[0]
	for _, m := range b.MessageSlices[1:] {
		base.Body.RichText.Elems = append(base.Body.RichText.Elems, m.Body.RichText.Elems...)
	}
	return base
}

func packRequestDataV3(data []byte) (r []byte) {
	r = append([]byte{0x0A}, data...)
	r = append(r, 0x0B)
	return
}

func genForwardTemplate(resId, preview, title, brief, source, summary string, ts int64) *message.SendingMessage {
	template := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?><msg serviceID="35" templateID="1" action="viewMultiMsg" brief="%s" m_resid="%s" m_fileName="%d" tSum="3" sourceMsgId="0" url="" flag="3" adverSign="0" multiMsgFlag="0"><item layout="1"><title color="#000000" size="34">%s</title> %s<hr></hr><summary size="26" color="#808080">%s</summary></item><source name="%s"></source></msg>`,
		brief, resId, ts, title, preview, summary, source,
	)
	return &message.SendingMessage{Elements: []message.IMessageElement{
		&message.ServiceElement{
			Id:      35,
			Content: template,
			ResId:   resId,
			SubType: "Forward",
		},
	}}
}

func genLongTemplate(resId, brief string, ts int64) *message.SendingMessage {
	limited := func() string {
		if len(brief) > 30 {
			return brief[:30] + "…"
		}
		return brief
	}()
	template := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?><msg serviceID="35" templateID="1" action="viewMultiMsg" brief="%s" m_resid="%s" m_fileName="%d" sourceMsgId="0" url="" flag="3" adverSign="0" multiMsgFlag="1"> <item layout="1"> <title>%s</title> <hr hidden="false" style="0"/> <summary>点击查看完整消息</summary> </item> <source name="聊天记录" icon="" action="" appid="-1"/> </msg>`,
		limited, resId, ts, limited,
	)
	return &message.SendingMessage{Elements: []message.IMessageElement{
		&message.ServiceElement{
			Id:      35,
			Content: template,
			ResId:   resId,
			SubType: "Long",
		},
	}}
}

func (c *QQClient) Info(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "INFO",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (c *QQClient) Error(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "ERROR",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (c *QQClient) Debug(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "DEBUG",
		Message: fmt.Sprintf(msg, args...),
	})
}
