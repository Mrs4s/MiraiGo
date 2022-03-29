package client

import (
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/msf"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	decoders["StatSvc.GetDevLoginInfo"] = decodeDevListResponse
	decoders["StatSvc.SvcReqMSFLoginNotify"] = decodeLoginNotifyPacket
	decoders["RegPrxySvc.getOffMsg"] = ignoreDecoder
	decoders["RegPrxySvc.GetMsgV2"] = ignoreDecoder
	decoders["RegPrxySvc.PbGetMsg"] = ignoreDecoder
	decoders["RegPrxySvc.NoticeEnd"] = ignoreDecoder
	decoders["RegPrxySvc.PushParam"] = decodePushParamPacket
	decoders["RegPrxySvc.PbSyncMsg"] = decodeMsgSyncResponse
	decoders["PbMessageSvc.PbMsgReadedReport"] = decodeMsgReadedResponse
	decoders["MessageSvc.PushReaded"] = ignoreDecoder
	decoders["OnlinePush.PbC2CMsgSync"] = decodeC2CSyncPacket
}

type (
	// SessionSyncResponse 会话同步结果
	SessionSyncResponse struct {
		GroupSessions []*GroupSessionInfo
	}

	// GroupSessionInfo 群会话信息
	GroupSessionInfo struct {
		GroupCode      int64
		UnreadCount    uint32
		LatestMessages []*message.GroupMessage
	}

	sessionSyncEvent struct {
		IsEnd         bool
		GroupNum      int32
		GroupSessions []*GroupSessionInfo
	}
)

// GetAllowedClients 获取已允许的其他客户端
func (c *QQClient) GetAllowedClients() ([]*OtherClientInfo, error) {
	i, err := c.sendAndWait(c.buildDeviceListRequestPacket())
	if err != nil {
		return nil, err
	}
	list := i.([]jce.SvcDevLoginInfo)
	ret := make([]*OtherClientInfo, 0, len(list))
	for _, l := range list {
		ret = append(ret, &OtherClientInfo{
			AppId:      l.AppId,
			DeviceName: l.DeviceName,
			DeviceKind: l.DeviceTypeInfo,
		})
	}
	return ret, nil
}

// RefreshStatus 刷新客户端状态
func (c *QQClient) RefreshStatus() error {
	_, err := c.sendAndWait(c.buildGetOfflineMsgRequestPacket())
	return err
}

// SyncSessions 同步会话列表
func (c *QQClient) SyncSessions() (*SessionSyncResponse, error) {
	ret := &SessionSyncResponse{}
	notifyChan := make(chan bool, 4)
	var groupNum int32 = -1
	stop := c.waitPacket("RegPrxySvc.PbSyncMsg", func(i any, err error) {
		if err != nil {
			return
		}
		e := i.(*sessionSyncEvent)
		if len(e.GroupSessions) > 0 {
			ret.GroupSessions = append(ret.GroupSessions, e.GroupSessions...)
		}
		if e.GroupNum != -1 {
			groupNum = e.GroupNum
		}
		c.debug("sync session %v/%v", len(ret.GroupSessions), groupNum)
		if groupNum != -1 && len(ret.GroupSessions) >= int(groupNum) {
			notifyChan <- true
		}
	})
	_, pkt := c.buildSyncMsgRequestPacket()
	if err := c.sendPacket(pkt); err != nil {
		stop()
		return nil, err
	}
	select {
	case <-notifyChan:
		stop()
	case <-time.After(time.Second * 3):
		stop()
	}
	return ret, nil
}

// MarkGroupMessageReaded 标记群消息已读, 适当调用应该能减少风控
func (c *QQClient) MarkGroupMessageReaded(groupCode, seq int64) {
	_, _ = c.sendAndWait(c.buildGroupMsgReadedPacket(groupCode, seq))
}

func (c *QQClient) MarkPrivateMessageReaded(uin, time int64) {
	_, _ = c.sendAndWait(c.buildPrivateMsgReadedPacket(uin, time))
}

// StatSvc.GetDevLoginInfo
func (c *QQClient) buildDeviceListRequestPacket() (uint16, []byte) {
	req := &jce.SvcReqGetDevLoginInfo{
		Guid:           c.deviceInfo.Guid,
		LoginType:      1,
		AppName:        "com.tencent.mobileqq",
		RequireMax:     20,
		GetDevListType: 2,
	}
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{"SvcReqGetDevLoginInfo": packUniRequestData(req.ToBytes())}}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "StatSvc",
		SFuncName:    "SvcReqGetDevLoginInfo",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	return c.uniPacket("StatSvc.GetDevLoginInfo", pkt.ToBytes())
}

// RegPrxySvc.getOffMsg
func (c *QQClient) buildGetOfflineMsgRequestPacket() (uint16, []byte) {
	regReq := &jce.SvcReqRegisterNew{
		RequestOptional: 0x101C2 | 32,
		C2CMsg: &jce.SvcReqGetMsgV2{
			Uin: c.Uin,
			DateTime: func() int32 {
				t := c.stat.LastMessageTime.Load()
				if t == 0 {
					return 1
				}
				return int32(t)
			}(),
			RecivePic:        1,
			Ability:          15,
			Channel:          4,
			Inst:             1,
			ChannelEx:        1,
			SyncCookie:       c.sig.SyncCookie,
			SyncFlag:         0, // START
			RambleFlag:       0,
			GeneralAbi:       1,
			PubAccountCookie: c.sig.PubAccountCookie,
		},
		GroupMsg: &jce.SvcReqPullGroupMsgSeq{
			VerifyType: 0,
			Filter:     1, // LIMIT_10_AND_IN_3_DAYS
		},
		EndSeq: time.Now().Unix(),
	}
	flag := msg.SyncFlag_START
	msgReq, _ := proto.Marshal(&msg.GetMessageRequest{
		SyncFlag:           &flag,
		SyncCookie:         c.sig.SyncCookie,
		RambleFlag:         proto.Int32(0),
		ContextFlag:        proto.Int32(1),
		OnlineSyncFlag:     proto.Int32(0),
		LatestRambleNumber: proto.Int32(20),
		OtherRambleNumber:  proto.Int32(3),
	})
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{
		"req_PbOffMsg": jce.NewJceWriter().WriteBytes(append([]byte{0, 0, 0, 0}, msgReq...), 0).Bytes(),
		"req_OffMsg":   packUniRequestData(regReq.ToBytes()),
	}}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "RegPrxySvc",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	return c.uniPacket("RegPrxySvc.getOffMsg", pkt.ToBytes())
}

// RegPrxySvc.PbSyncMsg
func (c *QQClient) buildSyncMsgRequestPacket() (uint16, []byte) {
	oidbReq, _ := proto.Marshal(&oidb.D769RspBody{
		ConfigList: []*oidb.D769ConfigSeq{
			{
				Type:    proto.Uint32(46),
				Version: proto.Uint32(0),
			},
			{
				Type:    proto.Uint32(283),
				Version: proto.Uint32(0),
			},
		},
	})
	regReq := &jce.SvcReqRegisterNew{
		RequestOptional:   128 | 64 | 256 | 2 | 8192 | 16384 | 65536,
		DisGroupMsgFilter: 1,
		C2CMsg: &jce.SvcReqGetMsgV2{
			Uin: c.Uin,
			DateTime: func() int32 {
				t := c.stat.LastMessageTime.Load()
				if t == 0 {
					return 1
				}
				return int32(t)
			}(),
			RecivePic:        1,
			Ability:          15,
			Channel:          4,
			Inst:             1,
			ChannelEx:        1,
			SyncCookie:       c.sig.SyncCookie,
			SyncFlag:         0, // START
			RambleFlag:       0,
			GeneralAbi:       1,
			PubAccountCookie: c.sig.PubAccountCookie,
		},
		GroupMask: 2,
		EndSeq:    int64(rand.Uint32()),
		O769Body:  oidbReq,
	}
	flag := msg.SyncFlag_START
	msgReq := &msg.GetMessageRequest{
		SyncFlag:           &flag,
		SyncCookie:         c.sig.SyncCookie,
		RambleFlag:         proto.Int32(0),
		ContextFlag:        proto.Int32(1),
		OnlineSyncFlag:     proto.Int32(0),
		LatestRambleNumber: proto.Int32(20),
		OtherRambleNumber:  proto.Int32(3),
		MsgReqType:         proto.Int32(1),
	}
	offMsg, _ := proto.Marshal(msgReq)
	msgReq.MsgReqType = proto.Int32(2)
	msgReq.SyncCookie = nil
	msgReq.PubaccountCookie = c.sig.PubAccountCookie
	pubMsg, _ := proto.Marshal(msgReq)
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{
		"req_PbOffMsg": jce.NewJceWriter().WriteBytes(append([]byte{0, 0, 0, 0}, offMsg...), 0).Bytes(),
		"req_PbPubMsg": jce.NewJceWriter().WriteBytes(append([]byte{0, 0, 0, 0}, pubMsg...), 0).Bytes(),
		"req_OffMsg":   packUniRequestData(regReq.ToBytes()),
	}}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "RegPrxySvc",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	return c.uniPacket("RegPrxySvc.infoSync", pkt.ToBytes())
}

// PbMessageSvc.PbMsgReadedReport
func (c *QQClient) buildGroupMsgReadedPacket(groupCode, msgSeq int64) (uint16, []byte) {
	req, _ := proto.Marshal(&msg.PbMsgReadedReportReq{GrpReadReport: []*msg.PbGroupReadedReportReq{{
		GroupCode:   proto.Uint64(uint64(groupCode)),
		LastReadSeq: proto.Uint64(uint64(msgSeq)),
	}}})
	return c.uniPacket("PbMessageSvc.PbMsgReadedReport", req)
}

func (c *QQClient) buildPrivateMsgReadedPacket(uin, time int64) (uint16, []byte) {
	req, _ := proto.Marshal(&msg.PbMsgReadedReportReq{C2CReadReport: &msg.PbC2CReadedReportReq{PairInfo: []*msg.UinPairReadInfo{
		{
			PeerUin:      proto.Uint64(uint64(uin)),
			LastReadTime: proto.Uint32(uint32(time)),
		},
	}, SyncCookie: c.sig.SyncCookie}})
	return c.uniPacket("PbMessageSvc.PbMsgReadedReport", req)
}

// StatSvc.GetDevLoginInfo
func decodeDevListResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	rsp := jce.NewJceReader(data.Map["SvcRspGetDevLoginInfo"]["QQService.SvcRspGetDevLoginInfo"][1:])
	d := rsp.ReadSvcDevLoginInfos(4)
	if len(d) > 0 {
		return d, nil
	}
	d = rsp.ReadSvcDevLoginInfos(5)
	if len(d) > 0 {
		return d, nil
	}
	d = rsp.ReadSvcDevLoginInfos(6)
	if len(d) > 0 {
		return d, nil
	}
	return nil, errors.New("not any device")
}

// RegPrxySvc.PushParam
func decodePushParamPacket(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	reader := jce.NewJceReader(data.Map["SvcRespParam"]["RegisterProxySvcPack.SvcRespParam"][1:])
	rsp := &jce.SvcRespParam{}
	rsp.ReadFrom(reader)
	allowedClients, _ := c.GetAllowedClients()
	c.OnlineClients = []*OtherClientInfo{}
	for _, i := range rsp.OnlineInfos {
		c.OnlineClients = append(c.OnlineClients, &OtherClientInfo{
			AppId: int64(i.InstanceId),
			DeviceName: func() string {
				for _, ac := range allowedClients {
					if ac.AppId == int64(i.InstanceId) {
						return ac.DeviceName
					}
				}
				return i.SubPlatform
			}(),
			DeviceKind: func() string {
				switch i.UClientType {
				case 65793:
					return "Windows"
				case 65805, 68104:
					return "aPad"
				case 66818, 66831, 81154:
					return "Mac"
				case 68361, 72194:
					return "iPad"
				case 75023, 78082, 78096:
					return "Watch"
				case 77313:
					return "Windows TIM"
				default:
					return i.SubPlatform
				}
			}(),
		})
	}
	return nil, nil
}

// RegPrxySvc.PbSyncMsg
func decodeMsgSyncResponse(c *QQClient, info *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := &msf.SvcRegisterProxyMsgResp{}
	if err := proto.Unmarshal(payload, rsp); err != nil {
		return nil, err
	}
	ret := &sessionSyncEvent{
		IsEnd:    (rsp.GetFlag() & 2) == 2,
		GroupNum: -1,
	}
	if rsp.Info != nil {
		ret.GroupNum = int32(rsp.Info.GetGroupNum())
	}
	if len(rsp.GroupMsg) > 0 {
		for _, gm := range rsp.GroupMsg {
			gmRsp := &msg.GetGroupMsgResp{}
			if err := proto.Unmarshal(gm.Content[4:], gmRsp); err != nil {
				continue
			}
			var latest []*message.GroupMessage
			for _, m := range gmRsp.Msg {
				if m.Head.GetFromUin() != 0 {
					pm := c.parseGroupMessage(m)
					if pm != nil {
						latest = append(latest, pm)
					}
				}
			}
			ret.GroupSessions = append(ret.GroupSessions, &GroupSessionInfo{
				GroupCode:      int64(gmRsp.GetGroupCode()),
				UnreadCount:    uint32(gmRsp.GetReturnEndSeq() - gm.GetMemberSeq()),
				LatestMessages: latest,
			})
		}
	}
	if len(rsp.C2CMsg) > 4 {
		c2cRsp := &msg.GetMessageResponse{}
		if proto.Unmarshal(rsp.C2CMsg[4:], c2cRsp) == nil {
			c.c2cMessageSyncProcessor(c2cRsp, info)
		}
	}
	return ret, nil
}

// OnlinePush.PbC2CMsgSync
func decodeC2CSyncPacket(c *QQClient, info *network.IncomingPacketInfo, payload []byte) (any, error) {
	m := msg.PbPushMsg{}
	if err := proto.Unmarshal(payload, &m); err != nil {
		return nil, err
	}
	_ = c.sendPacket(c.buildDeleteOnlinePushPacket(c.Uin, m.GetSvrip(), m.PushToken, info.SequenceId, nil))
	c.commMsgProcessor(m.Msg, info)
	return nil, nil
}

func decodeMsgReadedResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := msg.PbMsgReadedReportResp{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(rsp.GrpReadReport) > 0 {
		return rsp.GrpReadReport[0].GetResult() == 0, nil
	}
	return nil, nil
}

var loginNotifyLock sync.Mutex

// StatSvc.SvcReqMSFLoginNotify
func decodeLoginNotifyPacket(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	reader := jce.NewJceReader(data.Map["SvcReqMSFLoginNotify"]["QQService.SvcReqMSFLoginNotify"][1:])
	notify := &jce.SvcReqMSFLoginNotify{}
	notify.ReadFrom(reader)
	loginNotifyLock.Lock()
	defer loginNotifyLock.Unlock()
	if notify.Status == 1 {
		found := false
		for _, oc := range c.OnlineClients {
			if oc.AppId == notify.AppId {
				found = true
			}
		}
		if !found {
			allowedClients, _ := c.GetAllowedClients()
			for _, ac := range allowedClients {
				t := ac
				if ac.AppId == notify.AppId {
					c.OnlineClients = append(c.OnlineClients, t)
					c.OtherClientStatusChangedEvent.dispatch(c, &OtherClientStatusChangedEvent{
						Client: t,
						Online: true,
					})
					break
				}
			}
		}
	}
	if notify.Status == 2 {
		rmi := -1
		for i, oc := range c.OnlineClients {
			if oc.AppId == notify.AppId {
				rmi = i
			}
		}
		if rmi != -1 {
			rmc := c.OnlineClients[rmi]
			c.OnlineClients = append(c.OnlineClients[:rmi], c.OnlineClients[rmi+1:]...)
			c.OtherClientStatusChangedEvent.dispatch(c, &OtherClientStatusChangedEvent{
				Client: rmc,
				Online: false,
			})
		}
	}
	return nil, nil
}
