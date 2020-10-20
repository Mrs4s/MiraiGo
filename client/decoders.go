package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client/pb/notify"
	"github.com/Mrs4s/MiraiGo/client/pb/pttcenter"
	"github.com/Mrs4s/MiraiGo/client/pb/qweb"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x352"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/client/pb/structmsg"
	"github.com/golang/protobuf/proto"
)

var (
	groupJoinLock  sync.Mutex
	groupLeaveLock sync.Mutex
)

// wtlogin.login
func decodeLoginResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	reader := binary.NewReader(payload)
	reader.ReadUInt16() // sub command
	t := reader.ReadByte()
	reader.ReadUInt16()
	m := reader.ReadTlvMap(2)
	if t == 0 { // login success
		if t150, ok := m[0x150]; ok {
			c.t150 = t150
		}
		if t161, ok := m[0x161]; ok {
			c.decodeT161(t161)
		}
		c.decodeT119(m[0x119])
		return LoginResponse{
			Success: true,
		}, nil
	}
	if t == 2 {
		c.t104, _ = m[0x104]
		if m.Exists(0x192) { // slider, not supported yet
			return LoginResponse{
				Success:   false,
				VerifyUrl: string(m[0x192]),
				Error:     SliderNeededError,
			}, nil
		}
		if m.Exists(0x165) { // image
			imgData := binary.NewReader(m[0x105])
			signLen := imgData.ReadUInt16()
			imgData.ReadUInt16()
			sign := imgData.ReadBytes(int(signLen))
			return LoginResponse{
				Success:      false,
				Error:        NeedCaptcha,
				CaptchaImage: imgData.ReadAvailable(),
				CaptchaSign:  sign,
			}, nil
		} else {
			return LoginResponse{
				Success: false,
				Error:   UnknownLoginError,
			}, nil
		}
	} // need captcha

	if t == 40 {
		return LoginResponse{
			Success:      false,
			ErrorMessage: "账号被冻结",
			Error:        UnknownLoginError,
		}, nil
	}

	if t == 160 || t == 239 {
		if t174, ok := m[0x174]; ok { // 短信验证
			c.t104 = m[0x104]
			c.t174 = t174
			c.t402 = m[0x402]
			phone := func() string {
				r := binary.NewReader(m[0x178])
				return r.ReadStringLimit(int(r.ReadInt32()))
			}()
			if t204, ok := m[0x204]; ok { // 同时支持扫码验证 ?
				return LoginResponse{
					Success:      false,
					Error:        SMSOrVerifyNeededError,
					VerifyUrl:    string(t204),
					SMSPhone:     phone,
					ErrorMessage: string(m[0x17e]),
				}, nil
			}
			return LoginResponse{
				Success:      false,
				Error:        SMSNeededError,
				SMSPhone:     phone,
				ErrorMessage: string(m[0x17e]),
			}, nil
		}

		if _, ok := m[0x17b]; ok { // 二次验证
			c.t104 = m[0x104]
			return LoginResponse{
				Success: false,
				Error:   SMSNeededError,
			}, nil
		}

		if t204, ok := m[0x204]; ok { // 扫码验证
			return LoginResponse{
				Success:      false,
				Error:        UnsafeDeviceError,
				VerifyUrl:    string(t204),
				ErrorMessage: "",
			}, nil
		}

	}

	if t == 162 {
		return LoginResponse{
			Error: TooManySMSRequestError,
		}, nil
	}

	if t == 204 {
		c.t104 = m[0x104]
		return c.sendAndWait(c.buildDeviceLockLoginPacket(m[0x402]))
	} // drive lock

	if t149, ok := m[0x149]; ok {
		t149r := binary.NewReader(t149)
		t149r.ReadBytes(2)
		t149r.ReadStringShort() // title
		return LoginResponse{
			Success:      false,
			Error:        OtherLoginError,
			ErrorMessage: t149r.ReadStringShort(),
		}, nil
	}

	if t146, ok := m[0x146]; ok {
		t146r := binary.NewReader(t146)
		t146r.ReadBytes(4)      // ver and code
		t146r.ReadStringShort() // title
		return LoginResponse{
			Success:      false,
			Error:        OtherLoginError,
			ErrorMessage: t146r.ReadStringShort(),
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("unknown login response: %v", t)) // ?
}

// StatSvc.register
func decodeClientRegisterResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	return nil, nil
}

// wtlogin.exchange_emp
func decodeExchangeEmpResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	reader := binary.NewReader(payload)
	cmd := reader.ReadUInt16()
	t := reader.ReadByte()
	reader.ReadUInt16()
	m := reader.ReadTlvMap(2)
	if t != 0 {
		c.Error("exchange_emp error: %v", t)
		return nil, nil
	}
	if cmd == 15 { // TODO: 免密登录
		c.decodeT119R(m[0x119])
	}
	return nil, nil
}

// ConfigPushSvc.PushReq
func decodePushReqPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["PushReq"]["ConfigPush.PushReq"][1:])
	jceBuf := []byte{}
	t := r.ReadInt32(1)
	r.ReadSlice(&jceBuf, 2)
	if t == 1 && len(jceBuf) > 0 {
		ssoPkt := jce.NewJceReader(jceBuf)
		servers := []jce.SsoServerInfo{}
		ssoPkt.ReadSlice(&servers, 1)
		if len(servers) > 0 {
			var adds []*net.TCPAddr
			for _, s := range servers {
				if strings.Contains(s.Server, "com") {
					continue
				}
				c.Debug("got new server addr: %v location: %v", s.Server, s.Location)
				adds = append(adds, &net.TCPAddr{
					IP:   net.ParseIP(s.Server),
					Port: int(s.Port),
				})
			}
			c.SetCustomServer(adds)
			for _, e := range c.eventHandlers.serverUpdatedHandlers {
				cover(func() {
					e(c, &ServerUpdatedEvent{Servers: servers})
				})
			}
			return nil, nil
		}
	}
	seq := r.ReadInt64(3)
	_, pkt := c.buildConfPushRespPacket(t, seq, jceBuf)
	return nil, c.send(pkt)
}

// MessageSvc.PbGetMsg
func decodeMessageSvcPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.GetMessageResponse{}
	err := proto.Unmarshal(payload, &rsp)
	if err != nil {
		return nil, err
	}
	if rsp.Result != 0 {
		return nil, errors.New("message svc result unsuccessful")
	}
	c.syncCookie = rsp.SyncCookie
	c.pubAccountCookie = rsp.PubAccountCookie
	c.msgCtrlBuf = rsp.MsgCtrlBuf
	if rsp.UinPairMsgs == nil {
		return nil, nil
	}
	var delItems []*pb.MessageItem
	for _, pairMsg := range rsp.UinPairMsgs {
		for _, message := range pairMsg.Messages {
			// delete message
			delItem := &pb.MessageItem{
				FromUin: message.Head.FromUin,
				ToUin:   message.Head.ToUin,
				MsgType: 187,
				MsgSeq:  message.Head.MsgSeq,
				MsgUid:  message.Head.MsgUid,
			}
			delItems = append(delItems, delItem)
			if message.Head.ToUin != c.Uin {
				continue
			}
			if (int64(pairMsg.LastReadTime) & 4294967295) > int64(message.Head.MsgTime) {
				continue
			}
			strKey := fmt.Sprintf("%d%d%d%d", message.Head.FromUin, message.Head.ToUin, message.Head.MsgSeq, message.Head.MsgUid)
			if _, ok := c.msgSvcCache.Get(strKey); ok {
				continue
			}
			c.msgSvcCache.Add(strKey, "", time.Second*15)
			switch message.Head.MsgType {
			case 33: // 加群同步
				groupJoinLock.Lock()
				group := c.FindGroupByUin(message.Head.FromUin)
				if message.Head.AuthUin == c.Uin {
					if group == nil && c.ReloadGroupList() == nil {
						c.dispatchJoinGroupEvent(c.FindGroupByUin(message.Head.FromUin))
					}
				} else {
					if group != nil && group.FindMember(message.Head.AuthUin) == nil {
						mem := &GroupMemberInfo{
							Uin: message.Head.AuthUin,
							Nickname: func() string {
								if message.Head.AuthNick == "" {
									return message.Head.FromNick
								}
								return message.Head.AuthNick
							}(),
							JoinTime:   time.Now().Unix(),
							Permission: Member,
							Group:      group,
						}
						group.Update(func(info *GroupInfo) {
							info.Members = append(info.Members, mem)
						})
						c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
							Group:  group,
							Member: mem,
						})
					}
				}
				groupJoinLock.Unlock()
			case 84, 87:
				_, pkt := c.buildSystemMsgNewGroupPacket()
				_ = c.send(pkt)
			case 141: // 临时会话
				if message.Head.C2CTmpMsgHead == nil {
					continue
				}
				group := c.FindGroupByUin(message.Head.C2CTmpMsgHead.GroupUin)
				if group == nil {
					continue
				}
				if message.Head.FromUin == c.Uin {
					continue
				}
				c.dispatchTempMessage(c.parseTempMessage(message))
			case 166: // 好友消息
				if message.Head.FromUin == c.Uin {
					for {
						frdSeq := atomic.LoadInt32(&c.friendSeq)
						if frdSeq < message.Head.MsgSeq {
							if atomic.CompareAndSwapInt32(&c.friendSeq, frdSeq, message.Head.MsgSeq) {
								break
							}
						} else {
							break
						}
					}
				}
				if message.Body.RichText == nil || message.Body.RichText.Elems == nil {
					continue
				}
				c.dispatchFriendMessage(c.parsePrivateMessage(message))
			case 187:
				_, pkt := c.buildSystemMsgNewFriendPacket()
				_ = c.send(pkt)
			case 529:
				sub4 := msg.SubMsgType0X4Body{}
				if err := proto.Unmarshal(message.Body.MsgContent, &sub4); err != nil {
					c.Error("unmarshal sub msg 0x4 error: %v", err)
					continue
				}
				if sub4.NotOnlineFile != nil {
					rsp, err := c.sendAndWait(c.buildOfflineFileDownloadRequestPacket(sub4.NotOnlineFile.FileUuid)) // offline_file.go
					if err != nil {
						continue
					}
					c.dispatchOfflineFileEvent(&OfflineFileEvent{
						FileName:    string(sub4.NotOnlineFile.FileName),
						FileSize:    sub4.NotOnlineFile.FileSize,
						Sender:      message.Head.FromUin,
						DownloadUrl: rsp.(string),
					})
				}
			}
		}
	}
	_, _ = c.sendAndWait(c.buildDeleteMessageRequestPacket(delItems))
	if rsp.SyncFlag != msg.SyncFlag_STOP {
		c.Debug("continue sync with flag: %v", rsp.SyncFlag.String())
		_, _ = c.sendAndWait(c.buildGetMessageRequestPacket(rsp.SyncFlag, time.Now().Unix()))
	}
	return nil, err
}

// OnlinePush.PbPushGroupMsg
func decodeGroupMessagePacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkt := msg.PushMessagePacket{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, err
	}
	if pkt.Message.Head.FromUin == c.Uin {
		c.dispatchGroupMessageReceiptEvent(&groupMessageReceiptEvent{
			Rand: pkt.Message.Body.RichText.Attr.Random,
			Seq:  pkt.Message.Head.MsgSeq,
			Msg:  c.parseGroupMessage(pkt.Message),
		})
		return nil, nil
	}
	if pkt.Message.Content != nil && pkt.Message.Content.PkgNum > 1 {
		var builder *groupMessageBuilder // TODO: 支持多SEQ
		i, ok := c.groupMsgBuilders.Load(pkt.Message.Content.DivSeq)
		if !ok {
			builder = &groupMessageBuilder{}
			c.groupMsgBuilders.Store(pkt.Message.Content.DivSeq, builder)
		} else {
			builder = i.(*groupMessageBuilder)
		}
		builder.MessageSlices = append(builder.MessageSlices, pkt.Message)
		if int32(len(builder.MessageSlices)) >= pkt.Message.Content.PkgNum {
			c.groupMsgBuilders.Delete(pkt.Message.Content.DivSeq)
			c.dispatchGroupMessage(c.parseGroupMessage(builder.build()))
		}
		return nil, nil
	}
	c.dispatchGroupMessage(c.parseGroupMessage(pkt.Message))
	return nil, nil
}

// MessageSvc.PbSendMsg
func decodeMsgSendResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.SendMessageResponse{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if rsp.Result != 0 {
		c.Error("send msg error: %v %v", rsp.Result, rsp.ErrMsg)
	}
	return nil, nil
}

// MessageSvc.PushNotify
func decodeSvcNotify(c *QQClient, _ uint16, _ []byte) (interface{}, error) {
	c.msgSvcLock.Lock()
	defer c.msgSvcLock.Unlock()
	_, err := c.sendAndWait(c.buildGetMessageRequestPacket(msg.SyncFlag_START, time.Now().Unix()))
	return nil, err
}

// StatSvc.GetDevLoginInfo
func decodeDevListResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	rsp := jce.NewJceReader(data.Map["SvcRspGetDevLoginInfo"]["QQService.SvcRspGetDevLoginInfo"][1:])
	d := []jce.SvcDevLoginInfo{}
	ret := rsp.ReadInt64(3)
	rsp.ReadSlice(&d, 5)
	return ret, nil
}

// SummaryCard.ReqSummaryCard
func decodeSummaryCardResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	rsp := func() *jce.JceReader {
		if r, ok := data.Map["RespSummaryCard"]["SummaryCard.RespSummaryCard"]; ok {
			return jce.NewJceReader(r[1:])
		}
		return jce.NewJceReader(data.Map["RespSummaryCard"]["SummaryCard_Old.RespSummaryCard"][1:])
	}()
	return &SummaryCardInfo{
		Sex:       rsp.ReadByte(1),
		Age:       rsp.ReadByte(2),
		Nickname:  rsp.ReadString(3),
		Level:     rsp.ReadInt32(5),
		City:      rsp.ReadString(7),
		Sign:      rsp.ReadString(8),
		Mobile:    rsp.ReadString(11),
		Uin:       rsp.ReadInt64(23),
		LoginDays: rsp.ReadInt64(36),
	}, nil
}

// friendlist.getFriendGroupList
func decodeFriendGroupListResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["FLRESP"][1:])
	totalFriendCount := r.ReadInt16(5)
	friends := []jce.FriendInfo{}
	r.ReadSlice(&friends, 7)
	var l []*FriendInfo
	for _, f := range friends {
		l = append(l, &FriendInfo{
			Uin:      f.FriendUin,
			Nickname: f.Nick,
			Remark:   f.Remark,
			FaceId:   f.FaceId,
		})
	}
	rsp := FriendListResponse{
		TotalCount: int32(totalFriendCount),
		List:       l,
	}
	return rsp, nil
}

// friendlist.GetTroopListReqV2
func decodeGroupListResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GetTroopListRespV2"][1:])
	vecCookie := []byte{}
	groups := []jce.TroopNumber{}
	r.ReadSlice(&vecCookie, 4)
	r.ReadSlice(&groups, 5)
	var l []*GroupInfo
	for _, g := range groups {
		l = append(l, &GroupInfo{
			Uin:            g.GroupUin,
			Code:           g.GroupCode,
			Name:           g.GroupName,
			Memo:           g.GroupMemo,
			OwnerUin:       g.GroupOwnerUin,
			MemberCount:    uint16(g.MemberNum),
			MaxMemberCount: uint16(g.MaxGroupMemberNum),
			client:         c,
		})
	}
	if len(vecCookie) > 0 {
		rsp, err := c.sendAndWait(c.buildGroupListRequestPacket(vecCookie))
		if err != nil {
			return nil, err
		}
		l = append(l, rsp.([]*GroupInfo)...)
	}
	return l, nil
}

// OidbSvc.0x88d_0
func decodeGroupInfoResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D88DRspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, err
	}
	if len(rsp.RspGroupInfo) == 0 {
		return nil, errors.New(string(rsp.StrErrorInfo))
	}
	info := rsp.RspGroupInfo[0]
	return &GroupInfo{
		Uin:            int64(*info.GroupInfo.GroupUin),
		Code:           int64(*info.GroupCode),
		Name:           string(info.GroupInfo.GroupName),
		Memo:           string(info.GroupInfo.GroupMemo),
		OwnerUin:       int64(*info.GroupInfo.GroupOwner),
		MemberCount:    uint16(*info.GroupInfo.GroupMemberNum),
		MaxMemberCount: uint16(*info.GroupInfo.GroupMemberMaxNum),
		Members:        []*GroupMemberInfo{},
		client:         c,
	}, nil
}

// friendlist.GetTroopMemberListReq
func decodeGroupMemberListResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GTMLRESP"][1:])
	members := []jce.TroopMemberInfo{}
	r.ReadSlice(&members, 3)
	next := r.ReadInt64(4)
	var l []*GroupMemberInfo
	for _, m := range members {
		l = append(l, &GroupMemberInfo{
			Uin:                    m.MemberUin,
			Nickname:               m.Nick,
			Gender:                 m.Gender,
			CardName:               m.Name,
			Level:                  uint16(m.MemberLevel),
			JoinTime:               m.JoinTime,
			LastSpeakTime:          m.LastSpeakTime,
			SpecialTitle:           m.SpecialTitle,
			SpecialTitleExpireTime: m.SpecialTitleExpireTime,
			Permission: func() MemberPermission {
				if m.Flag == 1 {
					return Administrator
				}
				return Member
			}(),
		})
	}
	return groupMemberListResponse{
		NextUin: next,
		list:    l,
	}, nil
}

// group_member_card.get_group_member_card_info
func decodeGroupMemberInfoResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := pb.GroupMemberRspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if rsp.MemInfo.Nick == nil && rsp.MemInfo.Age == 0 {
		return nil, ErrMemberNotFound
	}
	group := c.FindGroup(rsp.GroupCode)
	return &GroupMemberInfo{
		Group:                  group,
		Uin:                    rsp.MemInfo.Uin,
		Gender:                 byte(rsp.MemInfo.Sex),
		Nickname:               string(rsp.MemInfo.Nick),
		CardName:               string(rsp.MemInfo.Card),
		Level:                  uint16(rsp.MemInfo.Level),
		JoinTime:               rsp.MemInfo.Join,
		LastSpeakTime:          rsp.MemInfo.LastSpeak,
		SpecialTitle:           string(rsp.MemInfo.SpecialTitle),
		SpecialTitleExpireTime: int64(rsp.MemInfo.SpecialTitleExpireTime),
		Permission: func() MemberPermission {
			if rsp.MemInfo.Uin == group.OwnerUin {
				return Owner
			}
			if rsp.MemInfo.Role == 1 {
				return Administrator
			}
			return Member
		}(),
	}, nil
}

// ImgStore.GroupPicUp
func decodeGroupImageStoreResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkt := pb.D388RespBody{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, err
	}
	rsp := pkt.MsgTryUpImgRsp[0]
	if rsp.Result != 0 {
		return imageUploadResponse{
			ResultCode: rsp.Result,
			Message:    rsp.FailMsg,
		}, nil
	}
	if rsp.BoolFileExit {
		if rsp.MsgImgInfo != nil {
			return imageUploadResponse{IsExists: true, FileId: rsp.Fid, Width: rsp.MsgImgInfo.FileWidth, Height: rsp.MsgImgInfo.FileHeight}, nil
		}
		return imageUploadResponse{IsExists: true, FileId: rsp.Fid}, nil
	}
	return imageUploadResponse{
		FileId:     rsp.Fid,
		UploadKey:  rsp.UpUkey,
		UploadIp:   rsp.Uint32UpIp,
		UploadPort: rsp.Uint32UpPort,
	}, nil
}

// LongConn.OffPicUp
func decodeOffPicUpResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := cmd0x352.RspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if rsp.FailMsg != "" {
		return imageUploadResponse{
			ResultCode: -1,
			Message:    rsp.FailMsg,
		}, nil
	}
	if rsp.Subcmd != 1 || len(rsp.MsgTryupImgRsp) == 0 {
		return imageUploadResponse{
			ResultCode: -2,
		}, nil
	}
	imgRsp := rsp.MsgTryupImgRsp[0]
	if imgRsp.Result != 0 {
		return imageUploadResponse{
			ResultCode: imgRsp.Result,
			Message:    imgRsp.FailMsg,
		}, nil
	}
	if imgRsp.BoolFileExit {
		return imageUploadResponse{
			IsExists:   true,
			ResourceId: imgRsp.UpResid,
		}, nil
	}
	return imageUploadResponse{
		ResourceId: imgRsp.UpResid,
		UploadKey:  imgRsp.UpUkey,
		UploadIp:   imgRsp.Uint32UpIp,
		UploadPort: imgRsp.Uint32UpPort,
	}, nil
}

// OnlinePush.ReqPush
func decodeOnlinePushReqPacket(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	jr := jce.NewJceReader(data.Map["req"]["OnlinePushPack.SvcReqPushMsg"][1:])
	msgInfos := []jce.PushMessageInfo{}
	uin := jr.ReadInt64(0)
	jr.ReadSlice(&msgInfos, 2)
	_ = c.send(c.buildDeleteOnlinePushPacket(uin, seq, msgInfos))
	seqExists := func(ms int16) bool {
		for _, s := range c.onlinePushCache {
			if ms == s {
				return true
			}
		}
		return false
	}
	for _, m := range msgInfos {
		if seqExists(m.MsgSeq) {
			continue
		}
		c.onlinePushCache = append(c.onlinePushCache, m.MsgSeq)
		if m.MsgType == 732 {
			r := binary.NewReader(m.VMsg)
			groupId := int64(uint32(r.ReadInt32()))
			iType := r.ReadByte()
			r.ReadByte()
			switch iType {
			case 0x0c: // 群内禁言
				operator := int64(uint32(r.ReadInt32()))
				if operator == c.Uin {
					continue
				}
				r.ReadBytes(6)
				target := int64(uint32(r.ReadInt32()))
				t := r.ReadInt32()
				c.dispatchGroupMuteEvent(&GroupMuteEvent{
					GroupCode:   groupId,
					OperatorUin: operator,
					TargetUin:   target,
					Time:        t,
				})
			case 0x10, 0x11, 0x14: // group notify msg
				r.ReadByte()
				b := notify.NotifyMsgBody{}
				_ = proto.Unmarshal(r.ReadAvailable(), &b)
				if b.OptMsgRecall != nil {
					for _, rm := range b.OptMsgRecall.RecalledMsgList {
						c.dispatchGroupMessageRecalledEvent(&GroupMessageRecalledEvent{
							GroupCode:   groupId,
							OperatorUin: b.OptMsgRecall.Uin,
							AuthorUin:   rm.AuthorUin,
							MessageId:   rm.Seq,
							Time:        rm.Time,
						})
					}
				}
				if b.OptGeneralGrayTip != nil {
					c.grayTipProcessor(groupId, b.OptGeneralGrayTip)
				}
				if b.OptMsgRedTips != nil {
					if b.OptMsgRedTips.LuckyFlag == 1 { // 运气王提示
						c.dispatchGroupNotifyEvent(&GroupRedBagLuckyKingNotifyEvent{
							GroupCode: groupId,
							Sender:    int64(b.OptMsgRedTips.SenderUin),
							LuckyKing: int64(b.OptMsgRedTips.ReceiverUin),
						})
					}
				}
			}
		}
		if m.MsgType == 528 {
			vr := jce.NewJceReader(m.VMsg)
			subType := vr.ReadInt64(0)
			probuf := vr.ReadAny(10).([]byte)
			switch subType {
			case 0x8A, 0x8B:
				s8a := pb.Sub8A{}
				if err := proto.Unmarshal(probuf, &s8a); err != nil {
					return nil, err
				}
				for _, m := range s8a.MsgInfo {
					if m.ToUin == c.Uin {
						c.dispatchFriendMessageRecalledEvent(&FriendMessageRecalledEvent{
							FriendUin: m.FromUin,
							MessageId: m.MsgSeq,
							Time:      m.MsgTime,
						})
					}
				}
			case 0xB3:
				b3 := pb.SubB3{}
				if err := proto.Unmarshal(probuf, &b3); err != nil {
					return nil, err
				}
				frd := &FriendInfo{
					Uin:      b3.MsgAddFrdNotify.Uin,
					Nickname: b3.MsgAddFrdNotify.Nick,
				}
				c.FriendList = append(c.FriendList, frd)
				c.dispatchNewFriendEvent(&NewFriendEvent{Friend: frd})
			case 0xD4:
				d4 := pb.SubD4{}
				if err := proto.Unmarshal(probuf, &d4); err != nil {
					return nil, err
				}
				groupLeaveLock.Lock()
				if g := c.FindGroup(d4.Uin); g != nil {
					if err := c.ReloadGroupList(); err != nil {
						groupLeaveLock.Unlock()
						return nil, err
					}
					c.dispatchLeaveGroupEvent(&GroupLeaveEvent{Group: g})
				}
				groupLeaveLock.Unlock()
			case 0x44:
				s44 := pb.Sub44{}
				if err := proto.Unmarshal(probuf, &s44); err != nil {
					return nil, err
				}
				if s44.GroupSyncMsg != nil {
					func() {
						groupJoinLock.Lock()
						defer groupJoinLock.Unlock()
						c.Debug("syncing groups.")
						old := c.GroupList
						any := func(code int64) bool {
							for _, g := range old {
								if g.Code == code {
									return true
								}
							}
							return false
						}
						if err := c.ReloadGroupList(); err == nil {
							for _, g := range c.GroupList {
								if !any(g.Code) {
									c.dispatchJoinGroupEvent(g)
								}
							}
						}
					}()
				}
			}
		}
	}

	return nil, nil
}

// OnlinePush.PbPushTransMsg
func decodeOnlinePushTransPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	info := msg.TransMsgInfo{}
	err := proto.Unmarshal(payload, &info)
	if err != nil {
		return nil, err
	}
	data := binary.NewReader(info.MsgData)
	idStr := strconv.FormatInt(info.MsgUid, 10)
	if _, ok := c.transCache.Get(idStr); ok {
		return nil, nil
	}
	c.transCache.Add(idStr, "", time.Second*15)
	if info.MsgType == 34 {
		data.ReadInt32()
		data.ReadByte()
		target := int64(uint32(data.ReadInt32()))
		typ := int32(data.ReadByte())
		operator := int64(uint32(data.ReadInt32()))
		if g := c.FindGroupByUin(info.FromUin); g != nil {
			groupLeaveLock.Lock()
			defer groupLeaveLock.Unlock()
			switch typ {
			case 0x02:
				if target == c.Uin {
					c.dispatchLeaveGroupEvent(&GroupLeaveEvent{Group: g})
				} else {
					if m := g.FindMember(target); m != nil {
						g.removeMember(target)
						c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
							Group:  g,
							Member: m,
						})
					}
				}
			case 0x03:
				if err = c.ReloadGroupList(); err != nil {
					return nil, err
				}
				if target == c.Uin {
					c.dispatchLeaveGroupEvent(&GroupLeaveEvent{
						Group:    g,
						Operator: g.FindMember(operator),
					})
				} else {
					if m := g.FindMember(target); m != nil {
						g.removeMember(target)
						c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
							Group:    g,
							Member:   m,
							Operator: g.FindMember(operator),
						})
					}
				}
			case 0x82:
				if m := g.FindMember(target); m != nil {
					g.removeMember(target)
					c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
						Group:  g,
						Member: m,
					})
				}
			case 0x83:
				if m := g.FindMember(target); m != nil {
					g.removeMember(target)
					c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
						Group:    g,
						Member:   m,
						Operator: g.FindMember(operator),
					})
				}
			}
		}
	}
	if info.MsgType == 44 {
		data.ReadBytes(5)
		var4 := int32(data.ReadByte())
		var var5 int64 = 0
		target := int64(uint32(data.ReadInt32()))
		if var4 != 0 && var4 != 1 {
			var5 = int64(uint32(data.ReadInt32()))
		}
		if g := c.FindGroupByUin(info.FromUin); g != nil {
			if var5 == 0 && data.Len() == 1 {
				newPermission := func() MemberPermission {
					if data.ReadByte() == 1 {
						return Administrator
					}
					return Member
				}()
				mem := g.FindMember(target)
				if mem.Permission != newPermission {
					old := mem.Permission
					mem.Permission = newPermission
					c.dispatchPermissionChanged(&MemberPermissionChangedEvent{
						Group:         g,
						Member:        mem,
						OldPermission: old,
						NewPermission: newPermission,
					})
				}
			}
		}
	}
	return nil, nil
}

// ProfileService.Pb.ReqSystemMsgNew.Group
func decodeSystemMsgGroupPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := structmsg.RspSystemMsgNew{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if len(rsp.Groupmsgs) == 0 {
		return nil, nil
	}
	st := rsp.Groupmsgs[0]
	if st.Msg != nil {
		if st.Msg.SubType == 1 {
			// 处理被邀请入群 或 处理成员入群申请
			switch st.Msg.GroupMsgType {
			case 1: // 成员申请
				c.dispatchJoinGroupRequest(&UserJoinGroupRequest{
					RequestId:     st.MsgSeq,
					Message:       st.Msg.MsgAdditional,
					RequesterUin:  st.ReqUin,
					RequesterNick: st.Msg.ReqUinNick,
					GroupCode:     st.Msg.GroupCode,
					GroupName:     st.Msg.GroupName,
					client:        c,
				})
			case 2: // 被邀请
				c.dispatchGroupInvitedEvent(&GroupInvitedRequest{
					RequestId:   st.MsgSeq,
					InvitorUin:  st.Msg.ActionUin,
					InvitorNick: st.Msg.ActionUinNick,
					GroupCode:   st.Msg.GroupCode,
					GroupName:   st.Msg.GroupName,
					client:      c,
				})
			default:
				log.Println("unknown group msg:", st)
			}
		} else if st.Msg.SubType == 2 {
			// 被邀请入群, 自动同意, 不需处理
		} else if st.Msg.SubType == 3 {
			// 已被其他管理员处理
		} else if st.Msg.SubType == 5 {
			// 成员退群消息
		} else {
			log.Println("unknown group msg:", st)
		}
	}
	return nil, nil
}

// ProfileService.Pb.ReqSystemMsgNew.Friend
func decodeSystemMsgFriendPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := structmsg.RspSystemMsgNew{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if len(rsp.Friendmsgs) == 0 {
		return nil, nil
	}
	st := rsp.Friendmsgs[0]
	if st.Msg != nil {
		c.dispatchNewFriendRequest(&NewFriendRequest{
			RequestId:     st.MsgSeq,
			Message:       st.Msg.MsgAdditional,
			RequesterUin:  st.ReqUin,
			RequesterNick: st.Msg.ReqUinNick,
			client:        c,
		})
	}
	return nil, nil
}

// MessageSvc.PushForceOffline
func decodeForceOfflinePacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["req_PushForceOffline"]["PushNotifyPack.RequestPushForceOffline"][1:])
	tips := r.ReadString(2)
	c.lastLostMsg = tips
	c.NetLooping = false
	c.Online = false
	return nil, nil
}

// StatSvc.ReqMSFOffline
func decodeMSFOfflinePacket(c *QQClient, _ uint16, _ []byte) (interface{}, error) {
	c.lastLostMsg = "服务器端强制下线."
	c.NetLooping = false
	c.Online = false
	return nil, nil
}

// OidbSvc.0xd79
func decodeWordSegmentation(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := &oidb.D79RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, rsp); err != nil {
		return nil, err
	}
	if rsp.Content != nil {
		return rsp.Content.SliceContent, nil
	}
	return nil, errors.New("no word receive")
}

// OidbSvc.0x6d6_2
func decodeOIDB6d6Response(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D6RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, err
	}
	ip := rsp.DownloadFileRsp.DownloadIp
	url := hex.EncodeToString(rsp.DownloadFileRsp.DownloadUrl)
	return fmt.Sprintf("http://%s/ftn_handler/%s/", ip, url), nil
}

// OidbSvc.0xe07_0
func decodeImageOcrResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.DE07RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, err
	}
	if rsp.Wording != "" {
		return nil, errors.New(rsp.Wording)
	}
	var texts []*TextDetection
	for _, text := range rsp.OcrRspBody.TextDetections {
		var points []*Coordinate
		for _, c := range text.Polygon.Coordinates {
			points = append(points, &Coordinate{
				X: c.X,
				Y: c.Y,
			})
		}
		texts = append(texts, &TextDetection{
			Text:        text.DetectedText,
			Confidence:  text.Confidence,
			Coordinates: points,
		})
	}
	return &OcrResponse{
		Texts:    texts,
		Language: rsp.OcrRspBody.Language,
	}, nil
}

// PttCenterSvr.ShortVideoDownReq
func decodePttShortVideoDownResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := pttcenter.ShortVideoRspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if rsp.PttShortVideoDownloadRsp == nil || rsp.PttShortVideoDownloadRsp.DownloadAddr == nil {
		return nil, errors.New("resp error")
	}
	return rsp.PttShortVideoDownloadRsp.DownloadAddr.Host[0] + rsp.PttShortVideoDownloadRsp.DownloadAddr.UrlArgs, nil
}

// LightAppSvc.mini_app_info.GetAppInfoById
func decodeAppInfoResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := qweb.QWebRsp{}
	rsp := qweb.GetAppInfoByIdRsp{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if pkg.RetCode != 0 {
		return nil, errors.New(pkg.ErrMsg)
	}
	if err := proto.Unmarshal(pkg.BusiBuff, &rsp); err != nil {
		return nil, err
	}
	return rsp.AppInfo, nil
}
