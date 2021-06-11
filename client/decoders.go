package client

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x352"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x6ff"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/client/pb/profilecard"
	"github.com/Mrs4s/MiraiGo/client/pb/qweb"
	"github.com/Mrs4s/MiraiGo/client/pb/structmsg"
	"github.com/Mrs4s/MiraiGo/utils"
)

var (
	groupJoinLock  sync.Mutex
	groupLeaveLock sync.Mutex
)

// wtlogin.login
func decodeLoginResponse(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	reader := binary.NewReader(payload)
	reader.ReadUInt16() // sub command
	t := reader.ReadByte()
	reader.ReadUInt16()
	m := reader.ReadTlvMap(2)
	if m.Exists(0x402) {
		c.dpwd = []byte(utils.RandomString(16))
		c.t402 = m[0x402]
		h := md5.Sum(append(append(SystemDeviceInfo.Guid, c.dpwd...), c.t402...))
		c.g = h[:]
	}
	if t == 0 { // login success
		if t150, ok := m[0x150]; ok {
			c.t150 = t150
		}
		if t161, ok := m[0x161]; ok {
			c.decodeT161(t161)
		}
		if m.Exists(0x403) {
			c.randSeed = m[0x403]
		}
		c.decodeT119(m[0x119], SystemDeviceInfo.TgtgtKey)
		return LoginResponse{
			Success: true,
		}, nil
	}
	if t == 2 {
		c.t104 = m[0x104]
		if m.Exists(0x192) {
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
			c.randSeed = m[0x403]
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
		c.randSeed = m[0x403]
		return c.sendAndWait(c.buildDeviceLockLoginPacket())
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
	c.Debug("unknown login response: %v", t)
	for k, v := range m {
		c.Debug("Type: %v Value: %v", strconv.FormatInt(int64(k), 16), hex.EncodeToString(v))
	}
	return nil, errors.Errorf("unknown login response: %v", t) // ?
}

// StatSvc.register
func decodeClientRegisterResponse(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	svcRsp := &jce.SvcRespRegister{}
	svcRsp.ReadFrom(jce.NewJceReader(data.Map["SvcRespRegister"]["QQService.SvcRespRegister"][1:]))
	if svcRsp.Result != "" || svcRsp.ReplyCode != 0 {
		if svcRsp.Result != "" {
			c.Error("reg error: %v", svcRsp.Result)
		}
		return nil, errors.New("reg failed")
	}
	return nil, nil
}

// wtlogin.exchange_emp
func decodeExchangeEmpResponse(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	reader := binary.NewReader(payload)
	cmd := reader.ReadUInt16()
	t := reader.ReadByte()
	reader.ReadUInt16()
	m := reader.ReadTlvMap(2)
	if t != 0 {
		c.Error("exchange_emp error: %v", t)
		return nil, errors.New("exchange_emp failed")
	}
	if cmd == 15 {
		c.decodeT119R(m[0x119])
	}
	if cmd == 11 {
		h := md5.Sum(c.sigInfo.d2Key)
		c.decodeT119(m[0x119], h[:])
	}
	return nil, nil
}

// wtlogin.trans_emp
func decodeTransEmpResponse(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	if len(payload) < 48 {
		return nil, errors.New("missing payload length")
	}
	reader := binary.NewReader(payload)
	reader.ReadBytes(5) // trans req head
	reader.ReadByte()
	reader.ReadUInt16()
	cmd := reader.ReadUInt16()
	reader.ReadBytes(21)
	reader.ReadByte()
	reader.ReadUInt16()
	reader.ReadUInt16()
	reader.ReadInt32()
	reader.ReadInt64()
	body := binary.NewReader(reader.ReadBytes(reader.Len() - 1))
	if cmd == 0x31 {
		body.ReadUInt16()
		body.ReadInt32()
		code := body.ReadByte()
		if code != 0 {
			return nil, errors.Errorf("wtlogin.trans_emp sub cmd 0x31 error: %v", code)
		}
		sig := body.ReadBytesShort()
		body.ReadUInt16()
		m := body.ReadTlvMap(2)
		if m.Exists(0x17) {
			return &QRCodeLoginResponse{
				State:     QRCodeImageFetch,
				ImageData: m[0x17],
				Sig:       sig,
			}, nil
		}
		return nil, errors.Errorf("wtlogin.trans_emp sub cmd 0x31 error: image not found")
	}
	if cmd == 0x12 {
		aVarLen := body.ReadUInt16()
		if aVarLen != 0 {
			aVarLen-- // 阴间的位移操作
			if body.ReadByte() == 2 {
				body.ReadInt64() // uin ?
				aVarLen -= 8
			}
		}
		if aVarLen > 0 {
			body.ReadBytes(int(aVarLen))
		}
		body.ReadInt32() // app id?
		code := body.ReadByte()
		if code != 0 {
			if code == 0x30 {
				return &QRCodeLoginResponse{State: QRCodeWaitingForScan}, nil
			}
			if code == 0x35 {
				return &QRCodeLoginResponse{State: QRCodeWaitingForConfirm}, nil
			}
			if code == 0x36 {
				return &QRCodeLoginResponse{State: QRCodeCanceled}, nil
			}
			if code == 0x11 {
				return &QRCodeLoginResponse{State: QRCodeTimeout}, nil
			}
			return nil, errors.Errorf("wtlogin.trans_emp sub cmd 0x12 error: %v", code)
		}
		c.Uin = body.ReadInt64()
		body.ReadInt32() // sig create time
		body.ReadUInt16()
		m := body.ReadTlvMap(2)
		if !m.Exists(0x18) || !m.Exists(0x1e) || !m.Exists(0x19) {
			return nil, errors.New("wtlogin.trans_emp sub cmd 0x12 error: tlv error")
		}
		SystemDeviceInfo.TgtgtKey = m[0x1e]
		return &QRCodeLoginResponse{State: QRCodeConfirmed, LoginInfo: &QRCodeLoginInfo{
			tmpPwd:      m[0x18],
			tmpNoPicSig: m[0x19],
			tgtQR:       m[0x65],
		}}, nil
	}
	return nil, errors.Errorf("unknown trans_emp response: %v", cmd)
}

// ConfigPushSvc.PushReq
func decodePushReqPacket(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["PushReq"]["ConfigPush.PushReq"][1:])
	t := r.ReadInt32(1)
	jceBuf := r.ReadAny(2).([]byte)
	if len(jceBuf) > 0 {
		switch t {
		case 1:
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
				f := true
				for _, e := range c.eventHandlers.serverUpdatedHandlers {
					cover(func() {
						if !e(c, &ServerUpdatedEvent{Servers: servers}) {
							f = false
						}
					})
				}
				if f {
					c.SetCustomServer(adds)
				}
				return nil, nil
			}
		case 2:
			fmtPkt := jce.NewJceReader(jceBuf)
			list := &jce.FileStoragePushFSSvcList{}
			list.ReadFrom(fmtPkt)
			c.Debug("got file storage svc push.")
			c.fileStorageInfo = list
			rsp := cmd0x6ff.C501RspBody{}
			if err := proto.Unmarshal(list.BigDataChannel.PbBuf, &rsp); err == nil && rsp.RspBody != nil {
				c.bigDataSession = &bigDataSessionInfo{
					SigSession: rsp.RspBody.SigSession,
					SessionKey: rsp.RspBody.SessionKey,
				}
				for _, srv := range rsp.RspBody.Addrs {
					if srv.GetServiceType() == 10 {
						for _, addr := range srv.Addrs {
							c.srvSsoAddrs = append(c.srvSsoAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(addr.GetIp()), addr.GetPort()))
						}
					}
					if srv.GetServiceType() == 21 {
						for _, addr := range srv.Addrs {
							c.otherSrvAddrs = append(c.otherSrvAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(addr.GetIp()), addr.GetPort()))
						}
					}
				}
			}
		}
	}

	seq := r.ReadInt64(3)
	_, pkt := c.buildConfPushRespPacket(t, seq, jceBuf)
	return nil, c.send(pkt)
}

// MessageSvc.PbGetMsg
func decodeMessageSvcPacket(c *QQClient, info *incomingPacketInfo, payload []byte) (interface{}, error) {
	rsp := msg.GetMessageResponse{}
	err := proto.Unmarshal(payload, &rsp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	c.c2cMessageSyncProcessor(&rsp, info)
	return nil, nil
}

// MessageSvc.PushNotify
func decodeSvcNotify(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload[15:]))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	if len(data.Map) == 0 {
		_, err := c.sendAndWait(c.buildGetMessageRequestPacket(msg.SyncFlag_START, time.Now().Unix()))
		return nil, err
	}
	notify := &jce.RequestPushNotify{}
	notify.ReadFrom(jce.NewJceReader(data.Map["req_PushNotify"]["PushNotifyPack.RequestPushNotify"][1:]))
	if _, ok := troopSystemMsgDecoders[notify.MsgType]; ok && notify.MsgType != 85 && notify.MsgType != 36 {
		c.exceptAndDispatchGroupSysMsg()
		return nil, nil
	}
	if _, ok := sysMsgDecoders[notify.MsgType]; ok {
		_, pkt := c.buildSystemMsgNewFriendPacket()
		return nil, c.send(pkt)
	}
	_, err := c.sendAndWait(c.buildGetMessageRequestPacket(msg.SyncFlag_START, time.Now().Unix()))
	return nil, err
}

// SummaryCard.ReqSummaryCard
func decodeSummaryCardResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
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
	info := &SummaryCardInfo{
		Sex:       rsp.ReadByte(1),
		Age:       rsp.ReadByte(2),
		Nickname:  rsp.ReadString(3),
		Level:     rsp.ReadInt32(5),
		City:      rsp.ReadString(7),
		Sign:      rsp.ReadString(8),
		Mobile:    rsp.ReadString(11),
		Uin:       rsp.ReadInt64(23),
		LoginDays: rsp.ReadInt64(36),
	}
	services := [][]byte{}
	rsp.ReadSlice(&services, 46)
	readService := func(buf []byte) (*profilecard.BusiComm, []byte) {
		r := binary.NewReader(buf)
		r.ReadByte()
		l1 := r.ReadInt32()
		l2 := r.ReadInt32()
		comm := r.ReadBytes(int(l1))
		d := r.ReadBytes(int(l2))
		c := &profilecard.BusiComm{}
		_ = proto.Unmarshal(comm, c)
		return c, d
	}
	for _, buf := range services {
		comm, payload := readService(buf)
		if comm.GetService() == 16 {
			rsp := profilecard.GateVaProfileGateRsp{}
			_ = proto.Unmarshal(payload, &rsp)
			if rsp.QidInfo != nil {
				info.Qid = rsp.QidInfo.GetQid()
			}
		}
	}
	return info, nil
}

// friendlist.getFriendGroupList
func decodeFriendGroupListResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["FLRESP"][1:])
	totalFriendCount := r.ReadInt16(5)
	friends := []jce.FriendInfo{}
	r.ReadSlice(&friends, 7)
	l := make([]*FriendInfo, 0, len(friends))
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

// friendlist.delFriend
func decodeFriendDeleteResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["DFRESP"][1:])
	if ret := r.ReadInt32(2); ret != 0 {
		return nil, errors.Errorf("delete friend error: %v", ret)
	}
	return nil, nil
}

// friendlist.GetTroopListReqV2
func decodeGroupListResponse(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GetTroopListRespV2"][1:])
	vecCookie := []byte{}
	groups := []jce.TroopNumber{}
	r.ReadSlice(&vecCookie, 4)
	r.ReadSlice(&groups, 5)
	l := make([]*GroupInfo, 0, len(groups))
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

// friendlist.GetTroopMemberListReq
func decodeGroupMemberListResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GTMLRESP"][1:])
	members := []jce.TroopMemberInfo{}
	r.ReadSlice(&members, 3)
	next := r.ReadInt64(4)
	l := make([]*GroupMemberInfo, 0, len(members))
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
func decodeGroupMemberInfoResponse(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	rsp := pb.GroupMemberRspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.MemInfo == nil || (rsp.MemInfo.Nick == nil && rsp.MemInfo.Age == 0) {
		return nil, errors.WithStack(ErrMemberNotFound)
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
			if rsp.MemInfo.Role == 2 {
				return Administrator
			}
			return Member
		}(),
	}, nil
}

// LongConn.OffPicUp
func decodeOffPicUpResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	rsp := cmd0x352.RspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.GetFailMsg() != nil {
		return imageUploadResponse{
			ResultCode: -1,
			Message:    string(rsp.FailMsg),
		}, nil
	}
	if rsp.GetSubcmd() != 1 || len(rsp.GetTryupImgRsp()) == 0 {
		return imageUploadResponse{
			ResultCode: -2,
		}, nil
	}
	imgRsp := rsp.GetTryupImgRsp()[0]
	if imgRsp.GetResult() != 0 {
		return imageUploadResponse{
			ResultCode: int32(*imgRsp.Result),
			Message:    string(imgRsp.GetFailMsg()),
		}, nil
	}
	if imgRsp.GetFileExit() {
		return imageUploadResponse{
			IsExists:   true,
			ResourceId: string(imgRsp.GetUpResid()),
		}, nil
	}
	return imageUploadResponse{
		ResourceId: string(imgRsp.GetUpResid()),
		UploadKey:  imgRsp.GetUpUkey(),
		UploadIp:   imgRsp.GetUpIp(),
		UploadPort: imgRsp.GetUpPort(),
	}, nil
}

// OnlinePush.PbPushTransMsg
func decodeOnlinePushTransPacket(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	info := msg.TransMsgInfo{}
	err := proto.Unmarshal(payload, &info)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	data := binary.NewReader(info.MsgData)
	idStr := strconv.FormatInt(info.GetMsgUid(), 10)
	if _, ok := c.transCache.Get(idStr); ok {
		return nil, nil
	}
	c.transCache.Add(idStr, "", time.Second*15)
	if info.GetMsgType() == 34 {
		data.ReadInt32()
		data.ReadByte()
		target := int64(uint32(data.ReadInt32()))
		typ := int32(data.ReadByte())
		operator := int64(uint32(data.ReadInt32()))
		if g := c.FindGroupByUin(info.GetFromUin()); g != nil {
			groupLeaveLock.Lock()
			defer groupLeaveLock.Unlock()
			switch typ {
			case 0x02:
				if target == c.Uin {
					c.dispatchLeaveGroupEvent(&GroupLeaveEvent{Group: g})
				} else if m := g.FindMember(target); m != nil {
					g.removeMember(target)
					c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
						Group:  g,
						Member: m,
					})
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
				} else if m := g.FindMember(target); m != nil {
					g.removeMember(target)
					c.dispatchMemberLeaveEvent(&MemberLeaveGroupEvent{
						Group:    g,
						Member:   m,
						Operator: g.FindMember(operator),
					})
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
	if info.GetMsgType() == 44 {
		data.ReadBytes(5)
		var4 := int32(data.ReadByte())
		var5 := int64(0)
		target := int64(uint32(data.ReadInt32()))
		if var4 != 0 && var4 != 1 {
			var5 = int64(uint32(data.ReadInt32()))
		}
		if g := c.FindGroupByUin(info.GetFromUin()); g != nil {
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

// ProfileService.Pb.ReqSystemMsgNew.Friend
func decodeSystemMsgFriendPacket(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	rsp := structmsg.RspSystemMsgNew{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
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
func decodeForceOfflinePacket(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["req_PushForceOffline"]["PushNotifyPack.RequestPushForceOffline"][1:])
	tips := r.ReadString(2)
	c.Disconnect()
	go c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: tips})
	return nil, nil
}

// StatSvc.ReqMSFOffline
func decodeMSFOfflinePacket(c *QQClient, _ *incomingPacketInfo, _ []byte) (interface{}, error) {
	// c.lastLostMsg = "服务器端强制下线."
	c.Disconnect()
	// 这个decoder不能消耗太多时间, event另起线程处理
	go c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "服务端强制下线."})
	return nil, nil
}

// OidbSvc.0xd79
func decodeWordSegmentation(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := &oidb.D79RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.Content != nil {
		return rsp.Content.SliceContent, nil
	}
	return nil, errors.New("no word received")
}

// LightAppSvc.mini_app_info.GetAppInfoById
func decodeAppInfoResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	pkg := qweb.QWebRsp{}
	rsp := qweb.GetAppInfoByIdRsp{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if pkg.RetCode != 0 {
		return nil, errors.New(pkg.ErrMsg)
	}
	if err := proto.Unmarshal(pkg.BusiBuff, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp.AppInfo, nil
}

func ignoreDecoder(_ *QQClient, _ *incomingPacketInfo, _ []byte) (interface{}, error) {
	return nil, nil
}
