// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/oidb/oidb0x88d.proto

package oidb

import (
	proto "github.com/RomiChan/protobuf/proto"
)

type D88DGroupHeadPortraitInfo struct {
	PicId proto.Option[uint32] `protobuf:"varint,1,opt"`
	_     [0]func()
}

type D88DGroupHeadPortrait struct {
	_ [0]func()
}

type D88DGroupExInfoOnly struct {
	_ [0]func()
}

type D88DGroupInfo struct {
	GroupOwner              proto.Option[uint64] `protobuf:"varint,1,opt"`
	GroupCreateTime         proto.Option[uint32] `protobuf:"varint,2,opt"`
	GroupFlag               proto.Option[uint32] `protobuf:"varint,3,opt"`
	GroupFlagExt            proto.Option[uint32] `protobuf:"varint,4,opt"`
	GroupMemberMaxNum       proto.Option[uint32] `protobuf:"varint,5,opt"`
	GroupMemberNum          proto.Option[uint32] `protobuf:"varint,6,opt"`
	GroupOption             proto.Option[uint32] `protobuf:"varint,7,opt"`
	GroupClassExt           proto.Option[uint32] `protobuf:"varint,8,opt"`
	GroupSpecialClass       proto.Option[uint32] `protobuf:"varint,9,opt"`
	GroupLevel              proto.Option[uint32] `protobuf:"varint,10,opt"`
	GroupFace               proto.Option[uint32] `protobuf:"varint,11,opt"`
	GroupDefaultPage        proto.Option[uint32] `protobuf:"varint,12,opt"`
	GroupInfoSeq            proto.Option[uint32] `protobuf:"varint,13,opt"`
	GroupRoamingTime        proto.Option[uint32] `protobuf:"varint,14,opt"`
	GroupName               []byte               `protobuf:"bytes,15,opt"`
	GroupMemo               []byte               `protobuf:"bytes,16,opt"`
	GroupFingerMemo         []byte               `protobuf:"bytes,17,opt"`
	GroupClassText          []byte               `protobuf:"bytes,18,opt"`
	GroupAllianceCode       []uint32             `protobuf:"varint,19,rep"`
	GroupExtraAadmNum       proto.Option[uint32] `protobuf:"varint,20,opt"`
	GroupUin                proto.Option[uint64] `protobuf:"varint,21,opt"`
	GroupCurMsgSeq          proto.Option[uint32] `protobuf:"varint,22,opt"`
	GroupLastMsgTime        proto.Option[uint32] `protobuf:"varint,23,opt"`
	GroupQuestion           []byte               `protobuf:"bytes,24,opt"`
	GroupAnswer             []byte               `protobuf:"bytes,25,opt"`
	GroupVisitorMaxNum      proto.Option[uint32] `protobuf:"varint,26,opt"`
	GroupVisitorCurNum      proto.Option[uint32] `protobuf:"varint,27,opt"`
	LevelNameSeq            proto.Option[uint32] `protobuf:"varint,28,opt"`
	GroupAdminMaxNum        proto.Option[uint32] `protobuf:"varint,29,opt"`
	GroupAioSkinTimestamp   proto.Option[uint32] `protobuf:"varint,30,opt"`
	GroupBoardSkinTimestamp proto.Option[uint32] `protobuf:"varint,31,opt"`
	GroupAioSkinUrl         []byte               `protobuf:"bytes,32,opt"`
	GroupBoardSkinUrl       []byte               `protobuf:"bytes,33,opt"`
	GroupCoverSkinTimestamp proto.Option[uint32] `protobuf:"varint,34,opt"`
	GroupCoverSkinUrl       []byte               `protobuf:"bytes,35,opt"`
	GroupGrade              proto.Option[uint32] `protobuf:"varint,36,opt"`
	ActiveMemberNum         proto.Option[uint32] `protobuf:"varint,37,opt"`
	CertificationType       proto.Option[uint32] `protobuf:"varint,38,opt"`
	CertificationText       []byte               `protobuf:"bytes,39,opt"`
	GroupRichFingerMemo     []byte               `protobuf:"bytes,40,opt"`
	// repeated D88DTagRecord tagRecord = 41;
	// optional D88DGroupGeoInfo groupGeoInfo = 42;
	HeadPortraitSeq       proto.Option[uint32]   `protobuf:"varint,43,opt"`
	MsgHeadPortrait       *D88DGroupHeadPortrait `protobuf:"bytes,44,opt"`
	ShutupTimestamp       proto.Option[uint32]   `protobuf:"varint,45,opt"`
	ShutupTimestampMe     proto.Option[uint32]   `protobuf:"varint,46,opt"`
	CreateSourceFlag      proto.Option[uint32]   `protobuf:"varint,47,opt"`
	CmduinMsgSeq          proto.Option[uint32]   `protobuf:"varint,48,opt"`
	CmduinJoinTime        proto.Option[uint32]   `protobuf:"varint,49,opt"`
	CmduinUinFlag         proto.Option[uint32]   `protobuf:"varint,50,opt"`
	CmduinFlagEx          proto.Option[uint32]   `protobuf:"varint,51,opt"`
	CmduinNewMobileFlag   proto.Option[uint32]   `protobuf:"varint,52,opt"`
	CmduinReadMsgSeq      proto.Option[uint32]   `protobuf:"varint,53,opt"`
	CmduinLastMsgTime     proto.Option[uint32]   `protobuf:"varint,54,opt"`
	GroupTypeFlag         proto.Option[uint32]   `protobuf:"varint,55,opt"`
	AppPrivilegeFlag      proto.Option[uint32]   `protobuf:"varint,56,opt"`
	StGroupExInfo         *D88DGroupExInfoOnly   `protobuf:"bytes,57,opt"`
	GroupSecLevel         proto.Option[uint32]   `protobuf:"varint,58,opt"`
	GroupSecLevelInfo     proto.Option[uint32]   `protobuf:"varint,59,opt"`
	CmduinPrivilege       proto.Option[uint32]   `protobuf:"varint,60,opt"`
	PoidInfo              []byte                 `protobuf:"bytes,61,opt"`
	CmduinFlagEx2         proto.Option[uint32]   `protobuf:"varint,62,opt"`
	ConfUin               proto.Option[uint64]   `protobuf:"varint,63,opt"`
	ConfMaxMsgSeq         proto.Option[uint32]   `protobuf:"varint,64,opt"`
	ConfToGroupTime       proto.Option[uint32]   `protobuf:"varint,65,opt"`
	PasswordRedbagTime    proto.Option[uint32]   `protobuf:"varint,66,opt"`
	SubscriptionUin       proto.Option[uint64]   `protobuf:"varint,67,opt"`
	MemberListChangeSeq   proto.Option[uint32]   `protobuf:"varint,68,opt"`
	MembercardSeq         proto.Option[uint32]   `protobuf:"varint,69,opt"`
	RootId                proto.Option[uint64]   `protobuf:"varint,70,opt"`
	ParentId              proto.Option[uint64]   `protobuf:"varint,71,opt"`
	TeamSeq               proto.Option[uint32]   `protobuf:"varint,72,opt"`
	HistoryMsgBeginTime   proto.Option[uint64]   `protobuf:"varint,73,opt"`
	InviteNoAuthNumLimit  proto.Option[uint64]   `protobuf:"varint,74,opt"`
	CmduinHistoryMsgSeq   proto.Option[uint32]   `protobuf:"varint,75,opt"`
	CmduinJoinMsgSeq      proto.Option[uint32]   `protobuf:"varint,76,opt"`
	GroupFlagext3         proto.Option[uint32]   `protobuf:"varint,77,opt"`
	GroupOpenAppid        proto.Option[uint32]   `protobuf:"varint,78,opt"`
	IsConfGroup           proto.Option[uint32]   `protobuf:"varint,79,opt"`
	IsModifyConfGroupFace proto.Option[uint32]   `protobuf:"varint,80,opt"`
	IsModifyConfGroupName proto.Option[uint32]   `protobuf:"varint,81,opt"`
	NoFingerOpenFlag      proto.Option[uint32]   `protobuf:"varint,82,opt"`
	NoCodeFingerOpenFlag  proto.Option[uint32]   `protobuf:"varint,83,opt"`
}

type ReqGroupInfo struct {
	GroupCode            proto.Option[uint64] `protobuf:"varint,1,opt"`
	Stgroupinfo          *D88DGroupInfo       `protobuf:"bytes,2,opt"`
	LastGetGroupNameTime proto.Option[uint32] `protobuf:"varint,3,opt"`
	_                    [0]func()
}

type D88DReqBody struct {
	AppId           proto.Option[uint32] `protobuf:"varint,1,opt"`
	ReqGroupInfo    []*ReqGroupInfo      `protobuf:"bytes,2,rep"`
	PcClientVersion proto.Option[uint32] `protobuf:"varint,3,opt"`
}

type RspGroupInfo struct {
	GroupCode proto.Option[uint64] `protobuf:"varint,1,opt"`
	Result    proto.Option[uint32] `protobuf:"varint,2,opt"`
	GroupInfo *D88DGroupInfo       `protobuf:"bytes,3,opt"`
	_         [0]func()
}

type D88DRspBody struct {
	RspGroupInfo []*RspGroupInfo `protobuf:"bytes,1,rep"`
	StrErrorInfo []byte          `protobuf:"bytes,2,opt"`
}

type D88DTagRecord struct {
	FromUin   proto.Option[uint64] `protobuf:"varint,1,opt"`
	GroupCode proto.Option[uint64] `protobuf:"varint,2,opt"`
	TagId     []byte               `protobuf:"bytes,3,opt"`
	SetTime   proto.Option[uint64] `protobuf:"varint,4,opt"`
	GoodNum   proto.Option[uint32] `protobuf:"varint,5,opt"`
	BadNum    proto.Option[uint32] `protobuf:"varint,6,opt"`
	TagLen    proto.Option[uint32] `protobuf:"varint,7,opt"`
	TagValue  []byte               `protobuf:"bytes,8,opt"`
}

type D88DGroupGeoInfo struct {
	Owneruin   proto.Option[uint64] `protobuf:"varint,1,opt"`
	Settime    proto.Option[uint32] `protobuf:"varint,2,opt"`
	Cityid     proto.Option[uint32] `protobuf:"varint,3,opt"`
	Longitude  proto.Option[int64]  `protobuf:"varint,4,opt"`
	Latitude   proto.Option[int64]  `protobuf:"varint,5,opt"`
	Geocontent []byte               `protobuf:"bytes,6,opt"`
	PoiId      proto.Option[uint64] `protobuf:"varint,7,opt"`
}
