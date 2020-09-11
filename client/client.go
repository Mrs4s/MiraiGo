package client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/longmsg"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/multimsg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
)

type QQClient struct {
	Uin          int64
	PasswordMd5  [16]byte
	CustomServer *net.TCPAddr

	Nickname   string
	Age        uint16
	Gender     uint16
	FriendList []*FriendInfo
	GroupList  []*GroupInfo
	Online     bool

	SequenceId              int32
	OutGoingPacketSessionId []byte
	RandomKey               []byte
	Conn                    net.Conn
	ConnectTime             time.Time

	decoders map[string]func(*QQClient, uint16, []byte) (interface{}, error)
	handlers sync.Map
	server   *net.TCPAddr

	syncCookie       []byte
	pubAccountCookie []byte
	msgCtrlBuf       []byte
	ksid             []byte
	t104             []byte
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
	groupMsgBuilders       sync.Map
	onlinePushCache        []int16 // reset on reconnect
	requestPacketRequestId int32
	groupSeq               int32
	friendSeq              int32
	groupDataTransSeq      int32
	highwayApplyUpSeq      int32
	eventHandlers          *eventHandlers

	groupListLock sync.Mutex
	msgSvcLock    sync.Mutex
}

type loginSigInfo struct {
	loginBitmap uint64
	tgt         []byte
	tgtKey      []byte

	userStKey          []byte
	userStWebSig       []byte
	sKey               []byte
	d2                 []byte
	d2Key              []byte
	wtSessionTicketKey []byte
	deviceToken        []byte

	psKeyMap    map[string][]byte
	pt4TokenMap map[string][]byte
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
		decoders: map[string]func(*QQClient, uint16, []byte) (interface{}, error){
			"wtlogin.login":                            decodeLoginResponse,
			"StatSvc.register":                         decodeClientRegisterResponse,
			"StatSvc.ReqMSFOffline":                    decodeMSFOfflinePacket,
			"MessageSvc.PushNotify":                    decodeSvcNotify,
			"OnlinePush.PbPushGroupMsg":                decodeGroupMessagePacket,
			"OnlinePush.ReqPush":                       decodeOnlinePushReqPacket,
			"OnlinePush.PbPushTransMsg":                decodeOnlinePushTransPacket,
			"ConfigPushSvc.PushReq":                    decodePushReqPacket,
			"MessageSvc.PbGetMsg":                      decodeMessageSvcPacket,
			"MessageSvc.PbSendMsg":                     decodeMsgSendResponse,
			"MessageSvc.PushForceOffline":              decodeForceOfflinePacket,
			"friendlist.getFriendGroupList":            decodeFriendGroupListResponse,
			"friendlist.GetTroopListReqV2":             decodeGroupListResponse,
			"friendlist.GetTroopMemberListReq":         decodeGroupMemberListResponse,
			"ImgStore.GroupPicUp":                      decodeGroupImageStoreResponse,
			"PttStore.GroupPttUp":                      decodeGroupPttStoreResponse,
			"LongConn.OffPicUp":                        decodeOffPicUpResponse,
			"ProfileService.Pb.ReqSystemMsgNew.Group":  decodeSystemMsgGroupPacket,
			"ProfileService.Pb.ReqSystemMsgNew.Friend": decodeSystemMsgFriendPacket,
			"MultiMsg.ApplyUp":                         decodeMultiApplyUpResponse,
			"MultiMsg.ApplyDown":                       decodeMultiApplyDownResponse,
			"OidbSvc.0x6d6_2":                          decodeOIDB6d6Response,
			"PttCenterSvr.ShortVideoDownReq":           decodePttShortVideoDownResponse,
		},
		sigInfo:                &loginSigInfo{},
		requestPacketRequestId: 1921334513,
		groupSeq:               int32(rand.Intn(20000)),
		friendSeq:              22911,
		highwayApplyUpSeq:      77918,
		ksid:                   []byte("|454001228437590|A8.2.7.27f6ea96"),
		eventHandlers:          &eventHandlers{},
		msgSvcCache:            utils.NewCache(time.Second * 15),
		transCache:             utils.NewCache(time.Second * 15),
	}
	rand.Read(cli.RandomKey)
	return cli
}

// Login send login request
func (c *QQClient) Login() (*LoginResponse, error) {
	if c.Online {
		return nil, ErrAlreadyOnline
	}
	c.server = nil
	err := c.connect()
	if err != nil {
		return nil, err
	}
	c.Online = true
	go c.netLoop()
	seq, packet := c.buildLoginPacket()
	rsp, err := c.sendAndWait(seq, packet)
	if err != nil {
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		c.lastLostMsg = ""
		c.registerClient()
		c.startHeartbeat()
	}
	return &l, nil
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

// SubmitCaptcha send captcha to server
func (c *QQClient) SubmitCaptcha(result string, sign []byte) (*LoginResponse, error) {
	seq, packet := c.buildCaptchaPacket(result, sign)
	rsp, err := c.sendAndWait(seq, packet)
	if err != nil {
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		c.registerClient()
		c.startHeartbeat()
	}
	return &l, nil
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

func (c *QQClient) GetGroupFileUrl(groupCode int64, fileId string, busId int32) string {
	i, err := c.sendAndWait(c.buildGroupFileDownloadReqPacket(groupCode, fileId, busId))
	if err != nil {
		return ""
	}
	url := i.(string)
	url += "?fname=" + hex.EncodeToString([]byte(fileId))
	return url
}

func (c *QQClient) SendGroupMessage(groupCode int64, m *message.SendingMessage, f ...bool) *message.GroupMessage {
	useFram := false
	if len(f) > 0 {
		useFram = f[0]
	}
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	if useFram {
		if m.Any(func(e message.IMessageElement) bool { return e.Type() == message.Reply }) {
			useFram = false
		}
	}
	msgLen := message.EstimateLength(m.Elements, 703)
	if msgLen > 5000 || imgCount > 50 {
		return nil
	}
	if (msgLen > 300 || imgCount > 2) && !useFram {
		ret := c.sendGroupLongOrForwardMessage(groupCode, true, &message.ForwardMessage{Nodes: []*message.ForwardNode{
			{
				SenderId:   c.Uin,
				SenderName: c.Nickname,
				Time:       int32(time.Now().Unix()),
				Message:    m.Elements,
			},
		}})
		if ret != nil && ret.Id == -1 {
			c.Error("long message send error. trying fragmented sending...")
			return c.SendGroupMessage(groupCode, m, true)
		}
		return ret
	}
	return c.sendGroupMessage(groupCode, false, m)
}

func (c *QQClient) sendGroupMessage(groupCode int64, forward bool, m *message.SendingMessage) *message.GroupMessage {
	eid := utils.RandomString(6)
	mr := int32(rand.Uint32())
	ch := make(chan int32)
	c.onGroupMessageReceipt(eid, func(c *QQClient, e *groupMessageReceiptEvent) {
		if e.Rand == mr {
			ch <- e.Seq
		}
	})
	defer c.onGroupMessageReceipt(eid)
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	msgLen := message.EstimateLength(m.Elements, 703)
	if (msgLen > 300 || imgCount > 2) && !forward && !m.Any(func(e message.IMessageElement) bool {
		_, ok := e.(*message.GroupVoiceElement)
		_, ok2 := e.(*message.ServiceElement)
		return ok || ok2
	}) {
		div := int32(rand.Uint32())
		fragmented := m.ToFragmented()
		for i, elems := range fragmented {
			_, pkt := c.buildGroupSendingPacket(groupCode, mr, int32(len(fragmented)), int32(i), div, forward, elems)
			_ = c.send(pkt)
		}
	} else {
		_, pkt := c.buildGroupSendingPacket(groupCode, mr, 1, 0, 0, forward, m.Elements)
		_ = c.send(pkt)
	}
	var mid int32
	ret := &message.GroupMessage{
		Id:         -1,
		InternalId: mr,
		GroupCode:  groupCode,
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Time:     int32(time.Now().Unix()),
		Elements: m.Elements,
	}
	select {
	case mid = <-ch:
	case <-time.After(time.Second * 5):
		return ret
	}
	ret.Id = mid
	return ret
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
			SenderId: m.Head.FromUin,
			SenderName: func() string {
				if m.Head.MsgType == 82 && m.Head.GroupInfo != nil {
					return m.Head.GroupInfo.GroupCard
				}
				return m.Head.FromNick
			}(),
			Time:    m.Head.MsgTime,
			Message: message.ParseMessageElems(m.Body.RichText.Elems),
		})
	}
	return ret
}

func (c *QQClient) SendGroupForwardMessage(groupCode int64, m *message.ForwardMessage) *message.GroupMessage {
	return c.sendGroupLongOrForwardMessage(groupCode, false, m)
}

func (c *QQClient) sendGroupLongOrForwardMessage(groupCode int64, isLong bool, m *message.ForwardMessage) *message.GroupMessage {
	if len(m.Nodes) >= 200 {
		return nil
	}
	ts := time.Now().Unix()
	seq := c.nextGroupSeq()
	data, hash := m.CalculateValidationData(seq, rand.Int31(), groupCode)
	i, err := c.sendAndWait(c.buildMultiApplyUpPacket(data, hash, func() int32 {
		if isLong {
			return 1
		} else {
			return 2
		}
	}(), utils.ToGroupUin(groupCode)))
	if err != nil {
		return nil
	}
	rsp := i.(*multimsg.MultiMsgApplyUpRsp)
	body, _ := proto.Marshal(&longmsg.LongReqBody{
		Subcmd:       1,
		TermType:     5,
		PlatformType: 9,
		MsgUpReq: []*longmsg.LongMsgUpReq{
			{
				MsgType:    3,
				DstUin:     utils.ToGroupUin(groupCode),
				MsgContent: data,
				StoreType:  2,
				MsgUkey:    rsp.MsgUkey,
			},
		},
	})
	for i, ip := range rsp.Uint32UpIp {
		err := c.highwayUploadImage(uint32(ip), int(rsp.Uint32UpPort[i]), rsp.MsgSig, body, 27)
		if err == nil {
			if !isLong {
				var pv string
				for i := 0; i < int(math.Min(4, float64(len(m.Nodes)))); i++ {
					pv += fmt.Sprintf(`<title size="26" color="#777777">%s: %s</title>`, m.Nodes[i].SenderName, message.ToReadableString(m.Nodes[i].Message))
				}
				return c.sendGroupMessage(groupCode, true, genForwardTemplate(rsp.MsgResid, pv, "群聊的聊天记录", "[聊天记录]", "聊天记录", fmt.Sprintf("查看 %d 条转发消息", len(m.Nodes)), ts))
			}
			bri := func() string {
				var r string
				for _, n := range m.Nodes {
					r += message.ToReadableString(n.Message)
					if len(r) >= 27 {
						break
					}
				}
				return r
			}()
			return c.sendGroupMessage(groupCode, false, genLongTemplate(rsp.MsgResid, bri, ts))
		}
	}
	return nil
}

func (c *QQClient) RecallGroupMessage(groupCode int64, msgId, msgInternalId int32) {
	_, pkt := c.buildGroupRecallPacket(groupCode, msgId, msgInternalId)
	_ = c.send(pkt)
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
		err := c.highwayUploadImage(uint32(ip), int(rsp.UploadPort[i]), rsp.UploadKey, img, 2)
		if err != nil {
			continue
		}
		goto ok
	}
	return nil, errors.New("upload failed")
ok:
	return message.NewGroupImage(binary.CalculateImageResourceId(h[:]), h[:], rsp.FileId), nil
}

func (c *QQClient) UploadPrivateImage(target int64, img []byte) (*message.FriendImageElement, error) {
	return c.uploadPrivateImage(target, img, 0)
}

func (c *QQClient) uploadPrivateImage(target int64, img []byte, count int) (*message.FriendImageElement, error) {
	count++
	h := md5.Sum(img)
	e, err := c.QueryFriendImage(target, h[:], int32(len(img)))
	if err != nil {
		// use group highway upload and query again for image id.
		if _, err = c.UploadGroupImage(target, img); err != nil {
			return nil, err
		}
		// safe
		if count >= 5 {
			return nil, errors.New("upload failed")
		}
		return c.uploadPrivateImage(target, img, count)
	}
	return e, nil
}

func (c *QQClient) UploadGroupPtt(groupCode int64, voice []byte) (*message.GroupVoiceElement, error) {
	h := md5.Sum(voice)
	seq, pkt := c.buildGroupPttStorePacket(groupCode, h[:], int32(len(voice)), 0, int32(len(voice)))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(pttUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		goto ok
	}
	for i, ip := range rsp.UploadIp {
		err := c.uploadGroupPtt(ip, rsp.UploadPort[i], rsp.UploadKey, rsp.FileKey, voice, h[:], 2)
		if err != nil {
			continue
		}
		goto ok
	}
	return nil, errors.New("upload failed")
ok:
	return &message.GroupVoiceElement{
		Ptt: &msg.Ptt{
			FileType:     4,
			SrcUin:       c.Uin,
			FileMd5:      h[:],
			FileName:     hex.EncodeToString(h[:]) + ".amr",
			FileSize:     int32(len(voice)),
			GroupFileKey: rsp.FileKey,
			BoolValid:    true,
			PbReserve:    []byte{8, 0, 40, 0, 56, 0},
		}}, nil
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
		return message.NewGroupImage(binary.CalculateImageResourceId(hash), hash, rsp.FileId), nil
	}
	return nil, errors.New("image not exists")
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
		return nil, errors.New("image not exists")
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

func (g *GroupInfo) SelfPermission() MemberPermission {
	return g.FindMember(g.client.Uin).Permission
}

func (g *GroupInfo) AdministratorOrOwner() bool {
	return g.SelfPermission() == Administrator || g.SelfPermission() == Owner
}

func (g *GroupInfo) FindMember(uin int64) *GroupMemberInfo {
	for _, m := range g.Members {
		f := m
		if f.Uin == uin {
			return f
		}
	}
	return nil
}

func (c *QQClient) getCookies() string {
	return fmt.Sprintf("uin=o%d; skey=%s;", c.Uin, c.sigInfo.sKey)
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

func (c *QQClient) kickGroupMember(groupCode, memberUin int64, msg string) {
	_, _ = c.sendAndWait(c.buildGroupKickPacket(groupCode, memberUin, msg))
}

func (g *GroupInfo) removeMember(uin int64) {
	if g.memLock == nil {
		g.memLock = new(sync.Mutex)
	}
	g.memLock.Lock()
	defer g.memLock.Unlock()
	for i, m := range g.Members {
		if m.Uin == uin {
			g.Members = append(g.Members[:i], g.Members[i+1:]...)
			break
		}
	}
}

var servers = []*net.TCPAddr{
	{IP: net.IP{42, 81, 169, 46}, Port: 8080},
	{IP: net.IP{42, 81, 172, 81}, Port: 80},
	{IP: net.IP{114, 221, 148, 59}, Port: 14000},
	{IP: net.IP{42, 81, 172, 147}, Port: 443},
	{IP: net.IP{125, 94, 60, 146}, Port: 80},
	{IP: net.IP{114, 221, 144, 215}, Port: 80},
	{IP: net.IP{42, 81, 172, 22}, Port: 80},
}

func (c *QQClient) connect() error {
	if c.server == nil {
		if c.CustomServer != nil {
			c.server = c.CustomServer
		} else {
			addrs, err := net.LookupIP("msfwifi.3g.qq.com")
			if err == nil && len(addrs) > 0 {
				c.server = &net.TCPAddr{
					IP:   addrs[rand.Intn(len(addrs))],
					Port: 8080,
				}
			} else {
				c.server = servers[rand.Intn(len(servers))]
			}
		}
	}
	c.Info("connect to server: %v", c.server.String())
	conn, err := net.DialTCP("tcp", nil, c.server)
	if err != nil {
		c.CustomServer = nil
		return err
	}
	c.ConnectTime = time.Now()
	c.Conn = conn
	c.onlinePushCache = []int16{}
	return nil
}

func (c *QQClient) registerClient() {
	_, packet := c.buildClientRegisterPacket()
	_ = c.send(packet)
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
	return err
}

func (c *QQClient) sendAndWait(seq uint16, pkt []byte) (interface{}, error) {
	type T struct {
		Response interface{}
		Error    error
	}
	_, err := c.Conn.Write(pkt)
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
				c.Error("packet %v timed out. retry.", seq)
				_ = c.send(pkt)
				continue
			}
			c.handlers.Delete(seq)
			c.Error("packet timed out, seq: %v", seq)
			//println("Packet Timed out")
			return nil, errors.New("timeout")
		}
	}
	return nil, nil
}

func (c *QQClient) netLoop() {
	reader := binary.NewNetworkReader(c.Conn)
	retry := 0
	for c.Online {
		l, err := reader.ReadInt32()
		if err == io.EOF || err == io.ErrClosedPipe {
			c.Error("connection dropped by server: %v", err)
			if c.ConnectTime.Sub(time.Now()) < time.Minute && c.CustomServer != nil {
				c.Error("custom server error.")
				c.CustomServer = nil
				c.server = nil
			}
			err = c.connect()
			if err != nil {
				break
			}
			reader = binary.NewNetworkReader(c.Conn)
			c.registerClient()
		}
		if l <= 0 {
			retry++
			time.Sleep(time.Second * 3)
			if retry > 10 {
				c.Online = false
			}
			continue
		}
		data, err := reader.ReadBytes(int(l) - 4)
		pkt, err := packets.ParseIncomingPacket(data, c.sigInfo.d2Key)
		if err != nil {
			c.Error("parse incoming packer error: %v", err)
			//log.Println("parse incoming packet error: " + err.Error())
			continue
		}
		payload := pkt.Payload
		if pkt.Flag2 == 2 {
			payload, err = pkt.DecryptPayload(c.RandomKey)
			if err != nil {
				continue
			}
		}
		retry = 0
		c.Debug("rev pkt: %v seq: %v", pkt.CommandName, pkt.SequenceId)
		go func() {
			defer func() {
				if pan := recover(); pan != nil {
					c.Error("panic on decoder: %v", pan)
					//fmt.Println("panic on decoder:", pan)
				}
			}()
			decoder, ok := c.decoders[pkt.CommandName]
			if !ok {
				if f, ok := c.handlers.Load(pkt.SequenceId); ok {
					c.handlers.Delete(pkt.SequenceId)
					f.(func(i interface{}, err error))(nil, nil)
				}
				return
			}
			rsp, err := decoder(c, pkt.SequenceId, payload)
			if err != nil {
				log.Println("decode", pkt.CommandName, "error:", err)
			}
			if f, ok := c.handlers.Load(pkt.SequenceId); ok {
				c.handlers.Delete(pkt.SequenceId)
				f.(func(i interface{}, err error))(rsp, err)
			}
		}()
	}
	c.Online = false
	_ = c.Conn.Close()
	if c.lastLostMsg == "" {
		c.lastLostMsg = "Connection lost."
	}
	c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: c.lastLostMsg})
}

func (c *QQClient) startHeartbeat() {
	time.AfterFunc(30*time.Second, c.doHeartbeat)
}

func (c *QQClient) doHeartbeat() {
	if c.Online {
		seq := c.nextSeq()
		sso := packets.BuildSsoPacket(seq, "Heartbeat.Alive", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, []byte{}, c.ksid)
		packet := packets.BuildLoginPacket(c.Uin, 0, []byte{}, sso, []byte{})
		_, _ = c.sendAndWait(seq, packet)
		time.AfterFunc(30*time.Second, c.doHeartbeat)
	}
}
