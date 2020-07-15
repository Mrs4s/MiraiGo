package client

import (
	"crypto/md5"
	"github.com/Mrs4s/MiraiGo/binary"
	devinfo "github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"google.golang.org/protobuf/proto"
	"math/rand"
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
	Version     *Version
}

type Version struct {
	Incremental []byte
	Release     []byte
	CodeName    []byte
	Sdk         uint32
}

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
	Version: &Version{
		Incremental: []byte("5891938"),
		Release:     []byte("10"),
		CodeName:    []byte("REL"),
		Sdk:         29,
	},
}

var EmptyBytes = []byte{}

func init() {
	r := make([]byte, 16)
	rand.Read(r)
	t := md5.Sum(r)
	SystemDeviceInfo.IMSIMd5 = t[:]
	SystemDeviceInfo.GenNewGuid()
	SystemDeviceInfo.GenNewTgtgtKey()
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
	if friend == nil {
		return nil
	}
	return &message.PrivateMessage{
		Id: msg.Head.MsgSeq,
		Sender: &message.Sender{
			Uin:      friend.Uin,
			Nickname: friend.Nickname,
		},
		Elements: parseMessageElems(msg.Body.RichText.Elems),
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
	g := &message.GroupMessage{
		Id:        m.Head.MsgSeq,
		GroupCode: group.Code,
		GroupName: string(m.Head.GroupInfo.GroupName),
		Sender:    sender,
		Elements:  parseMessageElems(m.Body.RichText.Elems),
	}
	return g
}

func parseMessageElems(elems []*msg.Elem) []message.IMessageElement {
	var res []message.IMessageElement
	for _, elem := range elems {
		if elem.SrcMsg != nil {
			if len(elem.SrcMsg.OrigSeqs) != 0 {
				r := &message.ReplyElement{
					ReplySeq: elem.SrcMsg.OrigSeqs[0],
					Elements: parseMessageElems(elem.SrcMsg.Elems),
				}
				res = append(res, r)
			}
			continue
		}
		if elem.Text != nil {
			if len(elem.Text.Attr6Buf) == 0 {
				res = append(res, message.NewText(elem.Text.Str))
			} else {
				att6 := binary.NewReader(elem.Text.Attr6Buf)
				att6.ReadBytes(7)
				target := int64(uint32(att6.ReadInt32()))
				res = append(res, message.NewAt(target, elem.Text.Str))
			}
		}
		if elem.CustomFace != nil {
			res = append(res, message.NewNetImage(elem.CustomFace.FilePath, "http://gchat.qpic.cn"+elem.CustomFace.OrigUrl))
		}
		if elem.NotOnlineImage != nil {
			var img string
			if elem.NotOnlineImage.OrigUrl != "" {
				img = "http://c2cpicdw.qpic.cn" + elem.NotOnlineImage.OrigUrl
			} else {
				img = "http://c2cpicdw.qpic.cn/offpic_new/0/" + elem.NotOnlineImage.ResId + "/0?term=2"
			}
			res = append(res, message.NewNetImage(elem.NotOnlineImage.FilePath, img))
		}
		if elem.Face != nil {
			res = append(res, message.NewFace(elem.Face.Index))
		}
	}
	return res
}
