package client

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	devinfo "github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
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
	MessageSlices []*msg.Message
}

type versionInfo struct {
	ApkSign         []byte
	ApkId           string
	SortVersionName string
	SdkVersion      string
	AppId           uint32
	BuildTime       uint32
	SSOVersion      uint32
	MiscBitmap      uint32
	SubSigmap       uint32
	MainSigMap      uint32
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
	Protocol:    IPad,
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

func genVersionInfo(p ClientProtocol) *versionInfo {
	switch p {
	case AndroidPhone: // Dumped by mirai from qq android v8.2.7
		return &versionInfo{
			ApkId:           "com.tencent.mobileqq",
			AppId:           537066419,
			SortVersionName: "8.4.18",
			BuildTime:       1604580615,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.2454",
			SSOVersion:      13,
			MiscBitmap:      184024956,
			SubSigmap:       0x10400,
			MainSigMap:      34869472,
		}
	case AndroidWatch:
		return &versionInfo{
			ApkId:           "com.tencent.mobileqq",
			AppId:           537061176,
			SortVersionName: "8.2.7",
			BuildTime:       1571193922,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.2413",
			SSOVersion:      5,
			MiscBitmap:      184024956,
			SubSigmap:       0x10400,
			MainSigMap:      34869472,
		}
	case IPad:
		return &versionInfo{
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537065739,
			SortVersionName: "5.8.9",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      12,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      1970400,
		}
	case MacOS:
		return &versionInfo{
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537064315,
			SortVersionName: "5.8.9",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      12,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      1970400,
		}
	}
	return nil
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
			case IPad:
				return 0
			case AndroidPhone:
				return 1
			case AndroidWatch:
				return 2
			case MacOS:
				return 3
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
		return errors.Wrap(err, "failed to unmarshal protobuf message")
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
	case 3:
		info.Protocol = MacOS
	default:
		info.Protocol = IPad
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
		panic(errors.Wrap(err, "failed to unmarshal protobuf message"))
	}
	return data
}

func getSSOAddress() ([]*net.TCPAddr, error) {
	protocol := genVersionInfo(SystemDeviceInfo.Protocol)
	key, _ := hex.DecodeString("F0441F5FF42DA58FDCF7949ABA62D411")
	payload := jce.NewJceWriter(). // see ServerConfig.d
					WriteInt64(0, 1).WriteInt64(0, 2).WriteByte(1, 3).
					WriteString("00000", 4).WriteInt32(100, 5).
					WriteInt32(int32(protocol.AppId), 6).WriteString(SystemDeviceInfo.IMEI, 7).
					WriteInt64(0, 8).WriteInt64(0, 9).WriteInt64(0, 10).
					WriteInt64(0, 11).WriteByte(0, 12).WriteInt64(0, 13).WriteByte(1, 14).Bytes()
	buf := &jce.RequestDataVersion2{
		Map: map[string]map[string][]byte{"HttpServerListReq": {"ConfigHttp.HttpServerListReq": packUniRequestData(payload)}},
	}
	pkt := &jce.RequestPacket{
		IVersion:     2,
		SServantName: "ConfigHttp",
		SFuncName:    "HttpServerListReq",
		SBuffer:      buf.ToBytes(),
	}
	tea := binary.NewTeaCipher(key)
	rsp, err := utils.HttpPostBytes("https://configsvr.msf.3g.qq.com/configsvr/serverlist.jsp", tea.Encrypt(binary.NewWriterF(func(w *binary.Writer) {
		w.WriteIntLvPacket(0, func(w *binary.Writer) {
			w.Write(pkt.ToBytes())
		})
	})))
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch server list")
	}
	rspPkt := &jce.RequestPacket{}
	data := &jce.RequestDataVersion2{}
	rspPkt.ReadFrom(jce.NewJceReader(tea.Decrypt(rsp)[4:]))
	data.ReadFrom(jce.NewJceReader(rspPkt.SBuffer))
	reader := jce.NewJceReader(data.Map["HttpServerListRes"]["ConfigHttp.HttpServerListRes"][1:])
	servers := []jce.SsoServerInfo{}
	reader.ReadSlice(&servers, 2)
	var adds []*net.TCPAddr
	for _, s := range servers {
		if strings.Contains(s.Server, "com") {
			continue
		}
		adds = append(adds, &net.TCPAddr{
			IP:   net.ParseIP(s.Server),
			Port: int(s.Port),
		})
	}
	return adds, nil
}

func qualityTest(addr *net.TCPAddr) (int64, error) {
	// see QualityTestManager
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr.String(), time.Second*5)
	if err != nil {
		return 0, errors.Wrap(err, "failed to connect to server during quality test")
	}
	_ = conn.Close()
	end := time.Now()
	return end.Sub(start).Milliseconds(), nil
}

func (c *QQClient) parsePrivateMessage(msg *msg.Message) *message.PrivateMessage {
	friend := c.FindFriend(msg.Head.GetFromUin())
	var sender *message.Sender
	if friend == nil {
		sender = &message.Sender{
			Uin:      msg.Head.GetFromUin(),
			Nickname: msg.Head.GetFromNick(),
			IsFriend: false,
		}
	} else {
		sender = &message.Sender{
			Uin:      friend.Uin,
			Nickname: friend.Nickname,
		}
	}
	ret := &message.PrivateMessage{
		Id:     msg.Head.GetMsgSeq(),
		Target: c.Uin,
		Time:   msg.Head.GetMsgTime(),
		Sender: sender,
		Elements: func() []message.IMessageElement {
			if msg.Body.RichText.Ptt != nil {
				return []message.IMessageElement{
					&message.VoiceElement{
						Name: msg.Body.RichText.Ptt.GetFileName(),
						Md5:  msg.Body.RichText.Ptt.FileMd5,
						Size: msg.Body.RichText.Ptt.GetFileSize(),
						Url:  string(msg.Body.RichText.Ptt.DownPara),
					},
				}
			}
			return message.ParseMessageElems(msg.Body.RichText.Elems)
		}(),
	}
	if msg.Body.RichText.Attr != nil {
		ret.InternalId = msg.Body.RichText.Attr.GetRandom()
	}
	return ret
}

func (c *QQClient) parseTempMessage(msg *msg.Message) *message.TempMessage {
	group := c.FindGroupByUin(msg.Head.C2CTmpMsgHead.GetGroupUin())
	mem := group.FindMember(msg.Head.GetFromUin())
	sender := &message.Sender{
		Uin:      msg.Head.GetFromUin(),
		Nickname: "Unknown",
		IsFriend: false,
	}
	if mem != nil {
		sender.Nickname = mem.Nickname
		sender.CardName = mem.CardName
	}
	return &message.TempMessage{
		Id:        msg.Head.GetMsgSeq(),
		GroupCode: group.Code,
		GroupName: group.Name,
		Sender:    sender,
		Elements:  message.ParseMessageElems(msg.Body.RichText.Elems),
	}
}

func (c *QQClient) parseGroupMessage(m *msg.Message) *message.GroupMessage {
	group := c.FindGroup(m.Head.GroupInfo.GetGroupCode())
	if group == nil {
		c.Debug("sync group %v.", m.Head.GroupInfo.GetGroupCode())
		info, err := c.GetGroupInfo(m.Head.GroupInfo.GetGroupCode())
		if err != nil {
			c.Error("error to sync group %v : %+v", m.Head.GroupInfo.GetGroupCode(), err)
			return nil
		}
		group = info
		c.GroupList = append(c.GroupList, info)
	}
	if len(group.Members) == 0 {
		mem, err := c.GetGroupMembers(group)
		if err != nil {
			c.Error("error to sync group %v member : %+v", m.Head.GroupInfo.GroupCode, err)
			return nil
		}
		group.Members = mem
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
		mem := group.FindMember(m.Head.GetFromUin())
		if mem == nil {
			group.Update(func(_ *GroupInfo) {
				if mem = group.FindMemberWithoutLock(m.Head.GetFromUin()); mem != nil {
					return
				}
				info, _ := c.getMemberInfo(group.Code, m.Head.GetFromUin())
				if info == nil {
					return
				}
				mem = info
				group.Members = append(group.Members, mem)
				go c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
					Group:  group,
					Member: info,
				})
			})
			if mem == nil {
				return nil
			}
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
		Id:             m.Head.GetMsgSeq(),
		GroupCode:      group.Code,
		GroupName:      string(m.Head.GroupInfo.GroupName),
		Sender:         sender,
		Time:           m.Head.GetMsgTime(),
		Elements:       message.ParseMessageElems(m.Body.RichText.Elems),
		OriginalObject: m,
	}
	var extInfo *msg.ExtraInfo
	// pre parse
	for _, elem := range m.Body.RichText.Elems {
		// is rich long msg
		if elem.GeneralFlags != nil && elem.GeneralFlags.GetLongTextResid() != "" {
			if f := c.GetForwardMessage(elem.GeneralFlags.GetLongTextResid()); f != nil && len(f.Nodes) == 1 {
				g = &message.GroupMessage{
					Id:             m.Head.GetMsgSeq(),
					GroupCode:      group.Code,
					GroupName:      string(m.Head.GroupInfo.GroupName),
					Sender:         sender,
					Time:           m.Head.GetMsgTime(),
					Elements:       f.Nodes[0].Message,
					OriginalObject: m,
				}
			}
		}
		if elem.ExtraInfo != nil {
			extInfo = elem.ExtraInfo
		}
	}
	if !sender.IsAnonymous() {
		mem := group.FindMember(m.Head.GetFromUin())
		groupCard := m.Head.GroupInfo.GetGroupCard()
		if extInfo != nil && len(extInfo.GroupCard) > 0 && extInfo.GroupCard[0] == 0x0A {
			buf := oidb.D8FCCommCardNameBuf{}
			if err := proto.Unmarshal(extInfo.GroupCard, &buf); err == nil && len(buf.RichCardName) > 0 {
				groupCard = ""
				for _, e := range buf.RichCardName {
					groupCard += string(e.Text)
				}
			}
		}
		if m.Head.GroupInfo != nil && groupCard != "" && mem.CardName != groupCard {
			old := mem.CardName
			if mem.Nickname == groupCard {
				mem.CardName = ""
			} else {
				mem.CardName = groupCard
			}
			if old != mem.CardName {
				go c.dispatchMemberCardUpdatedEvent(&MemberCardUpdatedEvent{
					Group:   group,
					OldCard: old,
					Member:  mem,
				})
			}
		}
	}
	if m.Body.RichText.Ptt != nil {
		g.Elements = []message.IMessageElement{
			&message.VoiceElement{
				Name: m.Body.RichText.Ptt.GetFileName(),
				Md5:  m.Body.RichText.Ptt.FileMd5,
				Size: m.Body.RichText.Ptt.GetFileSize(),
				Url:  "http://grouptalk.c2c.qq.com" + string(m.Body.RichText.Ptt.DownPara),
			},
		}
	}
	if m.Body.RichText.Attr != nil {
		g.InternalId = m.Body.RichText.Attr.GetRandom()
	}
	return g
}

func (b *groupMessageBuilder) build() *msg.Message {
	sort.Slice(b.MessageSlices, func(i, j int) bool {
		return b.MessageSlices[i].Content.GetPkgIndex() < b.MessageSlices[j].Content.GetPkgIndex()
	})
	base := b.MessageSlices[0]
	for _, m := range b.MessageSlices[1:] {
		base.Body.RichText.Elems = append(base.Body.RichText.Elems, m.Body.RichText.Elems...)
	}
	return base
}

func packUniRequestData(data []byte) (r []byte) {
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

func (c *QQClient) Error(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "ERROR",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (c *QQClient) Warning(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "WARNING",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (c *QQClient) Info(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "INFO",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (c *QQClient) Debug(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "DEBUG",
		Message: fmt.Sprintf(msg, args...),
	})
}

func (c *QQClient) Trace(msg string, args ...interface{}) {
	c.dispatchLogEvent(&LogEvent{
		Type:    "TRACE",
		Message: fmt.Sprintf(msg, args...),
	})
}
