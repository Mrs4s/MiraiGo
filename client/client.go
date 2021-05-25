package client

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math"
	"math/rand"
	"net"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/crypto"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
)

var json = jsoniter.ConfigFastest

//go:generate go run github.com/a8m/syncmap -o "handler_map_gen.go" -pkg client -name HandlerMap "map[uint16]*handlerInfo"

type QQClient struct {
	Uin         int64
	PasswordMd5 [16]byte
	AllowSlider bool

	Nickname      string
	Age           uint16
	Gender        uint16
	FriendList    []*FriendInfo
	GroupList     []*GroupInfo
	OnlineClients []*OtherClientInfo
	Online        bool
	QiDian        *QiDianAccountInfo
	// NetLooping    bool

	SequenceId              int32
	OutGoingPacketSessionId []byte
	RandomKey               []byte
	TCP                     *utils.TCPListener
	ConnectTime             time.Time

	handlers        HandlerMap
	waiters         sync.Map
	servers         []*net.TCPAddr
	currServerIndex int
	retryTimes      int
	version         *versionInfo

	dpwd             []byte
	syncCookie       []byte
	pubAccountCookie []byte
	msgCtrlBuf       []byte
	ksid             []byte
	t104             []byte
	t174             []byte
	g                []byte
	t402             []byte
	t150             []byte
	t149             []byte
	t528             []byte
	t530             []byte
	rollbackSig      []byte
	randSeed         []byte // t403
	timeDiff         int64
	sigInfo          *loginSigInfo
	bigDataSession   *bigDataSessionInfo
	srvSsoAddrs      []string
	otherSrvAddrs    []string
	fileStorageInfo  *jce.FileStoragePushFSSvcList
	pwdFlag          bool

	lastMessageSeq int32
	// lastMessageSeqTmp      sync.Map
	msgSvcCache            *utils.Cache
	lastC2CMsgTime         int64
	transCache             *utils.Cache
	lastLostMsg            string
	groupSysMsgCache       *GroupSystemMessages
	groupMsgBuilders       sync.Map
	onlinePushCache        *utils.Cache
	requestPacketRequestID int32
	groupSeq               int32
	friendSeq              int32
	heartbeatEnabled       bool
	groupDataTransSeq      int32
	highwayApplyUpSeq      int32
	eventHandlers          *eventHandlers
	stat                   *Statistics

	groupListLock sync.Mutex
}

type loginSigInfo struct {
	loginBitmap uint64
	tgt         []byte
	tgtKey      []byte

	srmToken           []byte // study room manager | 0x16a
	t133               []byte
	encryptedA1        []byte
	userStKey          []byte
	userStWebSig       []byte
	sKey               []byte
	sKeyExpiredTime    int64
	d2                 []byte
	d2Key              []byte
	wtSessionTicketKey []byte
	deviceToken        []byte

	psKeyMap    map[string][]byte
	pt4TokenMap map[string][]byte
}

type QiDianAccountInfo struct {
	MasterUin  int64
	ExtName    string
	CreateTime int64

	bigDataReqAddrs   []string
	bigDataReqSession *bigDataSessionInfo
}

type handlerInfo struct {
	fun    func(i interface{}, err error)
	params requestParams
}

var decoders = map[string]func(*QQClient, *incomingPacketInfo, []byte) (interface{}, error){
	"wtlogin.login":                                decodeLoginResponse,
	"wtlogin.exchange_emp":                         decodeExchangeEmpResponse,
	"wtlogin.trans_emp":                            decodeTransEmpResponse,
	"StatSvc.register":                             decodeClientRegisterResponse,
	"StatSvc.ReqMSFOffline":                        decodeMSFOfflinePacket,
	"MessageSvc.PushNotify":                        decodeSvcNotify,
	"OnlinePush.ReqPush":                           decodeOnlinePushReqPacket,
	"OnlinePush.PbPushTransMsg":                    decodeOnlinePushTransPacket,
	"ConfigPushSvc.PushReq":                        decodePushReqPacket,
	"MessageSvc.PbGetMsg":                          decodeMessageSvcPacket,
	"MessageSvc.PushForceOffline":                  decodeForceOfflinePacket,
	"PbMessageSvc.PbMsgWithDraw":                   decodeMsgWithDrawResponse,
	"friendlist.getFriendGroupList":                decodeFriendGroupListResponse,
	"friendlist.delFriend":                         decodeFriendDeleteResponse,
	"friendlist.GetTroopListReqV2":                 decodeGroupListResponse,
	"friendlist.GetTroopMemberListReq":             decodeGroupMemberListResponse,
	"group_member_card.get_group_member_card_info": decodeGroupMemberInfoResponse,
	"PttStore.GroupPttUp":                          decodeGroupPttStoreResponse,
	"LongConn.OffPicUp":                            decodeOffPicUpResponse,
	"ProfileService.Pb.ReqSystemMsgNew.Group":      decodeSystemMsgGroupPacket,
	"ProfileService.Pb.ReqSystemMsgNew.Friend":     decodeSystemMsgFriendPacket,
	"OidbSvc.0xd79":                                decodeWordSegmentation,
	"OidbSvc.0x990":                                decodeTranslateResponse,
	"SummaryCard.ReqSummaryCard":                   decodeSummaryCardResponse,
	"LightAppSvc.mini_app_info.GetAppInfoById":     decodeAppInfoResponse,
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// NewClient create new qq client
func NewClient(uin int64, password string) *QQClient {
	return NewClientMd5(uin, md5.Sum([]byte(password)))
}

func NewClientEmpty() *QQClient {
	return NewClient(0, "")
}

func NewClientMd5(uin int64, passwordMd5 [16]byte) *QQClient {
	crypto.ECDH.FetchPubKey(uin)
	cli := &QQClient{
		Uin:                     uin,
		PasswordMd5:             passwordMd5,
		SequenceId:              0x3635,
		AllowSlider:             true,
		RandomKey:               make([]byte, 16),
		OutGoingPacketSessionId: []byte{0x02, 0xB0, 0x5B, 0x8B},
		TCP:                     &utils.TCPListener{},
		sigInfo:                 &loginSigInfo{},
		requestPacketRequestID:  1921334513,
		groupSeq:                int32(rand.Intn(20000)),
		friendSeq:               22911,
		highwayApplyUpSeq:       77918,
		ksid:                    []byte(fmt.Sprintf("|%s|A8.2.7.27f6ea96", SystemDeviceInfo.IMEI)),
		eventHandlers:           &eventHandlers{},
		msgSvcCache:             utils.NewCache(time.Second * 15),
		transCache:              utils.NewCache(time.Second * 15),
		onlinePushCache:         utils.NewCache(time.Second * 15),
		version:                 genVersionInfo(SystemDeviceInfo.Protocol),
		servers:                 []*net.TCPAddr{},
		stat:                    &Statistics{},
	}
	sso, err := getSSOAddress()
	if err == nil && len(sso) > 0 {
		cli.servers = append(sso, cli.servers...)
	}
	adds, err := net.LookupIP("msfwifi.3g.qq.com") // host servers
	if err == nil && len(adds) > 0 {
		var hostAddrs []*net.TCPAddr
		for _, addr := range adds {
			hostAddrs = append(hostAddrs, &net.TCPAddr{
				IP:   addr,
				Port: 8080,
			})
		}
		cli.servers = append(hostAddrs, cli.servers...)
	}
	if len(cli.servers) == 0 {
		cli.servers = []*net.TCPAddr{ // default servers
			{IP: net.IP{42, 81, 172, 81}, Port: 80},
			{IP: net.IP{114, 221, 148, 59}, Port: 14000},
			{IP: net.IP{42, 81, 172, 147}, Port: 443},
			{IP: net.IP{125, 94, 60, 146}, Port: 80},
			{IP: net.IP{114, 221, 144, 215}, Port: 80},
			{IP: net.IP{42, 81, 172, 22}, Port: 80},
		}
	}
	pings := make([]int64, len(cli.servers))
	wg := sync.WaitGroup{}
	wg.Add(len(cli.servers))
	for i := range cli.servers {
		go func(index int) {
			defer wg.Done()
			p, err := qualityTest(cli.servers[index])
			if err != nil {
				pings[index] = 9999
				return
			}
			pings[index] = p
		}(i)
	}
	wg.Wait()
	sort.Slice(cli.servers, func(i, j int) bool {
		return pings[i] < pings[j]
	})
	if len(cli.servers) > 3 {
		cli.servers = cli.servers[0 : len(cli.servers)/2] // 保留ping值中位数以上的server
	}
	cli.TCP.PlannedDisconnect(cli.plannedDisconnect)
	cli.TCP.UnexpectedDisconnect(cli.unexpectedDisconnect)
	rand.Read(cli.RandomKey)
	go cli.netLoop()
	return cli
}

// Login send login request
func (c *QQClient) Login() (*LoginResponse, error) {
	if c.Online {
		return nil, ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return nil, err
	}
	rsp, err := c.sendAndWait(c.buildLoginPacket())
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		_ = c.init(false)
	}
	return &l, nil
}

func (c *QQClient) TokenLogin(token []byte) error {
	if c.Online {
		return ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return err
	}
	{
		r := binary.NewReader(token)
		c.Uin = r.ReadInt64()
		c.sigInfo.d2 = r.ReadBytesShort()
		c.sigInfo.d2Key = r.ReadBytesShort()
		c.sigInfo.tgt = r.ReadBytesShort()
		c.sigInfo.srmToken = r.ReadBytesShort()
		c.sigInfo.t133 = r.ReadBytesShort()
		c.sigInfo.encryptedA1 = r.ReadBytesShort()
		c.sigInfo.wtSessionTicketKey = r.ReadBytesShort()
		c.OutGoingPacketSessionId = r.ReadBytesShort()
		SystemDeviceInfo.TgtgtKey = r.ReadBytesShort()
	}
	_, err = c.sendAndWait(c.buildRequestChangeSigPacket())
	if err != nil {
		return err
	}
	return c.init(true)
}

func (c *QQClient) FetchQRCode() (*QRCodeLoginResponse, error) {
	if c.Online {
		return nil, ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return nil, err
	}
	i, err := c.sendAndWait(c.buildQRCodeFetchRequestPacket())
	if err != nil {
		return nil, errors.Wrap(err, "fetch qrcode error")
	}
	return i.(*QRCodeLoginResponse), nil
}

func (c *QQClient) QueryQRCodeStatus(sig []byte) (*QRCodeLoginResponse, error) {
	i, err := c.sendAndWait(c.buildQRCodeResultQueryRequestPacket(sig))
	if err != nil {
		return nil, errors.Wrap(err, "query result error")
	}
	return i.(*QRCodeLoginResponse), nil
}

func (c *QQClient) QRCodeLogin(info *QRCodeLoginInfo) (*LoginResponse, error) {
	i, err := c.sendAndWait(c.buildQRCodeLoginPacket(info.tmpPwd, info.tmpNoPicSig, info.tgtQR))
	if err != nil {
		return nil, errors.Wrap(err, "qrcode login error")
	}
	rsp := i.(LoginResponse)
	if rsp.Success {
		_ = c.init(false)
	}
	return &rsp, nil
}

// SubmitCaptcha send captcha to server
func (c *QQClient) SubmitCaptcha(result string, sign []byte) (*LoginResponse, error) {
	seq, packet := c.buildCaptchaPacket(result, sign)
	rsp, err := c.sendAndWait(seq, packet)
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		_ = c.init(false)
	}
	return &l, nil
}

func (c *QQClient) SubmitTicket(ticket string) (*LoginResponse, error) {
	seq, packet := c.buildTicketSubmitPacket(ticket)
	rsp, err := c.sendAndWait(seq, packet)
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		_ = c.init(false)
	}
	return &l, nil
}

func (c *QQClient) SubmitSMS(code string) (*LoginResponse, error) {
	rsp, err := c.sendAndWait(c.buildSMSCodeSubmitPacket(code))
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		_ = c.init(false)
	}
	return &l, nil
}

func (c *QQClient) RequestSMS() bool {
	rsp, err := c.sendAndWait(c.buildSMSRequestPacket())
	if err != nil {
		c.Error("request sms error: %v", err)
		return false
	}
	return rsp.(LoginResponse).Error == SMSNeededError
}

func (c *QQClient) init(tokenLogin bool) error {
	if len(c.g) == 0 {
		c.Warning("device lock is disable. http api may fail.")
	}
	if err := c.registerClient(); err != nil {
		return errors.Wrap(err, "register error")
	}
	if tokenLogin {
		notify := make(chan struct{})
		d := c.waitPacket("StatSvc.ReqMSFOffline", func(i interface{}, err error) {
			notify <- struct{}{}
		})
		d2 := c.waitPacket("MessageSvc.PushForceOffline", func(i interface{}, err error) {
			notify <- struct{}{}
		})
		select {
		case <-notify:
			d()
			d2()
			return errors.New("token failed")
		case <-time.After(time.Second):
			d()
			d2()
		}
	}
	c.groupSysMsgCache, _ = c.GetGroupSystemMessages()
	if !c.heartbeatEnabled {
		go c.doHeartbeat()
	}
	_ = c.RefreshStatus()
	if c.version.Protocol == QiDian {
		_, _ = c.sendAndWait(c.buildLoginExtraPacket())     // 小登录
		_, _ = c.sendAndWait(c.buildConnKeyRequestPacket()) // big data key 如果等待 config push 的话时间来不及
	}
	seq, pkt := c.buildGetMessageRequestPacket(msg.SyncFlag_START, time.Now().Unix())
	_, _ = c.sendAndWait(seq, pkt, requestParams{"used_reg_proxy": true, "init": true})
	c.stat.once.Do(func() {
		c.OnGroupMessage(func(_ *QQClient, _ *message.GroupMessage) {
			c.stat.MessageReceived++
			c.stat.LastMessageTime = time.Now().Unix()
		})
		c.OnPrivateMessage(func(_ *QQClient, _ *message.PrivateMessage) {
			c.stat.MessageReceived++
			c.stat.LastMessageTime = time.Now().Unix()
		})
		c.OnTempMessage(func(_ *QQClient, _ *TempMessageEvent) {
			c.stat.MessageReceived++
			c.stat.LastMessageTime = time.Now().Unix()
		})
		c.onGroupMessageReceipt("internal", func(_ *QQClient, _ *groupMessageReceiptEvent) {
			c.stat.MessageSent++
		})
	})
	return nil
}

func (c *QQClient) GenToken() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt64(uint64(c.Uin))
		w.WriteBytesShort(c.sigInfo.d2)
		w.WriteBytesShort(c.sigInfo.d2Key)
		w.WriteBytesShort(c.sigInfo.tgt)
		w.WriteBytesShort(c.sigInfo.srmToken)
		w.WriteBytesShort(c.sigInfo.t133)
		w.WriteBytesShort(c.sigInfo.encryptedA1)
		w.WriteBytesShort(c.sigInfo.wtSessionTicketKey)
		w.WriteBytesShort(c.OutGoingPacketSessionId)
		w.WriteBytesShort(SystemDeviceInfo.TgtgtKey)
	})
}

func (c *QQClient) SetOnlineStatus(s UserOnlineStatus) {
	if s < 1000 {
		_, _ = c.sendAndWait(c.buildStatusSetPacket(int32(s), 0))
		return
	}
	_, _ = c.sendAndWait(c.buildStatusSetPacket(11, int32(s)))
}

func (c *QQClient) GetVipInfo(target int64) (*VipInfo, error) {
	b, err := utils.HttpGetBytes(fmt.Sprintf("https://h5.vip.qq.com/p/mc/cardv2/other?platform=1&qq=%d&adtag=geren&aid=mvip.pingtai.mobileqq.androidziliaoka.fromqita", target), c.getCookiesWithDomain("h5.vip.qq.com"))
	if err != nil {
		return nil, err
	}
	ret := VipInfo{Uin: target}
	b = b[bytes.Index(b, []byte(`<span class="ui-nowrap">`))+24:]
	t := b[:bytes.Index(b, []byte(`</span>`))]
	ret.Name = string(t)
	b = b[bytes.Index(b, []byte(`<small>LV</small>`))+17:]
	t = b[:bytes.Index(b, []byte(`</p>`))]
	ret.Level, _ = strconv.Atoi(string(t))
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	t = b[:bytes.Index(b, []byte(`<small>倍`))]
	ret.LevelSpeed, _ = strconv.ParseFloat(string(t), 64)
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	st := string(b[:bytes.Index(b, []byte(`</p>`))])
	st = strings.Replace(st, "<small>", "", 1)
	st = strings.Replace(st, "</small>", "", 1)
	ret.VipLevel = st
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	t = b[:bytes.Index(b, []byte(`</p>`))]
	ret.VipGrowthSpeed, _ = strconv.Atoi(string(t))
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	t = b[:bytes.Index(b, []byte(`</p>`))]
	ret.VipGrowthTotal, _ = strconv.Atoi(string(t))
	return &ret, nil
}

func (c *QQClient) GetGroupHonorInfo(groupCode int64, honorType HonorType) (*GroupHonorInfo, error) {
	b, err := utils.HttpGetBytes(fmt.Sprintf("https://qun.qq.com/interactive/honorlist?gc=%d&type=%d", groupCode, honorType), c.getCookiesWithDomain("qun.qq.com"))
	if err != nil {
		return nil, err
	}
	b = b[bytes.Index(b, []byte(`window.__INITIAL_STATE__=`))+25:]
	b = b[:bytes.Index(b, []byte("</script>"))]
	ret := GroupHonorInfo{}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *QQClient) GetWordSegmentation(text string) ([]string, error) {
	rsp, err := c.sendAndWait(c.buildWordSegmentationPacket([]byte(text)))
	if err != nil {
		return nil, err
	}
	if data, ok := rsp.([][]byte); ok {
		var ret []string
		for _, val := range data {
			ret = append(ret, string(val))
		}
		return ret, nil
	}
	return nil, errors.New("decode error")
}

func (c *QQClient) GetSummaryInfo(target int64) (*SummaryCardInfo, error) {
	rsp, err := c.sendAndWait(c.buildSummaryCardRequestPacket(target))
	if err != nil {
		return nil, err
	}
	return rsp.(*SummaryCardInfo), nil
}

// ReloadFriendList refresh QQClient.FriendList field via GetFriendList()
func (c *QQClient) ReloadFriendList() error {
	rsp, err := c.GetFriendList()
	if err != nil {
		return err
	}
	c.FriendList = rsp.List
	return nil
}

// GetFriendList
// 当使用普通QQ时: 请求好友列表
// 当使用企点QQ时: 请求外部联系人列表
func (c *QQClient) GetFriendList() (*FriendListResponse, error) {
	if c.version.Protocol == QiDian {
		rsp, err := c.getQiDianAddressDetailList()
		if err != nil {
			return nil, err
		}
		return &FriendListResponse{TotalCount: int32(len(rsp)), List: rsp}, nil
	}
	curFriendCount := 0
	r := &FriendListResponse{}
	for {
		rsp, err := c.sendAndWait(c.buildFriendGroupListRequestPacket(int16(curFriendCount), 150, 0, 0))
		if err != nil {
			return nil, err
		}
		list := rsp.(FriendListResponse)
		r.TotalCount = list.TotalCount
		r.List = append(r.List, list.List...)
		curFriendCount += len(list.List)
		if int32(len(r.List)) >= r.TotalCount {
			break
		}
	}
	return r, nil
}

func (c *QQClient) GetForwardMessage(resID string) *message.ForwardMessage {
	m := c.DownloadForwardMessage(resID)
	if m == nil {
		return nil
	}
	var (
		item *msg.PbMultiMsgItem
		ret  = &message.ForwardMessage{Nodes: []*message.ForwardNode{}}
	)
	for _, iter := range m.Items {
		if iter.GetFileName() == m.FileName {
			item = iter
		}
	}
	if item == nil {
		return nil
	}
	for _, m := range item.GetBuffer().GetMsg() {
		ret.Nodes = append(ret.Nodes, &message.ForwardNode{
			SenderId: m.Head.GetFromUin(),
			SenderName: func() string {
				if m.Head.GetMsgType() == 82 && m.Head.GroupInfo != nil {
					return m.Head.GroupInfo.GetGroupCard()
				}
				return m.Head.GetFromNick()
			}(),
			Time:    m.Head.GetMsgTime(),
			Message: message.ParseMessageElems(m.Body.RichText.Elems),
		})
	}
	return ret
}

func (c *QQClient) DownloadForwardMessage(resId string) *message.ForwardElement {
	i, err := c.sendAndWait(c.buildMultiApplyDownPacket(resId))
	if err != nil {
		return nil
	}
	multiMsg := i.(*msg.PbMultiMsgTransmit)
	if multiMsg.GetPbItemList() == nil {
		return nil
	}
	var pv string
	for i := 0; i < int(math.Min(4, float64(len(multiMsg.GetMsg())))); i++ {
		m := multiMsg.Msg[i]
		pv += fmt.Sprintf(`<title size="26" color="#777777">%s: %s</title>`,
			func() string {
				if m.Head.GetMsgType() == 82 && m.Head.GroupInfo != nil {
					return m.Head.GroupInfo.GetGroupCard()
				}
				return m.Head.GetFromNick()
			}(),
			message.ToReadableString(
				message.ParseMessageElems(multiMsg.Msg[i].GetBody().GetRichText().Elems),
			),
		)
	}
	return genForwardTemplate(
		resId, pv, "群聊的聊天记录", "[聊天记录]", "聊天记录",
		fmt.Sprintf("查看 %d 条转发消息", len(multiMsg.GetMsg())),
		time.Now().UnixNano(),
		multiMsg.GetPbItemList(),
	)
}

func (c *QQClient) SendGroupPoke(groupCode, target int64) {
	_, _ = c.sendAndWait(c.buildGroupPokePacket(groupCode, target))
}

func (c *QQClient) SendFriendPoke(target int64) {
	_, _ = c.sendAndWait(c.buildFriendPokePacket(target))
}

func (c *QQClient) ReloadGroupList() error {
	c.groupListLock.Lock()
	defer c.groupListLock.Unlock()
	list, err := c.GetGroupList()
	if err != nil {
		return err
	}
	c.GroupList = list
	return nil
}

func (c *QQClient) GetGroupList() ([]*GroupInfo, error) {
	rsp, err := c.sendAndWait(c.buildGroupListRequestPacket(EmptyBytes))
	if err != nil {
		return nil, err
	}
	r := rsp.([]*GroupInfo)
	wg := sync.WaitGroup{}
	batch := 50
	for i := 0; i < len(r); i += batch {
		k := i + batch
		if k > len(r) {
			k = len(r)
		}
		wg.Add(k - i)
		for j := i; j < k; j++ {
			go func(g *GroupInfo, wg *sync.WaitGroup) {
				defer wg.Done()
				m, err := c.GetGroupMembers(g)
				if err != nil {
					return
				}
				g.Members = m
			}(r[j], &wg)
		}
		wg.Wait()
	}
	return r, nil
}

func (c *QQClient) GetGroupMembers(group *GroupInfo) ([]*GroupMemberInfo, error) {
	var nextUin int64
	var list []*GroupMemberInfo
	for {
		data, err := c.sendAndWait(c.buildGroupMemberListRequestPacket(group.Uin, group.Code, nextUin))
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, errors.New("group member list unavailable: rsp is nil")
		}
		rsp := data.(groupMemberListResponse)
		nextUin = rsp.NextUin
		for _, m := range rsp.list {
			m.Group = group
			if m.Uin == group.OwnerUin {
				m.Permission = Owner
			}
		}
		list = append(list, rsp.list...)
		if nextUin == 0 {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Uin < list[j].Uin
			})
			return list, nil
		}
	}
}

func (c *QQClient) GetMemberInfo(groupCode, memberUin int64) (*GroupMemberInfo, error) {
	info, err := c.sendAndWait(c.buildGroupMemberInfoRequestPacket(groupCode, memberUin))
	if err != nil {
		return nil, err
	}
	return info.(*GroupMemberInfo), nil
}

func (c *QQClient) FindFriend(uin int64) *FriendInfo {
	for _, t := range c.FriendList {
		f := t
		if f.Uin == uin {
			return f
		}
	}
	return nil
}

func (c *QQClient) DeleteFriend(uin int64) error {
	if c.FindFriend(uin) == nil {
		return errors.New("friend not found")
	}
	_, err := c.sendAndWait(c.buildFriendDeletePacket(uin))
	return errors.Wrap(err, "delete friend error")
}

func (c *QQClient) FindGroupByUin(uin int64) *GroupInfo {
	for _, g := range c.GroupList {
		f := g
		if f.Uin == uin {
			return f
		}
	}
	return nil
}

func (c *QQClient) FindGroup(code int64) *GroupInfo {
	for _, g := range c.GroupList {
		f := g
		if f.Code == code {
			return f
		}
	}
	return nil
}

func (c *QQClient) SolveGroupJoinRequest(i interface{}, accept, block bool, reason string) {
	if accept {
		block = false
		reason = ""
	}

	switch req := i.(type) {
	case *UserJoinGroupRequest:
		_, pkt := c.buildSystemMsgGroupActionPacket(req.RequestId, req.RequesterUin, req.GroupCode, func() int32 {
			if req.Suspicious {
				return 2
			} else {
				return 1
			}
		}(), false, accept, block, reason)
		_ = c.send(pkt)
	case *GroupInvitedRequest:
		_, pkt := c.buildSystemMsgGroupActionPacket(req.RequestId, req.InvitorUin, req.GroupCode, 1, true, accept, block, reason)
		_ = c.send(pkt)
	}
}

func (c *QQClient) SolveFriendRequest(req *NewFriendRequest, accept bool) {
	_, pkt := c.buildSystemMsgFriendActionPacket(req.RequestId, req.RequesterUin, accept)
	_ = c.send(pkt)
}

func (c *QQClient) getSKey() string {
	if c.sigInfo.sKeyExpiredTime < time.Now().Unix() && len(c.g) > 0 {
		c.Debug("skey expired. refresh...")
		_, _ = c.sendAndWait(c.buildRequestTgtgtNopicsigPacket())
	}
	return string(c.sigInfo.sKey)
}

func (c *QQClient) getCookies() string {
	return fmt.Sprintf("uin=o%d; skey=%s;", c.Uin, c.getSKey())
}

func (c *QQClient) getCookiesWithDomain(domain string) string {
	cookie := c.getCookies()

	if psKey, ok := c.sigInfo.psKeyMap[domain]; ok {
		return fmt.Sprintf("%s p_uin=o%d; p_skey=%s;", cookie, c.Uin, psKey)
	} else {
		return cookie
	}
}

func (c *QQClient) getCSRFToken() int {
	accu := 5381
	for _, b := range c.sigInfo.sKey {
		accu = accu + (accu << 5) + int(b)
	}
	return 2147483647 & accu
}

func (c *QQClient) editMemberCard(groupCode, memberUin int64, card string) {
	_, _ = c.sendAndWait(c.buildEditGroupTagPacket(groupCode, memberUin, card))
}

func (c *QQClient) editMemberSpecialTitle(groupCode, memberUin int64, title string) {
	_, _ = c.sendAndWait(c.buildEditSpecialTitlePacket(groupCode, memberUin, title))
}

func (c *QQClient) setGroupAdmin(groupCode, memberUin int64, flag bool) {
	_, _ = c.sendAndWait(c.buildGroupAdminSetPacket(groupCode, memberUin, flag))
}

func (c *QQClient) updateGroupName(groupCode int64, newName string) {
	_, _ = c.sendAndWait(c.buildGroupNameUpdatePacket(groupCode, newName))
}

func (c *QQClient) updateGroupMemo(groupCode int64, newMemo string) {
	_, _ = c.sendAndWait(c.buildGroupMemoUpdatePacket(groupCode, newMemo))
}

func (c *QQClient) groupMuteAll(groupCode int64, mute bool) {
	_, _ = c.sendAndWait(c.buildGroupMuteAllPacket(groupCode, mute))
}

func (c *QQClient) groupMute(groupCode, memberUin int64, time uint32) {
	_, _ = c.sendAndWait(c.buildGroupMutePacket(groupCode, memberUin, time))
}

func (c *QQClient) quitGroup(groupCode int64) {
	_, _ = c.sendAndWait(c.buildQuitGroupPacket(groupCode))
}

func (c *QQClient) kickGroupMember(groupCode, memberUin int64, msg string, block bool) {
	_, _ = c.sendAndWait(c.buildGroupKickPacket(groupCode, memberUin, msg, block))
}

func (g *GroupInfo) removeMember(uin int64) {
	g.Update(func(info *GroupInfo) {
		i := sort.Search(len(info.Members), func(i int) bool {
			return info.Members[i].Uin >= uin
		})
		if i >= len(info.Members) || info.Members[i].Uin != uin { // not found
			return
		}
		info.Members = append(info.Members[:i], info.Members[i+1:]...)
	})
}

func (c *QQClient) SetCustomServer(servers []*net.TCPAddr) {
	c.servers = append(servers, c.servers...)
}

func (c *QQClient) SendGroupGift(groupCode, uin uint64, gift message.GroupGift) {
	_, packet := c.sendGroupGiftPacket(groupCode, uin, gift)
	_ = c.send(packet)
}

func (c *QQClient) registerClient() error {
	_, err := c.sendAndWait(c.buildClientRegisterPacket())
	if err == nil {
		c.Online = true
	}
	return err
}

func (c *QQClient) nextSeq() uint16 {
	return uint16(atomic.AddInt32(&c.SequenceId, 1) & 0x7FFF)
}

func (c *QQClient) nextPacketSeq() int32 {
	return atomic.AddInt32(&c.requestPacketRequestID, 2)
}

func (c *QQClient) nextGroupSeq() int32 {
	return atomic.AddInt32(&c.groupSeq, 2)
}

func (c *QQClient) nextFriendSeq() int32 {
	return atomic.AddInt32(&c.friendSeq, 1)
}

func (c *QQClient) nextGroupDataTransSeq() int32 {
	return atomic.AddInt32(&c.groupDataTransSeq, 2)
}

func (c *QQClient) nextHighwayApplySeq() int32 {
	return atomic.AddInt32(&c.highwayApplyUpSeq, 2)
}

func (c *QQClient) send(pkt []byte) error {
	err := c.TCP.Write(pkt)
	if err != nil {
		c.stat.PacketLost++
	} else {
		c.stat.PacketSent++
	}
	return errors.Wrap(err, "Packet failed to send")
}

func (c *QQClient) sendAndWait(seq uint16, pkt []byte, params ...requestParams) (interface{}, error) {
	type T struct {
		Response interface{}
		Error    error
	}

	err := c.send(pkt)
	if err != nil {
		return nil, err
	}

	ch := make(chan T)
	defer close(ch)

	p := func() requestParams {
		if len(params) == 0 {
			return nil
		}
		return params[0]
	}()

	c.handlers.Store(seq, &handlerInfo{fun: func(i interface{}, err error) {
		ch <- T{
			Response: i,
			Error:    err,
		}
	}, params: p})

	retry := 0
	for {
		select {
		case rsp := <-ch:
			return rsp.Response, rsp.Error
		case <-time.After(time.Second * 15):
			retry++
			if retry < 2 {
				_ = c.send(pkt)
				continue
			}
			c.handlers.Delete(seq)
			// c.Error("packet timed out, seq: %v", seq)
			// println("Packet Timed out")
			return nil, errors.New("Packet timed out")
		}
	}
}

// 等待一个或多个数据包解析, 优先级低于 sendAndWait
// 返回终止解析函数
func (c *QQClient) waitPacket(cmd string, f func(interface{}, error)) func() {
	c.waiters.Store(cmd, f)
	return func() {
		c.waiters.Delete(cmd)
	}
}

func (c *QQClient) connect() error {
	c.Info("connect to server: %v", c.servers[c.currServerIndex].String())
	err := c.TCP.Connect(c.servers[c.currServerIndex])
	c.currServerIndex++
	if c.currServerIndex == len(c.servers) {
		c.currServerIndex = 0
	}
	if err != nil {
		c.retryTimes++
		if c.retryTimes > len(c.servers) {
			return errors.New("All servers are unreachable")
		}
		c.Error("connect server error: %v", err)
		return err
	}
	c.retryTimes = 0
	c.ConnectTime = time.Now()
	return nil
}

func (c *QQClient) quickReconnect() {
	c.Disconnect()
	time.Sleep(time.Millisecond * 200)
	if err := c.connect(); err != nil {
		c.Error("connect server error: %v", err)
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "quick reconnect failed"})
		return
	}
	if err := c.registerClient(); err != nil {
		c.Error("register client failed: %v", err)
		c.Disconnect()
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "register error"})
		return
	}
}

func (c *QQClient) Disconnect() {
	c.Online = false
	c.TCP.Close()
}

func (c *QQClient) plannedDisconnect(_ *utils.TCPListener) {
	c.Debug("planned disconnect.")
	c.stat.DisconnectTimes++
	c.Online = false
}

func (c *QQClient) unexpectedDisconnect(_ *utils.TCPListener, e error) {
	c.Error("unexpected disconnect: %v", e)
	c.stat.DisconnectTimes++
	c.Online = false
	if err := c.connect(); err != nil {
		c.Error("connect server error: %v", err)
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "connection dropped by server."})
		return
	}
	if err := c.registerClient(); err != nil {
		c.Error("register client failed: %v", err)
		c.Disconnect()
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "register error"})
		return
	}
}

func (c *QQClient) netLoop() {
	// todo: release this
	errCount := 0
	for {
		l, err := c.TCP.ReadInt32()
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		data, _ := c.TCP.ReadBytes(int(l) - 4)
		pkt, err := packets.ParseIncomingPacket(data, c.sigInfo.d2Key)
		if err != nil {
			c.Error("parse incoming packet error: %v", err)
			if errors.Is(err, packets.ErrSessionExpired) || errors.Is(err, packets.ErrPacketDropped) {
				c.Disconnect()
				go c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "session expired"})
				continue
			}
			errCount++
			if errCount > 2 {
				go c.quickReconnect()
				continue
			}
			continue
		}
		payload := pkt.Payload
		if pkt.Flag2 == 2 {
			payload, err = pkt.DecryptPayload(c.RandomKey, c.sigInfo.wtSessionTicketKey)
			if err != nil {
				c.Error("decrypt payload error: %v", err)
				continue
			}
		}
		errCount = 0
		c.Debug("rev pkt: %v seq: %v", pkt.CommandName, pkt.SequenceId)
		c.stat.PacketReceived++
		go func() {
			defer func() {
				if pan := recover(); pan != nil {
					c.Error("panic on decoder %v : %v\n%s", pkt.CommandName, pan, debug.Stack())
				}
			}()

			if decoder, ok := decoders[pkt.CommandName]; ok {
				// found predefined decoder
				info, ok := c.handlers.LoadAndDelete(pkt.SequenceId)
				rsp, err := decoder(c, &incomingPacketInfo{
					SequenceId:  pkt.SequenceId,
					CommandName: pkt.CommandName,
					Params: func() requestParams {
						if !ok {
							return nil
						}
						return info.params
					}(),
				}, payload)
				if err != nil {
					c.Debug("decode pkt %v error: %+v", pkt.CommandName, err)
				}
				if ok {
					info.fun(rsp, err)
				} else if f, ok := c.waiters.Load(pkt.CommandName); ok { // 在不存在handler的情况下触发wait
					f.(func(interface{}, error))(rsp, err)
				}
			} else if f, ok := c.handlers.LoadAndDelete(pkt.SequenceId); ok {
				// does not need decoder
				f.fun(nil, nil)
			} else {
				c.Debug("Unhandled Command: %s\nSeq: %d\nThis message can be ignored.", pkt.CommandName, pkt.SequenceId)
			}
		}()
	}
	/*
		c.NetLooping = false
		c.Online = false
		_ = c.TCP.Close()
		if c.lastLostMsg == "" {
			c.lastLostMsg = "Connection lost."
		}
		c.stat.LostTimes++
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: c.lastLostMsg})
	*/
}

func (c *QQClient) doHeartbeat() {
	c.heartbeatEnabled = true
	times := 0
	for c.Online {
		time.Sleep(time.Second * 30)
		seq := c.nextSeq()
		sso := packets.BuildSsoPacket(seq, c.version.AppId, c.version.SubAppId, "Heartbeat.Alive", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, []byte{}, c.ksid)
		packet := packets.BuildLoginPacket(c.Uin, 0, []byte{}, sso, []byte{})
		_, err := c.sendAndWait(seq, packet)
		if errors.Is(err, utils.ErrConnectionClosed) {
			continue
		}
		times++
		if times >= 7 {
			_ = c.registerClient()
			times = 0
		}
	}
	c.heartbeatEnabled = false
}
