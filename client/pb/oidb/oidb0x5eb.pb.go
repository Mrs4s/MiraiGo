// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/oidb/oidb0x5eb.proto

package oidb

type D5EBReqBody struct {
	Uins                              []uint64 `protobuf:"varint,1,rep"`
	MaxPackageSize                    *uint32  `protobuf:"varint,3,opt"`
	Openid                            [][]byte `protobuf:"bytes,4,rep"`
	Appid                             *uint32  `protobuf:"varint,5,opt"`
	ReqNick                           *uint32  `protobuf:"varint,20002,opt"`
	ReqCountry                        *uint32  `protobuf:"varint,20003,opt"`
	ReqProvince                       *int32   `protobuf:"varint,20004,opt"`
	ReqGender                         *int32   `protobuf:"varint,20009,opt"`
	ReqAllow                          *int32   `protobuf:"varint,20014,opt"`
	ReqFaceId                         *int32   `protobuf:"varint,20015,opt"`
	ReqCity                           *int32   `protobuf:"varint,20020,opt"`
	ReqConstellation                  *int32   `protobuf:"varint,20022,opt"`
	ReqCommonPlace1                   *int32   `protobuf:"varint,20027,opt"`
	ReqMss3Bitmapextra                *int32   `protobuf:"varint,20030,opt"`
	ReqBirthday                       *int32   `protobuf:"varint,20031,opt"`
	ReqCityId                         *int32   `protobuf:"varint,20032,opt"`
	ReqLang1                          *int32   `protobuf:"varint,20033,opt"`
	ReqLang2                          *int32   `protobuf:"varint,20034,opt"`
	ReqLang3                          *int32   `protobuf:"varint,20035,opt"`
	ReqAge                            *int32   `protobuf:"varint,20037,opt"`
	ReqCityZoneId                     *int32   `protobuf:"varint,20041,opt"`
	ReqOin                            *int32   `protobuf:"varint,20056,opt"`
	ReqBubbleId                       *int32   `protobuf:"varint,20059,opt"`
	ReqMss2Identity                   *int32   `protobuf:"varint,21001,opt"`
	ReqMss1Service                    *int32   `protobuf:"varint,21002,opt"`
	ReqLflag                          *int32   `protobuf:"varint,21003,opt"`
	ReqExtFlag                        *int32   `protobuf:"varint,21004,opt"`
	ReqBasicSvrFlag                   *int32   `protobuf:"varint,21006,opt"`
	ReqBasicCliFlag                   *int32   `protobuf:"varint,21007,opt"`
	ReqFullBirthday                   *int32   `protobuf:"varint,26004,opt"`
	ReqFullAge                        *int32   `protobuf:"varint,26005,opt"`
	ReqSimpleUpdateTime               *int32   `protobuf:"varint,26010,opt"`
	ReqMssUpdateTime                  *int32   `protobuf:"varint,26011,opt"`
	ReqPstnMultiCallTime              *int32   `protobuf:"varint,26012,opt"`
	ReqPstnMultiLastGuideRechargeTime *int32   `protobuf:"varint,26013,opt"`
	ReqPstnC2CCallTime                *int32   `protobuf:"varint,26014,opt"`
	ReqPstnC2CLastGuideRechargeTime   *int32   `protobuf:"varint,26015,opt"`
	ReqGroupMemCreditFlag             *int32   `protobuf:"varint,27022,opt"`
	ReqFaceAddonId                    *int32   `protobuf:"varint,27025,opt"`
	ReqMusicGene                      *int32   `protobuf:"varint,27026,opt"`
	ReqStrangerNick                   *int32   `protobuf:"varint,27034,opt"`
	ReqStrangerDeclare                *int32   `protobuf:"varint,27035,opt"`
	ReqLoveStatus                     *int32   `protobuf:"varint,27036,opt"`
	ReqProfession                     *int32   `protobuf:"varint,27037,opt"`
	ReqVasColorringFlag               *int32   `protobuf:"varint,27041,opt"`
	ReqCharm                          *int32   `protobuf:"varint,27052,opt"`
	ReqApolloTimestamp                *int32   `protobuf:"varint,27059,opt"`
	ReqVasFontIdFlag                  *int32   `protobuf:"varint,27201,opt"`
	ReqGlobalGroupLevel               *int32   `protobuf:"varint,27208,opt"`
	ReqInvite2GroupAutoAgreeFlag      *int32   `protobuf:"varint,40346,opt"`
	ReqSubaccountDisplayThirdQqFlag   *int32   `protobuf:"varint,40348,opt"`
	NotifyPartakeLikeRankingListFlag  *int32   `protobuf:"varint,40350,opt"`
	ReqLightalkSwitch                 *int32   `protobuf:"varint,40506,opt"`
	ReqMusicRingVisible               *int32   `protobuf:"varint,40507,opt"`
	ReqMusicRingAutoplay              *int32   `protobuf:"varint,40508,opt"`
	ReqMusicRingRedpoint              *int32   `protobuf:"varint,40509,opt"`
	TorchDisableFlag                  *int32   `protobuf:"varint,40525,opt"`
	ReqVasMagicfontFlag               *int32   `protobuf:"varint,40530,opt"`
	ReqVipFlag                        *int32   `protobuf:"varint,41756,opt"`
	ReqAuthFlag                       *int32   `protobuf:"varint,41783,opt"`
	ReqForbidFlag                     *int32   `protobuf:"varint,41784,opt"`
	ReqGodForbid                      *int32   `protobuf:"varint,41804,opt"`
	ReqGodFlag                        *int32   `protobuf:"varint,41805,opt"`
	ReqCharmLevel                     *int32   `protobuf:"varint,41950,opt"`
	ReqCharmShown                     *int32   `protobuf:"varint,41973,opt"`
	ReqFreshnewsNotifyFlag            *int32   `protobuf:"varint,41993,opt"`
	ReqApolloVipLevel                 *int32   `protobuf:"varint,41999,opt"`
	ReqApolloVipFlag                  *int32   `protobuf:"varint,42003,opt"`
	ReqPstnC2CVip                     *int32   `protobuf:"varint,42005,opt"`
	ReqPstnMultiVip                   *int32   `protobuf:"varint,42006,opt"`
	ReqPstnEverC2CVip                 *int32   `protobuf:"varint,42007,opt"`
	ReqPstnEverMultiVip               *int32   `protobuf:"varint,42008,opt"`
	ReqPstnMultiTryFlag               *int32   `protobuf:"varint,42011,opt"`
	ReqPstnC2CTryFlag                 *int32   `protobuf:"varint,42012,opt"`
	ReqSubscribeNearbyassistantSwitch *int32   `protobuf:"varint,42024,opt"`
	ReqTorchbearerFlag                *int32   `protobuf:"varint,42051,opt"`
	PreloadDisableFlag                *int32   `protobuf:"varint,42073,opt"`
	ReqMedalwallFlag                  *int32   `protobuf:"varint,42075,opt"`
	NotifyOnLikeRankingListFlag       *int32   `protobuf:"varint,42092,opt"`
	ReqApolloStatus                   *int32   `protobuf:"varint,42980,opt"`
}

func (x *D5EBReqBody) GetUins() []uint64 {
	if x != nil {
		return x.Uins
	}
	return nil
}

func (x *D5EBReqBody) GetMaxPackageSize() uint32 {
	if x != nil && x.MaxPackageSize != nil {
		return *x.MaxPackageSize
	}
	return 0
}

func (x *D5EBReqBody) GetOpenid() [][]byte {
	if x != nil {
		return x.Openid
	}
	return nil
}

func (x *D5EBReqBody) GetAppid() uint32 {
	if x != nil && x.Appid != nil {
		return *x.Appid
	}
	return 0
}

func (x *D5EBReqBody) GetReqNick() uint32 {
	if x != nil && x.ReqNick != nil {
		return *x.ReqNick
	}
	return 0
}

func (x *D5EBReqBody) GetReqCountry() uint32 {
	if x != nil && x.ReqCountry != nil {
		return *x.ReqCountry
	}
	return 0
}

func (x *D5EBReqBody) GetReqProvince() int32 {
	if x != nil && x.ReqProvince != nil {
		return *x.ReqProvince
	}
	return 0
}

func (x *D5EBReqBody) GetReqGender() int32 {
	if x != nil && x.ReqGender != nil {
		return *x.ReqGender
	}
	return 0
}

func (x *D5EBReqBody) GetReqAllow() int32 {
	if x != nil && x.ReqAllow != nil {
		return *x.ReqAllow
	}
	return 0
}

func (x *D5EBReqBody) GetReqFaceId() int32 {
	if x != nil && x.ReqFaceId != nil {
		return *x.ReqFaceId
	}
	return 0
}

func (x *D5EBReqBody) GetReqCity() int32 {
	if x != nil && x.ReqCity != nil {
		return *x.ReqCity
	}
	return 0
}

func (x *D5EBReqBody) GetReqConstellation() int32 {
	if x != nil && x.ReqConstellation != nil {
		return *x.ReqConstellation
	}
	return 0
}

func (x *D5EBReqBody) GetReqCommonPlace1() int32 {
	if x != nil && x.ReqCommonPlace1 != nil {
		return *x.ReqCommonPlace1
	}
	return 0
}

func (x *D5EBReqBody) GetReqMss3Bitmapextra() int32 {
	if x != nil && x.ReqMss3Bitmapextra != nil {
		return *x.ReqMss3Bitmapextra
	}
	return 0
}

func (x *D5EBReqBody) GetReqBirthday() int32 {
	if x != nil && x.ReqBirthday != nil {
		return *x.ReqBirthday
	}
	return 0
}

func (x *D5EBReqBody) GetReqCityId() int32 {
	if x != nil && x.ReqCityId != nil {
		return *x.ReqCityId
	}
	return 0
}

func (x *D5EBReqBody) GetReqLang1() int32 {
	if x != nil && x.ReqLang1 != nil {
		return *x.ReqLang1
	}
	return 0
}

func (x *D5EBReqBody) GetReqLang2() int32 {
	if x != nil && x.ReqLang2 != nil {
		return *x.ReqLang2
	}
	return 0
}

func (x *D5EBReqBody) GetReqLang3() int32 {
	if x != nil && x.ReqLang3 != nil {
		return *x.ReqLang3
	}
	return 0
}

func (x *D5EBReqBody) GetReqAge() int32 {
	if x != nil && x.ReqAge != nil {
		return *x.ReqAge
	}
	return 0
}

func (x *D5EBReqBody) GetReqCityZoneId() int32 {
	if x != nil && x.ReqCityZoneId != nil {
		return *x.ReqCityZoneId
	}
	return 0
}

func (x *D5EBReqBody) GetReqOin() int32 {
	if x != nil && x.ReqOin != nil {
		return *x.ReqOin
	}
	return 0
}

func (x *D5EBReqBody) GetReqBubbleId() int32 {
	if x != nil && x.ReqBubbleId != nil {
		return *x.ReqBubbleId
	}
	return 0
}

func (x *D5EBReqBody) GetReqMss2Identity() int32 {
	if x != nil && x.ReqMss2Identity != nil {
		return *x.ReqMss2Identity
	}
	return 0
}

func (x *D5EBReqBody) GetReqMss1Service() int32 {
	if x != nil && x.ReqMss1Service != nil {
		return *x.ReqMss1Service
	}
	return 0
}

func (x *D5EBReqBody) GetReqLflag() int32 {
	if x != nil && x.ReqLflag != nil {
		return *x.ReqLflag
	}
	return 0
}

func (x *D5EBReqBody) GetReqExtFlag() int32 {
	if x != nil && x.ReqExtFlag != nil {
		return *x.ReqExtFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqBasicSvrFlag() int32 {
	if x != nil && x.ReqBasicSvrFlag != nil {
		return *x.ReqBasicSvrFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqBasicCliFlag() int32 {
	if x != nil && x.ReqBasicCliFlag != nil {
		return *x.ReqBasicCliFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqFullBirthday() int32 {
	if x != nil && x.ReqFullBirthday != nil {
		return *x.ReqFullBirthday
	}
	return 0
}

func (x *D5EBReqBody) GetReqFullAge() int32 {
	if x != nil && x.ReqFullAge != nil {
		return *x.ReqFullAge
	}
	return 0
}

func (x *D5EBReqBody) GetReqSimpleUpdateTime() int32 {
	if x != nil && x.ReqSimpleUpdateTime != nil {
		return *x.ReqSimpleUpdateTime
	}
	return 0
}

func (x *D5EBReqBody) GetReqMssUpdateTime() int32 {
	if x != nil && x.ReqMssUpdateTime != nil {
		return *x.ReqMssUpdateTime
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnMultiCallTime() int32 {
	if x != nil && x.ReqPstnMultiCallTime != nil {
		return *x.ReqPstnMultiCallTime
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnMultiLastGuideRechargeTime() int32 {
	if x != nil && x.ReqPstnMultiLastGuideRechargeTime != nil {
		return *x.ReqPstnMultiLastGuideRechargeTime
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnC2CCallTime() int32 {
	if x != nil && x.ReqPstnC2CCallTime != nil {
		return *x.ReqPstnC2CCallTime
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnC2CLastGuideRechargeTime() int32 {
	if x != nil && x.ReqPstnC2CLastGuideRechargeTime != nil {
		return *x.ReqPstnC2CLastGuideRechargeTime
	}
	return 0
}

func (x *D5EBReqBody) GetReqGroupMemCreditFlag() int32 {
	if x != nil && x.ReqGroupMemCreditFlag != nil {
		return *x.ReqGroupMemCreditFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqFaceAddonId() int32 {
	if x != nil && x.ReqFaceAddonId != nil {
		return *x.ReqFaceAddonId
	}
	return 0
}

func (x *D5EBReqBody) GetReqMusicGene() int32 {
	if x != nil && x.ReqMusicGene != nil {
		return *x.ReqMusicGene
	}
	return 0
}

func (x *D5EBReqBody) GetReqStrangerNick() int32 {
	if x != nil && x.ReqStrangerNick != nil {
		return *x.ReqStrangerNick
	}
	return 0
}

func (x *D5EBReqBody) GetReqStrangerDeclare() int32 {
	if x != nil && x.ReqStrangerDeclare != nil {
		return *x.ReqStrangerDeclare
	}
	return 0
}

func (x *D5EBReqBody) GetReqLoveStatus() int32 {
	if x != nil && x.ReqLoveStatus != nil {
		return *x.ReqLoveStatus
	}
	return 0
}

func (x *D5EBReqBody) GetReqProfession() int32 {
	if x != nil && x.ReqProfession != nil {
		return *x.ReqProfession
	}
	return 0
}

func (x *D5EBReqBody) GetReqVasColorringFlag() int32 {
	if x != nil && x.ReqVasColorringFlag != nil {
		return *x.ReqVasColorringFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqCharm() int32 {
	if x != nil && x.ReqCharm != nil {
		return *x.ReqCharm
	}
	return 0
}

func (x *D5EBReqBody) GetReqApolloTimestamp() int32 {
	if x != nil && x.ReqApolloTimestamp != nil {
		return *x.ReqApolloTimestamp
	}
	return 0
}

func (x *D5EBReqBody) GetReqVasFontIdFlag() int32 {
	if x != nil && x.ReqVasFontIdFlag != nil {
		return *x.ReqVasFontIdFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqGlobalGroupLevel() int32 {
	if x != nil && x.ReqGlobalGroupLevel != nil {
		return *x.ReqGlobalGroupLevel
	}
	return 0
}

func (x *D5EBReqBody) GetReqInvite2GroupAutoAgreeFlag() int32 {
	if x != nil && x.ReqInvite2GroupAutoAgreeFlag != nil {
		return *x.ReqInvite2GroupAutoAgreeFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqSubaccountDisplayThirdQqFlag() int32 {
	if x != nil && x.ReqSubaccountDisplayThirdQqFlag != nil {
		return *x.ReqSubaccountDisplayThirdQqFlag
	}
	return 0
}

func (x *D5EBReqBody) GetNotifyPartakeLikeRankingListFlag() int32 {
	if x != nil && x.NotifyPartakeLikeRankingListFlag != nil {
		return *x.NotifyPartakeLikeRankingListFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqLightalkSwitch() int32 {
	if x != nil && x.ReqLightalkSwitch != nil {
		return *x.ReqLightalkSwitch
	}
	return 0
}

func (x *D5EBReqBody) GetReqMusicRingVisible() int32 {
	if x != nil && x.ReqMusicRingVisible != nil {
		return *x.ReqMusicRingVisible
	}
	return 0
}

func (x *D5EBReqBody) GetReqMusicRingAutoplay() int32 {
	if x != nil && x.ReqMusicRingAutoplay != nil {
		return *x.ReqMusicRingAutoplay
	}
	return 0
}

func (x *D5EBReqBody) GetReqMusicRingRedpoint() int32 {
	if x != nil && x.ReqMusicRingRedpoint != nil {
		return *x.ReqMusicRingRedpoint
	}
	return 0
}

func (x *D5EBReqBody) GetTorchDisableFlag() int32 {
	if x != nil && x.TorchDisableFlag != nil {
		return *x.TorchDisableFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqVasMagicfontFlag() int32 {
	if x != nil && x.ReqVasMagicfontFlag != nil {
		return *x.ReqVasMagicfontFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqVipFlag() int32 {
	if x != nil && x.ReqVipFlag != nil {
		return *x.ReqVipFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqAuthFlag() int32 {
	if x != nil && x.ReqAuthFlag != nil {
		return *x.ReqAuthFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqForbidFlag() int32 {
	if x != nil && x.ReqForbidFlag != nil {
		return *x.ReqForbidFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqGodForbid() int32 {
	if x != nil && x.ReqGodForbid != nil {
		return *x.ReqGodForbid
	}
	return 0
}

func (x *D5EBReqBody) GetReqGodFlag() int32 {
	if x != nil && x.ReqGodFlag != nil {
		return *x.ReqGodFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqCharmLevel() int32 {
	if x != nil && x.ReqCharmLevel != nil {
		return *x.ReqCharmLevel
	}
	return 0
}

func (x *D5EBReqBody) GetReqCharmShown() int32 {
	if x != nil && x.ReqCharmShown != nil {
		return *x.ReqCharmShown
	}
	return 0
}

func (x *D5EBReqBody) GetReqFreshnewsNotifyFlag() int32 {
	if x != nil && x.ReqFreshnewsNotifyFlag != nil {
		return *x.ReqFreshnewsNotifyFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqApolloVipLevel() int32 {
	if x != nil && x.ReqApolloVipLevel != nil {
		return *x.ReqApolloVipLevel
	}
	return 0
}

func (x *D5EBReqBody) GetReqApolloVipFlag() int32 {
	if x != nil && x.ReqApolloVipFlag != nil {
		return *x.ReqApolloVipFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnC2CVip() int32 {
	if x != nil && x.ReqPstnC2CVip != nil {
		return *x.ReqPstnC2CVip
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnMultiVip() int32 {
	if x != nil && x.ReqPstnMultiVip != nil {
		return *x.ReqPstnMultiVip
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnEverC2CVip() int32 {
	if x != nil && x.ReqPstnEverC2CVip != nil {
		return *x.ReqPstnEverC2CVip
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnEverMultiVip() int32 {
	if x != nil && x.ReqPstnEverMultiVip != nil {
		return *x.ReqPstnEverMultiVip
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnMultiTryFlag() int32 {
	if x != nil && x.ReqPstnMultiTryFlag != nil {
		return *x.ReqPstnMultiTryFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqPstnC2CTryFlag() int32 {
	if x != nil && x.ReqPstnC2CTryFlag != nil {
		return *x.ReqPstnC2CTryFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqSubscribeNearbyassistantSwitch() int32 {
	if x != nil && x.ReqSubscribeNearbyassistantSwitch != nil {
		return *x.ReqSubscribeNearbyassistantSwitch
	}
	return 0
}

func (x *D5EBReqBody) GetReqTorchbearerFlag() int32 {
	if x != nil && x.ReqTorchbearerFlag != nil {
		return *x.ReqTorchbearerFlag
	}
	return 0
}

func (x *D5EBReqBody) GetPreloadDisableFlag() int32 {
	if x != nil && x.PreloadDisableFlag != nil {
		return *x.PreloadDisableFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqMedalwallFlag() int32 {
	if x != nil && x.ReqMedalwallFlag != nil {
		return *x.ReqMedalwallFlag
	}
	return 0
}

func (x *D5EBReqBody) GetNotifyOnLikeRankingListFlag() int32 {
	if x != nil && x.NotifyOnLikeRankingListFlag != nil {
		return *x.NotifyOnLikeRankingListFlag
	}
	return 0
}

func (x *D5EBReqBody) GetReqApolloStatus() int32 {
	if x != nil && x.ReqApolloStatus != nil {
		return *x.ReqApolloStatus
	}
	return 0
}

type D5EBRspBody struct {
	UinData        []*UdcUinData `protobuf:"bytes,11,rep"`
	UnfinishedUins []int64       `protobuf:"varint,12,rep"`
}

func (x *D5EBRspBody) GetUinData() []*UdcUinData {
	if x != nil {
		return x.UinData
	}
	return nil
}

func (x *D5EBRspBody) GetUnfinishedUins() []int64 {
	if x != nil {
		return x.UnfinishedUins
	}
	return nil
}

type UdcUinData struct {
	Uin                              *int64 `protobuf:"varint,1,opt"`
	Openid                           []byte `protobuf:"bytes,4,opt"`
	Nick                             []byte `protobuf:"bytes,20002,opt"`
	Country                          []byte `protobuf:"bytes,20003,opt"`
	Province                         []byte `protobuf:"bytes,20004,opt"`
	Gender                           *int32 `protobuf:"varint,20009,opt"`
	Allow                            *int32 `protobuf:"varint,20014,opt"`
	FaceId                           *int32 `protobuf:"varint,20015,opt"`
	City                             []byte `protobuf:"bytes,20020,opt"`
	Constellation                    *int32 `protobuf:"varint,20022,opt"`
	CommonPlace1                     *int32 `protobuf:"varint,20027,opt"`
	Mss3Bitmapextra                  []byte `protobuf:"bytes,20030,opt"`
	Birthday                         []byte `protobuf:"bytes,20031,opt"`
	CityId                           []byte `protobuf:"bytes,20032,opt"`
	Lang1                            *int32 `protobuf:"varint,20033,opt"`
	Lang2                            *int32 `protobuf:"varint,20034,opt"`
	Lang3                            *int32 `protobuf:"varint,20035,opt"`
	Age                              *int32 `protobuf:"varint,20037,opt"`
	CityZoneId                       *int32 `protobuf:"varint,20041,opt"`
	Oin                              *int32 `protobuf:"varint,20056,opt"`
	BubbleId                         *int32 `protobuf:"varint,20059,opt"`
	Mss2Identity                     []byte `protobuf:"bytes,21001,opt"`
	Mss1Service                      []byte `protobuf:"bytes,21002,opt"`
	Lflag                            *int32 `protobuf:"varint,21003,opt"`
	ExtFlag                          *int32 `protobuf:"varint,21004,opt"`
	BasicSvrFlag                     []byte `protobuf:"bytes,21006,opt"`
	BasicCliFlag                     []byte `protobuf:"bytes,21007,opt"`
	FullBirthday                     []byte `protobuf:"bytes,26004,opt"`
	FullAge                          []byte `protobuf:"bytes,26005,opt"`
	SimpleUpdateTime                 *int32 `protobuf:"varint,26010,opt"`
	MssUpdateTime                    *int32 `protobuf:"varint,26011,opt"`
	PstnMultiCallTime                *int32 `protobuf:"varint,26012,opt"`
	PstnMultiLastGuideRechargeTime   *int32 `protobuf:"varint,26013,opt"`
	PstnC2CCallTime                  *int32 `protobuf:"varint,26014,opt"`
	PstnC2CLastGuideRechargeTime     *int32 `protobuf:"varint,26015,opt"`
	GroupMemCreditFlag               *int32 `protobuf:"varint,27022,opt"`
	FaceAddonId                      *int64 `protobuf:"varint,27025,opt"`
	MusicGene                        []byte `protobuf:"bytes,27026,opt"`
	StrangerNick                     []byte `protobuf:"bytes,27034,opt"`
	StrangerDeclare                  []byte `protobuf:"bytes,27035,opt"`
	LoveStatus                       *int32 `protobuf:"varint,27036,opt"`
	Profession                       *int32 `protobuf:"varint,27037,opt"`
	VasColorringId                   *int32 `protobuf:"varint,27041,opt"`
	Charm                            *int32 `protobuf:"varint,27052,opt"`
	ApolloTimestamp                  *int32 `protobuf:"varint,27059,opt"`
	VasFontId                        *int32 `protobuf:"varint,27201,opt"`
	GlobalGroupLevel                 *int32 `protobuf:"varint,27208,opt"`
	ReqInvite2GroupAutoAgreeFlag     *int32 `protobuf:"varint,40346,opt"`
	SubaccountDisplayThirdQqFlag     *int32 `protobuf:"varint,40348,opt"`
	NotifyPartakeLikeRankingListFlag *int32 `protobuf:"varint,40350,opt"`
	LightalkSwitch                   *int32 `protobuf:"varint,40506,opt"`
	MusicRingVisible                 *int32 `protobuf:"varint,40507,opt"`
	MusicRingAutoplay                *int32 `protobuf:"varint,40508,opt"`
	MusicRingRedpoint                *int32 `protobuf:"varint,40509,opt"`
	TorchDisableFlag                 *int32 `protobuf:"varint,40525,opt"`
	VasMagicfontFlag                 *int32 `protobuf:"varint,40530,opt"`
	VipFlag                          *int32 `protobuf:"varint,41756,opt"`
	AuthFlag                         *int32 `protobuf:"varint,41783,opt"`
	ForbidFlag                       *int32 `protobuf:"varint,41784,opt"`
	GodForbid                        *int32 `protobuf:"varint,41804,opt"`
	GodFlag                          *int32 `protobuf:"varint,41805,opt"`
	CharmLevel                       *int32 `protobuf:"varint,41950,opt"`
	CharmShown                       *int32 `protobuf:"varint,41973,opt"`
	FreshnewsNotifyFlag              *int32 `protobuf:"varint,41993,opt"`
	ApolloVipLevel                   *int32 `protobuf:"varint,41999,opt"`
	ApolloVipFlag                    *int32 `protobuf:"varint,42003,opt"`
	PstnC2CVip                       *int32 `protobuf:"varint,42005,opt"`
	PstnMultiVip                     *int32 `protobuf:"varint,42006,opt"`
	PstnEverC2CVip                   *int32 `protobuf:"varint,42007,opt"`
	PstnEverMultiVip                 *int32 `protobuf:"varint,42008,opt"`
	PstnMultiTryFlag                 *int32 `protobuf:"varint,42011,opt"`
	PstnC2CTryFlag                   *int32 `protobuf:"varint,42012,opt"`
	SubscribeNearbyassistantSwitch   *int32 `protobuf:"varint,42024,opt"`
	TorchbearerFlag                  *int32 `protobuf:"varint,42051,opt"`
	PreloadDisableFlag               *int32 `protobuf:"varint,42073,opt"`
	ReqMedalwallFlag                 *int32 `protobuf:"varint,42075,opt"`
	NotifyOnLikeRankingListFlag      *int32 `protobuf:"varint,42092,opt"`
	ApolloStatus                     *int32 `protobuf:"varint,42980,opt"`
}

func (x *UdcUinData) GetUin() int64 {
	if x != nil && x.Uin != nil {
		return *x.Uin
	}
	return 0
}

func (x *UdcUinData) GetOpenid() []byte {
	if x != nil {
		return x.Openid
	}
	return nil
}

func (x *UdcUinData) GetNick() []byte {
	if x != nil {
		return x.Nick
	}
	return nil
}

func (x *UdcUinData) GetCountry() []byte {
	if x != nil {
		return x.Country
	}
	return nil
}

func (x *UdcUinData) GetProvince() []byte {
	if x != nil {
		return x.Province
	}
	return nil
}

func (x *UdcUinData) GetGender() int32 {
	if x != nil && x.Gender != nil {
		return *x.Gender
	}
	return 0
}

func (x *UdcUinData) GetAllow() int32 {
	if x != nil && x.Allow != nil {
		return *x.Allow
	}
	return 0
}

func (x *UdcUinData) GetFaceId() int32 {
	if x != nil && x.FaceId != nil {
		return *x.FaceId
	}
	return 0
}

func (x *UdcUinData) GetCity() []byte {
	if x != nil {
		return x.City
	}
	return nil
}

func (x *UdcUinData) GetConstellation() int32 {
	if x != nil && x.Constellation != nil {
		return *x.Constellation
	}
	return 0
}

func (x *UdcUinData) GetCommonPlace1() int32 {
	if x != nil && x.CommonPlace1 != nil {
		return *x.CommonPlace1
	}
	return 0
}

func (x *UdcUinData) GetMss3Bitmapextra() []byte {
	if x != nil {
		return x.Mss3Bitmapextra
	}
	return nil
}

func (x *UdcUinData) GetBirthday() []byte {
	if x != nil {
		return x.Birthday
	}
	return nil
}

func (x *UdcUinData) GetCityId() []byte {
	if x != nil {
		return x.CityId
	}
	return nil
}

func (x *UdcUinData) GetLang1() int32 {
	if x != nil && x.Lang1 != nil {
		return *x.Lang1
	}
	return 0
}

func (x *UdcUinData) GetLang2() int32 {
	if x != nil && x.Lang2 != nil {
		return *x.Lang2
	}
	return 0
}

func (x *UdcUinData) GetLang3() int32 {
	if x != nil && x.Lang3 != nil {
		return *x.Lang3
	}
	return 0
}

func (x *UdcUinData) GetAge() int32 {
	if x != nil && x.Age != nil {
		return *x.Age
	}
	return 0
}

func (x *UdcUinData) GetCityZoneId() int32 {
	if x != nil && x.CityZoneId != nil {
		return *x.CityZoneId
	}
	return 0
}

func (x *UdcUinData) GetOin() int32 {
	if x != nil && x.Oin != nil {
		return *x.Oin
	}
	return 0
}

func (x *UdcUinData) GetBubbleId() int32 {
	if x != nil && x.BubbleId != nil {
		return *x.BubbleId
	}
	return 0
}

func (x *UdcUinData) GetMss2Identity() []byte {
	if x != nil {
		return x.Mss2Identity
	}
	return nil
}

func (x *UdcUinData) GetMss1Service() []byte {
	if x != nil {
		return x.Mss1Service
	}
	return nil
}

func (x *UdcUinData) GetLflag() int32 {
	if x != nil && x.Lflag != nil {
		return *x.Lflag
	}
	return 0
}

func (x *UdcUinData) GetExtFlag() int32 {
	if x != nil && x.ExtFlag != nil {
		return *x.ExtFlag
	}
	return 0
}

func (x *UdcUinData) GetBasicSvrFlag() []byte {
	if x != nil {
		return x.BasicSvrFlag
	}
	return nil
}

func (x *UdcUinData) GetBasicCliFlag() []byte {
	if x != nil {
		return x.BasicCliFlag
	}
	return nil
}

func (x *UdcUinData) GetFullBirthday() []byte {
	if x != nil {
		return x.FullBirthday
	}
	return nil
}

func (x *UdcUinData) GetFullAge() []byte {
	if x != nil {
		return x.FullAge
	}
	return nil
}

func (x *UdcUinData) GetSimpleUpdateTime() int32 {
	if x != nil && x.SimpleUpdateTime != nil {
		return *x.SimpleUpdateTime
	}
	return 0
}

func (x *UdcUinData) GetMssUpdateTime() int32 {
	if x != nil && x.MssUpdateTime != nil {
		return *x.MssUpdateTime
	}
	return 0
}

func (x *UdcUinData) GetPstnMultiCallTime() int32 {
	if x != nil && x.PstnMultiCallTime != nil {
		return *x.PstnMultiCallTime
	}
	return 0
}

func (x *UdcUinData) GetPstnMultiLastGuideRechargeTime() int32 {
	if x != nil && x.PstnMultiLastGuideRechargeTime != nil {
		return *x.PstnMultiLastGuideRechargeTime
	}
	return 0
}

func (x *UdcUinData) GetPstnC2CCallTime() int32 {
	if x != nil && x.PstnC2CCallTime != nil {
		return *x.PstnC2CCallTime
	}
	return 0
}

func (x *UdcUinData) GetPstnC2CLastGuideRechargeTime() int32 {
	if x != nil && x.PstnC2CLastGuideRechargeTime != nil {
		return *x.PstnC2CLastGuideRechargeTime
	}
	return 0
}

func (x *UdcUinData) GetGroupMemCreditFlag() int32 {
	if x != nil && x.GroupMemCreditFlag != nil {
		return *x.GroupMemCreditFlag
	}
	return 0
}

func (x *UdcUinData) GetFaceAddonId() int64 {
	if x != nil && x.FaceAddonId != nil {
		return *x.FaceAddonId
	}
	return 0
}

func (x *UdcUinData) GetMusicGene() []byte {
	if x != nil {
		return x.MusicGene
	}
	return nil
}

func (x *UdcUinData) GetStrangerNick() []byte {
	if x != nil {
		return x.StrangerNick
	}
	return nil
}

func (x *UdcUinData) GetStrangerDeclare() []byte {
	if x != nil {
		return x.StrangerDeclare
	}
	return nil
}

func (x *UdcUinData) GetLoveStatus() int32 {
	if x != nil && x.LoveStatus != nil {
		return *x.LoveStatus
	}
	return 0
}

func (x *UdcUinData) GetProfession() int32 {
	if x != nil && x.Profession != nil {
		return *x.Profession
	}
	return 0
}

func (x *UdcUinData) GetVasColorringId() int32 {
	if x != nil && x.VasColorringId != nil {
		return *x.VasColorringId
	}
	return 0
}

func (x *UdcUinData) GetCharm() int32 {
	if x != nil && x.Charm != nil {
		return *x.Charm
	}
	return 0
}

func (x *UdcUinData) GetApolloTimestamp() int32 {
	if x != nil && x.ApolloTimestamp != nil {
		return *x.ApolloTimestamp
	}
	return 0
}

func (x *UdcUinData) GetVasFontId() int32 {
	if x != nil && x.VasFontId != nil {
		return *x.VasFontId
	}
	return 0
}

func (x *UdcUinData) GetGlobalGroupLevel() int32 {
	if x != nil && x.GlobalGroupLevel != nil {
		return *x.GlobalGroupLevel
	}
	return 0
}

func (x *UdcUinData) GetReqInvite2GroupAutoAgreeFlag() int32 {
	if x != nil && x.ReqInvite2GroupAutoAgreeFlag != nil {
		return *x.ReqInvite2GroupAutoAgreeFlag
	}
	return 0
}

func (x *UdcUinData) GetSubaccountDisplayThirdQqFlag() int32 {
	if x != nil && x.SubaccountDisplayThirdQqFlag != nil {
		return *x.SubaccountDisplayThirdQqFlag
	}
	return 0
}

func (x *UdcUinData) GetNotifyPartakeLikeRankingListFlag() int32 {
	if x != nil && x.NotifyPartakeLikeRankingListFlag != nil {
		return *x.NotifyPartakeLikeRankingListFlag
	}
	return 0
}

func (x *UdcUinData) GetLightalkSwitch() int32 {
	if x != nil && x.LightalkSwitch != nil {
		return *x.LightalkSwitch
	}
	return 0
}

func (x *UdcUinData) GetMusicRingVisible() int32 {
	if x != nil && x.MusicRingVisible != nil {
		return *x.MusicRingVisible
	}
	return 0
}

func (x *UdcUinData) GetMusicRingAutoplay() int32 {
	if x != nil && x.MusicRingAutoplay != nil {
		return *x.MusicRingAutoplay
	}
	return 0
}

func (x *UdcUinData) GetMusicRingRedpoint() int32 {
	if x != nil && x.MusicRingRedpoint != nil {
		return *x.MusicRingRedpoint
	}
	return 0
}

func (x *UdcUinData) GetTorchDisableFlag() int32 {
	if x != nil && x.TorchDisableFlag != nil {
		return *x.TorchDisableFlag
	}
	return 0
}

func (x *UdcUinData) GetVasMagicfontFlag() int32 {
	if x != nil && x.VasMagicfontFlag != nil {
		return *x.VasMagicfontFlag
	}
	return 0
}

func (x *UdcUinData) GetVipFlag() int32 {
	if x != nil && x.VipFlag != nil {
		return *x.VipFlag
	}
	return 0
}

func (x *UdcUinData) GetAuthFlag() int32 {
	if x != nil && x.AuthFlag != nil {
		return *x.AuthFlag
	}
	return 0
}

func (x *UdcUinData) GetForbidFlag() int32 {
	if x != nil && x.ForbidFlag != nil {
		return *x.ForbidFlag
	}
	return 0
}

func (x *UdcUinData) GetGodForbid() int32 {
	if x != nil && x.GodForbid != nil {
		return *x.GodForbid
	}
	return 0
}

func (x *UdcUinData) GetGodFlag() int32 {
	if x != nil && x.GodFlag != nil {
		return *x.GodFlag
	}
	return 0
}

func (x *UdcUinData) GetCharmLevel() int32 {
	if x != nil && x.CharmLevel != nil {
		return *x.CharmLevel
	}
	return 0
}

func (x *UdcUinData) GetCharmShown() int32 {
	if x != nil && x.CharmShown != nil {
		return *x.CharmShown
	}
	return 0
}

func (x *UdcUinData) GetFreshnewsNotifyFlag() int32 {
	if x != nil && x.FreshnewsNotifyFlag != nil {
		return *x.FreshnewsNotifyFlag
	}
	return 0
}

func (x *UdcUinData) GetApolloVipLevel() int32 {
	if x != nil && x.ApolloVipLevel != nil {
		return *x.ApolloVipLevel
	}
	return 0
}

func (x *UdcUinData) GetApolloVipFlag() int32 {
	if x != nil && x.ApolloVipFlag != nil {
		return *x.ApolloVipFlag
	}
	return 0
}

func (x *UdcUinData) GetPstnC2CVip() int32 {
	if x != nil && x.PstnC2CVip != nil {
		return *x.PstnC2CVip
	}
	return 0
}

func (x *UdcUinData) GetPstnMultiVip() int32 {
	if x != nil && x.PstnMultiVip != nil {
		return *x.PstnMultiVip
	}
	return 0
}

func (x *UdcUinData) GetPstnEverC2CVip() int32 {
	if x != nil && x.PstnEverC2CVip != nil {
		return *x.PstnEverC2CVip
	}
	return 0
}

func (x *UdcUinData) GetPstnEverMultiVip() int32 {
	if x != nil && x.PstnEverMultiVip != nil {
		return *x.PstnEverMultiVip
	}
	return 0
}

func (x *UdcUinData) GetPstnMultiTryFlag() int32 {
	if x != nil && x.PstnMultiTryFlag != nil {
		return *x.PstnMultiTryFlag
	}
	return 0
}

func (x *UdcUinData) GetPstnC2CTryFlag() int32 {
	if x != nil && x.PstnC2CTryFlag != nil {
		return *x.PstnC2CTryFlag
	}
	return 0
}

func (x *UdcUinData) GetSubscribeNearbyassistantSwitch() int32 {
	if x != nil && x.SubscribeNearbyassistantSwitch != nil {
		return *x.SubscribeNearbyassistantSwitch
	}
	return 0
}

func (x *UdcUinData) GetTorchbearerFlag() int32 {
	if x != nil && x.TorchbearerFlag != nil {
		return *x.TorchbearerFlag
	}
	return 0
}

func (x *UdcUinData) GetPreloadDisableFlag() int32 {
	if x != nil && x.PreloadDisableFlag != nil {
		return *x.PreloadDisableFlag
	}
	return 0
}

func (x *UdcUinData) GetReqMedalwallFlag() int32 {
	if x != nil && x.ReqMedalwallFlag != nil {
		return *x.ReqMedalwallFlag
	}
	return 0
}

func (x *UdcUinData) GetNotifyOnLikeRankingListFlag() int32 {
	if x != nil && x.NotifyOnLikeRankingListFlag != nil {
		return *x.NotifyOnLikeRankingListFlag
	}
	return 0
}

func (x *UdcUinData) GetApolloStatus() int32 {
	if x != nil && x.ApolloStatus != nil {
		return *x.ApolloStatus
	}
	return 0
}
