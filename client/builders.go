package client

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x352"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/multimsg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/client/pb/structmsg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/crypto"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/protocol/tlv"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"strconv"
)

var (
	syncConst1 = rand.Int63()
	syncConst2 = rand.Int63()
)

func (c *QQClient) buildLoginPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x0810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(9)
		w.WriteUInt16(17)

		w.Write(tlv.T18(16, uint32(c.Uin)))
		w.Write(tlv.T1(uint32(c.Uin), SystemDeviceInfo.IpAddress))
		w.Write(tlv.T106(uint32(c.Uin), 0, c.PasswordMd5, true, SystemDeviceInfo.Guid, SystemDeviceInfo.TgtgtKey))
		w.Write(tlv.T116(184024956, 0x10400))
		w.Write(tlv.T100())
		w.Write(tlv.T107(0))
		w.Write(tlv.T142("com.tencent.mobileqq"))
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

		w.Write(tlv.T145(SystemDeviceInfo.Guid))
		w.Write(tlv.T147(16, []byte("8.2.7"), []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D}))
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
			"qzone.qq.com", "vip.qq.com", "qun.qq.com", "game.qq.com", "qqweb.qq.com",
			"office.qq.com", "ti.qq.com", "mail.qq.com", "qzone.com", "mma.qq.com",
		}))

		w.Write(tlv.T187(SystemDeviceInfo.MacAddress))
		w.Write(tlv.T188(SystemDeviceInfo.AndroidId))
		if len(SystemDeviceInfo.IMSIMd5) != 0 {
			w.Write(tlv.T194(SystemDeviceInfo.IMSIMd5))
		}
		w.Write(tlv.T191(0x82))
		if len(SystemDeviceInfo.WifiBSSID) != 0 && len(SystemDeviceInfo.WifiSSID) != 0 {
			w.Write(tlv.T202(SystemDeviceInfo.WifiBSSID, SystemDeviceInfo.WifiSSID))
		}
		w.Write(tlv.T177())
		w.Write(tlv.T516())
		w.Write(tlv.T521())
		w.Write(tlv.T525(tlv.T536([]byte{0x01, 0x00})))
	})
	sso := packets.BuildSsoPacket(seq, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
	return seq, packet
}

func (c *QQClient) buildDeviceLockLoginPacket(t402 []byte) (uint16, []byte) {
	seq := c.nextSeq()
	req := packets.BuildOicqRequestPacket(c.Uin, 0x0810, crypto.ECDH, c.RandomKey, func(w *binary.Writer) {
		w.WriteUInt16(20)
		w.WriteUInt16(4)

		w.Write(tlv.T8(2052))
		w.Write(tlv.T104(c.t104))
		w.Write(tlv.T116(150470524, 66560))
		h := md5.Sum(append(append(SystemDeviceInfo.Guid, []byte("stMNokHgxZUGhsYp")...), t402...))
		w.Write(tlv.T401(h[:]))
	})
	sso := packets.BuildSsoPacket(seq, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
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
		w.Write(tlv.T116(150470524, 66560))
	})
	sso := packets.BuildSsoPacket(seq, "wtlogin.login", SystemDeviceInfo.IMEI, []byte{}, c.OutGoingPacketSessionId, req, c.ksid)
	packet := packets.BuildLoginPacket(c.Uin, 2, make([]byte, 16), sso, []byte{})
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
		VendorName:   "MIUI",
		VendorOSName: "ONEPLUS A5000_23_17",
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
	sso := packets.BuildSsoPacket(seq, "StatSvc.register", SystemDeviceInfo.IMEI, c.sigInfo.tgt, c.OutGoingPacketSessionId, pkt.ToBytes(), c.ksid)
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
		Map: map[string][]byte{"PushResp": packRequestDataV3(req.Bytes())},
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
		IfReflush: func() byte { // fuck golang
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
		Map: map[string][]byte{"FL": packRequestDataV3(req.ToBytes())},
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

// friendlist.GetTroopListReqV2
func (c *QQClient) buildGroupListRequestPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := &jce.TroopListRequest{
		Uin:              c.Uin,
		GetMSFMsgFlag:    1,
		Cookies:          []byte{},
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

// MessageSvc.PbGetMsg
func (c *QQClient) buildGetMessageRequestPacket(flag msg.SyncFlag, msgTime int64) (uint16, []byte) {
	seq := c.nextSeq()
	cook := c.syncCookie
	if cook == nil {
		cook, _ = proto.Marshal(&msg.SyncCookie{
			Time:   msgTime,
			Ran1:   758330138,
			Ran2:   2480149246,
			Const1: 1167238020,
			Const2: 3913056418,
			Const3: 0x1D,
		})
	}
	req := &msg.GetMessageRequest{
		SyncFlag:           flag,
		SyncCookie:         cook,
		LatestRambleNumber: 20,
		OtherRambleNumber:  3,
		OnlineSyncFlag:     1,
		ContextFlag:        1,
		MsgReqType:         1,
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

// MessageSvc.PbSendMsg
func (c *QQClient) buildGroupSendingPacket(groupCode int64, r int32, forward bool, m *message.SendingMessage) (uint16, []byte) {
	seq := c.nextSeq()
	if m.Ptt != nil {
		m.Elements = []message.IMessageElement{}
	}

	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{Grp: &msg.Grp{GroupCode: groupCode}},
		ContentHead: &msg.ContentHead{PkgNum: 1},
		MsgBody: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m.Elements, true),
				Ptt:   m.Ptt,
			},
		},
		MsgSeq:     c.nextGroupSeq(),
		MsgRand:    r,
		SyncCookie: EmptyBytes,
		MsgVia:     1,
		MsgCtrl: func() *msg.MsgCtrl {
			if forward {
				return &msg.MsgCtrl{MsgFlag: 4}
			}
			return nil
		}(),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbSendMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MessageSvc.PbSendMsg
func (c *QQClient) buildFriendSendingPacket(target int64, msgSeq, r int32, time int64, m *message.SendingMessage) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{C2C: &msg.C2C{ToUin: target}},
		ContentHead: &msg.ContentHead{PkgNum: 1},
		MsgBody: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m.Elements, false),
			},
		},
		MsgSeq:  msgSeq,
		MsgRand: r,
		SyncCookie: func() []byte {
			cookie := &msg.SyncCookie{
				Time:   time,
				Ran1:   rand.Int63(),
				Ran2:   rand.Int63(),
				Const1: syncConst1,
				Const2: syncConst2,
				Const3: 0x1d,
			}
			b, _ := proto.Marshal(cookie)
			return b
		}(),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbSendMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
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

// ImgStore.GroupPicUp
func (c *QQClient) buildGroupImageStorePacket(groupCode int64, md5 []byte, size int32) (uint16, []byte) {
	seq := c.nextSeq()
	name := utils.RandomString(16) + ".gif"
	req := &pb.D388ReqBody{
		NetType: 3,
		Subcmd:  1,
		MsgTryUpImgReq: []*pb.TryUpImgReq{
			{
				GroupCode:    groupCode,
				SrcUin:       c.Uin,
				FileMd5:      md5,
				FileSize:     int64(size),
				FileName:     name,
				SrcTerm:      5,
				PlatformType: 9,
				BuType:       1,
				PicType:      1000,
				BuildVer:     "8.2.7.4410",
				AppPicType:   1006,
				FileIndex:    EmptyBytes,
				TransferUrl:  EmptyBytes,
			},
		},
		Extension: EmptyBytes,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ImgStore.GroupPicUp", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildImageUploadPacket(data, updKey []byte, commandId int32, fmd5 [16]byte) (r [][]byte) {
	offset := 0
	binary.ToChunkedBytesF(data, 8192*1024, func(chunked []byte) {
		w := binary.NewWriter()
		cmd5 := md5.Sum(chunked)
		head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
			MsgBasehead: &pb.DataHighwayHead{
				Version: 1,
				Uin:     strconv.FormatInt(c.Uin, 10),
				Command: "PicUp.DataUp",
				Seq: func() int32 {
					if commandId == 2 {
						return c.nextGroupDataTransSeq()
					}
					if commandId == 27 {
						return c.nextHighwayApplySeq()
					}
					return c.nextGroupDataTransSeq()
				}(),
				Appid:     537062409,
				Dataflag:  4096,
				CommandId: commandId,
				LocaleId:  2052,
			},
			MsgSeghead: &pb.SegHead{
				Filesize:      int64(len(data)),
				Dataoffset:    int64(offset),
				Datalength:    int32(len(chunked)),
				Serviceticket: updKey,
				Md5:           cmd5[:],
				FileMd5:       fmd5[:],
			},
			ReqExtendinfo: EmptyBytes,
		})
		offset += len(chunked)
		w.WriteByte(40)
		w.WriteUInt32(uint32(len(head)))
		w.WriteUInt32(uint32(len(chunked)))
		w.Write(head)
		w.Write(chunked)
		w.WriteByte(41)
		r = append(r, w.Bytes())
	})
	return
}

// PttStore.GroupPttUp
func (c *QQClient) buildGroupPttStorePacket(groupCode int64, md5 []byte, size, voiceLength int32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &pb.D388ReqBody{
		NetType: 3,
		Subcmd:  3,
		MsgTryUpPttReq: []*pb.TryUpPttReq{
			{
				GroupCode:     groupCode,
				SrcUin:        c.Uin,
				FileMd5:       md5,
				FileSize:      int64(size),
				FileName:      md5,
				SrcTerm:       5,
				PlatformType:  9,
				BuType:        4,
				InnerIp:       0,
				BuildVer:      "6.5.5.663",
				VoiceLength:   voiceLength,
				Codec:         0,
				VoiceType:     1,
				BoolNewUpChan: true,
			},
		},
		Extension: EmptyBytes,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "PttStore.GroupPttUp", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ProfileService.Pb.ReqSystemMsgNew.Group
func (c *QQClient) buildSystemMsgNewGroupPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := &structmsg.ReqSystemMsgNew{
		MsgNum:    5,
		Version:   100,
		Checktype: 3,
		Flag: &structmsg.FlagInfo{
			GrpMsgKickAdmin:                   1,
			GrpMsgHiddenGrp:                   1,
			GrpMsgWordingDown:                 1,
			GrpMsgGetOfficialAccount:          1,
			GrpMsgGetPayInGroup:               1,
			FrdMsgDiscuss2ManyChat:            1,
			GrpMsgNotAllowJoinGrpInviteNotFrd: 1,
			FrdMsgNeedWaitingMsg:              1,
			FrdMsgUint32NeedAllUnreadMsg:      1,
			GrpMsgNeedAutoAdminWording:        1,
			GrpMsgGetTransferGroupMsgFlag:     1,
			GrpMsgGetQuitPayGroupMsgFlag:      1,
			GrpMsgSupportInviteAutoJoin:       1,
			GrpMsgMaskInviteAutoJoin:          1,
			GrpMsgGetDisbandedByAdmin:         1,
			GrpMsgGetC2CInviteJoinGroup:       1,
		},
		FriendMsgTypeFlag: 1,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.Pb.ReqSystemMsgNew.Group", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
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

// ProfileService.Pb.ReqSystemMsgAction.Group
func (c *QQClient) buildSystemMsgGroupActionPacket(reqId, requester, group int64, isInvite, accept, block bool) (uint16, []byte) {
	seq := c.nextSeq()
	req := &structmsg.ReqSystemMsgAction{
		MsgType: 1,
		MsgSeq:  reqId,
		ReqUin:  requester,
		SubType: 1,
		SrcId:   3,
		SubSrcId: func() int32 {
			if isInvite {
				return 10016
			}
			return 31
		}(),
		GroupMsgType: func() int32 {
			if isInvite {
				return 2
			}
			return 1
		}(),
		ActionInfo: &structmsg.SystemMsgActionInfo{
			Type: func() int32 {
				if accept {
					return 11
				}
				return 12
			}(),
			GroupCode: group,
			Blacklist: block,
			Sig:       EmptyBytes,
		},
		Language: 1000,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.Pb.ReqSystemMsgAction.Group", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ProfileService.Pb.ReqSystemMsgAction.Friend
func (c *QQClient) buildSystemMsgFriendActionPacket(reqId, requester int64, accept bool) (uint16, []byte) {
	seq := c.nextSeq()
	req := &structmsg.ReqSystemMsgAction{
		MsgType:  1,
		MsgSeq:   reqId,
		ReqUin:   requester,
		SubType:  1,
		SrcId:    6,
		SubSrcId: 7,
		ActionInfo: &structmsg.SystemMsgActionInfo{
			Type: func() int32 {
				if accept {
					return 2
				}
				return 3
			}(),
			Blacklist:    false,
			AddFrdSNInfo: &structmsg.AddFrdSNInfo{},
		},
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.Pb.ReqSystemMsgAction.Friend", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// PbMessageSvc.PbMsgWithDraw
func (c *QQClient) buildGroupRecallPacket(groupCode int64, msgSeq, msgRan int32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.MsgWithDrawReq{
		GroupWithDraw: []*msg.GroupMsgWithDrawReq{
			{
				SubCmd:    1,
				GroupCode: groupCode,
				MsgList: []*msg.GroupMsgInfo{
					{
						MsgSeq:    msgSeq,
						MsgRandom: msgRan,
						MsgType:   0,
					},
				},
				UserDef: []byte{0x08, 0x00},
			},
		},
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "PbMessageSvc.PbMsgWithDraw", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
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
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{"MGCREQ": packRequestDataV3(req.ToBytes())}}
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
		GroupCode: groupCode,
		MemLevelInfo: []*oidb.D8FCMemberInfo{
			{
				Uin:                    memberUin,
				UinName:                []byte(newTitle),
				SpecialTitle:           []byte(newTitle),
				SpecialTitleExpireTime: -1,
			},
		},
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:     2300,
		ServiceType: 2,
		Bodybuffer:  b,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x8fc_2", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0x89a_0
func (c *QQClient) buildGroupOperationPacket(body *oidb.D89AReqBody) (uint16, []byte) {
	seq := c.nextSeq()
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:    2202,
		Bodybuffer: b,
	}
	payload, _ := proto.Marshal(req)
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
func (c *QQClient) buildGroupKickPacket(groupCode, memberUin int64, kickMsg string) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D8A0ReqBody{
		OptUint64GroupCode: groupCode,
		MsgKickList: []*oidb.D8A0KickMemberInfo{
			{
				OptUint32Operate:   5,
				OptUint64MemberUin: memberUin,
				OptUint32Flag:      1,
			},
		},
		KickMsg: []byte(kickMsg),
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:    2208,
		Bodybuffer: b,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x8a0_0", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0x570_8
func (c *QQClient) buildGroupMutePacket(groupCode, memberUin int64, time uint32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &oidb.OIDBSSOPkg{
		Command:     1392,
		ServiceType: 8,
		Bodybuffer: binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt32(uint32(groupCode))
			w.WriteByte(32)
			w.WriteUInt16(1)
			w.WriteUInt32(uint32(memberUin))
			w.WriteUInt32(time)
		}),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x570_8", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MultiMsg.ApplyUp
func (c *QQClient) buildMultiApplyUpPacket(data, hash []byte, buType int32, groupUin int64) (uint16, []byte) {
	seq := c.nextSeq()
	req := &multimsg.MultiReqBody{
		Subcmd:       1,
		TermType:     5,
		PlatformType: 9,
		NetType:      3,
		BuildVer:     "8.2.0.1296",
		MultimsgApplyupReq: []*multimsg.MultiMsgApplyUpReq{
			{
				DstUin:  groupUin,
				MsgSize: int64(len(data)),
				MsgMd5:  hash,
				MsgType: 3,
			},
		},
		BuType: buType,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MultiMsg.ApplyUp", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// MultiMsg.ApplyDown
func (c *QQClient) buildMultiApplyDownPacket(resId string) (uint16, []byte) {
	seq := c.nextSeq()
	req := &multimsg.MultiReqBody{
		Subcmd:       2,
		TermType:     5,
		PlatformType: 9,
		NetType:      3,
		BuildVer:     "8.2.0.1296",
		MultimsgApplydownReq: []*multimsg.MultiMsgApplyDownReq{
			{
				MsgResid: []byte(resId),
				MsgType:  3,
			},
		},
		BuType:         2,
		ReqChannelType: 2,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MultiMsg.ApplyDown", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
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
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{"GroupMngReq": packRequestDataV3(jw.Bytes())}}
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

// OidbSvc.0x6d6_2
func (c *QQClient) buildGroupFileDownloadReqPacket(groupCode int64, fileId string, busId int32) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D6D6ReqBody{
		DownloadFileReq: &oidb.DownloadFileReqBody{
			GroupCode: groupCode,
			AppId:     3,
			BusId:     busId,
			FileId:    fileId,
		},
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:     1750,
		ServiceType: 2,
		Bodybuffer:  b,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x6d6_2", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}
