package client

import (
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"sync"
	"time"
)

func init() {
	decoders["StatSvc.GetDevLoginInfo"] = decodeDevListResponse
	decoders["StatSvc.SvcReqMSFLoginNotify"] = decodeLoginNotifyPacket
	decoders["RegPrxySvc.getOffMsg"] = ignoreDecoder
	decoders["RegPrxySvc.GetMsgV2"] = ignoreDecoder
	decoders["RegPrxySvc.PbGetMsg"] = ignoreDecoder
	decoders["RegPrxySvc.NoticeEnd"] = ignoreDecoder
	decoders["RegPrxySvc.PushParam"] = decodePushParamPacket
}

// GetAllowedClients 获取已允许的其他客户端
func (c *QQClient) GetAllowedClients() ([]*OtherClientInfo, error) {
	i, err := c.sendAndWait(c.buildDeviceListRequestPacket())
	if err != nil {
		return nil, err
	}
	list := i.([]jce.SvcDevLoginInfo)
	var ret []*OtherClientInfo
	for _, l := range list {
		ret = append(ret, &OtherClientInfo{
			AppId:      l.AppId,
			DeviceName: l.DeviceName,
			DeviceKind: l.DeviceTypeInfo,
		})
	}
	return ret, nil
}

// RefreshClientStatus 刷新客户端状态
func (c *QQClient) RefreshStatus() error {
	_, err := c.sendAndWait(c.buildGetOfflineMsgRequest())
	return err
}

// StatSvc.GetDevLoginInfo
func (c *QQClient) buildDeviceListRequestPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := &jce.SvcReqGetDevLoginInfo{
		Guid:           SystemDeviceInfo.Guid,
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
	packet := packets.BuildUniPacket(c.Uin, seq, "StatSvc.GetDevLoginInfo", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// RegPrxySvc.getOffMsg
func (c *QQClient) buildGetOfflineMsgRequest() (uint16, []byte) {
	seq := c.nextSeq()
	regReq := &jce.SvcReqRegisterNew{
		RequestOptional: 0x101C2 | 32,
		C2CMsg: &jce.SvcReqGetMsgV2{
			Uin: c.Uin,
			DateTime: func() int32 {
				if c.stat.LastMessageTime == 0 {
					return 1
				}
				return int32(c.stat.LastMessageTime)
			}(),
			RecivePic:        1,
			Ability:          15,
			Channel:          4,
			Inst:             1,
			ChannelEx:        1,
			SyncCookie:       c.syncCookie,
			SyncFlag:         0, // START
			RambleFlag:       0,
			GeneralAbi:       1,
			PubAccountCookie: c.pubAccountCookie,
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
		SyncCookie:         c.syncCookie,
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
	packet := packets.BuildUniPacket(c.Uin, seq, "RegPrxySvc.getOffMsg", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// StatSvc.GetDevLoginInfo
func decodeDevListResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	rsp := jce.NewJceReader(data.Map["SvcRspGetDevLoginInfo"]["QQService.SvcRspGetDevLoginInfo"][1:])
	d := []jce.SvcDevLoginInfo{}
	rsp.ReadSlice(&d, 4)
	if len(d) > 0 {
		return d, nil
	}
	rsp.ReadSlice(&d, 5)
	if len(d) > 0 {
		return d, nil
	}
	rsp.ReadSlice(&d, 6)
	if len(d) > 0 {
		return d, nil
	}
	return nil, errors.New("not any device")
}

// RegPrxySvc.PushParam
func decodePushParamPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
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

var loginNotifyLock sync.Mutex

// StatSvc.SvcReqMSFLoginNotify
func decodeLoginNotifyPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
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
					c.dispatchOtherClientStatusChangedEvent(&OtherClientStatusChangedEvent{
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
			c.dispatchOtherClientStatusChangedEvent(&OtherClientStatusChangedEvent{
				Client: rmc,
				Online: false,
			})
		}
	}
	return nil, nil
}
