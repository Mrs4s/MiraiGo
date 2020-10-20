package jce

type IJceStruct interface {
	//ToBytes() []byte
	ReadFrom(*JceReader)
}

type (
	RequestPacket struct {
		IVersion     int16             `jceId:"1"`
		CPacketType  byte              `jceId:"2"`
		IMessageType int32             `jceId:"3"`
		IRequestId   int32             `jceId:"4"`
		SServantName string            `jceId:"5"`
		SFuncName    string            `jceId:"6"`
		SBuffer      []byte            `jceId:"7"`
		ITimeout     int32             `jceId:"8"`
		Context      map[string]string `jceId:"9"`
		Status       map[string]string `jceId:"10"`
	}

	RequestDataVersion3 struct {
		Map map[string][]byte `jceId:"0"`
	}

	RequestDataVersion2 struct {
		Map map[string]map[string][]byte `jceId:"0"`
	}

	SsoServerInfo struct {
		Server   string `jceId:"1"`
		Port     int32  `jceId:"2"`
		Location string `jceId:"8"`
	}

	SvcReqRegister struct {
		IJceStruct
		Uin                int64  `jceId:"0"`
		Bid                int64  `jceId:"1"`
		ConnType           byte   `jceId:"2"`
		Other              string `jceId:"3"`
		Status             int32  `jceId:"4"`
		OnlinePush         byte   `jceId:"5"`
		IsOnline           byte   `jceId:"6"`
		IsShowOnline       byte   `jceId:"7"`
		KickPC             byte   `jceId:"8"`
		KickWeak           byte   `jceId:"9"`
		Timestamp          int64  `jceId:"10"`
		IOSVersion         int64  `jceId:"11"`
		NetType            byte   `jceId:"12"`
		BuildVer           string `jceId:"13"`
		RegType            byte   `jceId:"14"`
		DevParam           []byte `jceId:"15"`
		Guid               []byte `jceId:"16"`
		LocaleId           int32  `jceId:"17"`
		SilentPush         byte   `jceId:"18"`
		DevName            string `jceId:"19"`
		DevType            string `jceId:"20"`
		OSVer              string `jceId:"21"`
		OpenPush           byte   `jceId:"22"`
		LargeSeq           int64  `jceId:"23"`
		LastWatchStartTime int64  `jceId:"24"`
		OldSSOIp           int64  `jceId:"26"`
		NewSSOIp           int64  `jceId:"27"`
		ChannelNo          string `jceId:"28"`
		CPID               int64  `jceId:"29"`
		VendorName         string `jceId:"30"`
		VendorOSName       string `jceId:"31"`
		IOSIdfa            string `jceId:"32"`
		B769               []byte `jceId:"33"`
		IsSetStatus        byte   `jceId:"34"`
		ServerBuf          []byte `jceId:"35"`
		SetMute            byte   `jceId:"36"`
	}

	PushMessageInfo struct {
		FromUin        int64  `jceId:"0"`
		MsgTime        int64  `jceId:"1"`
		MsgType        int16  `jceId:"2"`
		MsgSeq         int16  `jceId:"3"`
		Msg            string `jceId:"4"`
		RealMsgTime    int32  `jceId:"5"`
		VMsg           []byte `jceId:"6"`
		AppShareID     int64  `jceId:"7"`
		MsgCookies     []byte `jceId:"8"`
		AppShareCookie []byte `jceId:"9"`
		MsgUid         int64  `jceId:"10"`
		LastChangeTime int64  `jceId:"11"`
		FromInstId     int64  `jceId:"14"`
		RemarkOfSender []byte `jceId:"15"`
		FromMobile     string `jceId:"16"`
		FromName       string `jceId:"17"`
	}

	SvcRespPushMsg struct {
		IJceStruct
		Uin         int64        `jceId:"0"`
		DelInfos    []IJceStruct `jceId:"1"`
		Svrip       int32        `jceId:"2"`
		PushToken   []byte       `jceId:"3"`
		ServiceType int32        `jceId:"4"`
	}

	SvcReqGetDevLoginInfo struct {
		IJceStruct
		Guid           []byte `jceId:"0"`
		AppName        string `jceId:"1"`
		LoginType      int64  `jceId:"2"`
		Timestamp      int64  `jceId:"3"`
		NextItemIndex  int64  `jceId:"4"`
		RequireMax     int64  `jceId:"5"`
		GetDevListType int64  `jceId:"6"` // 1: getLoginDevList 2: getRecentLoginDevList 4: getAuthLoginDevList
	}

	SvcDevLoginInfo struct {
		AppId          int64
		Guid           []byte
		LoginTime      int64
		LoginPlatform  int64
		LoginLocation  string
		DeviceName     string
		DeviceTypeInfo string
		TerType        int64
		ProductType    int64
		CanBeKicked    int64
	}

	DelMsgInfo struct {
		IJceStruct
		FromUin    int64  `jceId:"0"`
		MsgTime    int64  `jceId:"1"`
		MsgSeq     int16  `jceId:"2"`
		MsgCookies []byte `jceId:"3"`
		Cmd        int16  `jceId:"4"`
		MsgType    int64  `jceId:"5"`
		AppId      int64  `jceId:"6"`
		SendTime   int64  `jceId:"7"`
		SsoSeq     int32  `jceId:"8"`
		SsoIp      int32  `jceId:"9"`
		ClientIp   int32  `jceId:"10"`
	}

	FriendListRequest struct {
		IJceStruct
		Reqtype         int32   `jceId:"0"`
		IfReflush       byte    `jceId:"1"`
		Uin             int64   `jceId:"2"`
		StartIndex      int16   `jceId:"3"`
		FriendCount     int16   `jceId:"4"`
		GroupId         byte    `jceId:"5"`
		IfGetGroupInfo  byte    `jceId:"6"`
		GroupStartIndex byte    `jceId:"7"`
		GroupCount      byte    `jceId:"8"`
		IfGetMSFGroup   byte    `jceId:"9"`
		IfShowTermType  byte    `jceId:"10"`
		Version         int64   `jceId:"11"`
		UinList         []int64 `jceId:"12"`
		AppType         int32   `jceId:"13"`
		IfGetDOVId      byte    `jceId:"14"`
		IfGetBothFlag   byte    `jceId:"15"`
		D50             []byte  `jceId:"16"`
		D6B             []byte  `jceId:"17"`
		SnsTypeList     []int64 `jceId:"18"`
	}

	FriendInfo struct {
		FriendUin           int64  `jceId:"0"`
		GroupId             byte   `jceId:"1"`
		FaceId              int16  `jceId:"2"`
		Remark              string `jceId:"3"`
		QQType              byte   `jceId:"4"`
		Status              byte   `jceId:"5"`
		MemberLevel         byte   `jceId:"6"`
		IsMqqOnLine         byte   `jceId:"7"`
		QQOnlineState       byte   `jceId:"8"`
		IsIphoneOnline      byte   `jceId:"9"`
		DetailStatusFlag    byte   `jceId:"10"`
		QQOnlineStateV2     byte   `jceId:"11"`
		ShowName            string `jceId:"12"`
		IsRemark            byte   `jceId:"13"`
		Nick                string `jceId:"14"`
		SpecialFlag         byte   `jceId:"15"`
		IMGroupID           []byte `jceId:"16"`
		MSFGroupID          []byte `jceId:"17"`
		TermType            int32  `jceId:"18"`
		Network             byte   `jceId:"20"`
		Ring                []byte `jceId:"21"`
		AbiFlag             int64  `jceId:"22"`
		FaceAddonId         int64  `jceId:"23"`
		NetworkType         int32  `jceId:"24"`
		VipFont             int64  `jceId:"25"`
		IconType            int32  `jceId:"26"`
		TermDesc            string `jceId:"27"`
		ColorRing           int64  `jceId:"28"`
		ApolloFlag          byte   `jceId:"29"`
		ApolloTimestamp     int64  `jceId:"30"`
		Sex                 byte   `jceId:"31"`
		FounderFont         int64  `jceId:"32"`
		EimId               string `jceId:"33"`
		EimMobile           string `jceId:"34"`
		OlympicTorch        byte   `jceId:"35"`
		ApolloSignTime      int64  `jceId:"36"`
		LaviUin             int64  `jceId:"37"`
		TagUpdateTime       int64  `jceId:"38"`
		GameLastLoginTime   int64  `jceId:"39"`
		GameAppId           int64  `jceId:"40"`
		CardID              []byte `jceId:"41"`
		BitSet              int64  `jceId:"42"`
		KingOfGloryFlag     byte   `jceId:"43"`
		KingOfGloryRank     int64  `jceId:"44"`
		MasterUin           string `jceId:"45"`
		LastMedalUpdateTime int64  `jceId:"46"`
		FaceStoreId         int64  `jceId:"47"`
		FontEffect          int64  `jceId:"48"`
		DOVId               string `jceId:"49"`
		BothFlag            int64  `jceId:"50"`
		CentiShow3DFlag     byte   `jceId:"51"`
		IntimateInfo        []byte `jceId:"52"`
		ShowNameplate       byte   `jceId:"53"`
		NewLoverDiamondFlag byte   `jceId:"54"`
		ExtSnsFrdData       []byte `jceId:"55"`
		MutualMarkData      []byte `jceId:"56"`
	}

	TroopListRequest struct {
		IJceStruct
		Uin              int64   `jceId:"0"`
		GetMSFMsgFlag    byte    `jceId:"1"`
		Cookies          []byte  `jceId:"2"`
		GroupInfo        []int64 `jceId:"3"`
		GroupFlagExt     byte    `jceId:"4"`
		Version          int32   `jceId:"5"`
		CompanyId        int64   `jceId:"6"`
		VersionNum       int64   `jceId:"7"`
		GetLongGroupName byte    `jceId:"8"`
	}

	TroopNumber struct {
		GroupUin              int64  `jceId:"0"`
		GroupCode             int64  `jceId:"1"`
		Flag                  byte   `jceId:"2"`
		GroupInfoSeq          int64  `jceId:"3"`
		GroupName             string `jceId:"4"`
		GroupMemo             string `jceId:"5"`
		GroupFlagExt          int64  `jceId:"6"`
		GroupRankSeq          int64  `jceId:"7"`
		CertificationType     int64  `jceId:"8"`
		ShutUpTimestamp       int64  `jceId:"9"`
		MyShutUpTimestamp     int64  `jceId:"10"`
		CmdUinUinFlag         int64  `jceId:"11"`
		AdditionalFlag        int64  `jceId:"12"`
		GroupTypeFlag         int64  `jceId:"13"`
		GroupSecType          int64  `jceId:"14"`
		GroupSecTypeInfo      int64  `jceId:"15"`
		GroupClassExt         int64  `jceId:"16"`
		AppPrivilegeFlag      int64  `jceId:"17"`
		SubscriptionUin       int64  `jceId:"18"`
		MemberNum             int64  `jceId:"19"`
		MemberNumSeq          int64  `jceId:"20"`
		MemberCardSeq         int64  `jceId:"21"`
		GroupFlagExt3         int64  `jceId:"22"`
		GroupOwnerUin         int64  `jceId:"23"`
		IsConfGroup           byte   `jceId:"24"`
		IsModifyConfGroupFace byte   `jceId:"25"`
		IsModifyConfGroupName byte   `jceId:"26"`
		CmdUinJoinTime        int64  `jceId:"27"`
		CompanyId             int64  `jceId:"28"`
		MaxGroupMemberNum     int64  `jceId:"29"`
		CmdUinGroupMask       int64  `jceId:"30"`
		GuildAppId            int64  `jceId:"31"`
		GuildSubType          int64  `jceId:"32"`
		CmdUinRingtoneID      int64  `jceId:"33"`
		CmdUinFlagEx2         int64  `jceId:"34"`
	}

	TroopMemberListRequest struct {
		IJceStruct
		Uin                int64 `jceId:"0"`
		GroupCode          int64 `jceId:"1"`
		NextUin            int64 `jceId:"2"`
		GroupUin           int64 `jceId:"3"`
		Version            int64 `jceId:"4"`
		ReqType            int64 `jceId:"5"`
		GetListAppointTime int64 `jceId:"6"`
		RichCardNameVer    byte  `jceId:"7"`
	}

	TroopMemberInfo struct {
		MemberUin              int64  `jceId:"0"`
		FaceId                 int16  `jceId:"1"`
		Age                    byte   `jceId:"2"`
		Gender                 byte   `jceId:"3"`
		Nick                   string `jceId:"4"`
		Status                 byte   `jceId:"5"`
		ShowName               string `jceId:"6"`
		Name                   string `jceId:"8"`
		Memo                   string `jceId:"12"`
		AutoRemark             string `jceId:"13"`
		MemberLevel            int64  `jceId:"14"`
		JoinTime               int64  `jceId:"15"`
		LastSpeakTime          int64  `jceId:"16"`
		CreditLevel            int64  `jceId:"17"`
		Flag                   int64  `jceId:"18"`
		FlagExt                int64  `jceId:"19"`
		Point                  int64  `jceId:"20"`
		Concerned              byte   `jceId:"21"`
		Shielded               byte   `jceId:"22"`
		SpecialTitle           string `jceId:"23"`
		SpecialTitleExpireTime int64  `jceId:"24"`
		Job                    string `jceId:"25"`
		ApolloFlag             byte   `jceId:"26"`
		ApolloTimestamp        int64  `jceId:"27"`
		GlobalGroupLevel       int64  `jceId:"28"`
		TitleId                int64  `jceId:"29"`
		ShutUpTimestap         int64  `jceId:"30"`
		GlobalGroupPoint       int64  `jceId:"31"`
		RichCardNameVer        byte   `jceId:"33"`
		VipType                int64  `jceId:"34"`
		VipLevel               int64  `jceId:"35"`
		BigClubLevel           int64  `jceId:"36"`
		BigClubFlag            int64  `jceId:"37"`
		Nameplate              int64  `jceId:"38"`
		GroupHonor             []byte `jceId:"39"`
	}

	ModifyGroupCardRequest struct {
		IJceStruct
		Zero      int64        `jceId:"0"`
		GroupCode int64        `jceId:"1"`
		NewSeq    int64        `jceId:"2"`
		UinInfo   []IJceStruct `jceId:"3"`
	}

	UinInfo struct {
		IJceStruct
		Uin    int64  `jceId:"0"`
		Flag   int64  `jceId:"1"`
		Name   string `jceId:"2"`
		Gender byte   `jceId:"3"`
		Phone  string `jceId:"4"`
		Email  string `jceId:"5"`
		Remark string `jceId:"6"`
	}

	SummaryCardReq struct {
		IJceStruct
		Uin                int64 `jceId:"0"`
		ComeFrom           int32 `jceId:"1"`
		QzoneFeedTimestamp int64 `jceId:"2"`
		IsFriend           byte  `jceId:"3"`
		GroupCode          int64 `jceId:"4"`
		GroupUin           int64 `jceId:"5"`
		//Seed               []byte`jceId:"6"`
		//SearchName         string`jceId:"7"`
		GetControl       int64   `jceId:"8"`
		AddFriendSource  int32   `jceId:"9"`
		SecureSig        []byte  `jceId:"10"`
		TinyId           int64   `jceId:"15"`
		LikeSource       int64   `jceId:"16"`
		ReqMedalWallInfo byte    `jceId:"18"`
		Req0x5ebFieldId  []int64 `jceId:"19"`
		ReqNearbyGodInfo byte    `jceId:"20"`
		ReqExtendCard    byte    `jceId:"22"`
	}
)

func (pkt *RequestPacket) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *RequestPacket) ReadFrom(r *JceReader) {
	pkt.SBuffer = []byte{}
	pkt.Context = make(map[string]string)
	pkt.Status = make(map[string]string)
	pkt.IVersion = r.ReadInt16(1)
	pkt.CPacketType = r.ReadByte(2)
	pkt.IMessageType = r.ReadInt32(3)
	pkt.IRequestId = r.ReadInt32(4)
	pkt.SServantName = r.ReadString(5)
	pkt.SFuncName = r.ReadString(6)
	r.ReadSlice(&pkt.SBuffer, 7)
	pkt.ITimeout = r.ReadInt32(8)
	r.ReadMapF(9, func(k interface{}, v interface{}) { pkt.Context[k.(string)] = v.(string) })
	r.ReadMapF(10, func(k interface{}, v interface{}) { pkt.Status[k.(string)] = v.(string) })
}

func (pkt *RequestDataVersion3) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *RequestDataVersion3) ReadFrom(r *JceReader) {
	pkt.Map = make(map[string][]byte)
	r.ReadMapF(0, func(k interface{}, v interface{}) {
		pkt.Map[k.(string)] = v.([]byte)
	})
}

func (pkt *RequestDataVersion2) ReadFrom(r *JceReader) {
	pkt.Map = make(map[string]map[string][]byte)
	r.ReadMapF(0, func(k interface{}, v interface{}) {
		pkt.Map[k.(string)] = make(map[string][]byte)
		for k2, v := range v.(map[interface{}]interface{}) {
			pkt.Map[k.(string)][k2.(string)] = v.([]byte)
		}
	})
}

func (pkt *SsoServerInfo) ReadFrom(r *JceReader) {
	pkt.Server = r.ReadString(1)
	pkt.Port = r.ReadInt32(2)
	pkt.Location = r.ReadString(8)
}

func (pkt *SvcReqRegister) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *FriendListRequest) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *SummaryCardReq) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *FriendInfo) ReadFrom(r *JceReader) {
	pkt.FriendUin = r.ReadInt64(0)
	pkt.GroupId = r.ReadByte(1)
	pkt.FaceId = r.ReadInt16(2)
	pkt.Remark = r.ReadString(3)
	pkt.Status = r.ReadByte(5)
	pkt.MemberLevel = r.ReadByte(6)
	pkt.Nick = r.ReadString(14)
	pkt.Network = r.ReadByte(20)
	pkt.NetworkType = r.ReadInt32(24)
	pkt.CardID = []byte{}
	r.ReadObject(&pkt.CardID, 41)
}

func (pkt *TroopListRequest) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *TroopNumber) ReadFrom(r *JceReader) {
	pkt.GroupUin = r.ReadInt64(0)
	pkt.GroupCode = r.ReadInt64(1)
	pkt.GroupName = r.ReadString(4)
	pkt.GroupMemo = r.ReadString(5)
	pkt.MemberNum = r.ReadInt64(19)
	pkt.GroupOwnerUin = r.ReadInt64(23)
	pkt.MaxGroupMemberNum = r.ReadInt64(29)
}

func (pkt *TroopMemberListRequest) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *TroopMemberInfo) ReadFrom(r *JceReader) {
	pkt.MemberUin = r.ReadInt64(0)
	pkt.FaceId = r.ReadInt16(1)
	pkt.Gender = r.ReadByte(3)
	pkt.Nick = r.ReadString(4)
	pkt.ShowName = r.ReadString(6)
	pkt.Name = r.ReadString(8)
	pkt.AutoRemark = r.ReadString(13)
	pkt.MemberLevel = r.ReadInt64(14)
	pkt.JoinTime = r.ReadInt64(15)
	pkt.LastSpeakTime = r.ReadInt64(16)
	pkt.Flag = r.ReadInt64(18)
	pkt.SpecialTitle = r.ReadString(23)
	pkt.SpecialTitleExpireTime = r.ReadInt64(24)
}

func (pkt *PushMessageInfo) ReadFrom(r *JceReader) {
	pkt.FromUin = r.ReadInt64(0)
	pkt.MsgTime = r.ReadInt64(1)
	pkt.MsgType = r.ReadInt16(2)
	pkt.MsgSeq = r.ReadInt16(3)
	pkt.Msg = r.ReadString(4)
	pkt.VMsg = r.ReadAny(6).([]byte)
	pkt.MsgCookies = r.ReadAny(8).([]byte)
	pkt.MsgUid = r.ReadInt64(10)
	pkt.FromMobile = r.ReadString(16)
	pkt.FromName = r.ReadString(17)
}

func (pkt *SvcDevLoginInfo) ReadFrom(r *JceReader) {
	pkt.AppId = r.ReadInt64(0)
	pkt.Guid = []byte{}
	r.ReadSlice(&pkt.Guid, 1)
	pkt.LoginTime = r.ReadInt64(2)
	pkt.LoginPlatform = r.ReadInt64(3)
	pkt.LoginLocation = r.ReadString(4)
	pkt.DeviceName = r.ReadString(5)
	pkt.DeviceTypeInfo = r.ReadString(6)
	pkt.TerType = r.ReadInt64(8)
	pkt.ProductType = r.ReadInt64(9)
	pkt.CanBeKicked = r.ReadInt64(10)
}

func (pkt *SvcRespPushMsg) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *ModifyGroupCardRequest) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}

func (pkt *SvcReqGetDevLoginInfo) ToBytes() []byte {
	w := NewJceWriter()
	w.WriteJceStructRaw(pkt)
	return w.Bytes()
}
