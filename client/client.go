package client

import (
	"crypto/md5"
	"errors"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type QQClient struct {
	Uin         int64
	PasswordMd5 [16]byte

	Nickname   string
	Age        uint16
	Gender     uint16
	FriendList []*FriendInfo
	GroupList  []*GroupInfo

	SequenceId              uint16
	OutGoingPacketSessionId []byte
	RandomKey               []byte
	Conn                    net.Conn

	decoders map[string]func(*QQClient, uint16, []byte) (interface{}, error)
	handlers map[uint16]func(interface{}, error)

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
	running          bool

	lastMessageSeq         int32
	lastMessageSeqTmp      sync.Map
	groupMsgBuilders       sync.Map
	onlinePushCache        []int16 // reset on reconnect
	requestPacketRequestId int32
	messageSeq             int32
	groupDataTransSeq      int32
	eventHandlers          *eventHandlers

	groupListLock *sync.Mutex
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
			"MessageSvc.PushNotify":                    decodeSvcNotify,
			"OnlinePush.PbPushGroupMsg":                decodeGroupMessagePacket,
			"OnlinePush.ReqPush":                       decodeOnlinePushReqPacket,
			"OnlinePush.PbPushTransMsg":                decodeOnlinePushTransPacket,
			"ConfigPushSvc.PushReq":                    decodePushReqPacket,
			"MessageSvc.PbGetMsg":                      decodeMessageSvcPacket,
			"friendlist.getFriendGroupList":            decodeFriendGroupListResponse,
			"friendlist.GetTroopListReqV2":             decodeGroupListResponse,
			"friendlist.GetTroopMemberListReq":         decodeGroupMemberListResponse,
			"ImgStore.GroupPicUp":                      decodeGroupImageStoreResponse,
			"ProfileService.Pb.ReqSystemMsgNew.Group":  decodeSystemMsgGroupPacket,
			"ProfileService.Pb.ReqSystemMsgNew.Friend": decodeSystemMsgFriendPacket,
			//"MultiMsg.ApplyDown":                       decodeMultiMsgDownPacket,
		},
		handlers:               map[uint16]func(interface{}, error){},
		sigInfo:                &loginSigInfo{},
		requestPacketRequestId: 1921334513,
		messageSeq:             22911,
		ksid:                   []byte("|454001228437590|A8.2.7.27f6ea96"),
		eventHandlers:          &eventHandlers{},
		groupListLock:          new(sync.Mutex),
	}
	rand.Read(cli.RandomKey)
	return cli
}

// Login send login request
func (c *QQClient) Login() (*LoginResponse, error) {
	if c.running {
		return nil, ErrAlreadyRunning
	}
	err := c.connect()
	if err != nil {
		return nil, err
	}
	c.running = true
	go c.loop()
	seq, packet := c.buildLoginPacket()
	rsp, err := c.sendAndWait(seq, packet)
	if err != nil {
		return nil, err
	}
	l := rsp.(LoginResponse)
	if l.Success {
		c.registerClient()
		go c.heartbeat()
	}
	return &l, nil
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
		go c.heartbeat()
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
		if int32(curFriendCount) >= r.TotalCount {
			break
		}
	}
	return r, nil
}

func (c *QQClient) SendGroupMessage(groupCode int64, m *message.SendingMessage) int32 {
	eid := utils.RandomString(6)
	mr := int32(rand.Uint32())
	ch := make(chan int32)
	c.onGroupMessageReceipt(eid, func(c *QQClient, e *groupMessageReceiptEvent) {
		if e.Rand == mr {
			ch <- e.Seq
			c.onGroupMessageReceipt(eid)
		}
	})
	_, pkt := c.buildGroupSendingPacket(groupCode, mr, m)
	_ = c.send(pkt)
	var mid int32
	select {
	case mid = <-ch:
	case <-time.After(time.Second * 5):
		c.onGroupMessageReceipt(eid)
		return -1
	}
	return mid
}

func (c *QQClient) UploadGroupImage(groupCode int64, img []byte) (*message.GroupImageElement, error) {
	h := md5.Sum(img)
	seq, pkt := c.buildGroupImageStorePacket(groupCode, h[:], int32(len(img)))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(groupImageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		return message.NewGroupImage(binary.CalculateImageResourceId(h[:]), h[:]), nil
	}
	for i, ip := range rsp.UploadIp {
		updServer := binary.UInt32ToIPV4Address(uint32(ip))
		conn, err := net.DialTimeout("tcp", updServer+":"+strconv.FormatInt(int64(rsp.UploadPort[i]), 10), time.Second*5)
		if err != nil {
			continue
		}
		if conn.SetDeadline(time.Now().Add(time.Second*10)) != nil {
			_ = conn.Close()
			continue
		}
		pkt := c.buildImageUploadPacket(img, rsp.UploadKey, 2, h)
		for _, p := range pkt {
			_, err = conn.Write(p)
		}
		if err != nil {
			continue
		}
		r := binary.NewNetworkReader(conn)
		_, err = r.ReadByte()
		if err != nil {
			continue
		}
		hl, _ := r.ReadInt32()
		_, _ = r.ReadBytes(4)
		payload, _ := r.ReadBytes(int(hl))
		_ = conn.Close()
		rsp := pb.RspDataHighwayHead{}
		if proto.Unmarshal(payload, &rsp) != nil {
			continue
		}
		if rsp.ErrorCode != 0 {
			return nil, errors.New("upload failed")
		}
		return message.NewGroupImage(binary.CalculateImageResourceId(h[:]), h[:]), nil
	}
	return nil, errors.New("upload failed")
}

func (c *QQClient) QueryGroupImage(groupCode int64, hash []byte, size int32) (*message.GroupImageElement, error) {
	r, err := c.sendAndWait(c.buildGroupImageStorePacket(groupCode, hash, size))
	if err != nil {
		return nil, err
	}
	rsp := r.(groupImageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		return message.NewGroupImage(binary.CalculateImageResourceId(hash), hash), nil
	}
	return nil, errors.New("image not exists")
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
	rsp, err := c.sendAndWait(c.buildGroupListRequestPacket())
	if err != nil {
		return nil, err
	}
	r := rsp.([]*GroupInfo)
	for _, group := range r {
		m, err := c.GetGroupMembers(group)
		if err != nil {
			continue
		}
		group.Members = m
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
			if m.Uin == group.OwnerUin {
				m.Permission = Owner
				break
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

func (c *QQClient) SolveGroupJoinRequest(i interface{}, accept bool) {
	switch req := i.(type) {
	case *UserJoinGroupRequest:
		_, pkt := c.buildSystemMsgGroupActionPacket(req.RequestId, req.RequesterUin, req.GroupCode, false, accept, false)
		_ = c.send(pkt)
	case *GroupInvitedRequest:
		_, pkt := c.buildSystemMsgGroupActionPacket(req.RequestId, req.InvitorUin, req.GroupCode, true, accept, false)
		_ = c.send(pkt)
	}
}

func (c *QQClient) SolveFriendRequest(req *NewFriendRequest, accept bool) {
	_, pkt := c.buildSystemMsgFriendActionPacket(req.RequestId, req.RequesterUin, accept)
	_ = c.send(pkt)
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

func (c *QQClient) connect() error {
	conn, err := net.Dial("tcp", "125.94.60.146:80") //TODO: more servers
	if err != nil {
		return err
	}
	c.Conn = conn
	c.onlinePushCache = []int16{}
	return nil
}

func (c *QQClient) registerClient() {
	_, packet := c.buildClientRegisterPacket()
	_ = c.send(packet)
}

func (c *QQClient) nextSeq() uint16 {
	c.SequenceId++
	c.SequenceId &= 0x7FFF
	if c.SequenceId == 0 {
		c.SequenceId++
	}
	return c.SequenceId
}

func (c *QQClient) nextPacketSeq() int32 {
	s := atomic.LoadInt32(&c.requestPacketRequestId)
	atomic.AddInt32(&c.requestPacketRequestId, 2)
	return s
}

func (c *QQClient) nextMessageSeq() int32 {
	s := atomic.LoadInt32(&c.messageSeq)
	atomic.AddInt32(&c.messageSeq, 2)
	return s
}

func (c *QQClient) nextGroupDataTransSeq() int32 {
	s := atomic.LoadInt32(&c.groupDataTransSeq)
	atomic.AddInt32(&c.groupDataTransSeq, 2)
	return s
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
	c.handlers[seq] = func(i interface{}, err error) {
		ch <- T{
			Response: i,
			Error:    err,
		}
	}
	rsp := <-ch
	return rsp.Response, rsp.Error
}

func (c *QQClient) loop() {
	reader := binary.NewNetworkReader(c.Conn)
	for c.running {
		l, err := reader.ReadInt32()
		if err == io.EOF || err == io.ErrClosedPipe {
			err = c.connect()
			if err != nil {
				c.running = false
				return
			}
			reader = binary.NewNetworkReader(c.Conn)
			c.registerClient()
		}
		if l <= 0 {
			continue
		}
		data, err := reader.ReadBytes(int(l) - 4)
		pkt, err := packets.ParseIncomingPacket(data, c.sigInfo.d2Key)
		if err != nil {
			log.Println("parse incoming packet error: " + err.Error())
			continue
		}
		payload := pkt.Payload
		if pkt.Flag2 == 2 {
			payload, err = pkt.DecryptPayload(c.RandomKey)
			if err != nil {
				continue
			}
		}
		//fmt.Println(pkt.CommandName)
		go func() {
			decoder, ok := c.decoders[pkt.CommandName]
			if !ok {
				if f, ok := c.handlers[pkt.SequenceId]; ok {
					delete(c.handlers, pkt.SequenceId)
					f(nil, nil)
				}
				return
			}
			rsp, err := decoder(c, pkt.SequenceId, payload)
			if err != nil {
				log.Println("decode", pkt.CommandName, "error:", err)
			}
			if f, ok := c.handlers[pkt.SequenceId]; ok {
				delete(c.handlers, pkt.SequenceId)
				f(rsp, err)
			}
		}()
	}
}

func (c *QQClient) heartbeat() {
	for c.running {
		time.Sleep(time.Second * 30)
		seq := c.nextSeq()
		sso := packets.BuildSsoPacket(seq, "Heartbeat.Alive", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, []byte{}, c.ksid)
		packet := packets.BuildLoginPacket(c.Uin, 0, []byte{}, sso, []byte{})
		_, _ = c.sendAndWait(seq, packet)
	}
}
