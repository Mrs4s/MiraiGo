package client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	devinfo "github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

type (
	DeviceInfo struct {
		Display      string
		Product      string
		Device       string
		Board        string
		Brand        string
		Model        string
		Bootloader   string
		FingerPrint  string
		BootId       string
		ProcVersion  string
		BaseBand     string
		SimInfo      string
		OSType       string
		MacAddress   string
		IpAddress    []byte
		WifiBSSID    string
		WifiSSID     string
		IMSIMd5      []byte
		IMEI         string
		AndroidId    string
		APN          string
		VendorName   string
		VendorOSName string
		Guid         []byte
		TgtgtKey     []byte
		Protocol     ClientProtocol
		Version      *Version
	}

	Version struct {
		Incremental string
		Release     string
		CodeName    string
		Sdk         uint32
	}

	DeviceInfoFile struct {
		Display      string       `json:"display"`
		Product      string       `json:"product"`
		Device       string       `json:"device"`
		Board        string       `json:"board"`
		Model        string       `json:"model"`
		FingerPrint  string       `json:"finger_print"`
		BootId       string       `json:"boot_id"`
		ProcVersion  string       `json:"proc_version"`
		Protocol     int          `json:"protocol"` // 0: Pad 1: Phone 2: Watch
		IMEI         string       `json:"imei"`
		Brand        string       `json:"brand"`
		Bootloader   string       `json:"bootloader"`
		BaseBand     string       `json:"base_band"`
		Version      *VersionFile `json:"version"`
		SimInfo      string       `json:"sim_info"`
		OsType       string       `json:"os_type"`
		MacAddress   string       `json:"mac_address"`
		IpAddress    []int32      `json:"ip_address"`
		WifiBSSID    string       `json:"wifi_bssid"`
		WifiSSID     string       `json:"wifi_ssid"`
		ImsiMd5      string       `json:"imsi_md5"`
		AndroidId    string       `json:"android_id"`
		Apn          string       `json:"apn"`
		VendorName   string       `json:"vendor_name"`
		VendorOSName string       `json:"vendor_os_name"`
	}

	VersionFile struct {
		Incremental string `json:"incremental"`
		Release     string `json:"release"`
		Codename    string `json:"codename"`
		Sdk         uint32 `json:"sdk"`
	}

	groupMessageBuilder struct {
		MessageSlices []*msg.Message
	}

	versionInfo struct {
		ApkSign         []byte
		ApkId           string
		SortVersionName string
		SdkVersion      string
		AppId           uint32
		SubAppId        uint32
		BuildTime       uint32
		SSOVersion      uint32
		MiscBitmap      uint32
		SubSigmap       uint32
		MainSigMap      uint32
		Protocol        ClientProtocol
	}

	incomingPacketInfo struct {
		CommandName string
		SequenceId  uint16
		Params      requestParams
	}

	requestParams map[string]interface{}
)

var (
	EmptyBytes  = []byte{}
	NumberRange = "0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewDeviceInfo() *DeviceInfo {
	info := &DeviceInfo{
		Display:      "MIRAI.123456.001",
		Product:      "mirai",
		Device:       "mirai",
		Board:        "mirai",
		Brand:        "mamoe",
		Model:        "mirai",
		Bootloader:   "unknown",
		FingerPrint:  "mamoe/mirai/mirai:10/MIRAI.200122.001/1234567:user/release-keys",
		BootId:       "cb886ae2-00b6-4d68-a230-787f111d12c7",
		ProcVersion:  "Linux version 3.0.31-cb886ae2 (android-build@xxx.xxx.xxx.xxx.com)",
		BaseBand:     "",
		SimInfo:      "T-Mobile",
		OSType:       "android",
		MacAddress:   "00:50:56:C0:00:08",
		IpAddress:    []byte{10, 0, 1, 3}, // 10.0.1.3
		WifiBSSID:    "00:50:56:C0:00:08",
		WifiSSID:     "<unknown ssid>",
		IMEI:         "468356291846738",
		AndroidId:    "MIRAI.123456.001",
		APN:          "wifi",
		VendorName:   "MIUI",
		VendorOSName: "mirai",
		Protocol:     IPad,
		Version: &Version{
			Incremental: "5891938",
			Release:     "10",
			CodeName:    "REL",
			Sdk:         29,
		},
	}
	r := make([]byte, 16)
	rand.Read(r)
	t := md5.Sum(r)
	info.IMSIMd5 = t[:]
	info.RandomDevice()
	return info
}

func (info *DeviceInfo) RandomDevice() {
	r := make([]byte, 16)
	rand.Read(r)
	info.Display = "MIRAI." + utils.RandomStringRange(6, NumberRange) + ".001"
	info.FingerPrint = "mamoe/mirai/mirai:10/MIRAI.200122.001/" + utils.RandomStringRange(7, NumberRange) + ":user/release-keys"
	info.BootId = string(binary.GenUUID(r))
	info.ProcVersion = "Linux version 3.0.31-" + utils.RandomString(8) + " (android-build@xxx.xxx.xxx.xxx.com)"
	rand.Read(r)
	t := md5.Sum(r)
	info.IMSIMd5 = t[:]
	info.IMEI = GenIMEI()
	r = make([]byte, 8)
	rand.Read(r)
	hex.Encode([]byte(info.AndroidId), r)
	info.GenNewGuid()
	info.GenNewTgtgtKey()
}

func genVersionInfo(p ClientProtocol) *versionInfo {
	switch p {
	case AndroidPhone: // Dumped by mirai from qq android v8.2.7
		return &versionInfo{
			ApkId:           "com.tencent.mobileqq",
			AppId:           537066738,
			SubAppId:        537066738,
			SortVersionName: "8.5.0",
			BuildTime:       1607689988,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.2454",
			SSOVersion:      15,
			MiscBitmap:      184024956,
			SubSigmap:       0x10400,
			MainSigMap:      34869472,
			Protocol:        p,
		}
	case AndroidWatch:
		return &versionInfo{
			ApkId:           "com.tencent.qqlite",
			AppId:           537064446,
			SubAppId:        537064446,
			SortVersionName: "2.0.5",
			BuildTime:       1559564731,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.236",
			SSOVersion:      5,
			MiscBitmap:      16252796,
			SubSigmap:       0x10400,
			MainSigMap:      34869472,
			Protocol:        p,
		}
	case IPad:
		return &versionInfo{
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537065739,
			SubAppId:        537065739,
			SortVersionName: "5.8.9",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      12,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      1970400,
			Protocol:        p,
		}
	case MacOS:
		return &versionInfo{
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537064315,
			SubAppId:        537064315,
			SortVersionName: "5.8.9",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      12,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      1970400,
			Protocol:        p,
		}
	case QiDian:
		return &versionInfo{
			ApkId:           "com.tencent.qidian",
			AppId:           537061386,
			SubAppId:        537036590,
			SortVersionName: "3.8.6",
			BuildTime:       1556628836,
			ApkSign:         []byte{160, 30, 236, 171, 133, 233, 227, 186, 43, 15, 106, 21, 140, 133, 92, 41},
			SdkVersion:      "6.0.0.2365",
			SSOVersion:      5,
			MiscBitmap:      49807228,
			SubSigmap:       66560,
			MainSigMap:      34869472,
			Protocol:        p,
		}
	}
	return nil
}

func (info *DeviceInfo) ToJson() []byte {
	f := &DeviceInfoFile{
		Display:     info.Display,
		Product:     info.Product,
		Device:      info.Device,
		Board:       info.Board,
		Model:       string(info.Model),
		FingerPrint: string(info.FingerPrint),
		BootId:      string(info.BootId),
		ProcVersion: string(info.ProcVersion),
		IMEI:        info.IMEI,
		Brand:       string(info.Brand),
		Bootloader:  string(info.Bootloader),
		BaseBand:    string(info.BaseBand),
		AndroidId:   string(info.AndroidId),
		Version: &VersionFile{
			Incremental: string(info.Version.Incremental),
			Release:     string(info.Version.Release),
			Codename:    string(info.Version.CodeName),
			Sdk:         info.Version.Sdk,
		},
		SimInfo:      string(info.SimInfo),
		OsType:       string(info.OSType),
		MacAddress:   string(info.MacAddress),
		IpAddress:    []int32{int32(info.IpAddress[0]), int32(info.IpAddress[1]), int32(info.IpAddress[2]), int32(info.IpAddress[3])},
		WifiBSSID:    string(info.WifiBSSID),
		WifiSSID:     string(info.WifiSSID),
		ImsiMd5:      hex.EncodeToString(info.IMSIMd5),
		Apn:          string(info.APN),
		VendorName:   string(info.VendorName),
		VendorOSName: string(info.VendorOSName),
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
			case QiDian:
				return 4
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
		return errors.Wrap(err, "failed to unmarshal json message")
	}

	SetIfNotEmpty := func(trg *string, str string) {
		if str != "" {
			*trg = str
		}
	}
	SetIfNotEmpty(&info.Display, f.Display)
	SetIfNotEmpty(&info.Product, f.Product)
	SetIfNotEmpty(&info.Device, f.Device)
	SetIfNotEmpty(&info.Board, f.Board)
	SetIfNotEmpty(&info.Brand, f.Brand)
	SetIfNotEmpty(&info.Model, f.Model)
	SetIfNotEmpty(&info.Bootloader, f.Bootloader)
	SetIfNotEmpty(&info.FingerPrint, f.FingerPrint)
	SetIfNotEmpty(&info.BootId, f.BootId)
	SetIfNotEmpty(&info.ProcVersion, f.ProcVersion)
	SetIfNotEmpty(&info.BaseBand, f.BaseBand)
	SetIfNotEmpty(&info.SimInfo, f.SimInfo)
	SetIfNotEmpty(&info.OSType, f.OsType)
	SetIfNotEmpty(&info.MacAddress, f.MacAddress)
	if len(f.IpAddress) == 4 {
		info.IpAddress = []byte{byte(f.IpAddress[0]), byte(f.IpAddress[1]), byte(f.IpAddress[2]), byte(f.IpAddress[3])}
	}
	SetIfNotEmpty(&info.WifiBSSID, f.WifiBSSID)
	SetIfNotEmpty(&info.WifiSSID, f.WifiSSID)
	if len(f.ImsiMd5) != 0 {
		imsiMd5, err := hex.DecodeString(f.ImsiMd5)
		if err != nil {
			info.IMSIMd5 = imsiMd5
		}
	}
	if f.IMEI != "" {
		info.IMEI = f.IMEI
	}
	SetIfNotEmpty(&info.APN, f.Apn)
	SetIfNotEmpty(&info.VendorName, f.VendorName)
	SetIfNotEmpty(&info.VendorOSName, f.VendorOSName)

	SetIfNotEmpty(&info.AndroidId, f.AndroidId)
	if f.AndroidId == "" {
		info.AndroidId = info.Display // ?
	}

	switch f.Protocol {
	case 1:
		info.Protocol = AndroidPhone
	case 2:
		info.Protocol = AndroidWatch
	case 3:
		info.Protocol = MacOS
	case 4:
		info.Protocol = QiDian
	default:
		info.Protocol = IPad
	}
	info.GenNewGuid()
	info.GenNewTgtgtKey()
	return nil
}

func (info *DeviceInfo) GenNewGuid() {
	t := md5.Sum([]byte(info.AndroidId + info.MacAddress))
	info.Guid = t[:]
}

func (info *DeviceInfo) GenNewTgtgtKey() {
	r := make([]byte, 16)
	rand.Read(r)
	h := md5.New()
	h.Write(r)
	h.Write(info.Guid)
	info.TgtgtKey = h.Sum(nil)
}

func (info *DeviceInfo) GenDeviceInfoData() []byte {
	m := &devinfo.DeviceInfo{
		Bootloader:   info.Bootloader,
		ProcVersion:  info.ProcVersion,
		Codename:     info.Version.CodeName,
		Incremental:  info.Version.Incremental,
		Fingerprint:  info.FingerPrint,
		BootId:       info.BootId,
		AndroidId:    info.AndroidId,
		BaseBand:     info.BaseBand,
		InnerVersion: info.Version.Incremental,
	}
	data, err := proto.Marshal(m)
	if err != nil {
		panic(errors.Wrap(err, "failed to unmarshal protobuf message"))
	}
	return data
}

func GenIMEI() string {
	sum := 0 // the control sum of digits
	var final strings.Builder

	for i := 0; i < 14; i++ { // generating all the base digits
		toAdd := rand.Intn(10)
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

func getSSOAddress(deviceInfo *DeviceInfo) ([]*net.TCPAddr, error) {
	protocol := genVersionInfo(deviceInfo.Protocol)
	key, _ := hex.DecodeString("F0441F5FF42DA58FDCF7949ABA62D411")
	payload := jce.NewJceWriter(). // see ServerConfig.d
		WriteInt64(0, 1).WriteInt64(0, 2).WriteByte(1, 3).
		WriteString("00000", 4).WriteInt32(100, 5).
		WriteInt32(int32(protocol.AppId), 6).WriteString(deviceInfo.IMEI, 7).
		WriteInt64(0, 8).WriteInt64(0, 9).WriteInt64(0, 10).
		WriteInt64(0, 11).WriteByte(0, 12).WriteInt64(0, 13).WriteByte(1, 14).Bytes()
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"HttpServerListReq": packUniRequestData(payload)},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
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
	data := &jce.RequestDataVersion3{}
	rspPkt.ReadFrom(jce.NewJceReader(tea.Decrypt(rsp)[4:]))
	data.ReadFrom(jce.NewJceReader(rspPkt.SBuffer))
	reader := jce.NewJceReader(data.Map["HttpServerListRes"][1:])
	servers := []jce.SsoServerInfo{}
	reader.ReadSlice(&servers, 2)
	adds := make([]*net.TCPAddr, 0, len(servers))
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

func XmlEscape(c string) string {
	buf := new(bytes.Buffer)
	_ = xml.EscapeText(buf, []byte(c))
	return buf.String()
}

func genForwardTemplate(resID, preview, title, brief, source, summary string, ts int64, items []*msg.PbMultiMsgItem) *message.ForwardElement {
	template := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?><msg serviceID="35" templateID="1" action="viewMultiMsg" brief="%s" m_resid="%s" m_fileName="%d" tSum="3" sourceMsgId="0" url="" flag="3" adverSign="0" multiMsgFlag="0"><item layout="1"><title color="#000000" size="34">%s</title> %s<hr></hr><summary size="26" color="#808080">%s</summary></item><source name="%s"></source></msg>`,
		brief, resID, ts, title, preview, summary, source,
	)
	for _, item := range items {
		if item.GetFileName() == "MultiMsg" {
			*item.FileName = strconv.FormatInt(ts, 10)
		}
	}
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
		XmlEscape(limited), resID, ts, XmlEscape(limited),
	)
	return &message.ServiceElement{
		Id:      35,
		Content: template,
		ResId:   resID,
		SubType: "Long",
	}
}

func (p requestParams) bool(k string) bool {
	if p == nil {
		return false
	}
	i, ok := p[k]
	if !ok {
		return false
	}
	return i.(bool)
}

func (p requestParams) int32(k string) int32 {
	if p == nil {
		return 0
	}
	i, ok := p[k]
	if !ok {
		return 0
	}
	return i.(int32)
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

func (c *QQClient) packOIDBPackageProto(cmd, serviceType int32, msg proto.Message) []byte {
	b, _ := proto.Marshal(msg)
	return c.packOIDBPackage(cmd, serviceType, b)
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
