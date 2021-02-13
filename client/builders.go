package client

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x352"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/client/pb/profilecard"
	"github.com/Mrs4s/MiraiGo/client/pb/qweb"
	"github.com/Mrs4s/MiraiGo/client/pb/structmsg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/crypto"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/protocol/tlv"
	"github.com/golang/protobuf/proto"
)

var (
	syncConst1 = rand.Int63()
	syncConst2 = rand.Int63()
)

func (c *QQClient) buildLoginPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x0810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(9)
		if c.AllowSlider {
			w.WriteUInt16(0x17)
		} else {
			w.WriteUInt16(0x16)
		}

		w.Write(tlv.T18(16, uint32(c.Uin)))
		w.Write(tlv.T1(uint32(c.Uin), SystemDeviceInfo.IpAddress))
		w.Write(tlv.T106(uint32(c.Uin), 0, c.version.AppId, c.version.SSOVersion, c.PasswordMd5, true, SystemDeviceInfo.Guid, SystemDeviceInfo.TgtgtKey, 0))
		w.Write(tlv.T116(c.version.MiscBitmap, c.version.SubSigmap))
		w.Write(tlv.T100(c.version.SSOVersion, c.version.AppId, c.version.MainSigMap))
		w.Write(tlv.T107(0))
		w.Write(tlv.T142(c.version.ApkId))
		w.Write(tlv.T144(
			[]byte(SystemDeviceInfo.IMEI),
			SystemDeviceInfo.GenDeviceInfoData(),
			SystemDeviceInfo.OSType,
			SystemDeviceInfo.Version.Release,
			SystemDeviceInfo.SimInfo,
			SystemDeviceInfo.APN,
			false, true, false, tlv.GuidFlag(),
			SystemDeviceInfo.Model,
			SystemDeviceInfo.Guid,
			SystemDeviceInfo.Brand,
			SystemDeviceInfo.TgtgtKey,
		))

		w.Write(tlv.T145(SystemDeviceInfo.Guid))
		w.Write(tlv.T147(16, []byte(c.version.SortVersionName), c.version.ApkSign))
		/*
			if (miscBitMap & 0x80) != 0{
				w.Write(tlv.T166(1))
			}
		*/
		w.Write(tlv.T154(seq))
		w.Write(tlv.T141(SystemDeviceInfo.SimInfo, SystemDeviceInfo.APN))
		w.Write(tlv.T8(2052))
		w.Write(tlv.T511([]string{
			"tenpay.com", "openmobile.qq.com", "docs.qq.com", "connect.qq.com",
			"qzone.qq.com", "vip.qq.com", "gamecenter.qq.com", "qun.qq.com", "game.qq.com",
			"qqweb.qq.com", "office.qq.com", "ti.qq.com", "mail.qq.com", "mma.qq.com",
		}))

		w.Write(tlv.T187(SystemDeviceInfo.MacAddress))
		w.Write(tlv.T188(SystemDeviceInfo.AndroidId))
		if len(SystemDeviceInfo.IMSIMd5) != 0 {
			w.Write(tlv.T194(SystemDeviceInfo.IMSIMd5))
		}
		if c.AllowSlider {
			w.Write(tlv.T191(0x82))
		}
		if len(SystemDeviceInfo.WifiBSSID) != 0 && len(SystemDeviceInfo.WifiSSID) != 0 {
			w.Write(tlv.T202(SystemDeviceInfo.WifiBSSID, SystemDeviceInfo.WifiSSID))
		}
		w.Write(tlv.T177(c.version.BuildTime, c.version.SdkVersion))
		w.Write(tlv.T516())
		w.Write(tlv.T521())
		w.Write(tlv.T525(tlv.T536([]byte{0x01, 0x00})))
	})
	sso := packets.BuildSsoPacket(seq, c.version.AppId, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
	return seq, packet
}

func (c *QQClient) buildDeviceLockLoginPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x0810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(20)
		w.WriteUInt16(4)

		w.Write(tlv.T8(2052))
		w.Write(tlv.T104(c.t104))
		w.Write(tlv.T116(c.version.MiscBitmap, c.version.SubSigmap))
		w.Write(tlv.T401(c.g))
	})
	sso := packets.BuildSsoPacket(seq, c.version.AppId, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
	return seq, packet
}

func (c *QQClient) buildCaptchaPacket(result string, sign []byte) (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(2) // sub command
		w.WriteUInt16(4)

		w.Write(tlv.T2(result, sign))
		w.Write(tlv.T8(2052))
		w.Write(tlv.T104(c.t104))
		w.Write(tlv.T116(c.version.MiscBitmap, c.version.SubSigmap))
	})
	sso := packets.BuildSsoPacket(seq, c.version.AppId, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
	return seq, packet
}

func (c *QQClient) buildSMSRequestPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(8)
		w.WriteUInt16(6)

		w.Write(tlv.T8(2052))
		w.Write(tlv.T104(c.t104))
		w.Write(tlv.T116(c.version.MiscBitmap, c.version.SubSigmap))
		w.Write(tlv.T174(c.t174))
		w.Write(tlv.T17A(9))
		w.Write(tlv.T197())
	})
	sso := packets.BuildSsoPacket(seq, c.version.AppId, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
	return seq, packet
}

func (c *QQClient) buildSMSCodeSubmitPacket(code string) (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(7)
		w.WriteUInt16(7)

		w.Write(tlv.T8(2052))
		w.Write(tlv.T104(c.t104))
		w.Write(tlv.T116(c.version.MiscBitmap, c.version.SubSigmap))
		w.Write(tlv.T174(c.t174))
		w.Write(tlv.T17C(code))
		w.Write(tlv.T401(c.g))
		w.Write(tlv.T198())
	})
	sso := packets.BuildSsoPacket(seq, c.version.AppId, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
	return seq, packet
}

func (c *QQClient) buildTicketSubmitPacket(ticket string) (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(2)
		w.WriteUInt16(4)

		w.Write(tlv.T193(ticket))
		w.Write(tlv.T8(2052))
		w.Write(tlv.T104(c.t104))
		w.Write(tlv.T116(c.version.MiscBitmap, c.version.SubSigmap))
	})
	sso := packets.BuildSsoPacket(seq, c.version.AppId, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
	return seq, packet
}

func (c *QQClient) buildRequestTgtgtNopicsigPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x0810, crypto.NewEncryptSession(c.sigInfo.t133), c.sigInfo.wtSessionTicketKey, func(w *binary.Writer) {
		w.WriteUInt16(15)
		w.WriteUInt16(24)

		w.Write(tlv.T18(16, uint32(c.Uin)))
		w.Write(tlv.T1(uint32(c.Uin), SystemDeviceInfo.IpAddress))
		w.Write(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt16(0x106)
			w.WriteTlv(c.sigInfo.encryptedA1)
		}))
		w.Write(tlv.T116(c.version.MiscBitmap, c.version.SubSigmap))
		w.Write(tlv.T100(c.version.SSOVersion, 2, c.version.MainSigMap))
		w.Write(tlv.T107(0))
		w.Write(tlv.T144(
			SystemDeviceInfo.AndroidId,
			SystemDeviceInfo.GenDeviceInfoData(),
			SystemDeviceInfo.OSType,
			SystemDeviceInfo.Version.Release,
			SystemDeviceInfo.SimInfo,
			SystemDeviceInfo.APN,
			false, true, false, tlv.GuidFlag(),
			SystemDeviceInfo.Model,
			SystemDeviceInfo.Guid,
			SystemDeviceInfo.Brand,
			SystemDeviceInfo.TgtgtKey,
		))
		w.Write(tlv.T142(c.version.ApkId))
		w.Write(tlv.T145(SystemDeviceInfo.Guid))
		w.Write(tlv.T16A(c.sigInfo.srmToken))
		w.Write(tlv.T154(seq))
		w.Write(tlv.T141(SystemDeviceInfo.SimInfo, SystemDeviceInfo.APN))
		w.Write(tlv.T8(2052))
		w.Write(tlv.T511([]string{
			"tenpay.com", "openmobile.qq.com", "docs.qq.com", "connect.qq.com",
			"qzone.qq.com", "vip.qq.com", "qun.qq.com", "game.qq.com", "qqweb.qq.com",
			"office.qq.com", "ti.qq.com", "mail.qq.com", "qzone.com", "mma.qq.com",
		}))
		w.Write(tlv.T147(16, []byte(c.version.SortVersionName), c.version.ApkSign))
		w.Write(tlv.T177(c.version.BuildTime, c.version.SdkVersion))
		w.Write(tlv.T400(c.g, c.Uin, SystemDeviceInfo.Guid, c.dpwd, 1, 16, c.randSeed))
		w.Write(tlv.T187(SystemDeviceInfo.MacAddress))
		w.Write(tlv.T188(SystemDeviceInfo.AndroidId))
		w.Write(tlv.T194(SystemDeviceInfo.IMSIMd5))
		w.Write(tlv.T202(SystemDeviceInfo.WifiBSSID, SystemDeviceInfo.WifiSSID))
		w.Write(tlv.T516())
		w.Write(tlv.T521())
		w.Write(tlv.T525(tlv.T536([]byte{0x01, 0x00})))

	})
	packet := packets.BuildUniPacket(c.Uin, seq, "wtlogin.exchange_emp", 2, c.OutGoingPacketSessionId, []byte{}, make([]byte, 16), req)
	return seq, packet
}

// StatSvc.register
func (c *QQClient) buildClientRegisterPacket() (uint16, []byte) {
	seq := c.nextSeq()
	svc := &jce.SvcReqRegister{
		ConnType:     0,
		Uin:          c.Uin,
		Bid:          1 | 2 | 4,
		Status:       11,
		KickPC:       0,
		KickWeak:     0,
		IOSVersion:   int64(SystemDeviceInfo.Version.Sdk),
		NetType:      1,
		RegType:      0,
		Guid:         SystemDeviceInfo.Guid,
		IsSetStatus:  0,
		LocaleId:     2052,
		DevName:      string(SystemDeviceInfo.Model),
		DevType:      string(SystemDeviceInfo.Model),
		OSVer:        string(SystemDeviceInfo.Version.Release),
		OpenPush:     1,
		LargeSeq:     1551,
		OldSSOIp:     0,
		NewSSOIp:     31806887127679168,
		ChannelNo:    "",
		CPID:         0,
		VendorName:   string(SystemDeviceInfo.VendorName),
		VendorOSName: string(SystemDeviceInfo.VendorOSName),
		B769:         []byte{0x0A, 0x04, 0x08, 0x2E, 0x10, 0x00, 0x0A, 0x05, 0x08, 0x9B, 0x02, 0x10, 0x00},
		SetMute:      0,
	}
	b := append([]byte{0x0A}, svc.ToBytes()...)
	b = append(b, 0x0B)
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"SvcReqRegister": b},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "PushService",
		SFuncName:    "SvcReqRegister",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	sso := packets.BuildSsoPacket(seq, c.version.AppId, "StatSvc.register", SystemDeviceInfo.IMEI, c.sigInfo.tgt, c.OutGoingPacketSessionId, pkt.ToBytes(), c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 1, c.sigInfo.d2Key, sso, c.sigInfo.d2)
	return seq, packet
}

// ConfigPushSvc.PushResp
func (c *QQClient) buildConfPushRespPacket(t int32, pktSeq int64, jceBuf []byte) (uint16, []byte) {
	seq := c.nextSeq()
	req := jce.NewJceWriter()
	req.WriteInt32(t, 1)
	req.WriteInt64(pktSeq, 2)
	req.WriteBytes(jceBuf, 3)
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"PushResp": packUniRequestData(req.Bytes())},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "QQService.ConfigPushSvc.MainServant",
		SFuncName:    "PushResp",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "ConfigPushSvc.PushResp", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// friendlist.getFriendGroupList
func (c *QQClient) buildFriendGroupListRequestPacket(friendStartIndex, friendListCount, groupStartIndex, groupListCount int16) (uint16, []byte) {
	seq := c.nextSeq()
	d50, _ := proto.Marshal(&pb.D50ReqBody{
		Appid:                   1002,
		ReqMusicSwitch:          1,
		ReqMutualmarkAlienation: 1,
		ReqKsingSwitch:          1,
		ReqMutualmarkLbsshare:   1,
	})
	req := &jce.FriendListRequest{
		Reqtype: 3,
		IfReflush: func() byte {
			if friendStartIndex <= 0 {
				return 0
			}
			return 1
		}(),
		Uin:         c.Uin,
		StartIndex:  friendStartIndex,
		FriendCount: friendListCount,
		GroupId:     0,
		IfGetGroupInfo: func() byte {
			if groupListCount <= 0 {
				return 0
			}
			return 1
		}(),
		GroupStartIndex: byte(groupStartIndex),
		GroupCount:      byte(groupListCount),
		IfGetMSFGroup:   0,
		IfShowTermType:  1,
		Version:         27,
		UinList:         nil,
		AppType:         0,
		IfGetDOVId:      0,
		IfGetBothFlag:   0,
		D50:             d50,
		D6B:             []byte{},
		SnsTypeList:     []int64{13580, 13581, 13582},
	}
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"FL": packUniRequestData(req.ToBytes())},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		CPacketType:  0x003,
		IRequestId:   1921334514,
		SServantName: "mqq.IMService.FriendListServiceServantObj",
		SFuncName:    "GetFriendListReq",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "friendlist.getFriendGroupList", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// SummaryCard.ReqSummaryCard
func (c *QQClient) buildSummaryCardRequestPacket(target int64) (uint16, []byte) {
	seq := c.nextSeq()
	packBusinessBuf := func(t int32, buf []byte) []byte {
		return binary.NewWriterF(func(w *binary.Writer) {
			comm, _ := proto.Marshal(&profilecard.BusiComm{
				Ver:      proto.Int32(1),
				Seq:      proto.Int32(int32(seq)),
				Fromuin:  &c.Uin,
				Touin:    &target,
				Service:  &t,
				Platform: proto.Int32(2),
				Qqver:    proto.String("8.4.18.4945"),
				Build:    proto.Int32(4945),
			})
			w.WriteByte(40)
			w.WriteUInt32(uint32(len(comm)))
			w.WriteUInt32(uint32(len(buf)))
			w.Write(comm)
			w.Write(buf)
			w.WriteByte(41)
		})
	}
	gate, _ := proto.Marshal(&profilecard.GateVaProfileGateReq{
		UCmd:           proto.Int32(3),
		StPrivilegeReq: &profilecard.GatePrivilegeBaseInfoReq{UReqUin: &target},
		StGiftReq:      &profilecard.GateGetGiftListReq{Uin: proto.Int32(int32(target))},
		StVipCare:      &profilecard.GateGetVipCareReq{Uin: &target},
		OidbFlag: []*profilecard.GateOidbFlagInfo{
			{
				Fieled: proto.Int32(42334),
			},
			{
				Fieled: proto.Int32(42340),
			},
			{
				Fieled: proto.Int32(42344),
			},
			{
				Fieled: proto.Int32(42354),
			},
		},
	})
	/*
		e5b, _ := proto.Marshal(&oidb.DE5BReqBody{
			Uin:                   proto.Uint64(uint64(target)),
			MaxCount:              proto.Uint32(10),
			ReqAchievementContent: proto.Bool(false),
		})
		ec4, _ := proto.Marshal(&oidb.DEC4ReqBody{
			Uin:       proto.Uint64(uint64(target)),
			QuestNum:  proto.Uint64(10),
			FetchType: proto.Uint32(1),
		})
	*/
	req := &jce.SummaryCardReq{
		Uin:              target,
		ComeFrom:         31,
		GetControl:       69181,
		AddFriendSource:  3001,
		SecureSig:        []byte{0x00},
		ReqMedalWallInfo: 0,
		Req0x5ebFieldId:  []int64{27225, 27224, 42122, 42121, 27236, 27238, 42167, 42172, 40324, 42284, 42326, 42325, 42356, 42363, 42361, 42367, 42377, 42425, 42505, 42488},
		ReqServices:      [][]byte{packBusinessBuf(16, gate)},
		ReqNearbyGodInfo: 1,
		ReqExtendCard:    1,
	}
	head := jce.NewJceWriter()
	head.WriteInt32(2, 0)
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{
		"ReqHead":        packUniRequestData(head.Bytes()),
		"ReqSummaryCard": packUniRequestData(req.ToBytes()),
	}}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "SummaryCardServantObj",
		SFuncName:    "ReqSummaryCard",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "SummaryCard.ReqSummaryCard", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// friendlist.GetTroopListReqV2
func (c *QQClient) buildGroupListRequestPacket(vecCookie []byte) (uint16, []byte) {
	seq := c.nextSeq()
	req := &jce.TroopListRequest{
		Uin:              c.Uin,
		GetMSFMsgFlag:    1,
		Cookies:          vecCookie,
		GroupInfo:        []int64{},
		GroupFlagExt:     1,
		Version:          7,
		CompanyId:        0,
		VersionNum:       1,
		GetLongGroupName: 1,
	}
	b := append([]byte{0x0A}, req.ToBytes()...)
	b = append(b, 0x0B)
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"GetTroopListReqV2Simplify": b},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		CPacketType:  0x00,
		IRequestId:   c.nextPacketSeq(),
		SServantName: "mqq.IMService.FriendListServiceServantObj",
		SFuncName:    "GetTroopListReqV2Simplify",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "friendlist.GetTroopListReqV2", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// friendlist.GetTroopMemberListReq
func (c *QQClient) buildGroupMemberListRequestPacket(groupUin, groupCode, nextUin int64) (uint16, []byte) {
	seq := c.nextSeq()
	req := &jce.TroopMemberListRequest{
		Uin:       c.Uin,
		GroupCode: groupCode,
		NextUin:   nextUin,
		GroupUin:  groupUin,
		Version:   2,
	}
	b := append([]byte{0x0A}, req.ToBytes()...)
	b = append(b, 0x0B)
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"GTML": b},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		IRequestId:   c.nextPacketSeq(),
		SServantName: "mqq.IMService.FriendListServiceServantObj",
		SFuncName:    "GetTroopMemberListReq",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "friendlist.GetTroopMemberListReq", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// group_member_card.get_group_member_card_info
func (c *QQClient) buildGroupMemberInfoRequestPacket(groupCode, uin int64) (uint16, []byte) {
	seq := c.nextSeq()
	req := &pb.GroupMemberReqBody{
		GroupCode:       groupCode,
		Uin:             uin,
		NewClient:       true,
		ClientType:      1,
		RichCardNameVer: 1,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "group_member_card.get_group_member_card_info", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MessageSvc.PbGetMsg
func (c *QQClient) buildGetMessageRequestPacket(flag msg.SyncFlag, msgTime int64) (uint16, []byte) {
	seq := c.nextSeq()
	cook := c.syncCookie
	if cook == nil {
		cook, _ = proto.Marshal(&msg.SyncCookie{
			Time:   &msgTime,
			Ran1:   proto.Int64(758330138),
			Ran2:   proto.Int64(2480149246),
			Const1: proto.Int64(1167238020),
			Const2: proto.Int64(3913056418),
			Const3: proto.Int64(0x1D),
		})
	}
	req := &msg.GetMessageRequest{
		SyncFlag:           &flag,
		SyncCookie:         cook,
		LatestRambleNumber: proto.Int32(20),
		OtherRambleNumber:  proto.Int32(3),
		OnlineSyncFlag:     proto.Int32(1),
		ContextFlag:        proto.Int32(1),
		MsgReqType:         proto.Int32(1),
		PubaccountCookie:   []byte{},
		MsgCtrlBuf:         []byte{},
		ServerBuf:          []byte{},
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbGetMsg", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildStopGetMessagePacket(msgTime int64) []byte {
	_, pkt := c.buildGetMessageRequestPacket(msg.SyncFlag_STOP, msgTime)
	return pkt
}

// MessageSvc.PbDeleteMsg
func (c *QQClient) buildDeleteMessageRequestPacket(msg []*pb.MessageItem) (uint16, []byte) {
	seq := c.nextSeq()
	req := &pb.DeleteMessageRequest{Items: msg}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbDeleteMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OnlinePush.RespPush
func (c *QQClient) buildDeleteOnlinePushPacket(uin int64, seq uint16, delMsg []jce.PushMessageInfo) []byte {
	req := &jce.SvcRespPushMsg{Uin: uin}
	for _, m := range delMsg {
		req.DelInfos = append(req.DelInfos, &jce.DelMsgInfo{
			FromUin:    m.FromUin,
			MsgSeq:     m.MsgSeq,
			MsgCookies: m.MsgCookies,
			MsgTime:    m.MsgTime,
		})
	}
	b := append([]byte{0x0A}, req.ToBytes()...)
	b = append(b, 0x0B)
	buf := &jce.RequestDataVersion3{
		Map: map[string][]byte{"resp": b},
	}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		IRequestId:   int32(seq),
		SServantName: "OnlinePush",
		SFuncName:    "SvcRespPushMsg",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	return packets.BuildUniPacket(c.Uin, seq, "OnlinePush.RespPush", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
}

// LongConn.OffPicUp
func (c *QQClient) buildOffPicUpPacket(target int64, md5 []byte, size int32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &cmd0x352.ReqBody{
		Subcmd: 1,
		MsgTryupImgReq: []*cmd0x352.D352TryUpImgReq{
			{
				SrcUin:       int32(c.Uin),
				DstUin:       int32(target),
				FileMd5:      md5,
				FileSize:     size,
				Filename:     hex.EncodeToString(md5) + ".jpg",
				SrcTerm:      5,
				PlatformType: 9,
				BuType:       1,
				ImgOriginal:  1,
				ImgType:      1000,
				BuildVer:     "8.2.7.4410",
				FileIndex:    EmptyBytes,
				SrvUpload:    1,
				TransferUrl:  EmptyBytes,
			},
		},
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "LongConn.OffPicUp", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ProfileService.Pb.ReqSystemMsgNew.Friend
func (c *QQClient) buildSystemMsgNewFriendPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := &structmsg.ReqSystemMsgNew{
		MsgNum:    20,
		Version:   1000,
		Checktype: 2,
		Flag: &structmsg.FlagInfo{
			FrdMsgDiscuss2ManyChat:       1,
			FrdMsgGetBusiCard:            1,
			FrdMsgNeedWaitingMsg:         1,
			FrdMsgUint32NeedAllUnreadMsg: 1,
			GrpMsgMaskInviteAutoJoin:     1,
		},
		FriendMsgTypeFlag: 1,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.Pb.ReqSystemMsgNew.Friend", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// friendlist.ModifyGroupCardReq
func (c *QQClient) buildEditGroupTagPacket(groupCode, memberUin int64, newTag string) (uint16, []byte) {
	seq := c.nextSeq()
	req := &jce.ModifyGroupCardRequest{
		GroupCode: groupCode,
		UinInfo: []jce.IJceStruct{
			&jce.UinInfo{
				Uin:  memberUin,
				Flag: 31,
				Name: newTag,
			},
		},
	}
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{"MGCREQ": packUniRequestData(req.ToBytes())}}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		IRequestId:   c.nextPacketSeq(),
		SServantName: "mqq.IMService.FriendListServiceServantObj",
		SFuncName:    "ModifyGroupCardReq",
		SBuffer:      buf.ToBytes(),
		Context:      map[string]string{},
		Status:       map[string]string{},
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "friendlist.ModifyGroupCardReq", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// OidbSvc.0x8fc_2
func (c *QQClient) buildEditSpecialTitlePacket(groupCode, memberUin int64, newTitle string) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D8FCReqBody{
		GroupCode: &groupCode,
		MemLevelInfo: []*oidb.D8FCMemberInfo{
			{
				Uin:                    &memberUin,
				UinName:                []byte(newTitle),
				SpecialTitle:           []byte(newTitle),
				SpecialTitleExpireTime: proto.Int32(-1),
			},
		},
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(2300, 2, b)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x8fc_2", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0x89a_0
func (c *QQClient) buildGroupOperationPacket(body *oidb.D89AReqBody) (uint16, []byte) {
	seq := c.nextSeq()
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(2202, 0, b)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x89a_0", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0x89a_0
func (c *QQClient) buildGroupNameUpdatePacket(groupCode int64, newName string) (uint16, []byte) {
	body := &oidb.D89AReqBody{
		GroupCode: groupCode,
		StGroupInfo: &oidb.D89AGroupinfo{
			IngGroupName: []byte(newName),
		},
	}
	return c.buildGroupOperationPacket(body)
}

func (c *QQClient) buildGroupMemoUpdatePacket(groupCode int64, newMemo string) (uint16, []byte) {
	body := &oidb.D89AReqBody{
		GroupCode: groupCode,
		StGroupInfo: &oidb.D89AGroupinfo{
			IngGroupMemo: []byte(newMemo),
		},
	}
	return c.buildGroupOperationPacket(body)
}

// OidbSvc.0x89a_0
func (c *QQClient) buildGroupMuteAllPacket(groupCode int64, mute bool) (uint16, []byte) {
	body := &oidb.D89AReqBody{
		GroupCode: groupCode,
		StGroupInfo: &oidb.D89AGroupinfo{
			ShutupTime: &oidb.D89AGroupinfo_Val{Val: func() int32 {
				if mute {
					return 1
				}
				return 0
			}()},
		},
	}
	return c.buildGroupOperationPacket(body)
}

// OidbSvc.0x8a0_0
func (c *QQClient) buildGroupKickPacket(groupCode, memberUin int64, kickMsg string, block bool) (uint16, []byte) {
	seq := c.nextSeq()
	flagBlock := 0
	if block {
		flagBlock = 1
	}
	body := &oidb.D8A0ReqBody{
		OptUint64GroupCode: groupCode,
		MsgKickList: []*oidb.D8A0KickMemberInfo{
			{
				OptUint32Operate:   5,
				OptUint64MemberUin: memberUin,
				OptUint32Flag:      int32(flagBlock),
			},
		},
		KickMsg: []byte(kickMsg),
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(2208, 0, b)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x8a0_0", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0x570_8
func (c *QQClient) buildGroupMutePacket(groupCode, memberUin int64, time uint32) (uint16, []byte) {
	seq := c.nextSeq()
	payload := c.packOIDBPackage(1392, 8, binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt32(uint32(groupCode))
		w.WriteByte(32)
		w.WriteUInt16(1)
		w.WriteUInt32(uint32(memberUin))
		w.WriteUInt32(time)
	}))
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x570_8", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0xed3
func (c *QQClient) buildGroupPokePacket(groupCode, target int64) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.DED3ReqBody{
		ToUin:     target,
		GroupCode: groupCode,
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(3795, 1, b)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0xed3", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0xed3
func (c *QQClient) buildFriendPokePacket(target int64) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.DED3ReqBody{
		ToUin:  target,
		AioUin: target,
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(3795, 1, b)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0xed3", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0x55c_1
func (c *QQClient) buildGroupAdminSetPacket(groupCode, member int64, flag bool) (uint16, []byte) {
	seq := c.nextSeq()
	payload := c.packOIDBPackage(1372, 1, binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt32(uint32(groupCode))
		w.WriteUInt32(uint32(member))
		w.WriteByte(func() byte {
			if flag {
				return 1
			}
			return 0
		}())
	}))
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x55c_1", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ProfileService.GroupMngReq
func (c *QQClient) buildQuitGroupPacket(groupCode int64) (uint16, []byte) {
	seq := c.nextSeq()
	jw := jce.NewJceWriter()
	jw.WriteInt32(2, 0)
	jw.WriteInt64(c.Uin, 1)
	jw.WriteBytes(binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt32(uint32(c.Uin))
		w.WriteUInt32(uint32(groupCode))
	}), 2)
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{"GroupMngReq": packUniRequestData(jw.Bytes())}}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		IRequestId:   c.nextPacketSeq(),
		SServantName: "KQQ.ProfileService.ProfileServantObj",
		SFuncName:    "GroupMngReq",
		SBuffer:      buf.ToBytes(),
		Context:      map[string]string{},
		Status:       map[string]string{},
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.GroupMngReq", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// OidbSvc.0xe07_0
func (c *QQClient) buildImageOcrRequestPacket(url, md5 string, size, weight, height int32) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.DE07ReqBody{
		Version:  1,
		Entrance: 3,
		OcrReqBody: &oidb.OCRReqBody{
			ImageUrl:              url,
			OriginMd5:             md5,
			AfterCompressMd5:      md5,
			AfterCompressFileSize: size,
			AfterCompressWeight:   weight,
			AfterCompressHeight:   height,
			IsCut:                 false,
		},
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(3591, 0, b)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0xe07_0", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// LightAppSvc.mini_app_info.GetAppInfoById
func (c *QQClient) buildAppInfoRequestPacket(id string) (uint16, []byte) {
	seq := c.nextSeq()
	req := &qweb.GetAppInfoByIdReq{
		AppId:           id,
		NeedVersionInfo: 1,
	}
	b, _ := proto.Marshal(req)
	body := &qweb.QWebReq{
		Seq:        1,
		Qua:        "V1_AND_SQ_8.4.8_1492_YYB_D",
		DeviceInfo: fmt.Sprintf("i=865166025905020&imsi=460002478794049&mac=02:00:00:00:00:00&m=%v&o=7.1.2&a=25&sc=1&sd=0&p=900*1600&f=nubia&mm=3479&cf=2407&cc=4&aid=086bbf84a7d5fbb3&qimei=865166023450458&sharpP=1&n=wifi", string(SystemDeviceInfo.Model)),
		BusiBuff:   b,
		TraceId:    fmt.Sprintf("%v_%v_%v", c.Uin, time.Now().Format("0102150405"), rand.Int63()),
	}
	payload, _ := proto.Marshal(body)
	packet := packets.BuildUniPacket(c.Uin, seq, "LightAppSvc.mini_app_info.GetAppInfoById", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildWordSegmentationPacket(data []byte) (uint16, []byte) {
	seq := c.nextSeq()
	payload := c.packOIDBPackageProto(3449, 1, &oidb.D79ReqBody{
		Uin:     uint64(c.Uin),
		Content: data,
		Qua:     []byte("and_537065262_8.4.5"),
	})
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0xd79", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0xdad_1
func (c *QQClient) sendGroupGiftPacket(groupCode, uin uint64, productId message.GroupGift) (uint16, []byte) {
	seq := c.nextSeq()
	payload := c.packOIDBPackageProto(3501, 1, &oidb.DADReqBody{
		Client:    1,
		ProductId: uint64(productId),
		ToUin:     uin,
		Gc:        groupCode,
		Version:   "V 8.4.5.4745",
		Sig: &oidb.DADLoginSig{
			Type: 1,
			Sig:  []byte(c.getSKey()),
		},
	})
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0xdad_1", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}
