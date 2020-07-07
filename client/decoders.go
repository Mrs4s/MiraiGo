package client

import (
	"errors"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/golang/protobuf/proto"
	"time"
)

func decodeLoginResponse(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
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
				Success: false,
				Error:   UnknownLoginError,
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

	if t == 160 {

	}

	if t == 204 {
		return LoginResponse{
			Success: false,
			Error:   DeviceLockError,
		}, nil
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

	return nil, nil // ?
}

func decodeClientRegisterResponse(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	return nil, nil
}

func decodePushReqPacket(c *QQClient, s uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["PushReq"]["ConfigPush.PushReq"][1:])
	jceBuf := []byte{}
	t := r.ReadInt32(1)
	r.ReadSlice(&jceBuf, 2)
	seq := r.ReadInt64(3)
	_, pkt := c.buildConfPushRespPacket(t, seq, jceBuf)
	return nil, c.send(pkt)
}

func decodeMessageSvcPacket(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
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
			if message.Body.RichText == nil || message.Body.RichText.Elems == nil {
				continue
			}
			if c.lastMessageSeq >= message.Head.MsgSeq {
				continue
			}
			c.lastMessageSeq = message.Head.MsgSeq

			c.dispatchFriendMessage(c.parsePrivateMessage(message))
		}
	}
	_, _ = c.sendAndWait(c.buildDeleteMessageRequestPacket(delItems))
	if rsp.SyncFlag != msg.SyncFlag_STOP {
		seq, pkt := c.buildGetMessageRequestPacket(rsp.SyncFlag, time.Now().Unix())
		_, _ = c.sendAndWait(seq, pkt)
	}
	return nil, err
}

func decodeGroupMessagePacket(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	pkt := msg.PushMessagePacket{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, err
	}
	if pkt.Message.Head.FromUin == c.Uin {
		return nil, nil
	}
	c.dispatchGroupMessage(c.parseGroupMessage(pkt.Message))
	return nil, nil
}

func decodeSvcNotify(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	_, pkt := c.buildGetMessageRequestPacket(msg.SyncFlag_START, time.Now().Unix())
	return nil, c.send(pkt)
}

func decodeFriendGroupListResponse(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
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

func decodeGroupListResponse(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GetTroopListRespV2"][1:])
	groups := []jce.TroopNumber{}
	r.ReadSlice(&groups, 5)
	var l []*GroupInfo
	for _, g := range groups {
		l = append(l, &GroupInfo{
			Uin:            g.GroupUin,
			Code:           g.GroupCode,
			Name:           g.GroupName,
			Memo:           g.GroupMemo,
			OwnerUin:       uint32(g.GroupOwnerUin),
			MemberCount:    uint16(g.MemberNum),
			MaxMemberCount: uint16(g.MaxGroupMemberNum),
		})
	}
	return l, nil
}

func decodeGroupMemberListResponse(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion3{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	r := jce.NewJceReader(data.Map["GTMLRESP"][1:])
	members := []jce.TroopMemberInfo{}
	r.ReadSlice(&members, 3)
	next := r.ReadInt64(4)
	var l []GroupMemberInfo
	for _, m := range members {
		l = append(l, GroupMemberInfo{
			Uin:                    m.MemberUin,
			Nickname:               m.Nick,
			CardName:               m.Name,
			Level:                  uint16(m.MemberLevel),
			JoinTime:               m.JoinTime,
			LastSpeakTime:          m.LastSpeakTime,
			SpecialTitle:           m.SpecialTitle,
			SpecialTitleExpireTime: m.SpecialTitleExpireTime,
			Job:                    m.Job,
		})
	}
	return groupMemberListResponse{
		NextUin: next,
		list:    l,
	}, nil
}

func decodeGroupImageStoreResponse(c *QQClient, seq uint16, payload []byte) (interface{}, error) {
	pkt := pb.D388RespBody{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, err
	}
	rsp := pkt.MsgTryupImgRsp[0]
	if rsp.Result != 0 {
		return groupImageUploadResponse{
			ResultCode: rsp.Result,
			Message:    rsp.FailMsg,
		}, nil
	}
	if rsp.BoolFileExit {
		return groupImageUploadResponse{IsExists: true}, nil
	}
	return groupImageUploadResponse{
		UploadKey:  rsp.UpUkey,
		UploadIp:   rsp.Uint32UpIp,
		UploadPort: rsp.Uint32UpPort,
	}, nil
}

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
					GroupUin:    groupId,
					OperatorUin: operator,
					TargetUin:   target,
					Time:        t,
				})
			case 0x11: // 撤回消息
				r.ReadByte()
				b := pb.NotifyMsgBody{}
				_ = proto.Unmarshal(r.ReadAvailable(), &b)
				if b.OptMsgRecall == nil {
					continue
				}
				for _, rm := range b.OptMsgRecall.RecalledMsgList {
					c.dispatchGroupMessageRecalledEvent(&GroupMessageRecalledEvent{
						GroupUin:    groupId,
						OperatorUin: b.OptMsgRecall.Uin,
						AuthorUin:   rm.AuthorUin,
						MessageId:   rm.Seq,
						Time:        rm.Time,
					})
				}
			}
		}
	}

	return nil, nil
}
