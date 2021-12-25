package client

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/atomic"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/internal/auth"
	"github.com/Mrs4s/MiraiGo/client/internal/highway"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/internal/oicq"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/utils"
)

type QQClient struct {
	Uin         int64
	PasswordMd5 [16]byte

	stat Statistics
	once sync.Once

	// option
	AllowSlider bool

	// account info
	Online        atomic.Bool
	Nickname      string
	Age           uint16
	Gender        uint16
	FriendList    []*FriendInfo
	GroupList     []*GroupInfo
	OnlineClients []*OtherClientInfo
	QiDian        *QiDianAccountInfo
	GuildService  *GuildService

	// protocol public field
	SequenceId  atomic.Int32
	SessionId   []byte
	RandomKey   []byte
	TCP         *network.TCPListener
	ConnectTime time.Time

	// todo: combine net conn, transport, pending into one struct
	pendingMu sync.Mutex
	pending   map[uint16]*network.Call
	transport *network.Transport
	oicq      *oicq.Codec

	// internal state
	waiters         sync.Map
	servers         []*net.TCPAddr
	currServerIndex int
	retryTimes      int
	version         *auth.AppVersion
	deviceInfo      *auth.Device
	alive           bool

	// session info
	qwebSeq        atomic.Int64
	sig            *auth.SigInfo
	highwaySession *highway.Session
	// pwdFlag        bool
	// timeDiff       int64

	// address
	otherSrvAddrs   []string
	fileStorageInfo *jce.FileStoragePushFSSvcList

	// message state
	msgSvcCache            *utils.Cache
	lastC2CMsgTime         int64
	transCache             *utils.Cache
	groupSysMsgCache       *GroupSystemMessages
	groupMsgBuilders       sync.Map
	onlinePushCache        *utils.Cache
	heartbeatEnabled       bool
	requestPacketRequestID atomic.Int32
	groupSeq               atomic.Int32
	friendSeq              atomic.Int32
	highwayApplyUpSeq      atomic.Int32
	eventHandlers          *eventHandlers

	groupListLock sync.Mutex
}

type QiDianAccountInfo struct {
	MasterUin  int64
	ExtName    string
	CreateTime int64

	bigDataReqAddrs   []string
	bigDataReqSession *bigDataSessionInfo
}

var decoders = map[string]func(*QQClient, *network.Response) (interface{}, error){
	"StatSvc.ReqMSFOffline":       decodeMSFOfflinePacket,
	"MessageSvc.PushNotify":       decodeSvcNotify,
	"OnlinePush.ReqPush":          decodeOnlinePushReqPacket,
	"OnlinePush.PbPushTransMsg":   decodeOnlinePushTransPacket,
	"OnlinePush.SidTicketExpired": decodeSidExpiredPacket,
	"ConfigPushSvc.PushReq":       decodePushReqPacket,
	"MessageSvc.PushForceOffline": decodeForceOfflinePacket,
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// NewClient create new qq client
func NewClient(uin int64, password string) *QQClient {
	return NewClientMd5(uin, md5.Sum([]byte(password)))
}

func NewClientEmpty() *QQClient {
	return NewClientMd5(0, [16]byte{})
}

func NewClientMd5(uin int64, passwordMd5 [16]byte) *QQClient {
	cli := &QQClient{
		Uin:         uin,
		PasswordMd5: passwordMd5,
		AllowSlider: true,
		TCP:         &network.TCPListener{},
		sig: &auth.SigInfo{
			OutPacketSessionID: []byte{0x02, 0xB0, 0x5B, 0x8B},
		},
		eventHandlers:   &eventHandlers{},
		msgSvcCache:     utils.NewCache(time.Second * 15),
		transCache:      utils.NewCache(time.Second * 15),
		onlinePushCache: utils.NewCache(time.Second * 15),
		servers:         []*net.TCPAddr{},
		alive:           true,
		highwaySession:  new(highway.Session),

		pending:    make(map[uint16]*network.Call),
		version:    new(auth.AppVersion),
		deviceInfo: new(auth.Device),
	}

	cli.transport = &network.Transport{
		Sig:     cli.sig,
		Version: cli.version,
		Device:  cli.deviceInfo,
	}
	cli.oicq = oicq.NewCodec(cli.Uin)
	{ // init atomic values
		cli.SequenceId.Store(0x3635)
		cli.requestPacketRequestID.Store(1921334513)
		cli.groupSeq.Store(int32(rand.Intn(20000)))
		cli.friendSeq.Store(22911)
		cli.highwayApplyUpSeq.Store(77918)
	}
	cli.highwaySession.Uin = strconv.FormatInt(cli.Uin, 10)
	cli.GuildService = &GuildService{c: cli}
	cli.UseDevice(SystemDeviceInfo)
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
			{IP: net.IP{42, 81, 172, 22}, Port: 80},
			{IP: net.IP{42, 81, 172, 81}, Port: 80},
			{IP: net.IP{42, 81, 172, 147}, Port: 443},
			{IP: net.IP{114, 221, 144, 215}, Port: 80},
			{IP: net.IP{114, 221, 148, 59}, Port: 14000},
			{IP: net.IP{125, 94, 60, 146}, Port: 80},
		}
	}
	/*pings := make([]int64, len(cli.servers))
	wg := sync.WaitGroup{}
	wg.Add(len(cli.servers))
	for i := range cli.servers {
		go func(index int) {
			defer wg.Done()
			p, err := qualityTest(cli.servers[index].String())
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
	}*/
	cli.TCP.PlannedDisconnect = cli.plannedDisconnect
	cli.TCP.UnexpectedDisconnect = cli.unexpectedDisconnect
	rand.Read(cli.RandomKey)
	return cli
}

func (c *QQClient) UseDevice(info *auth.Device) {
	*c.version = *info.Protocol.Version()
	*c.deviceInfo = *info
	c.highwaySession.AppID = int32(c.version.AppId)
	c.sig.Ksid = []byte(fmt.Sprintf("|%s|A8.2.7.27f6ea96", info.IMEI))
}

func (c *QQClient) Release() {
	if c.Online.Load() {
		c.Disconnect()
	}
	c.alive = false
}

// Login send login request
func (c *QQClient) Login() (*LoginResponse, error) {
	if c.Online.Load() {
		return nil, ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return nil, err
	}
	rsp, err := c.callAndDecode(c.buildLoginRequest(), decodeLoginResponse)
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		err = c.init(false)
	}
	return &l, err
}

func (c *QQClient) TokenLogin(token []byte) error {
	if c.Online.Load() {
		return ErrAlreadyOnline
	}
	err := c.LoadToken(token)
	if err != nil {
		return err
	}
	return c.ReLogin()
}

func (c *QQClient) ReLogin() error {
	if c.Online.Load() {
		return ErrAlreadyOnline
	}
	err := c.connectFastest()
	if err != nil {
		return err
	}
	_, err = c.callAndDecode(c.buildRequestChangeSigRequest(c.version.MainSigMap), decodeExchangeEmpResponse)
	if err != nil {
		return err
	}
	err = c.init(true)
	// 登录失败
	if err != nil {
		c.Disconnect()
	}
	return err
}

func (c *QQClient) DumpToken() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt64(uint64(c.Uin))
		w.WriteBytesShort(c.sig.D2)
		w.WriteBytesShort(c.sig.D2Key)
		w.WriteBytesShort(c.sig.TGT)
		w.WriteBytesShort(c.sig.SrmToken)
		w.WriteBytesShort(c.sig.T133)
		w.WriteBytesShort(c.sig.EncryptedA1)
		w.WriteBytesShort(c.oicq.WtSessionTicketKey)
		w.WriteBytesShort(c.sig.OutPacketSessionID)
		w.WriteBytesShort(c.deviceInfo.TgtgtKey)
	})
}

func (c *QQClient) LoadToken(token []byte) error {
	return utils.CoverError(func() {
		r := binary.NewReader(token)
		c.Uin = r.ReadInt64()
		c.sig.D2 = r.ReadBytesShort()
		c.sig.D2Key = r.ReadBytesShort()
		c.sig.TGT = r.ReadBytesShort()
		c.sig.SrmToken = r.ReadBytesShort()
		c.sig.T133 = r.ReadBytesShort()
		c.sig.EncryptedA1 = r.ReadBytesShort()
		c.oicq.WtSessionTicketKey = r.ReadBytesShort()
		c.sig.OutPacketSessionID = r.ReadBytesShort()
		// SystemDeviceInfo.TgtgtKey = r.ReadBytesShort()
		c.deviceInfo.TgtgtKey = r.ReadBytesShort()
		copy(SystemDeviceInfo.TgtgtKey, c.deviceInfo.TgtgtKey)
	})
}

func (c *QQClient) DumpDevice() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteBytesShort(c.deviceInfo.Display)
		w.WriteBytesShort(c.deviceInfo.Product)
		w.WriteBytesShort(c.deviceInfo.Device)
		w.WriteBytesShort(c.deviceInfo.Board)
		w.WriteBytesShort(c.deviceInfo.Brand)
		w.WriteBytesShort(c.deviceInfo.Model)
		w.WriteBytesShort(c.deviceInfo.Bootloader)
		w.WriteBytesShort(c.deviceInfo.FingerPrint)
		w.WriteBytesShort(c.deviceInfo.BootId)
		w.WriteBytesShort(c.deviceInfo.ProcVersion)
		w.WriteBytesShort(c.deviceInfo.BaseBand)
		w.WriteBytesShort(c.deviceInfo.SimInfo)
		w.WriteBytesShort(c.deviceInfo.OSType)
		w.WriteBytesShort(c.deviceInfo.MacAddress)
		w.WriteBytesShort(c.deviceInfo.IpAddress)
		w.WriteBytesShort(c.deviceInfo.WifiBSSID)
		w.WriteBytesShort(c.deviceInfo.WifiSSID)
		w.WriteBytesShort(c.deviceInfo.IMSIMd5)
		w.WriteStringShort(c.deviceInfo.IMEI)
		w.WriteBytesShort(c.deviceInfo.APN)
		w.WriteBytesShort(c.deviceInfo.VendorName)
		w.WriteBytesShort(c.deviceInfo.VendorOSName)
		w.WriteBytesShort(c.deviceInfo.AndroidId)

		w.Write(c.PasswordMd5[:])
	})
}

func (c *QQClient) LoadDevice(device []byte) error {
	return utils.CoverError(func() {
		r := binary.NewReader(device)
		c.deviceInfo.Display = r.ReadBytesShort()
		c.deviceInfo.Product = r.ReadBytesShort()
		c.deviceInfo.Device = r.ReadBytesShort()
		c.deviceInfo.Board = r.ReadBytesShort()
		c.deviceInfo.Brand = r.ReadBytesShort()
		c.deviceInfo.Model = r.ReadBytesShort()
		c.deviceInfo.Bootloader = r.ReadBytesShort()
		c.deviceInfo.FingerPrint = r.ReadBytesShort()
		c.deviceInfo.BootId = r.ReadBytesShort()
		c.deviceInfo.ProcVersion = r.ReadBytesShort()
		c.deviceInfo.BaseBand = r.ReadBytesShort()
		c.deviceInfo.SimInfo = r.ReadBytesShort()
		c.deviceInfo.OSType = r.ReadBytesShort()
		c.deviceInfo.MacAddress = r.ReadBytesShort()
		c.deviceInfo.IpAddress = r.ReadBytesShort()
		c.deviceInfo.WifiBSSID = r.ReadBytesShort()
		c.deviceInfo.WifiSSID = r.ReadBytesShort()
		c.deviceInfo.IMSIMd5 = r.ReadBytesShort()
		c.deviceInfo.IMEI = r.ReadStringShort()
		c.deviceInfo.APN = r.ReadBytesShort()
		c.deviceInfo.VendorName = r.ReadBytesShort()
		c.deviceInfo.VendorOSName = r.ReadBytesShort()
		c.deviceInfo.AndroidId = r.ReadBytesShort()

		copy(c.PasswordMd5[:], r.ReadBytes(md5.Size))
	})
}

// FetchQRCode 以默认值获取登录二维码
// 函数已被弃用 请使用FetchQRCodeCustomSize获得更可控结果
// 但该兼容函数不会被删除
// Deprecated use FetchQRCodeCustomSize(3, 4, 2) instead
func (c *QQClient) FetchQRCode() (*QRCodeLoginResponse, error) {
	return c.FetchQRCodeCustomSize(3, 4, 2)
}

// FetchQRCodeCustomSize 以特定参数获取登录二维码
// size: 块尺寸 默认值3 即单个黑/白块大小为3x3像素
// margin: 与图片边界的距离 默认值4 即二维码主体至图片边界有4像素白色填充
// ecLevel: 纠错等级 可用值：1,2,3 默认值2
func (c *QQClient) FetchQRCodeCustomSize(size, margin, ecLevel uint32) (*QRCodeLoginResponse, error) {
	if c.Online.Load() {
		return nil, ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return nil, err
	}
	c.transport.Version = auth.AndroidWatch.Version()
	i, err := c.callAndDecode(c.buildQRCodeFetchRequest(size, margin, ecLevel), decodeTransEmpResponse)
	c.transport.Version = c.version
	if err != nil {
		return nil, errors.Wrap(err, "fetch qrcode error")
	}
	return i.(*QRCodeLoginResponse), nil
}

func (c *QQClient) QueryQRCodeStatus(sig []byte) (*QRCodeLoginResponse, error) {
	c.transport.Version = auth.AndroidWatch.Version()
	i, err := c.callAndDecode(c.buildQRCodeResultQueryRequest(sig), decodeTransEmpResponse)
	c.transport.Version = c.version
	if err != nil {
		return nil, errors.Wrap(err, "query result error")
	}
	return i.(*QRCodeLoginResponse), nil
}

func (c *QQClient) QRCodeLogin(info *QRCodeLoginInfo) (*LoginResponse, error) {
	i, err := c.callAndDecode(c.buildQRCodeLoginRequest(info.tmpPwd, info.tmpNoPicSig, info.tgtQR), decodeLoginResponse)
	if err != nil {
		return nil, errors.Wrap(err, "qrcode login error")
	}
	rsp := i.(LoginResponse)
	if rsp.Success {
		err = c.init(false)
	}
	return &rsp, err
}

// SubmitCaptcha send captcha to server
func (c *QQClient) SubmitCaptcha(result string, sign []byte) (*LoginResponse, error) {
	req := c.buildCaptchaRequest(result, sign)
	rsp, err := c.callAndDecode(req, decodeLoginResponse)
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		err = c.init(false)
	}
	return &l, err
}

func (c *QQClient) SubmitTicket(ticket string) (*LoginResponse, error) {
	req := c.buildTicketSubmitRequest(ticket)
	rsp, err := c.callAndDecode(req, decodeLoginResponse)
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		err = c.init(false)
	}
	return &l, err
}

func (c *QQClient) SubmitSMS(code string) (*LoginResponse, error) {
	rsp, err := c.callAndDecode(c.buildSMSCodeSubmitRequest(code), decodeLoginResponse)
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		err = c.init(false)
	}
	return &l, err
}

func (c *QQClient) RequestSMS() bool {
	rsp, err := c.callAndDecode(c.buildSMSRequest(), decodeLoginResponse)
	if err != nil {
		c.Error("request sms error: %v", err)
		return false
	}
	return rsp.(LoginResponse).Error == SMSNeededError
}

func (c *QQClient) init(tokenLogin bool) error {
	if len(c.sig.G) == 0 {
		c.Warning("device lock is disable. http api may fail.")
	}
	c.highwaySession.Uin = strconv.FormatInt(c.Uin, 10)
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
	go c.doHeartbeat()
	_ = c.RefreshStatus()
	if c.version.Protocol == auth.QiDian {
		_, _ = c.callAndDecode(c.buildLoginExtraPacket(), decodeLoginExtraResponse)  // 小登录
		_, _ = c.callAndDecode(c.buildConnKeyRequestPacket(), decodeConnKeyResponse) // big data key 如果等待 config push 的话时间来不及
	}
	c.groupSysMsgCache, _ = c.GetGroupSystemMessages()
	req := c.buildGetMessageRequest(msg.SyncFlag_START, time.Now().Unix())
	req.Params = network.Params{"used_reg_proxy": true, "init": true}
	_, _ = c.callAndDecode(req, decodeMessageSvcPacket)
	c.syncChannelFirstView()
	return nil
}

func (c *QQClient) GenToken() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt64(uint64(c.Uin))
		w.WriteBytesShort(c.sig.D2)
		w.WriteBytesShort(c.sig.D2Key)
		w.WriteBytesShort(c.sig.TGT)
		w.WriteBytesShort(c.sig.SrmToken)
		w.WriteBytesShort(c.sig.T133)
		w.WriteBytesShort(c.sig.EncryptedA1)
		w.WriteBytesShort(c.oicq.WtSessionTicketKey)
		w.WriteBytesShort(c.sig.OutPacketSessionID)
		w.WriteBytesShort(c.deviceInfo.TgtgtKey)
	})
}

func (c *QQClient) SetOnlineStatus(s UserOnlineStatus) {
	if s < 1000 {
		_, _ = c.call(c.buildStatusSetPacket(int32(s), 0))
		return
	}
	_, _ = c.call(c.buildStatusSetPacket(11, int32(s)))
}

func (c *QQClient) GetWordSegmentation(text string) ([]string, error) {
	rsp, err := c.callAndDecode(c.buildWordSegmentationPacket([]byte(text)), decodeWordSegmentation)
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
	rsp, err := c.callAndDecode(c.buildSummaryCardRequest(target), decodeSummaryCardResponse)
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
	if c.version.Protocol == auth.QiDian {
		rsp, err := c.getQiDianAddressDetailList()
		if err != nil {
			return nil, err
		}
		return &FriendListResponse{TotalCount: int32(len(rsp)), List: rsp}, nil
	}
	curFriendCount := 0
	r := &FriendListResponse{}
	for {
		call := c.buildFriendGroupListRequest(int16(curFriendCount), 150, 0, 0)
		rsp, err := c.callAndDecode(call, decodeFriendGroupListResponse)
		if err != nil {
			return nil, err
		}
		list := rsp.(*FriendListResponse)
		r.TotalCount = list.TotalCount
		r.List = append(r.List, list.List...)
		curFriendCount += len(list.List)
		if int32(len(r.List)) >= r.TotalCount {
			break
		}
	}
	return r, nil
}

func (c *QQClient) SendGroupPoke(groupCode, target int64) {
	_, _ = c.call(c.buildGroupPokeRequest(groupCode, target))
}

func (c *QQClient) SendFriendPoke(target int64) {
	_, _ = c.call(c.buildFriendPokeRequest(target))
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
	rsp, err := c.callAndDecode(c.buildGroupListRequest(EmptyBytes), decodeGroupListResponse)
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
		data, err := c.callAndDecode(c.buildGroupMemberListRequest(group.Uin, group.Code, nextUin), decodeGroupMemberListResponse)
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, errors.New("group member list unavailable: rsp is nil")
		}
		rsp := data.(*groupMemberListResponse)
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
	info, err := c.callAndDecode(c.buildGroupMemberInfoRequest(groupCode, memberUin), decodeGroupMemberInfoResponse)
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
	_, err := c.callAndDecode(c.buildFriendDeletePacket(uin), decodeFriendDeleteResponse)
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
		if g.Code == code {
			return g
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
		call := c.buildSystemMsgGroupActionPacket(req.RequestId, req.RequesterUin, req.GroupCode, func() int32 {
			if req.Suspicious {
				return 2
			} else {
				return 1
			}
		}(), false, accept, block, reason)
		_, _ = c.call(call)
	case *GroupInvitedRequest:
		call := c.buildSystemMsgGroupActionPacket(req.RequestId, req.InvitorUin, req.GroupCode, 1, true, accept, block, reason)
		_, _ = c.call(call)
	}
}

func (c *QQClient) SolveFriendRequest(req *NewFriendRequest, accept bool) {
	_, _ = c.call(c.buildSystemMsgFriendActionPacket(req.RequestId, req.RequesterUin, accept))
}

func (c *QQClient) getSKey() string {
	if c.sig.SKeyExpiredTime < time.Now().Unix() && len(c.sig.G) > 0 {
		c.Debug("skey expired. refresh...")
		_, _ = c.callAndDecode(c.buildRequestTgtgtNopicsigRequest(), decodeExchangeEmpResponse)
	}
	return string(c.sig.SKey)
}

func (c *QQClient) getCookies() string {
	return fmt.Sprintf("uin=o%d; skey=%s;", c.Uin, c.getSKey())
}

func (c *QQClient) getCookiesWithDomain(domain string) string {
	cookie := c.getCookies()

	if psKey, ok := c.sig.PsKeyMap[domain]; ok {
		return fmt.Sprintf("%s p_uin=o%d; p_skey=%s;", cookie, c.Uin, psKey)
	} else {
		return cookie
	}
}

func (c *QQClient) getCSRFToken() int {
	accu := 5381
	for _, b := range []byte(c.getSKey()) {
		accu = accu + (accu << 5) + int(b)
	}
	return 2147483647 & accu
}

func (c *QQClient) editMemberCard(groupCode, memberUin int64, card string) {
	_, _ = c.call(c.buildEditGroupTagPacket(groupCode, memberUin, card))
}

func (c *QQClient) editMemberSpecialTitle(groupCode, memberUin int64, title string) {
	_, _ = c.call(c.buildEditSpecialTitlePacket(groupCode, memberUin, title))
}

func (c *QQClient) setGroupAdmin(groupCode, memberUin int64, flag bool) {
	_, _ = c.call(c.buildGroupAdminSetPacket(groupCode, memberUin, flag))
}

func (c *QQClient) updateGroupName(groupCode int64, newName string) {
	_, _ = c.call(c.buildGroupNameUpdateRequest(groupCode, newName))
}

func (c *QQClient) updateGroupMemo(groupCode int64, newMemo string) {
	_, _ = c.call(c.buildGroupMemoUpdatePacket(groupCode, newMemo))
}

func (c *QQClient) groupMuteAll(groupCode int64, mute bool) {
	_, _ = c.call(c.buildGroupMuteAllPacket(groupCode, mute))
}

func (c *QQClient) groupMute(groupCode, memberUin int64, time uint32) {
	_, _ = c.call(c.buildGroupMutePacket(groupCode, memberUin, time))
}

func (c *QQClient) quitGroup(groupCode int64) {
	_, _ = c.call(c.buildQuitGroupPacket(groupCode))
}

func (c *QQClient) kickGroupMember(groupCode, memberUin int64, msg string, block bool) {
	_, _ = c.call(c.buildGroupKickPacket(groupCode, memberUin, msg, block))
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

func (c *QQClient) registerClient() error {
	_, err := c.callAndDecode(c.buildClientRegisterPacket(), decodeClientRegisterResponse)
	if err == nil {
		c.Online.Store(true)
	}
	return err
}

func (c *QQClient) nextSeq() uint16 {
	return uint16(c.SequenceId.Add(1) & 0x7FFF)
}

func (c *QQClient) nextPacketSeq() int32 {
	return c.requestPacketRequestID.Add(2)
}

func (c *QQClient) nextGroupSeq() int32 {
	return c.groupSeq.Add(2)
}

func (c *QQClient) nextFriendSeq() int32 {
	return c.friendSeq.Add(1)
}

func (c *QQClient) nextQWebSeq() int64 {
	return c.qwebSeq.Add(1)
}

func (c *QQClient) nextHighwayApplySeq() int32 {
	return c.highwayApplyUpSeq.Add(2)
}

func (c *QQClient) doHeartbeat() {
	// 不需要atomic/锁
	if c.heartbeatEnabled {
		return
	}
	c.heartbeatEnabled = true
	defer func() {
		c.heartbeatEnabled = false
	}()
	times := 0
	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		if !c.Online.Load() {
			ticker.Stop()
			return // 下线停止goroutine，for gc
		}
		seq := c.nextSeq()
		req := network.Request{
			Type:        network.RequestTypeLogin,
			EncryptType: network.EncryptTypeNoEncrypt,
			SequenceID:  int32(seq),
			Uin:         c.Uin,
			CommandName: "wtlogin.login",
			Body:        EmptyBytes,
		}
		_, err := c.call(&req)
		//if err != nil {
		//	if errors.Is(err, ErrNotConnected) {
		//		continue
		//	}
		//}
		times++
		if times >= 7 {
			_ = c.registerClient()
			times = 0
		}
	}
}
