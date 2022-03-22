package client

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/internal/auth"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

type (
	DeviceInfo = auth.Device
	Version    = auth.OSVersion
)

var SystemDeviceInfo = &DeviceInfo{
	Display:      []byte("MIRAI.123456.001"),
	Product:      []byte("mirai"),
	Device:       []byte("mirai"),
	Board:        []byte("mirai"),
	Brand:        []byte("mamoe"),
	Model:        []byte("mirai"),
	Bootloader:   []byte("unknown"),
	FingerPrint:  []byte("mamoe/mirai/mirai:10/MIRAI.200122.001/1234567:user/release-keys"),
	BootId:       []byte("cb886ae2-00b6-4d68-a230-787f111d12c7"),
	ProcVersion:  []byte("Linux version 3.0.31-cb886ae2 (android-build@xxx.xxx.xxx.xxx.com)"),
	BaseBand:     EmptyBytes,
	SimInfo:      []byte("T-Mobile"),
	OSType:       []byte("android"),
	MacAddress:   []byte("00:50:56:C0:00:08"),
	IpAddress:    []byte{10, 0, 1, 3}, // 10.0.1.3
	WifiBSSID:    []byte("00:50:56:C0:00:08"),
	WifiSSID:     []byte("<unknown ssid>"),
	IMEI:         "468356291846738",
	AndroidId:    []byte("MIRAI.123456.001"),
	APN:          []byte("wifi"),
	VendorName:   []byte("MIUI"),
	VendorOSName: []byte("mirai"),
	Protocol:     IPad,
	Version: &Version{
		Incremental: []byte("5891938"),
		Release:     []byte("10"),
		CodeName:    []byte("REL"),
		SDK:         29,
	},
}

var EmptyBytes = make([]byte, 0)

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
	const numberRange = "0123456789"
	SystemDeviceInfo.Display = []byte("MIRAI." + utils.RandomStringRange(6, numberRange) + ".001")
	SystemDeviceInfo.FingerPrint = []byte("mamoe/mirai/mirai:10/MIRAI.200122.001/" + utils.RandomStringRange(7, numberRange) + ":user/release-keys")
	SystemDeviceInfo.BootId = binary.GenUUID(r)
	SystemDeviceInfo.ProcVersion = []byte("Linux version 3.0.31-" + utils.RandomString(8) + " (android-build@xxx.xxx.xxx.xxx.com)")
	rand.Read(r)
	t := md5.Sum(r)
	SystemDeviceInfo.IMSIMd5 = t[:]
	SystemDeviceInfo.IMEI = GenIMEI()
	r = make([]byte, 8)
	rand.Read(r)
	hex.Encode(SystemDeviceInfo.AndroidId, r)
	SystemDeviceInfo.GenNewGuid()
	SystemDeviceInfo.GenNewTgtgtKey()
}

func GenIMEI() string {
	sum := 0 // the control sum of digits
	var final strings.Builder

	randSrc := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSrc)

	for i := 0; i < 14; i++ { // generating all the base digits
		toAdd := randGen.Intn(10)
		if (i+1)%2 == 0 { // special proc for every 2nd one
			toAdd *= 2
			if toAdd >= 10 {
				toAdd = (toAdd % 10) + 1
			}
		}
		sum += toAdd
		final.WriteString(fmt.Sprintf("%d", toAdd)) // and even printing them here!
	}
	ctrlDigit := (sum * 9) % 10 // calculating the control digit
	final.WriteString(fmt.Sprintf("%d", ctrlDigit))
	return final.String()
}

func getSSOAddress() ([]netip.AddrPort, error) {
	protocol := SystemDeviceInfo.Protocol.Version()
	key, _ := hex.DecodeString("F0441F5FF42DA58FDCF7949ABA62D411")
	payload := jce.NewJceWriter(). // see ServerConfig.d
					WriteInt64(0, 1).WriteInt64(0, 2).WriteByte(1, 3).
					WriteString("00000", 4).WriteInt32(100, 5).
					WriteInt32(int32(protocol.AppId), 6).WriteString(SystemDeviceInfo.IMEI, 7).
					WriteInt64(0, 8).WriteInt64(0, 9).WriteInt64(0, 10).
					WriteInt64(0, 11).WriteByte(0, 12).WriteInt64(0, 13).Bytes()
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"HttpServerListReq": packUniRequestData(payload)},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "ConfigHttp",
		SFuncName:    "HttpServerListReq",
		SBuffer:      buf.ToBytes(),
	}
	b, cl := binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteIntLvPacket(0, func(w *binary.Writer) {
			w.Write(pkt.ToBytes())
		})
	})
	tea := binary.NewTeaCipher(key)
	encpkt := tea.Encrypt(b)
	cl()
	rsp, err := utils.HttpPostBytes("https://configsvr.msf.3g.qq.com/configsvr/serverlist.jsp?mType=getssolist", encpkt)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch server list")
	}
	rspPkt := &jce.RequestPacket{}
	data := &jce.RequestDataVersion3{}
	rspPkt.ReadFrom(jce.NewJceReader(tea.Decrypt(rsp)[4:]))
	data.ReadFrom(jce.NewJceReader(rspPkt.SBuffer))
	reader := jce.NewJceReader(data.Map["HttpServerListRes"][1:])
	servers := reader.ReadSsoServerInfos(2)
	adds := make([]netip.AddrPort, 0, len(servers))
	for _, s := range servers {
		if strings.Contains(s.Server, "com") {
			continue
		}
		ip, ok := netip.AddrFromSlice(net.ParseIP(s.Server))
		if ok {
			adds = append(adds, netip.AddrPortFrom(ip, uint16(s.Port)))
		}
	}
	return adds, nil
}

func qualityTest(addr string) (int64, error) {
	// see QualityTestManager
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		return 0, errors.Wrap(err, "failed to connect to server during quality test")
	}
	_ = conn.Close()
	return time.Since(start).Milliseconds(), nil
}

func (c *QQClient) parsePrivateMessage(msg *msg.Message) *message.PrivateMessage {
	friend := c.FindFriend(msg.Head.GetFromUin())
	var sender *message.Sender
	if friend == nil {
		sender = &message.Sender{
			Uin:      msg.Head.GetFromUin(),
			Nickname: msg.Head.GetFromNick(),
		}
	} else {
		sender = &message.Sender{
			Uin:      friend.Uin,
			Nickname: friend.Nickname,
			IsFriend: true,
		}
	}
	ret := &message.PrivateMessage{
		Id:     msg.Head.GetMsgSeq(),
		Target: msg.Head.GetToUin(),
		Time:   msg.Head.GetMsgTime(),
		Sender: sender,
		Self:   c.Uin,
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
	var groupCode int64
	var groupName string
	group := c.FindGroupByUin(msg.Head.C2CTmpMsgHead.GetGroupUin())
	sender := &message.Sender{
		Uin:      msg.Head.GetFromUin(),
		Nickname: "Unknown",
		IsFriend: false,
	}
	if group != nil {
		groupCode = group.Code
		groupName = group.Name
		mem := group.FindMember(msg.Head.GetFromUin())
		if mem != nil {
			sender.Nickname = mem.Nickname
			sender.CardName = mem.CardName
		}
	}
	return &message.TempMessage{
		Id:        msg.Head.GetMsgSeq(),
		GroupCode: groupCode,
		GroupName: groupName,
		Self:      c.Uin,
		Sender:    sender,
		Elements:  message.ParseMessageElems(msg.Body.RichText.Elems),
	}
}

func (c *QQClient) messageBuilder(seq int32) *messageBuilder {
	actual, ok := c.msgBuilders.Load(seq)
	if !ok {
		builder := &messageBuilder{}
		actual, _ = c.msgBuilders.LoadOrStore(seq, builder)
		time.AfterFunc(time.Minute, func() {
			c.msgBuilders.Delete(seq) // delete avoid memory leak
		})
	}
	return actual
}

type messageBuilder struct {
	lock   sync.Mutex
	slices []*msg.Message
}

func (b *messageBuilder) append(msg *msg.Message) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.slices = append(b.slices, msg)
}

func (b *messageBuilder) len() int32 {
	b.lock.Lock()
	x := len(b.slices)
	b.lock.Unlock()
	return int32(x)
}

func (b *messageBuilder) build() *msg.Message {
	b.lock.Lock()
	defer b.lock.Unlock()
	sort.Slice(b.slices, func(i, j int) bool {
		return b.slices[i].Content.GetPkgIndex() < b.slices[j].Content.GetPkgIndex()
	})
	base := b.slices[0]
	for _, m := range b.slices[1:] {
		base.Body.RichText.Elems = append(base.Body.RichText.Elems, m.Body.RichText.Elems...)
	}
	return base
}

func packUniRequestData(data []byte) []byte {
	r := make([]byte, 0, len(data)+2)
	r = append(r, 0x0a)
	r = append(r, data...)
	r = append(r, 0x0B)
	return r
}

func genForwardTemplate(resID, preview, summary string, ts int64, items []*msg.PbMultiMsgItem) *message.ForwardElement {
	template := forwardDisplay(resID, strconv.FormatInt(ts, 10), preview, summary)
	return &message.ForwardElement{
		FileName: strconv.FormatInt(ts, 10),
		Content:  template,
		ResId:    resID,
		Items:    items,
	}
}

func genLongTemplate(resID, brief string, ts int64) *message.ServiceElement {
	limited := func() string {
		if len(brief) > 30 {
			return brief[:30] + "…"
		}
		return brief
	}()
	template := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?><msg serviceID="35" templateID="1" action="viewMultiMsg" brief="%s" m_resid="%s" m_fileName="%d" sourceMsgId="0" url="" flag="3" adverSign="0" multiMsgFlag="1"> <item layout="1"> <title>%s</title> <hr hidden="false" style="0"/> <summary>点击查看完整消息</summary> </item> <source name="聊天记录" icon="" action="" appid="-1"/> </msg>`,
		utils.XmlEscape(limited), resID, ts, utils.XmlEscape(limited),
	)
	return &message.ServiceElement{
		Id:      35,
		Content: template,
		ResId:   resID,
		SubType: "Long",
	}
}

func (c *QQClient) getWebDeviceInfo() (i string) {
	qimei := strings.ToLower(utils.RandomString(36))
	i += fmt.Sprintf("i=%v&imsi=&mac=%v&m=%v&o=%v&", c.deviceInfo.IMEI, utils.B2S(c.deviceInfo.MacAddress), utils.B2S(c.deviceInfo.Device), utils.B2S(c.deviceInfo.Version.Release))
	i += fmt.Sprintf("a=%v&sd=0&c64=0&sc=1&p=1080*2210&aid=%v&", c.deviceInfo.Version.SDK, c.deviceInfo.IMEI)
	i += fmt.Sprintf("f=%v&mm=%v&cf=%v&cc=%v&", c.deviceInfo.Brand, 5629 /* Total Memory*/, 1725 /* CPU Frequency */, 8 /* CPU Core Count */)
	i += fmt.Sprintf("qimei=%v&qimei36=%v&", qimei, qimei)
	i += "sharpP=1&n=wifi&support_xsj_live=true&client_mod=default&timezone=Asia/Shanghai&material_sdk_version=2.9.0&vh265=null&refreshrate=60"
	return
}

func (c *QQClient) packOIDBPackage(cmd, serviceType int32, body []byte) []byte {
	pkg := &oidb.OIDBSSOPkg{
		Command:       cmd,
		ServiceType:   serviceType,
		Bodybuffer:    body,
		ClientVersion: "Android " + c.version.SortVersionName,
	}
	r, _ := proto.Marshal(pkg)
	return r
}

func (c *QQClient) packOIDBPackageDynamically(cmd, serviceType int32, msg proto.DynamicMessage) []byte {
	return c.packOIDBPackage(cmd, serviceType, msg.Encode())
}

func (c *QQClient) packOIDBPackageProto(cmd, serviceType int32, msg proto.Message) []byte {
	b, _ := proto.Marshal(msg)
	return c.packOIDBPackage(cmd, serviceType, b)
}

func unpackOIDBPackage(payload []byte, rsp proto.Message) error {
	pkg := new(oidb.OIDBSSOPkg)
	if err := proto.Unmarshal(payload, pkg); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if pkg.Result != 0 {
		return errors.Errorf("oidb result unsuccessful: %v msg: %v", pkg.Result, pkg.ErrorMsg)
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, rsp); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return nil
}
