// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/msg/msg.proto

package msg

import (
	proto "github.com/RomiChan/protobuf/proto"
)

type SyncFlag = int32

const (
	SyncFlag_START     SyncFlag = 0
	SyncFlag_CONTINUME SyncFlag = 1
	SyncFlag_STOP      SyncFlag = 2
)

type GetMessageRequest struct {
	SyncFlag           proto.Option[SyncFlag] `protobuf:"varint,1,opt"`
	SyncCookie         []byte                 `protobuf:"bytes,2,opt"`
	RambleFlag         proto.Option[int32]    `protobuf:"varint,3,opt"`
	LatestRambleNumber proto.Option[int32]    `protobuf:"varint,4,opt"`
	OtherRambleNumber  proto.Option[int32]    `protobuf:"varint,5,opt"`
	OnlineSyncFlag     proto.Option[int32]    `protobuf:"varint,6,opt"`
	ContextFlag        proto.Option[int32]    `protobuf:"varint,7,opt"`
	WhisperSessionId   proto.Option[int32]    `protobuf:"varint,8,opt"`
	MsgReqType         proto.Option[int32]    `protobuf:"varint,9,opt"`
	PubaccountCookie   []byte                 `protobuf:"bytes,10,opt"`
	MsgCtrlBuf         []byte                 `protobuf:"bytes,11,opt"`
	ServerBuf          []byte                 `protobuf:"bytes,12,opt"`
}

type SendMessageRequest struct {
	RoutingHead *RoutingHead        `protobuf:"bytes,1,opt"`
	ContentHead *ContentHead        `protobuf:"bytes,2,opt"`
	MsgBody     *MessageBody        `protobuf:"bytes,3,opt"`
	MsgSeq      proto.Option[int32] `protobuf:"varint,4,opt"`
	MsgRand     proto.Option[int32] `protobuf:"varint,5,opt"`
	SyncCookie  []byte              `protobuf:"bytes,6,opt"`
	//MsgComm.AppShareInfo? appShare = 7;
	MsgVia      proto.Option[int32] `protobuf:"varint,8,opt"`
	DataStatist proto.Option[int32] `protobuf:"varint,9,opt"`
	//MultiMsgAssist? multiMsgAssist = 10;
	//PbInputNotifyInfo? inputNotifyInfo = 11;
	MsgCtrl *MsgCtrl `protobuf:"bytes,12,opt"`
	//ImReceipt.ReceiptReq? receiptReq = 13;
	MultiSendSeq proto.Option[int32] `protobuf:"varint,14,opt"`
}

type SendMessageResponse struct {
	Result proto.Option[int32]  `protobuf:"varint,1,opt"`
	ErrMsg proto.Option[string] `protobuf:"bytes,2,opt"`
}

type MsgWithDrawReq struct {
	C2CWithDraw   []*C2CMsgWithDrawReq   `protobuf:"bytes,1,rep"`
	GroupWithDraw []*GroupMsgWithDrawReq `protobuf:"bytes,2,rep"`
}

type C2CMsgWithDrawReq struct {
	MsgInfo         []*C2CMsgInfo       `protobuf:"bytes,1,rep"`
	LongMessageFlag proto.Option[int32] `protobuf:"varint,2,opt"`
	Reserved        []byte              `protobuf:"bytes,3,opt"`
	SubCmd          proto.Option[int32] `protobuf:"varint,4,opt"`
}

type GroupMsgWithDrawReq struct {
	SubCmd    proto.Option[int32] `protobuf:"varint,1,opt"`
	GroupType proto.Option[int32] `protobuf:"varint,2,opt"`
	GroupCode proto.Option[int64] `protobuf:"varint,3,opt"`
	MsgList   []*GroupMsgInfo     `protobuf:"bytes,4,rep"`
	UserDef   []byte              `protobuf:"bytes,5,opt"`
}

type MsgWithDrawResp struct {
	C2CWithDraw   []*C2CMsgWithDrawResp   `protobuf:"bytes,1,rep"`
	GroupWithDraw []*GroupMsgWithDrawResp `protobuf:"bytes,2,rep"`
}

type C2CMsgWithDrawResp struct {
	Result proto.Option[int32]  `protobuf:"varint,1,opt"`
	ErrMsg proto.Option[string] `protobuf:"bytes,2,opt"`
}

type GroupMsgWithDrawResp struct {
	Result proto.Option[int32]  `protobuf:"varint,1,opt"`
	ErrMsg proto.Option[string] `protobuf:"bytes,2,opt"`
}

type GroupMsgInfo struct {
	MsgSeq    proto.Option[int32] `protobuf:"varint,1,opt"`
	MsgRandom proto.Option[int32] `protobuf:"varint,2,opt"`
	MsgType   proto.Option[int32] `protobuf:"varint,3,opt"`
}

type C2CMsgInfo struct {
	FromUin     proto.Option[int64] `protobuf:"varint,1,opt"`
	ToUin       proto.Option[int64] `protobuf:"varint,2,opt"`
	MsgSeq      proto.Option[int32] `protobuf:"varint,3,opt"`
	MsgUid      proto.Option[int64] `protobuf:"varint,4,opt"`
	MsgTime     proto.Option[int64] `protobuf:"varint,5,opt"`
	MsgRandom   proto.Option[int32] `protobuf:"varint,6,opt"`
	PkgNum      proto.Option[int32] `protobuf:"varint,7,opt"`
	PkgIndex    proto.Option[int32] `protobuf:"varint,8,opt"`
	DivSeq      proto.Option[int32] `protobuf:"varint,9,opt"`
	MsgType     proto.Option[int32] `protobuf:"varint,10,opt"`
	RoutingHead *RoutingHead        `protobuf:"bytes,20,opt"`
}

type RoutingHead struct {
	C2C    *C2C    `protobuf:"bytes,1,opt"`
	Grp    *Grp    `protobuf:"bytes,2,opt"`
	GrpTmp *GrpTmp `protobuf:"bytes,3,opt"`
	WpaTmp *WPATmp `protobuf:"bytes,6,opt"`
}

type WPATmp struct {
	ToUin proto.Option[uint64] `protobuf:"varint,1,opt"`
	Sig   []byte               `protobuf:"bytes,2,opt"`
}

type C2C struct {
	ToUin proto.Option[int64] `protobuf:"varint,1,opt"`
}

type Grp struct {
	GroupCode proto.Option[int64] `protobuf:"varint,1,opt"`
}

type GrpTmp struct {
	GroupUin proto.Option[int64] `protobuf:"varint,1,opt"`
	ToUin    proto.Option[int64] `protobuf:"varint,2,opt"`
}

type MsgCtrl struct {
	MsgFlag proto.Option[int32] `protobuf:"varint,1,opt"`
}

type GetMessageResponse struct {
	Result           proto.Option[int32]    `protobuf:"varint,1,opt"`
	ErrorMessage     proto.Option[string]   `protobuf:"bytes,2,opt"`
	SyncCookie       []byte                 `protobuf:"bytes,3,opt"`
	SyncFlag         proto.Option[SyncFlag] `protobuf:"varint,4,opt"`
	UinPairMsgs      []*UinPairMessage      `protobuf:"bytes,5,rep"`
	BindUin          proto.Option[int64]    `protobuf:"varint,6,opt"`
	MsgRspType       proto.Option[int32]    `protobuf:"varint,7,opt"`
	PubAccountCookie []byte                 `protobuf:"bytes,8,opt"`
	IsPartialSync    proto.Option[bool]     `protobuf:"varint,9,opt"`
	MsgCtrlBuf       []byte                 `protobuf:"bytes,10,opt"`
}

type PushMessagePacket struct {
	Message     *Message            `protobuf:"bytes,1,opt"`
	Svrip       proto.Option[int32] `protobuf:"varint,2,opt"`
	PushToken   []byte              `protobuf:"bytes,3,opt"`
	PingFLag    proto.Option[int32] `protobuf:"varint,4,opt"`
	GeneralFlag proto.Option[int32] `protobuf:"varint,9,opt"`
}

type UinPairMessage struct {
	LastReadTime proto.Option[int32] `protobuf:"varint,1,opt"`
	PeerUin      proto.Option[int64] `protobuf:"varint,2,opt"`
	MsgCompleted proto.Option[int32] `protobuf:"varint,3,opt"`
	Messages     []*Message          `protobuf:"bytes,4,rep"`
}

type Message struct {
	Head    *MessageHead `protobuf:"bytes,1,opt"`
	Content *ContentHead `protobuf:"bytes,2,opt"`
	Body    *MessageBody `protobuf:"bytes,3,opt"`
}

type MessageBody struct {
	RichText          *RichText `protobuf:"bytes,1,opt"`
	MsgContent        []byte    `protobuf:"bytes,2,opt"`
	MsgEncryptContent []byte    `protobuf:"bytes,3,opt"`
}

type RichText struct {
	Attr          *Attr          `protobuf:"bytes,1,opt"`
	Elems         []*Elem        `protobuf:"bytes,2,rep"`
	NotOnlineFile *NotOnlineFile `protobuf:"bytes,3,opt"`
	Ptt           *Ptt           `protobuf:"bytes,4,opt"`
}

type Elem struct {
	Text           *Text           `protobuf:"bytes,1,opt"`
	Face           *Face           `protobuf:"bytes,2,opt"`
	OnlineImage    *OnlineImage    `protobuf:"bytes,3,opt"`
	NotOnlineImage *NotOnlineImage `protobuf:"bytes,4,opt"`
	TransElemInfo  *TransElem      `protobuf:"bytes,5,opt"`
	MarketFace     *MarketFace     `protobuf:"bytes,6,opt"`
	//ElemFlags elemFlags = 7;
	CustomFace *CustomFace `protobuf:"bytes,8,opt"`
	ElemFlags2 *ElemFlags2 `protobuf:"bytes,9,opt"`
	//FunFace funFace = 10;
	//SecretFileMsg secretFile = 11;
	RichMsg   *RichMsg   `protobuf:"bytes,12,opt"`
	GroupFile *GroupFile `protobuf:"bytes,13,opt"`
	//PubGroup pubGroup = 14;
	//MarketTrans marketTrans = 15;
	ExtraInfo *ExtraInfo `protobuf:"bytes,16,opt"`
	//ShakeWindow? shakeWindow = 17;
	//PubAccount? pubAccount = 18;
	VideoFile *VideoFile `protobuf:"bytes,19,opt"`
	//TipsInfo? tipsInfo = 20;
	AnonGroupMsg *AnonymousGroupMessage `protobuf:"bytes,21,opt"`
	//QQLiveOld? qqLiveOld = 22;
	//LifeOnlineAccount? lifeOnline = 23;
	QQWalletMsg *QQWalletMsg `protobuf:"bytes,24,opt"`
	//CrmElem? crmElem = 25;
	//ConferenceTipsInfo? conferenceTipsInfo = 26;
	//RedBagInfo? redbagInfo = 27;
	//LowVersionTips? lowVersionTips = 28;
	//bytes bankcodeCtrlInfo = 29;
	//NearByMessageType? nearByMsg = 30;
	CustomElem *CustomElem `protobuf:"bytes,31,opt"`
	//LocationInfo? locationInfo = 32;
	//PubAccInfo? pubAccInfo = 33;
	//SmallEmoji? smallEmoji = 34;
	//FSJMessageElem? fsjMsgElem = 35;
	//ArkAppElem? arkApp = 36;
	GeneralFlags *GeneralFlags `protobuf:"bytes,37,opt"`
	//CustomFace? hcFlashPic = 38;
	//DeliverGiftMsg? deliverGiftMsg = 39;
	//BitAppMsg? bitappMsg = 40;
	//OpenQQData? openQqData = 41;
	//ApolloActMsg? apolloMsg = 42;
	//GroupPubAccountInfo? groupPubAccInfo = 43;
	//BlessingMessage? blessMsg = 44;
	SrcMsg *SourceMsg `protobuf:"bytes,45,opt"`
	//LolaMsg? lolaMsg = 46;
	//GroupBusinessMsg? groupBusinessMsg = 47;
	//WorkflowNotifyMsg? msgWorkflowNotify = 48;
	//PatsElem? patElem = 49;
	//GroupPostElem? groupPostElem = 50;
	LightApp *LightAppElem `protobuf:"bytes,51,opt"`
	//EIMInfo? eimInfo = 52;
	CommonElem *CommonElem `protobuf:"bytes,53,opt"`
}

type MarketFace struct {
	FaceName    []byte               `protobuf:"bytes,1,opt"`
	ItemType    proto.Option[uint32] `protobuf:"varint,2,opt"`
	FaceInfo    proto.Option[uint32] `protobuf:"varint,3,opt"`
	FaceId      []byte               `protobuf:"bytes,4,opt"`
	TabId       proto.Option[uint32] `protobuf:"varint,5,opt"`
	SubType     proto.Option[uint32] `protobuf:"varint,6,opt"`
	Key         []byte               `protobuf:"bytes,7,opt"`
	Param       []byte               `protobuf:"bytes,8,opt"`
	MediaType   proto.Option[uint32] `protobuf:"varint,9,opt"`
	ImageWidth  proto.Option[uint32] `protobuf:"varint,10,opt"`
	ImageHeight proto.Option[uint32] `protobuf:"varint,11,opt"`
	Mobileparam []byte               `protobuf:"bytes,12,opt"`
	PbReserve   []byte               `protobuf:"bytes,13,opt"`
}

type ElemFlags2 struct {
	ColorTextId      proto.Option[uint32] `protobuf:"varint,1,opt"`
	MsgId            proto.Option[uint64] `protobuf:"varint,2,opt"`
	WhisperSessionId proto.Option[uint32] `protobuf:"varint,3,opt"`
	PttChangeBit     proto.Option[uint32] `protobuf:"varint,4,opt"`
	VipStatus        proto.Option[uint32] `protobuf:"varint,5,opt"`
	CompatibleId     proto.Option[uint32] `protobuf:"varint,6,opt"`
	Insts            []*ElemFlags2_Inst   `protobuf:"bytes,7,rep"`
	MsgRptCnt        proto.Option[uint32] `protobuf:"varint,8,opt"`
	SrcInst          *ElemFlags2_Inst     `protobuf:"bytes,9,opt"`
	Longtitude       proto.Option[uint32] `protobuf:"varint,10,opt"`
	Latitude         proto.Option[uint32] `protobuf:"varint,11,opt"`
	CustomFont       proto.Option[uint32] `protobuf:"varint,12,opt"`
	PcSupportDef     *PcSupportDef        `protobuf:"bytes,13,opt"`
	CrmFlags         proto.Option[uint32] `protobuf:"varint,14,opt"`
}

type PcSupportDef struct {
	PcPtlBegin     proto.Option[uint32] `protobuf:"varint,1,opt"`
	PcPtlEnd       proto.Option[uint32] `protobuf:"varint,2,opt"`
	MacPtlBegin    proto.Option[uint32] `protobuf:"varint,3,opt"`
	MacPtlEnd      proto.Option[uint32] `protobuf:"varint,4,opt"`
	PtlsSupport    []uint32             `protobuf:"varint,5,rep"`
	PtlsNotSupport []uint32             `protobuf:"varint,6,rep"`
}

type CommonElem struct {
	ServiceType  proto.Option[int32] `protobuf:"varint,1,opt"`
	PbElem       []byte              `protobuf:"bytes,2,opt"`
	BusinessType proto.Option[int32] `protobuf:"varint,3,opt"`
}

type QQWalletMsg struct {
	AioBody *QQWalletAioBody `protobuf:"bytes,1,opt"`
}

type QQWalletAioBody struct {
	SendUin     proto.Option[uint64] `protobuf:"varint,1,opt"`
	Sender      *QQWalletAioElem     `protobuf:"bytes,2,opt"`
	Receiver    *QQWalletAioElem     `protobuf:"bytes,3,opt"`
	ChannelId   proto.Option[int32]  `protobuf:"zigzag32,4,opt"`
	TemplateId  proto.Option[int32]  `protobuf:"zigzag32,5,opt"`
	Resend      proto.Option[uint32] `protobuf:"varint,6,opt"`
	MsgPriority proto.Option[uint32] `protobuf:"varint,7,opt"`
	RedType     proto.Option[int32]  `protobuf:"zigzag32,8,opt"`
	BillNo      []byte               `protobuf:"bytes,9,opt"`
	AuthKey     []byte               `protobuf:"bytes,10,opt"`
	SessionType proto.Option[int32]  `protobuf:"zigzag32,11,opt"`
	MsgType     proto.Option[int32]  `protobuf:"zigzag32,12,opt"`
	EnvelOpeId  proto.Option[int32]  `protobuf:"zigzag32,13,opt"`
	Name        []byte               `protobuf:"bytes,14,opt"`
	ConfType    proto.Option[int32]  `protobuf:"zigzag32,15,opt"`
	MsgFrom     proto.Option[int32]  `protobuf:"zigzag32,16,opt"`
	PcBody      []byte               `protobuf:"bytes,17,opt"`
	Index       []byte               `protobuf:"bytes,18,opt"`
	RedChannel  proto.Option[uint32] `protobuf:"varint,19,opt"`
	GrapUin     []uint64             `protobuf:"varint,20,rep"`
	PbReserve   []byte               `protobuf:"bytes,21,opt"`
}

type QQWalletAioElem struct {
	Background      proto.Option[uint32] `protobuf:"varint,1,opt"`
	Icon            proto.Option[uint32] `protobuf:"varint,2,opt"`
	Title           proto.Option[string] `protobuf:"bytes,3,opt"`
	Subtitle        proto.Option[string] `protobuf:"bytes,4,opt"`
	Content         proto.Option[string] `protobuf:"bytes,5,opt"`
	LinkUrl         []byte               `protobuf:"bytes,6,opt"`
	BlackStripe     []byte               `protobuf:"bytes,7,opt"`
	Notice          []byte               `protobuf:"bytes,8,opt"`
	TitleColor      proto.Option[uint32] `protobuf:"varint,9,opt"`
	SubtitleColor   proto.Option[uint32] `protobuf:"varint,10,opt"`
	ActionsPriority []byte               `protobuf:"bytes,11,opt"`
	JumpUrl         []byte               `protobuf:"bytes,12,opt"`
	NativeIos       []byte               `protobuf:"bytes,13,opt"`
	NativeAndroid   []byte               `protobuf:"bytes,14,opt"`
	IconUrl         []byte               `protobuf:"bytes,15,opt"`
	ContentColor    proto.Option[uint32] `protobuf:"varint,16,opt"`
	ContentBgColor  proto.Option[uint32] `protobuf:"varint,17,opt"`
	AioImageLeft    []byte               `protobuf:"bytes,18,opt"`
	AioImageRight   []byte               `protobuf:"bytes,19,opt"`
	CftImage        []byte               `protobuf:"bytes,20,opt"`
	PbReserve       []byte               `protobuf:"bytes,21,opt"`
}

type RichMsg struct {
	Template1 []byte              `protobuf:"bytes,1,opt"`
	ServiceId proto.Option[int32] `protobuf:"varint,2,opt"`
	MsgResId  []byte              `protobuf:"bytes,3,opt"`
	Rand      proto.Option[int32] `protobuf:"varint,4,opt"`
	Seq       proto.Option[int32] `protobuf:"varint,5,opt"`
}

type CustomElem struct {
	Desc     []byte              `protobuf:"bytes,1,opt"`
	Data     []byte              `protobuf:"bytes,2,opt"`
	EnumType proto.Option[int32] `protobuf:"varint,3,opt"`
	Ext      []byte              `protobuf:"bytes,4,opt"`
	Sound    []byte              `protobuf:"bytes,5,opt"`
}

type Text struct {
	Str       proto.Option[string] `protobuf:"bytes,1,opt"`
	Link      proto.Option[string] `protobuf:"bytes,2,opt"`
	Attr6Buf  []byte               `protobuf:"bytes,3,opt"`
	Attr7Buf  []byte               `protobuf:"bytes,4,opt"`
	Buf       []byte               `protobuf:"bytes,11,opt"`
	PbReserve []byte               `protobuf:"bytes,12,opt"`
}

type Attr struct {
	CodePage       proto.Option[int32]  `protobuf:"varint,1,opt"`
	Time           proto.Option[int32]  `protobuf:"varint,2,opt"`
	Random         proto.Option[int32]  `protobuf:"varint,3,opt"`
	Color          proto.Option[int32]  `protobuf:"varint,4,opt"`
	Size           proto.Option[int32]  `protobuf:"varint,5,opt"`
	Effect         proto.Option[int32]  `protobuf:"varint,6,opt"`
	CharSet        proto.Option[int32]  `protobuf:"varint,7,opt"`
	PitchAndFamily proto.Option[int32]  `protobuf:"varint,8,opt"`
	FontName       proto.Option[string] `protobuf:"bytes,9,opt"`
	ReserveData    []byte               `protobuf:"bytes,10,opt"`
}

type Ptt struct {
	FileType      proto.Option[int32]  `protobuf:"varint,1,opt"`
	SrcUin        proto.Option[int64]  `protobuf:"varint,2,opt"`
	FileUuid      []byte               `protobuf:"bytes,3,opt"`
	FileMd5       []byte               `protobuf:"bytes,4,opt"`
	FileName      proto.Option[string] `protobuf:"bytes,5,opt"`
	FileSize      proto.Option[int32]  `protobuf:"varint,6,opt"`
	Reserve       []byte               `protobuf:"bytes,7,opt"`
	FileId        proto.Option[int32]  `protobuf:"varint,8,opt"`
	ServerIp      proto.Option[int32]  `protobuf:"varint,9,opt"`
	ServerPort    proto.Option[int32]  `protobuf:"varint,10,opt"`
	BoolValid     proto.Option[bool]   `protobuf:"varint,11,opt"`
	Signature     []byte               `protobuf:"bytes,12,opt"`
	Shortcut      []byte               `protobuf:"bytes,13,opt"`
	FileKey       []byte               `protobuf:"bytes,14,opt"`
	MagicPttIndex proto.Option[int32]  `protobuf:"varint,15,opt"`
	VoiceSwitch   proto.Option[int32]  `protobuf:"varint,16,opt"`
	PttUrl        []byte               `protobuf:"bytes,17,opt"`
	GroupFileKey  []byte               `protobuf:"bytes,18,opt"`
	Time          proto.Option[int32]  `protobuf:"varint,19,opt"`
	DownPara      []byte               `protobuf:"bytes,20,opt"`
	Format        proto.Option[int32]  `protobuf:"varint,29,opt"`
	PbReserve     []byte               `protobuf:"bytes,30,opt"`
	BytesPttUrls  [][]byte             `protobuf:"bytes,31,rep"`
	DownloadFlag  proto.Option[int32]  `protobuf:"varint,32,opt"`
}

type OnlineImage struct {
	Guid           []byte `protobuf:"bytes,1,opt"`
	FilePath       []byte `protobuf:"bytes,2,opt"`
	OldVerSendFile []byte `protobuf:"bytes,3,opt"`
}

type NotOnlineImage struct {
	FilePath       proto.Option[string] `protobuf:"bytes,1,opt"`
	FileLen        proto.Option[int32]  `protobuf:"varint,2,opt"`
	DownloadPath   proto.Option[string] `protobuf:"bytes,3,opt"`
	OldVerSendFile []byte               `protobuf:"bytes,4,opt"`
	ImgType        proto.Option[int32]  `protobuf:"varint,5,opt"`
	PreviewsImage  []byte               `protobuf:"bytes,6,opt"`
	PicMd5         []byte               `protobuf:"bytes,7,opt"`
	PicHeight      proto.Option[int32]  `protobuf:"varint,8,opt"`
	PicWidth       proto.Option[int32]  `protobuf:"varint,9,opt"`
	ResId          proto.Option[string] `protobuf:"bytes,10,opt"`
	Flag           []byte               `protobuf:"bytes,11,opt"`
	ThumbUrl       proto.Option[string] `protobuf:"bytes,12,opt"`
	Original       proto.Option[int32]  `protobuf:"varint,13,opt"`
	BigUrl         proto.Option[string] `protobuf:"bytes,14,opt"`
	OrigUrl        proto.Option[string] `protobuf:"bytes,15,opt"`
	BizType        proto.Option[int32]  `protobuf:"varint,16,opt"`
	Result         proto.Option[int32]  `protobuf:"varint,17,opt"`
	Index          proto.Option[int32]  `protobuf:"varint,18,opt"`
	OpFaceBuf      []byte               `protobuf:"bytes,19,opt"`
	OldPicMd5      proto.Option[bool]   `protobuf:"varint,20,opt"`
	ThumbWidth     proto.Option[int32]  `protobuf:"varint,21,opt"`
	ThumbHeight    proto.Option[int32]  `protobuf:"varint,22,opt"`
	FileId         proto.Option[int32]  `protobuf:"varint,23,opt"`
	ShowLen        proto.Option[int32]  `protobuf:"varint,24,opt"`
	DownloadLen    proto.Option[int32]  `protobuf:"varint,25,opt"`
	PbReserve      []byte               `protobuf:"bytes,29,opt"`
}

type NotOnlineFile struct {
	FileType      proto.Option[int32] `protobuf:"varint,1,opt"`
	Sig           []byte              `protobuf:"bytes,2,opt"`
	FileUuid      []byte              `protobuf:"bytes,3,opt"`
	FileMd5       []byte              `protobuf:"bytes,4,opt"`
	FileName      []byte              `protobuf:"bytes,5,opt"`
	FileSize      proto.Option[int64] `protobuf:"varint,6,opt"`
	Note          []byte              `protobuf:"bytes,7,opt"`
	Reserved      proto.Option[int32] `protobuf:"varint,8,opt"`
	Subcmd        proto.Option[int32] `protobuf:"varint,9,opt"`
	MicroCloud    proto.Option[int32] `protobuf:"varint,10,opt"`
	BytesFileUrls [][]byte            `protobuf:"bytes,11,rep"`
	DownloadFlag  proto.Option[int32] `protobuf:"varint,12,opt"`
	DangerEvel    proto.Option[int32] `protobuf:"varint,50,opt"`
	LifeTime      proto.Option[int32] `protobuf:"varint,51,opt"`
	UploadTime    proto.Option[int32] `protobuf:"varint,52,opt"`
	AbsFileType   proto.Option[int32] `protobuf:"varint,53,opt"`
	ClientType    proto.Option[int32] `protobuf:"varint,54,opt"`
	ExpireTime    proto.Option[int32] `protobuf:"varint,55,opt"`
	PbReserve     []byte              `protobuf:"bytes,56,opt"`
}

type TransElem struct {
	ElemType  proto.Option[int32] `protobuf:"varint,1,opt"`
	ElemValue []byte              `protobuf:"bytes,2,opt"`
}

type ExtraInfo struct {
	Nick          []byte              `protobuf:"bytes,1,opt"`
	GroupCard     []byte              `protobuf:"bytes,2,opt"`
	Level         proto.Option[int32] `protobuf:"varint,3,opt"`
	Flags         proto.Option[int32] `protobuf:"varint,4,opt"`
	GroupMask     proto.Option[int32] `protobuf:"varint,5,opt"`
	MsgTailId     proto.Option[int32] `protobuf:"varint,6,opt"`
	SenderTitle   []byte              `protobuf:"bytes,7,opt"`
	ApnsTips      []byte              `protobuf:"bytes,8,opt"`
	Uin           proto.Option[int64] `protobuf:"varint,9,opt"`
	MsgStateFlag  proto.Option[int32] `protobuf:"varint,10,opt"`
	ApnsSoundType proto.Option[int32] `protobuf:"varint,11,opt"`
	NewGroupFlag  proto.Option[int32] `protobuf:"varint,12,opt"`
}

type GroupFile struct {
	Filename    []byte              `protobuf:"bytes,1,opt"`
	FileSize    proto.Option[int64] `protobuf:"varint,2,opt"`
	FileId      []byte              `protobuf:"bytes,3,opt"`
	BatchId     []byte              `protobuf:"bytes,4,opt"`
	FileKey     []byte              `protobuf:"bytes,5,opt"`
	Mark        []byte              `protobuf:"bytes,6,opt"`
	Sequence    proto.Option[int64] `protobuf:"varint,7,opt"`
	BatchItemId []byte              `protobuf:"bytes,8,opt"`
	FeedMsgTime proto.Option[int32] `protobuf:"varint,9,opt"`
	PbReserve   []byte              `protobuf:"bytes,10,opt"`
}

type AnonymousGroupMessage struct {
	Flags        proto.Option[int32] `protobuf:"varint,1,opt"`
	AnonId       []byte              `protobuf:"bytes,2,opt"`
	AnonNick     []byte              `protobuf:"bytes,3,opt"`
	HeadPortrait proto.Option[int32] `protobuf:"varint,4,opt"`
	ExpireTime   proto.Option[int32] `protobuf:"varint,5,opt"`
	BubbleId     proto.Option[int32] `protobuf:"varint,6,opt"`
	RankColor    []byte              `protobuf:"bytes,7,opt"`
}

type VideoFile struct {
	FileUuid               []byte              `protobuf:"bytes,1,opt"`
	FileMd5                []byte              `protobuf:"bytes,2,opt"`
	FileName               []byte              `protobuf:"bytes,3,opt"`
	FileFormat             proto.Option[int32] `protobuf:"varint,4,opt"`
	FileTime               proto.Option[int32] `protobuf:"varint,5,opt"`
	FileSize               proto.Option[int32] `protobuf:"varint,6,opt"`
	ThumbWidth             proto.Option[int32] `protobuf:"varint,7,opt"`
	ThumbHeight            proto.Option[int32] `protobuf:"varint,8,opt"`
	ThumbFileMd5           []byte              `protobuf:"bytes,9,opt"`
	Source                 []byte              `protobuf:"bytes,10,opt"`
	ThumbFileSize          proto.Option[int32] `protobuf:"varint,11,opt"`
	BusiType               proto.Option[int32] `protobuf:"varint,12,opt"`
	FromChatType           proto.Option[int32] `protobuf:"varint,13,opt"`
	ToChatType             proto.Option[int32] `protobuf:"varint,14,opt"`
	BoolSupportProgressive proto.Option[bool]  `protobuf:"varint,15,opt"`
	FileWidth              proto.Option[int32] `protobuf:"varint,16,opt"`
	FileHeight             proto.Option[int32] `protobuf:"varint,17,opt"`
	SubBusiType            proto.Option[int32] `protobuf:"varint,18,opt"`
	VideoAttr              proto.Option[int32] `protobuf:"varint,19,opt"`
	BytesThumbFileUrls     [][]byte            `protobuf:"bytes,20,rep"`
	BytesVideoFileUrls     [][]byte            `protobuf:"bytes,21,rep"`
	ThumbDownloadFlag      proto.Option[int32] `protobuf:"varint,22,opt"`
	VideoDownloadFlag      proto.Option[int32] `protobuf:"varint,23,opt"`
	PbReserve              []byte              `protobuf:"bytes,24,opt"`
}

type SourceMsg struct {
	OrigSeqs  []int32             `protobuf:"varint,1,rep"`
	SenderUin proto.Option[int64] `protobuf:"varint,2,opt"`
	Time      proto.Option[int32] `protobuf:"varint,3,opt"`
	Flag      proto.Option[int32] `protobuf:"varint,4,opt"`
	Elems     []*Elem             `protobuf:"bytes,5,rep"`
	Type      proto.Option[int32] `protobuf:"varint,6,opt"`
	RichMsg   []byte              `protobuf:"bytes,7,opt"`
	PbReserve []byte              `protobuf:"bytes,8,opt"`
	SrcMsg    []byte              `protobuf:"bytes,9,opt"`
	ToUin     proto.Option[int64] `protobuf:"varint,10,opt"`
	TroopName []byte              `protobuf:"bytes,11,opt"`
}

type Face struct {
	Index proto.Option[int32] `protobuf:"varint,1,opt"`
	Old   []byte              `protobuf:"bytes,2,opt"`
	Buf   []byte              `protobuf:"bytes,11,opt"`
}

type LightAppElem struct {
	Data     []byte `protobuf:"bytes,1,opt"`
	MsgResid []byte `protobuf:"bytes,2,opt"`
}

type CustomFace struct {
	Guid        []byte               `protobuf:"bytes,1,opt"`
	FilePath    proto.Option[string] `protobuf:"bytes,2,opt"`
	Shortcut    proto.Option[string] `protobuf:"bytes,3,opt"`
	Buffer      []byte               `protobuf:"bytes,4,opt"`
	Flag        []byte               `protobuf:"bytes,5,opt"`
	OldData     []byte               `protobuf:"bytes,6,opt"`
	FileId      proto.Option[int32]  `protobuf:"varint,7,opt"`
	ServerIp    proto.Option[int32]  `protobuf:"varint,8,opt"`
	ServerPort  proto.Option[int32]  `protobuf:"varint,9,opt"`
	FileType    proto.Option[int32]  `protobuf:"varint,10,opt"`
	Signature   []byte               `protobuf:"bytes,11,opt"`
	Useful      proto.Option[int32]  `protobuf:"varint,12,opt"`
	Md5         []byte               `protobuf:"bytes,13,opt"`
	ThumbUrl    proto.Option[string] `protobuf:"bytes,14,opt"`
	BigUrl      proto.Option[string] `protobuf:"bytes,15,opt"`
	OrigUrl     proto.Option[string] `protobuf:"bytes,16,opt"`
	BizType     proto.Option[int32]  `protobuf:"varint,17,opt"`
	RepeatIndex proto.Option[int32]  `protobuf:"varint,18,opt"`
	RepeatImage proto.Option[int32]  `protobuf:"varint,19,opt"`
	ImageType   proto.Option[int32]  `protobuf:"varint,20,opt"`
	Index       proto.Option[int32]  `protobuf:"varint,21,opt"`
	Width       proto.Option[int32]  `protobuf:"varint,22,opt"`
	Height      proto.Option[int32]  `protobuf:"varint,23,opt"`
	Source      proto.Option[int32]  `protobuf:"varint,24,opt"`
	Size        proto.Option[int32]  `protobuf:"varint,25,opt"`
	Origin      proto.Option[int32]  `protobuf:"varint,26,opt"`
	ThumbWidth  proto.Option[int32]  `protobuf:"varint,27,opt"`
	ThumbHeight proto.Option[int32]  `protobuf:"varint,28,opt"`
	ShowLen     proto.Option[int32]  `protobuf:"varint,29,opt"`
	DownloadLen proto.Option[int32]  `protobuf:"varint,30,opt"`
	X400Url     proto.Option[string] `protobuf:"bytes,31,opt"`
	X400Width   proto.Option[int32]  `protobuf:"varint,32,opt"`
	X400Height  proto.Option[int32]  `protobuf:"varint,33,opt"`
	PbReserve   []byte               `protobuf:"bytes,34,opt"`
}

type ContentHead struct {
	PkgNum    proto.Option[int32] `protobuf:"varint,1,opt"`
	PkgIndex  proto.Option[int32] `protobuf:"varint,2,opt"`
	DivSeq    proto.Option[int32] `protobuf:"varint,3,opt"`
	AutoReply proto.Option[int32] `protobuf:"varint,4,opt"`
}

type MessageHead struct {
	FromUin                    proto.Option[int64]  `protobuf:"varint,1,opt"`
	ToUin                      proto.Option[int64]  `protobuf:"varint,2,opt"`
	MsgType                    proto.Option[int32]  `protobuf:"varint,3,opt"`
	C2CCmd                     proto.Option[int32]  `protobuf:"varint,4,opt"`
	MsgSeq                     proto.Option[int32]  `protobuf:"varint,5,opt"`
	MsgTime                    proto.Option[int32]  `protobuf:"varint,6,opt"`
	MsgUid                     proto.Option[int64]  `protobuf:"varint,7,opt"`
	C2CTmpMsgHead              *C2CTempMessageHead  `protobuf:"bytes,8,opt"`
	GroupInfo                  *GroupInfo           `protobuf:"bytes,9,opt"`
	FromAppid                  proto.Option[int32]  `protobuf:"varint,10,opt"`
	FromInstid                 proto.Option[int32]  `protobuf:"varint,11,opt"`
	UserActive                 proto.Option[int32]  `protobuf:"varint,12,opt"`
	DiscussInfo                *DiscussInfo         `protobuf:"bytes,13,opt"`
	FromNick                   proto.Option[string] `protobuf:"bytes,14,opt"`
	AuthUin                    proto.Option[int64]  `protobuf:"varint,15,opt"`
	AuthNick                   proto.Option[string] `protobuf:"bytes,16,opt"`
	MsgFlag                    proto.Option[int32]  `protobuf:"varint,17,opt"`
	AuthRemark                 proto.Option[string] `protobuf:"bytes,18,opt"`
	GroupName                  proto.Option[string] `protobuf:"bytes,19,opt"`
	MutiltransHead             *MutilTransHead      `protobuf:"bytes,20,opt"`
	MsgInstCtrl                *InstCtrl            `protobuf:"bytes,21,opt"`
	PublicAccountGroupSendFlag proto.Option[int32]  `protobuf:"varint,22,opt"`
	WseqInC2CMsghead           proto.Option[int32]  `protobuf:"varint,23,opt"`
	Cpid                       proto.Option[int64]  `protobuf:"varint,24,opt"`
	ExtGroupKeyInfo            *ExtGroupKeyInfo     `protobuf:"bytes,25,opt"`
	MultiCompatibleText        proto.Option[string] `protobuf:"bytes,26,opt"`
	AuthSex                    proto.Option[int32]  `protobuf:"varint,27,opt"`
	IsSrcMsg                   proto.Option[bool]   `protobuf:"varint,28,opt"`
}

type GroupInfo struct {
	GroupCode     proto.Option[int64]  `protobuf:"varint,1,opt"`
	GroupType     proto.Option[int32]  `protobuf:"varint,2,opt"`
	GroupInfoSeq  proto.Option[int64]  `protobuf:"varint,3,opt"`
	GroupCard     proto.Option[string] `protobuf:"bytes,4,opt"`
	GroupRank     []byte               `protobuf:"bytes,5,opt"`
	GroupLevel    proto.Option[int32]  `protobuf:"varint,6,opt"`
	GroupCardType proto.Option[int32]  `protobuf:"varint,7,opt"`
	GroupName     []byte               `protobuf:"bytes,8,opt"`
}

type DiscussInfo struct {
	DiscussUin     proto.Option[int64] `protobuf:"varint,1,opt"`
	DiscussType    proto.Option[int32] `protobuf:"varint,2,opt"`
	DiscussInfoSeq proto.Option[int64] `protobuf:"varint,3,opt"`
	DiscussRemark  []byte              `protobuf:"bytes,4,opt"`
	DiscussName    []byte              `protobuf:"bytes,5,opt"`
}

type MutilTransHead struct {
	Status proto.Option[int32] `protobuf:"varint,1,opt"`
	MsgId  proto.Option[int32] `protobuf:"varint,2,opt"`
}

type C2CTempMessageHead struct {
	C2CType       proto.Option[int32]  `protobuf:"varint,1,opt"`
	ServiceType   proto.Option[int32]  `protobuf:"varint,2,opt"`
	GroupUin      proto.Option[int64]  `protobuf:"varint,3,opt"`
	GroupCode     proto.Option[int64]  `protobuf:"varint,4,opt"`
	Sig           []byte               `protobuf:"bytes,5,opt"`
	SigType       proto.Option[int32]  `protobuf:"varint,6,opt"`
	FromPhone     proto.Option[string] `protobuf:"bytes,7,opt"`
	ToPhone       proto.Option[string] `protobuf:"bytes,8,opt"`
	LockDisplay   proto.Option[int32]  `protobuf:"varint,9,opt"`
	DirectionFlag proto.Option[int32]  `protobuf:"varint,10,opt"`
	Reserved      []byte               `protobuf:"bytes,11,opt"`
}

type InstCtrl struct {
	MsgSendToInst  []*InstInfo `protobuf:"bytes,1,rep"`
	MsgExcludeInst []*InstInfo `protobuf:"bytes,2,rep"`
	MsgFromInst    *InstInfo   `protobuf:"bytes,3,opt"`
}

type InstInfo struct {
	Apppid         proto.Option[int32] `protobuf:"varint,1,opt"`
	Instid         proto.Option[int32] `protobuf:"varint,2,opt"`
	Platform       proto.Option[int32] `protobuf:"varint,3,opt"`
	EnumDeviceType proto.Option[int32] `protobuf:"varint,10,opt"`
}

type ExtGroupKeyInfo struct {
	CurMaxSeq proto.Option[int32] `protobuf:"varint,1,opt"`
	CurTime   proto.Option[int64] `protobuf:"varint,2,opt"`
}

type SyncCookie struct {
	Time1        proto.Option[int64] `protobuf:"varint,1,opt"`
	Time         proto.Option[int64] `protobuf:"varint,2,opt"`
	Ran1         proto.Option[int64] `protobuf:"varint,3,opt"`
	Ran2         proto.Option[int64] `protobuf:"varint,4,opt"`
	Const1       proto.Option[int64] `protobuf:"varint,5,opt"`
	Const2       proto.Option[int64] `protobuf:"varint,11,opt"`
	Const3       proto.Option[int64] `protobuf:"varint,12,opt"`
	LastSyncTime proto.Option[int64] `protobuf:"varint,13,opt"`
	Const4       proto.Option[int64] `protobuf:"varint,14,opt"`
}

type TransMsgInfo struct {
	FromUin         proto.Option[int64]  `protobuf:"varint,1,opt"`
	ToUin           proto.Option[int64]  `protobuf:"varint,2,opt"`
	MsgType         proto.Option[int32]  `protobuf:"varint,3,opt"`
	MsgSubtype      proto.Option[int32]  `protobuf:"varint,4,opt"`
	MsgSeq          proto.Option[int32]  `protobuf:"varint,5,opt"`
	MsgUid          proto.Option[int64]  `protobuf:"varint,6,opt"`
	MsgTime         proto.Option[int32]  `protobuf:"varint,7,opt"`
	RealMsgTime     proto.Option[int32]  `protobuf:"varint,8,opt"`
	NickName        proto.Option[string] `protobuf:"bytes,9,opt"`
	MsgData         []byte               `protobuf:"bytes,10,opt"`
	SvrIp           proto.Option[int32]  `protobuf:"varint,11,opt"`
	ExtGroupKeyInfo *ExtGroupKeyInfo     `protobuf:"bytes,12,opt"`
	GeneralFlag     proto.Option[int32]  `protobuf:"varint,17,opt"`
}

type GeneralFlags struct {
	BubbleDiyTextId     proto.Option[int32]  `protobuf:"varint,1,opt"`
	GroupFlagNew        proto.Option[int32]  `protobuf:"varint,2,opt"`
	Uin                 proto.Option[int64]  `protobuf:"varint,3,opt"`
	RpId                []byte               `protobuf:"bytes,4,opt"`
	PrpFold             proto.Option[int32]  `protobuf:"varint,5,opt"`
	LongTextFlag        proto.Option[int32]  `protobuf:"varint,6,opt"`
	LongTextResid       proto.Option[string] `protobuf:"bytes,7,opt"`
	GroupType           proto.Option[int32]  `protobuf:"varint,8,opt"`
	ToUinFlag           proto.Option[int32]  `protobuf:"varint,9,opt"`
	GlamourLevel        proto.Option[int32]  `protobuf:"varint,10,opt"`
	MemberLevel         proto.Option[int32]  `protobuf:"varint,11,opt"`
	GroupRankSeq        proto.Option[int64]  `protobuf:"varint,12,opt"`
	OlympicTorch        proto.Option[int32]  `protobuf:"varint,13,opt"`
	BabyqGuideMsgCookie []byte               `protobuf:"bytes,14,opt"`
	Uin32ExpertFlag     proto.Option[int32]  `protobuf:"varint,15,opt"`
	BubbleSubId         proto.Option[int32]  `protobuf:"varint,16,opt"`
	PendantId           proto.Option[int64]  `protobuf:"varint,17,opt"`
	RpIndex             []byte               `protobuf:"bytes,18,opt"`
	PbReserve           []byte               `protobuf:"bytes,19,opt"`
}

type PbMultiMsgItem struct {
	FileName proto.Option[string] `protobuf:"bytes,1,opt"`
	Buffer   *PbMultiMsgNew       `protobuf:"bytes,2,opt"`
}

type PbMultiMsgNew struct {
	Msg []*Message `protobuf:"bytes,1,rep"`
}

type PbMultiMsgTransmit struct {
	Msg        []*Message        `protobuf:"bytes,1,rep"`
	PbItemList []*PbMultiMsgItem `protobuf:"bytes,2,rep"`
}

type MsgElemInfoServtype3 struct {
	FlashTroopPic *CustomFace     `protobuf:"bytes,1,opt"`
	FlashC2CPic   *NotOnlineImage `protobuf:"bytes,2,opt"`
}

type MsgElemInfoServtype33 struct {
	Index  proto.Option[uint32] `protobuf:"varint,1,opt"`
	Text   []byte               `protobuf:"bytes,2,opt"`
	Compat []byte               `protobuf:"bytes,3,opt"`
	Buf    []byte               `protobuf:"bytes,4,opt"`
}

type MsgElemInfoServtype38 struct {
	ReactData []byte `protobuf:"bytes,1,opt"`
	ReplyData []byte `protobuf:"bytes,2,opt"`
}

type SubMsgType0X4Body struct {
	NotOnlineFile              *NotOnlineFile       `protobuf:"bytes,1,opt"`
	MsgTime                    proto.Option[uint32] `protobuf:"varint,2,opt"`
	OnlineFileForPolyToOffline proto.Option[uint32] `protobuf:"varint,3,opt"` // fileImageInfo
}

type ResvAttr struct {
	ImageBizType proto.Option[uint32] `protobuf:"varint,1,opt"`
	ImageShow    *AnimationImageShow  `protobuf:"bytes,7,opt"`
}

type AnimationImageShow struct {
	EffectId       proto.Option[int32] `protobuf:"varint,1,opt"`
	AnimationParam []byte              `protobuf:"bytes,2,opt"`
}

type UinTypeUserDef struct {
	FromUinType   proto.Option[int32]  `protobuf:"varint,1,opt"`
	FromGroupCode proto.Option[int64]  `protobuf:"varint,2,opt"`
	FileUuid      proto.Option[string] `protobuf:"bytes,3,opt"`
}

type GetGroupMsgReq struct {
	GroupCode       proto.Option[uint64] `protobuf:"varint,1,opt"`
	BeginSeq        proto.Option[uint64] `protobuf:"varint,2,opt"`
	EndSeq          proto.Option[uint64] `protobuf:"varint,3,opt"`
	Filter          proto.Option[uint32] `protobuf:"varint,4,opt"`
	MemberSeq       proto.Option[uint64] `protobuf:"varint,5,opt"`
	PublicGroup     proto.Option[bool]   `protobuf:"varint,6,opt"`
	ShieldFlag      proto.Option[uint32] `protobuf:"varint,7,opt"`
	SaveTrafficFlag proto.Option[uint32] `protobuf:"varint,8,opt"`
}

type GetGroupMsgResp struct {
	Result         proto.Option[uint32] `protobuf:"varint,1,opt"`
	Errmsg         proto.Option[string] `protobuf:"bytes,2,opt"`
	GroupCode      proto.Option[uint64] `protobuf:"varint,3,opt"`
	ReturnBeginSeq proto.Option[uint64] `protobuf:"varint,4,opt"`
	ReturnEndSeq   proto.Option[uint64] `protobuf:"varint,5,opt"`
	Msg            []*Message           `protobuf:"bytes,6,rep"`
}

type PbGetOneDayRoamMsgReq struct {
	PeerUin     proto.Option[uint64] `protobuf:"varint,1,opt"`
	LastMsgTime proto.Option[uint64] `protobuf:"varint,2,opt"`
	Random      proto.Option[uint64] `protobuf:"varint,3,opt"`
	ReadCnt     proto.Option[uint32] `protobuf:"varint,4,opt"`
}

type PbGetOneDayRoamMsgResp struct {
	Result      proto.Option[uint32] `protobuf:"varint,1,opt"`
	ErrMsg      proto.Option[string] `protobuf:"bytes,2,opt"`
	PeerUin     proto.Option[uint64] `protobuf:"varint,3,opt"`
	LastMsgTime proto.Option[uint64] `protobuf:"varint,4,opt"`
	Random      proto.Option[uint64] `protobuf:"varint,5,opt"`
	Msg         []*Message           `protobuf:"bytes,6,rep"`
	IsComplete  proto.Option[uint32] `protobuf:"varint,7,opt"`
}

type PbPushMsg struct {
	Msg         *Message             `protobuf:"bytes,1,opt"`
	Svrip       proto.Option[int32]  `protobuf:"varint,2,opt"`
	PushToken   []byte               `protobuf:"bytes,3,opt"`
	PingFlag    proto.Option[uint32] `protobuf:"varint,4,opt"`
	GeneralFlag proto.Option[uint32] `protobuf:"varint,9,opt"`
	BindUin     proto.Option[uint64] `protobuf:"varint,10,opt"`
}

type MsgElemInfoServtype37 struct {
	Packid      []byte               `protobuf:"bytes,1,opt"`
	Stickerid   []byte               `protobuf:"bytes,2,opt"`
	Qsid        proto.Option[uint32] `protobuf:"varint,3,opt"`
	Sourcetype  proto.Option[uint32] `protobuf:"varint,4,opt"`
	Stickertype proto.Option[uint32] `protobuf:"varint,5,opt"`
	Resultid    []byte               `protobuf:"bytes,6,opt"`
	Text        []byte               `protobuf:"bytes,7,opt"`
	Surpriseid  []byte               `protobuf:"bytes,8,opt"`
	Randomtype  proto.Option[uint32] `protobuf:"varint,9,opt"`
}

type ElemFlags2_Inst struct {
	AppId  proto.Option[uint32] `protobuf:"varint,1,opt"`
	InstId proto.Option[uint32] `protobuf:"varint,2,opt"`
}
