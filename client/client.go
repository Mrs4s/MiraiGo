package client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"math/rand"
	"net"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
)

//go:generate go run github.com/a8m/syncmap -o "handler_map_gen.go" -pkg client -name HandlerMap "map[uint16]func(i interface{}, err error)"

type QQClient struct {
	Uin         int64
	PasswordMd5 [16]byte
	AllowSlider bool

	Nickname   string
	Age        uint16
	Gender     uint16
	FriendList []*FriendInfo
	GroupList  []*GroupInfo
	Online     bool
	NetLooping bool

	SequenceId              int32
	OutGoingPacketSessionId []byte
	RandomKey               []byte
	Conn                    net.Conn
	ConnectTime             time.Time

	handlers        HandlerMap
	servers         []*net.TCPAddr
	currServerIndex int
	retryTimes      int
	version         *versionInfo

	syncCookie       []byte
	pubAccountCookie []byte
	msgCtrlBuf       []byte
	ksid             []byte
	t104             []byte
	t174             []byte
	t402             []byte // only for sms
	t150             []byte
	t149             []byte
	t528             []byte
	t530             []byte
	rollbackSig      []byte
	timeDiff         int64
	sigInfo          *loginSigInfo
	pwdFlag          bool

	lastMessageSeq int32
	//lastMessageSeqTmp      sync.Map
	msgSvcCache            *utils.Cache
	transCache             *utils.Cache
	lastLostMsg            string
	groupSysMsgCache       *GroupSystemMessages
	groupMsgBuilders       sync.Map
	onlinePushCache        *utils.Cache
	requestPacketRequestId int32
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

var decoders = map[string]func(*QQClient, uint16, []byte) (interface{}, error){
	"wtlogin.login":                                      decodeLoginResponse,
	"wtlogin.exchange_emp":                               decodeExchangeEmpResponse,
	"StatSvc.register":                                   decodeClientRegisterResponse,
	"StatSvc.ReqMSFOffline":                              decodeMSFOfflinePacket,
	"StatSvc.GetDevLoginInfo":                            decodeDevListResponse,
	"MessageSvc.PushNotify":                              decodeSvcNotify,
	"OnlinePush.ReqPush":                                 decodeOnlinePushReqPacket,
	"OnlinePush.PbPushTransMsg":                          decodeOnlinePushTransPacket,
	"ConfigPushSvc.PushReq":                              decodePushReqPacket,
	"MessageSvc.PbGetMsg":                                decodeMessageSvcPacket,
	"MessageSvc.PushForceOffline":                        decodeForceOfflinePacket,
	"PbMessageSvc.PbMsgWithDraw":                         decodeMsgWithDrawResponse,
	"friendlist.getFriendGroupList":                      decodeFriendGroupListResponse,
	"friendlist.GetTroopListReqV2":                       decodeGroupListResponse,
	"friendlist.GetTroopMemberListReq":                   decodeGroupMemberListResponse,
	"group_member_card.get_group_member_card_info":       decodeGroupMemberInfoResponse,
	"ImgStore.GroupPicUp":                                decodeGroupImageStoreResponse,
	"PttStore.GroupPttUp":                                decodeGroupPttStoreResponse,
	"LongConn.OffPicUp":                                  decodeOffPicUpResponse,
	"ProfileService.Pb.ReqSystemMsgNew.Group":            decodeSystemMsgGroupPacket,
	"ProfileService.Pb.ReqSystemMsgNew.Friend":           decodeSystemMsgFriendPacket,
	"MultiMsg.ApplyUp":                                   decodeMultiApplyUpResponse,
	"MultiMsg.ApplyDown":                                 decodeMultiApplyDownResponse,
	"OidbSvc.0xe07_0":                                    decodeImageOcrResponse,
	"OidbSvc.0xd79":                                      decodeWordSegmentation,
	"OidbSvc.0x990":                                      decodeTranslateResponse,
	"SummaryCard.ReqSummaryCard":                         decodeSummaryCardResponse,
	"PttCenterSvr.ShortVideoDownReq":                     decodePttShortVideoDownResponse,
	"LightAppSvc.mini_app_info.GetAppInfoById":           decodeAppInfoResponse,
	"PttCenterSvr.pb_pttCenter_CMD_REQ_APPLY_UPLOAD-500": decodePrivatePttStoreResponse,
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// NewClient create new qq client
func NewClient(uin int64, password string) *QQClient {
	return NewClientMd5(uin, md5.Sum([]byte(password)))
}

func NewClientMd5(uin int64, passwordMd5 [16]byte) *QQClient {
	cli := &QQClient{
		Uin:                     uin,
		PasswordMd5:             passwordMd5,
		SequenceId:              0x3635,
		RandomKey:               make([]byte, 16),
		OutGoingPacketSessionId: []byte{0x02, 0xB0, 0x5B, 0x8B},
		sigInfo:                 &loginSigInfo{},
		requestPacketRequestId:  1921334513,
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
	rand.Read(cli.RandomKey)
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
	go c.netLoop()
	rsp, err := c.sendAndWait(c.buildLoginPacket())
	if err != nil {
		c.Disconnect()
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		c.init()
	}
	return &l, nil
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
		c.init()
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
		c.init()
	}
	return &l, nil
}

func (c *QQClient) init() {
	c.Online = true
	_ = c.registerClient()
	c.groupSysMsgCache, _ = c.GetGroupSystemMessages()
	if !c.heartbeatEnabled {
		go c.doHeartbeat()
	}
	c.stat.once.Do(func() {
		c.OnGroupMessage(func(_ *QQClient, _ *message.GroupMessage) {
			c.stat.MessageReceived++
		})
		c.OnPrivateMessage(func(_ *QQClient, _ *message.PrivateMessage) {
			c.stat.MessageReceived++
		})
		c.OnTempMessage(func(_ *QQClient, _ *message.TempMessage) {
			c.stat.MessageReceived++
		})
		c.onGroupMessageReceipt("internal", func(_ *QQClient, _ *groupMessageReceiptEvent) {
			c.stat.MessageSent++
		})
	})
}

func (c *QQClient) RequestSMS() bool {
	rsp, err := c.sendAndWait(c.buildSMSRequestPacket())
	if err != nil {
		c.Error("request sms error: %v", err)
		return false
	}
	return rsp.(LoginResponse).Error == SMSNeededError
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

// GetFriendList request friend list
func (c *QQClient) GetFriendList() (*FriendListResponse, error) {
	var curFriendCount = 0
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

func (c *QQClient) GetShortVideoUrl(uuid, md5 []byte) string {
	i, err := c.sendAndWait(c.buildPttShortVideoDownReqPacket(uuid, md5))
	if err != nil {
		return ""
	}
	return i.(string)
}

func (c *QQClient) SendPrivateMessage(target int64, m *message.SendingMessage) *message.PrivateMessage {
	mr := int32(rand.Uint32())
	seq := c.nextFriendSeq()
	t := time.Now().Unix()
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	msgLen := message.EstimateLength(m.Elements, 703)
	if msgLen > 5000 || imgCount > 50 {
		return nil
	}
	if msgLen > 300 || imgCount > 2 {
		div := int32(rand.Uint32())
		fragmented := m.ToFragmented()
		for i, elems := range fragmented {
			_, pkt := c.buildFriendSendingPacket(target, c.nextFriendSeq(), mr, int32(len(fragmented)), int32(i), div, t, elems)
			_ = c.send(pkt)
		}
	} else {
		_, pkt := c.buildFriendSendingPacket(target, seq, mr, 1, 0, 0, t, m.Elements)
		_ = c.send(pkt)
	}
	c.stat.MessageSent++
	return &message.PrivateMessage{
		Id:         seq,
		InternalId: mr,
		Target:     target,
		Time:       int32(t),
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Elements: m.Elements,
	}
}

func (c *QQClient) SendTempMessage(groupCode, target int64, m *message.SendingMessage) *message.TempMessage {
	group := c.FindGroup(groupCode)
	if group == nil {
		return nil
	}
	if c.FindFriend(target) != nil {
		pm := c.SendPrivateMessage(target, m)
		return &message.TempMessage{
			Id:        pm.Id,
			GroupCode: group.Code,
			GroupName: group.Name,
			Sender:    pm.Sender,
			Elements:  m.Elements,
		}
	}
	mr := int32(rand.Uint32())
	seq := c.nextFriendSeq()
	t := time.Now().Unix()
	_, pkt := c.buildTempSendingPacket(group.Uin, target, seq, mr, t, m)
	_ = c.send(pkt)
	c.stat.MessageSent++
	return &message.TempMessage{
		Id:        seq,
		GroupCode: group.Code,
		GroupName: group.Name,
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Elements: m.Elements,
	}
}

func (c *QQClient) GetForwardMessage(resId string) *message.ForwardMessage {
	i, err := c.sendAndWait(c.buildMultiApplyDownPacket(resId))
	if err != nil {
		return nil
	}
	multiMsg := i.(*msg.PbMultiMsgTransmit)
	ret := &message.ForwardMessage{}
	for _, m := range multiMsg.Msg {
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

func (c *QQClient) sendGroupPoke(groupCode, target int64) {
	_, _ = c.sendAndWait(c.buildGroupPokePacket(groupCode, target))
}

func (c *QQClient) SendFriendPoke(target int64) {
	_, _ = c.sendAndWait(c.buildFriendPokePacket(target))
}

func (c *QQClient) UploadGroupImage(groupCode int64, img []byte) (*message.GroupImageElement, error) {
	h := md5.Sum(img)
	seq, pkt := c.buildGroupImageStorePacket(groupCode, h[:], int32(len(img)))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		goto ok
	}
	for i, ip := range rsp.UploadIp {
		err := c.highwayUpload(uint32(ip), int(rsp.UploadPort[i]), rsp.UploadKey, img, 2)
		if err != nil {
			continue
		}
		goto ok
	}
	return nil, errors.New("upload failed")
ok:
	i, _, _ := image.DecodeConfig(bytes.NewReader(img))
	var imageType int32 = 1000
	if bytes.HasPrefix(img, []byte{0x47, 0x49, 0x46, 0x38}) {
		imageType = 2000
	}
	return message.NewGroupImage(binary.CalculateImageResourceId(h[:]), h[:], rsp.FileId, int32(len(img)), int32(i.Width), int32(i.Height), imageType), nil
}

func (c *QQClient) UploadPrivateImage(target int64, img []byte) (*message.FriendImageElement, error) {
	return c.uploadPrivateImage(target, img, 0)
}

func (c *QQClient) uploadPrivateImage(target int64, img []byte, count int) (*message.FriendImageElement, error) {
	count++
	h := md5.Sum(img)
	e, err := c.QueryFriendImage(target, h[:], int32(len(img)))
	if errors.Is(err, ErrNotExists) {
		// use group highway upload and query again for image id.
		if _, err = c.UploadGroupImage(target, img); err != nil {
			return nil, err
		}
		if count >= 5 {
			return e, nil
		}
		return c.uploadPrivateImage(target, img, count)
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (c *QQClient) ImageOcr(img interface{}) (*OcrResponse, error) {
	switch e := img.(type) {
	case *message.GroupImageElement:
		rsp, err := c.sendAndWait(c.buildImageOcrRequestPacket(e.Url, strings.ToUpper(hex.EncodeToString(e.Md5)), e.Size, e.Width, e.Height))
		if err != nil {
			return nil, err
		}
		return rsp.(*OcrResponse), nil
	case *message.ImageElement:
		rsp, err := c.sendAndWait(c.buildImageOcrRequestPacket(e.Url, strings.ToUpper(hex.EncodeToString(e.Md5)), e.Size, e.Width, e.Height))
		if err != nil {
			return nil, err
		}
		return rsp.(*OcrResponse), nil
	}
	return nil, errors.New("image error")
}

func (c *QQClient) QueryGroupImage(groupCode int64, hash []byte, size int32) (*message.GroupImageElement, error) {
	r, err := c.sendAndWait(c.buildGroupImageStorePacket(groupCode, hash, size))
	if err != nil {
		return nil, err
	}
	rsp := r.(imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		return message.NewGroupImage(binary.CalculateImageResourceId(hash), hash, rsp.FileId, size, rsp.Width, rsp.Height, 1000), nil
	}
	return nil, errors.New("image does not exist")
}

func (c *QQClient) QueryFriendImage(target int64, hash []byte, size int32) (*message.FriendImageElement, error) {
	i, err := c.sendAndWait(c.buildOffPicUpPacket(target, hash, size))
	if err != nil {
		return nil, err
	}
	rsp := i.(imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if !rsp.IsExists {
		return &message.FriendImageElement{
			ImageId: rsp.ResourceId,
			Md5:     hash,
		}, errors.WithStack(ErrNotExists)
	}
	return &message.FriendImageElement{
		ImageId: rsp.ResourceId,
		Md5:     hash,
	}, nil
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
			return list, nil
		}
	}
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
		_, pkt := c.buildSystemMsgGroupActionPacket(req.RequestId, req.RequesterUin, req.GroupCode, false, accept, block, reason)
		_ = c.send(pkt)
	case *GroupInvitedRequest:
		_, pkt := c.buildSystemMsgGroupActionPacket(req.RequestId, req.InvitorUin, req.GroupCode, true, accept, block, reason)
		_ = c.send(pkt)
	}
}

func (c *QQClient) SolveFriendRequest(req *NewFriendRequest, accept bool) {
	_, pkt := c.buildSystemMsgFriendActionPacket(req.RequestId, req.RequesterUin, accept)
	_ = c.send(pkt)
}

func (c *QQClient) getSKey() string {
	if c.sigInfo.sKeyExpiredTime < time.Now().Unix() {
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

func (c *QQClient) getMemberInfo(groupCode, memberUin int64) (*GroupMemberInfo, error) {
	info, err := c.sendAndWait(c.buildGroupMemberInfoRequestPacket(groupCode, memberUin))
	if err != nil {
		return nil, err
	}
	return info.(*GroupMemberInfo), nil
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
		for i, m := range info.Members {
			if m.Uin == uin {
				info.Members = append(info.Members[:i], info.Members[i+1:]...)
				break
			}
		}
	})
}

func (c *QQClient) connect() error {
	c.Info("connect to server: %v", c.servers[c.currServerIndex].String())
	conn, err := net.DialTCP("tcp", nil, c.servers[c.currServerIndex])
	c.currServerIndex++
	if c.currServerIndex == len(c.servers) {
		c.currServerIndex = 0
	}
	if err != nil || conn == nil {
		c.retryTimes++
		if c.retryTimes > len(c.servers) {
			return errors.New("All servers are unreachable")
		}
		c.Error("connect server error: %v", err)
		if err = c.connect(); err != nil {
			return err
		}
		return nil
	}
	c.retryTimes = 0
	c.ConnectTime = time.Now()
	c.Conn = conn
	return nil
}

func (c *QQClient) Disconnect() {
	c.NetLooping = false
	c.Online = false
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
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
	return err
}

func (c *QQClient) nextSeq() uint16 {
	return uint16(atomic.AddInt32(&c.SequenceId, 1) & 0x7FFF)
}

func (c *QQClient) nextPacketSeq() int32 {
	return atomic.AddInt32(&c.requestPacketRequestId, 2)
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
	_, err := c.Conn.Write(pkt)
	if err != nil {
		c.stat.PacketLost++
	} else {
		c.stat.PacketSent++
	}
	return errors.Wrap(err, "Packet failed to send")
}

func (c *QQClient) sendAndWait(seq uint16, pkt []byte) (interface{}, error) {
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
	c.handlers.Store(seq, func(i interface{}, err error) {
		ch <- T{
			Response: i,
			Error:    err,
		}
	})

	retry := 0
	for true {
		select {
		case rsp := <-ch:
			return rsp.Response, rsp.Error
		case <-time.After(time.Second * 30):
			retry++
			if retry < 2 {
				_ = c.send(pkt)
				continue
			}
			c.handlers.Delete(seq)
			//c.Error("packet timed out, seq: %v", seq)
			//println("Packet Timed out")
			return nil, errors.New("Packet timed out")
		}
	}
	return nil, nil
}

func (c *QQClient) netLoop() {
	if c.NetLooping {
		return
	}
	c.NetLooping = true
	reader := binary.NewNetworkReader(c.Conn)
	retry := 0
	errCount := 0
	for c.NetLooping {
		l, err := reader.ReadInt32()
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
			c.Error("connection dropped by server: %v", err)
			c.stat.DisconnectTimes++
			err = c.connect()
			if err != nil {
				c.Error("connect server error: %v", err)
				break
			}
			reader = binary.NewNetworkReader(c.Conn)
			if e := c.registerClient(); e != nil && e.Error() != "Packet timed out" { // 掉线在心跳已经有判断了, 只需要处理返回值
				c.Disconnect()
				c.lastLostMsg = "register client failed: " + e.Error()
				c.Error("reconnect failed: " + e.Error())
				break
			}
		}
		if l <= 0 {
			retry++
			time.Sleep(time.Second * 3)
			if retry > 10 {
				break
			}
			continue
		}
		data, err := reader.ReadBytes(int(l) - 4)
		pkt, err := packets.ParseIncomingPacket(data, c.sigInfo.d2Key)
		if err != nil {
			c.Error("parse incoming packet error: %v", err)
			if errors.Is(err, packets.ErrSessionExpired) || errors.Is(err, packets.ErrPacketDropped) {
				break
			}
			errCount++
			if errCount > 2 {
				break
			}
			//log.Println("parse incoming packet error: " + err.Error())
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
		retry = 0
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
				rsp, err := decoder(c, pkt.SequenceId, payload)
				if err != nil {
					c.Debug("decode pkt %v error: %+v", pkt.CommandName, err)
				}
				if f, ok := c.handlers.LoadAndDelete(pkt.SequenceId); ok {
					f(rsp, err)
				}
			} else if f, ok := c.handlers.LoadAndDelete(pkt.SequenceId); ok {
				// does not need decoder
				f(nil, nil)
			} else {
				c.Debug("\nUnhandled Command: %s\nSeq: %d\nThis message can be ignored.", pkt.CommandName, pkt.SequenceId)
			}
		}()
	}
	c.NetLooping = false
	c.Online = false
	_ = c.Conn.Close()
	if c.lastLostMsg == "" {
		c.lastLostMsg = "Connection lost."
	}
	c.stat.LostTimes++
	c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: c.lastLostMsg})
}

func (c *QQClient) doHeartbeat() {
	c.heartbeatEnabled = true
	times := 0
	for c.Online {
		seq := c.nextSeq()
		sso := packets.BuildSsoPacket(seq, c.version.AppId, "Heartbeat.Alive", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, []byte{}, c.ksid)
		packet := packets.BuildLoginPacket(c.Uin, 0, []byte{}, sso, []byte{})
		_, err := c.sendAndWait(seq, packet)
		if err != nil {
			c.lastLostMsg = "Heartbeat failed: " + err.Error()
			c.Disconnect()
			break
		}
		times++
		if times >= 7 {
			_ = c.registerClient()
			times = 0
		}
		time.Sleep(time.Second * 30)
	}
	c.heartbeatEnabled = false
}
