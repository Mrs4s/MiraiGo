// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/cmd0x346/cmd0x346.proto

package cmd0x346

type ApplyCleanTrafficRsp struct {
	RetCode int32  `protobuf:"varint,10,opt"`
	RetMsg  string `protobuf:"bytes,20,opt"`
}

type ApplyCopyFromReq struct {
	SrcUin          int64  `protobuf:"varint,10,opt"`
	SrcGroup        int64  `protobuf:"varint,20,opt"`
	SrcSvcid        int32  `protobuf:"varint,30,opt"`
	SrcParentfolder []byte `protobuf:"bytes,40,opt"`
	SrcUuid         []byte `protobuf:"bytes,50,opt"`
	FileMd5         []byte `protobuf:"bytes,60,opt"`
	DstUin          int64  `protobuf:"varint,70,opt"`
	FileSize        int64  `protobuf:"varint,80,opt"`
	FileName        string `protobuf:"bytes,90,opt"`
	DangerLevel     int32  `protobuf:"varint,100,opt"`
	TotalSpace      int64  `protobuf:"varint,110,opt"`
}

type ApplyCopyFromRsp struct {
	RetCode    int32  `protobuf:"varint,10,opt"`
	RetMsg     string `protobuf:"bytes,20,opt"`
	Uuid       []byte `protobuf:"bytes,30,opt"`
	TotalSpace int64  `protobuf:"varint,40,opt"`
}

type ApplyCopyToReq struct {
	DstId         int64  `protobuf:"varint,10,opt"`
	DstUin        int64  `protobuf:"varint,20,opt"`
	DstSvcid      int32  `protobuf:"varint,30,opt"`
	SrcUin        int64  `protobuf:"varint,40,opt"`
	FileSize      int64  `protobuf:"varint,50,opt"`
	FileName      string `protobuf:"bytes,60,opt"`
	LocalFilepath string `protobuf:"bytes,70,opt"`
	Uuid          []byte `protobuf:"bytes,80,opt"`
}

type ApplyCopyToRsp struct {
	RetCode int32  `protobuf:"varint,10,opt"`
	RetMsg  string `protobuf:"bytes,20,opt"`
	FileKey string `protobuf:"bytes,30,opt"`
}

type ApplyDownloadAbsReq struct {
	Uin  int64  `protobuf:"varint,10,opt"`
	Uuid []byte `protobuf:"bytes,20,opt"`
}

type ApplyDownloadAbsRsp struct {
	RetCode      int32         `protobuf:"varint,10,opt"`
	RetMsg       string        `protobuf:"bytes,20,opt"`
	DownloadInfo *DownloadInfo `protobuf:"bytes,30,opt"`
}

type ApplyDownloadReq struct {
	Uin       int64  `protobuf:"varint,10,opt"`
	Uuid      []byte `protobuf:"bytes,20,opt"`
	OwnerType int32  `protobuf:"varint,30,opt"`
	ExtIntype int32  `protobuf:"varint,500,opt"`
}

type ApplyDownloadRsp struct {
	RetCode      int32         `protobuf:"varint,10,opt"`
	RetMsg       string        `protobuf:"bytes,20,opt"`
	DownloadInfo *DownloadInfo `protobuf:"bytes,30,opt"`
	FileInfo     *FileInfo     `protobuf:"bytes,40,opt"`
}

type ApplyForwardFileReq struct {
	SenderUin   int64  `protobuf:"varint,10,opt"`
	RecverUin   int64  `protobuf:"varint,20,opt"`
	Uuid        []byte `protobuf:"bytes,30,opt"`
	DangerLevel int32  `protobuf:"varint,40,opt"`
	TotalSpace  int64  `protobuf:"varint,50,opt"`
}

type ApplyForwardFileRsp struct {
	RetCode    int32  `protobuf:"varint,10,opt"`
	RetMsg     string `protobuf:"bytes,20,opt"`
	TotalSpace int64  `protobuf:"varint,30,opt"`
	UsedSpace  int64  `protobuf:"varint,40,opt"`
	Uuid       []byte `protobuf:"bytes,50,opt"`
}

type ApplyGetTrafficReq struct {
}

type ApplyGetTrafficRsp struct {
	RetCode     int32  `protobuf:"varint,10,opt"`
	RetMsg      string `protobuf:"bytes,20,opt"`
	UseFileSize int64  `protobuf:"varint,30,opt"`
	UseFileNum  int32  `protobuf:"varint,40,opt"`
	AllFileSize int64  `protobuf:"varint,50,opt"`
	AllFileNum  int32  `protobuf:"varint,60,opt"`
}

type ApplyListDownloadReq struct {
	Uin        int64 `protobuf:"varint,10,opt"`
	BeginIndex int32 `protobuf:"varint,20,opt"`
	ReqCount   int32 `protobuf:"varint,30,opt"`
}

type ApplyListDownloadRsp struct {
	RetCode    int32       `protobuf:"varint,10,opt"`
	RetMsg     string      `protobuf:"bytes,20,opt"`
	TotalCount int32       `protobuf:"varint,30,opt"`
	BeginIndex int32       `protobuf:"varint,40,opt"`
	RspCount   int32       `protobuf:"varint,50,opt"`
	IsEnd      int32       `protobuf:"varint,60,opt"`
	FileList   []*FileInfo `protobuf:"bytes,70,rep"`
}

type ApplyUploadHitReq struct {
	SenderUin     int64  `protobuf:"varint,10,opt"`
	RecverUin     int64  `protobuf:"varint,20,opt"`
	FileSize      int64  `protobuf:"varint,30,opt"`
	FileName      string `protobuf:"bytes,40,opt"`
	Bytes_10MMd5  []byte `protobuf:"bytes,50,opt"`
	LocalFilepath string `protobuf:"bytes,60,opt"`
	DangerLevel   int32  `protobuf:"varint,70,opt"`
	TotalSpace    int64  `protobuf:"varint,80,opt"`
}

type ApplyUploadHitReqV2 struct {
	SenderUin     int64  `protobuf:"varint,10,opt"`
	RecverUin     int64  `protobuf:"varint,20,opt"`
	FileSize      int64  `protobuf:"varint,30,opt"`
	FileName      string `protobuf:"bytes,40,opt"`
	Bytes_10MMd5  []byte `protobuf:"bytes,50,opt"`
	Bytes_3Sha    []byte `protobuf:"bytes,60,opt"`
	Sha           []byte `protobuf:"bytes,70,opt"`
	LocalFilepath string `protobuf:"bytes,80,opt"`
	DangerLevel   int32  `protobuf:"varint,90,opt"`
	TotalSpace    int64  `protobuf:"varint,100,opt"`
}

type ApplyUploadHitReqV3 struct {
	SenderUin     int64  `protobuf:"varint,10,opt"`
	RecverUin     int64  `protobuf:"varint,20,opt"`
	FileSize      int64  `protobuf:"varint,30,opt"`
	FileName      string `protobuf:"bytes,40,opt"`
	Bytes_10MMd5  []byte `protobuf:"bytes,50,opt"`
	Sha           []byte `protobuf:"bytes,60,opt"`
	LocalFilepath string `protobuf:"bytes,70,opt"`
	DangerLevel   int32  `protobuf:"varint,80,opt"`
	TotalSpace    int64  `protobuf:"varint,90,opt"`
}

type ApplyUploadHitRsp struct {
	RetCode      int32  `protobuf:"varint,10,opt"`
	RetMsg       string `protobuf:"bytes,20,opt"`
	UploadIp     string `protobuf:"bytes,30,opt"`
	UploadPort   int32  `protobuf:"varint,40,opt"`
	UploadDomain string `protobuf:"bytes,50,opt"`
	Uuid         []byte `protobuf:"bytes,60,opt"`
	UploadKey    []byte `protobuf:"bytes,70,opt"`
	TotalSpace   int64  `protobuf:"varint,80,opt"`
	UsedSpace    int64  `protobuf:"varint,90,opt"`
}

type ApplyUploadHitRspV2 struct {
	RetCode      int32  `protobuf:"varint,10,opt"`
	RetMsg       string `protobuf:"bytes,20,opt"`
	UploadIp     string `protobuf:"bytes,30,opt"`
	UploadPort   int32  `protobuf:"varint,40,opt"`
	UploadDomain string `protobuf:"bytes,50,opt"`
	Uuid         []byte `protobuf:"bytes,60,opt"`
	UploadKey    []byte `protobuf:"bytes,70,opt"`
	TotalSpace   int64  `protobuf:"varint,80,opt"`
	UsedSpace    int64  `protobuf:"varint,90,opt"`
}

type ApplyUploadHitRspV3 struct {
	RetCode      int32  `protobuf:"varint,10,opt"`
	RetMsg       string `protobuf:"bytes,20,opt"`
	UploadIp     string `protobuf:"bytes,30,opt"`
	UploadPort   int32  `protobuf:"varint,40,opt"`
	UploadDomain string `protobuf:"bytes,50,opt"`
	Uuid         []byte `protobuf:"bytes,60,opt"`
	UploadKey    []byte `protobuf:"bytes,70,opt"`
	TotalSpace   int64  `protobuf:"varint,80,opt"`
	UsedSpace    int64  `protobuf:"varint,90,opt"`
}

type ApplyUploadReq struct {
	SenderUin     int64  `protobuf:"varint,10,opt"`
	RecverUin     int64  `protobuf:"varint,20,opt"`
	FileType      int32  `protobuf:"varint,30,opt"`
	FileSize      int64  `protobuf:"varint,40,opt"`
	FileName      string `protobuf:"bytes,50,opt"`
	Bytes_10MMd5  []byte `protobuf:"bytes,60,opt"`
	LocalFilepath string `protobuf:"bytes,70,opt"`
	DangerLevel   int32  `protobuf:"varint,80,opt"`
	TotalSpace    int64  `protobuf:"varint,90,opt"`
}

type ApplyUploadReqV2 struct {
	SenderUin     int64  `protobuf:"varint,10,opt"`
	RecverUin     int64  `protobuf:"varint,20,opt"`
	FileSize      int64  `protobuf:"varint,30,opt"`
	FileName      string `protobuf:"bytes,40,opt"`
	Bytes_10MMd5  []byte `protobuf:"bytes,50,opt"`
	Bytes_3Sha    []byte `protobuf:"bytes,60,opt"`
	LocalFilepath string `protobuf:"bytes,70,opt"`
	DangerLevel   int32  `protobuf:"varint,80,opt"`
	TotalSpace    int64  `protobuf:"varint,90,opt"`
}

type ApplyUploadReqV3 struct {
	SenderUin     int64  `protobuf:"varint,10,opt"`
	RecverUin     int64  `protobuf:"varint,20,opt"`
	FileSize      int64  `protobuf:"varint,30,opt"`
	FileName      string `protobuf:"bytes,40,opt"`
	Bytes_10MMd5  []byte `protobuf:"bytes,50,opt"`
	Sha           []byte `protobuf:"bytes,60,opt"`
	LocalFilepath string `protobuf:"bytes,70,opt"`
	DangerLevel   int32  `protobuf:"varint,80,opt"`
	TotalSpace    int64  `protobuf:"varint,90,opt"`
}

type ApplyUploadRsp struct {
	RetCode       int32    `protobuf:"varint,10,opt"`
	RetMsg        string   `protobuf:"bytes,20,opt"`
	TotalSpace    int64    `protobuf:"varint,30,opt"`
	UsedSpace     int64    `protobuf:"varint,40,opt"`
	UploadedSize  int64    `protobuf:"varint,50,opt"`
	UploadIp      string   `protobuf:"bytes,60,opt"`
	UploadDomain  string   `protobuf:"bytes,70,opt"`
	UploadPort    int32    `protobuf:"varint,80,opt"`
	Uuid          []byte   `protobuf:"bytes,90,opt"`
	UploadKey     []byte   `protobuf:"bytes,100,opt"`
	BoolFileExist bool     `protobuf:"varint,110,opt"`
	PackSize      int32    `protobuf:"varint,120,opt"`
	UploadipList  []string `protobuf:"bytes,130,rep"`
}

type ApplyUploadRspV2 struct {
	RetCode       int32    `protobuf:"varint,10,opt"`
	RetMsg        string   `protobuf:"bytes,20,opt"`
	TotalSpace    int64    `protobuf:"varint,30,opt"`
	UsedSpace     int64    `protobuf:"varint,40,opt"`
	UploadedSize  int64    `protobuf:"varint,50,opt"`
	UploadIp      string   `protobuf:"bytes,60,opt"`
	UploadDomain  string   `protobuf:"bytes,70,opt"`
	UploadPort    int32    `protobuf:"varint,80,opt"`
	Uuid          []byte   `protobuf:"bytes,90,opt"`
	UploadKey     []byte   `protobuf:"bytes,100,opt"`
	BoolFileExist bool     `protobuf:"varint,110,opt"`
	PackSize      int32    `protobuf:"varint,120,opt"`
	UploadipList  []string `protobuf:"bytes,130,rep"`
	HttpsvrApiVer int32    `protobuf:"varint,140,opt"`
	Sha           []byte   `protobuf:"bytes,141,opt"`
}

type ApplyUploadRspV3 struct {
	RetCode           int32    `protobuf:"varint,10,opt"`
	RetMsg            string   `protobuf:"bytes,20,opt"`
	TotalSpace        int64    `protobuf:"varint,30,opt"`
	UsedSpace         int64    `protobuf:"varint,40,opt"`
	UploadedSize      int64    `protobuf:"varint,50,opt"`
	UploadIp          string   `protobuf:"bytes,60,opt"`
	UploadDomain      string   `protobuf:"bytes,70,opt"`
	UploadPort        int32    `protobuf:"varint,80,opt"`
	Uuid              []byte   `protobuf:"bytes,90,opt"`
	UploadKey         []byte   `protobuf:"bytes,100,opt"`
	BoolFileExist     bool     `protobuf:"varint,110,opt"`
	PackSize          int32    `protobuf:"varint,120,opt"`
	UploadIpList      []string `protobuf:"bytes,130,rep"`
	UploadHttpsPort   int32    `protobuf:"varint,140,opt"`
	UploadHttpsDomain string   `protobuf:"bytes,150,opt"`
	UploadDns         string   `protobuf:"bytes,160,opt"`
	UploadLanip       string   `protobuf:"bytes,170,opt"`
}

type DelMessageReq struct {
	UinSender   int64 `protobuf:"varint,1,opt"`
	UinReceiver int64 `protobuf:"varint,2,opt"`
	Time        int32 `protobuf:"varint,10,opt"`
	Random      int32 `protobuf:"varint,20,opt"`
	SeqNo       int32 `protobuf:"varint,30,opt"`
}

type DeleteFileReq struct {
	Uin        int64  `protobuf:"varint,10,opt"`
	PeerUin    int64  `protobuf:"varint,20,opt"`
	DeleteType int32  `protobuf:"varint,30,opt"`
	Uuid       []byte `protobuf:"bytes,40,opt"`
}

type DeleteFileRsp struct {
	RetCode int32  `protobuf:"varint,10,opt"`
	RetMsg  string `protobuf:"bytes,20,opt"`
}

type DownloadInfo struct {
	DownloadKey    []byte   `protobuf:"bytes,10,opt"`
	DownloadIp     string   `protobuf:"bytes,20,opt"`
	DownloadDomain string   `protobuf:"bytes,30,opt"`
	Port           int32    `protobuf:"varint,40,opt"`
	DownloadUrl    string   `protobuf:"bytes,50,opt"`
	DownloadipList []string `protobuf:"bytes,60,rep"`
	Cookie         string   `protobuf:"bytes,70,opt"`
}

type DownloadSuccReq struct {
	Uin  int64  `protobuf:"varint,10,opt"`
	Uuid []byte `protobuf:"bytes,20,opt"`
}

type DownloadSuccRsp struct {
	RetCode  int32  `protobuf:"varint,10,opt"`
	RetMsg   string `protobuf:"bytes,20,opt"`
	DownStat int32  `protobuf:"varint,30,opt"`
}

type ExtensionReq struct {
	Id               int64          `protobuf:"varint,1,opt"`
	Type             int64          `protobuf:"varint,2,opt"`
	DstPhonenum      string         `protobuf:"bytes,3,opt"`
	PhoneConvertType int32          `protobuf:"varint,4,opt"`
	Sig              []byte         `protobuf:"bytes,20,opt"`
	RouteId          int64          `protobuf:"varint,100,opt"`
	DelMessageReq    *DelMessageReq `protobuf:"bytes,90100,opt"`
	DownloadUrlType  int32          `protobuf:"varint,90200,opt"`
	PttFormat        int32          `protobuf:"varint,90300,opt"`
	IsNeedInnerIp    int32          `protobuf:"varint,90400,opt"`
	NetType          int32          `protobuf:"varint,90500,opt"`
	VoiceType        int32          `protobuf:"varint,90600,opt"`
	FileType         int32          `protobuf:"varint,90700,opt"`
	PttTime          int32          `protobuf:"varint,90800,opt"`
}

type ExtensionRsp struct {
}

type FileInfo struct {
	Uin          int64  `protobuf:"varint,1,opt"`
	DangerEvel   int32  `protobuf:"varint,2,opt"`
	FileSize     int64  `protobuf:"varint,3,opt"`
	LifeTime     int32  `protobuf:"varint,4,opt"`
	UploadTime   int32  `protobuf:"varint,5,opt"`
	Uuid         []byte `protobuf:"bytes,6,opt"`
	FileName     string `protobuf:"bytes,7,opt"`
	AbsFileType  int32  `protobuf:"varint,90,opt"`
	Bytes_10MMd5 []byte `protobuf:"bytes,100,opt"`
	Sha          []byte `protobuf:"bytes,101,opt"`
	ClientType   int32  `protobuf:"varint,110,opt"`
	OwnerUin     int64  `protobuf:"varint,120,opt"`
	PeerUin      int64  `protobuf:"varint,121,opt"`
	ExpireTime   int32  `protobuf:"varint,130,opt"`
}

type FileQueryReq struct {
	Uin  int64  `protobuf:"varint,10,opt"`
	Uuid []byte `protobuf:"bytes,20,opt"`
}

type FileQueryRsp struct {
	RetCode  int32     `protobuf:"varint,10,opt"`
	RetMsg   string    `protobuf:"bytes,20,opt"`
	FileInfo *FileInfo `protobuf:"bytes,30,opt"`
}

type RecallFileReq struct {
	Uin  int64  `protobuf:"varint,1,opt"`
	Uuid []byte `protobuf:"bytes,2,opt"`
}

type RecallFileRsp struct {
	RetCode int32  `protobuf:"varint,1,opt"`
	RetMsg  string `protobuf:"bytes,2,opt"`
}

type RecvListQueryReq struct {
	Uin        int64 `protobuf:"varint,1,opt"`
	BeginIndex int32 `protobuf:"varint,2,opt"`
	ReqCount   int32 `protobuf:"varint,3,opt"`
}

type RecvListQueryRsp struct {
	RetCode      int32       `protobuf:"varint,1,opt"`
	RetMsg       string      `protobuf:"bytes,2,opt"`
	FileTotCount int32       `protobuf:"varint,3,opt"`
	BeginIndex   int32       `protobuf:"varint,4,opt"`
	RspFileCount int32       `protobuf:"varint,5,opt"`
	IsEnd        int32       `protobuf:"varint,6,opt"`
	FileList     []*FileInfo `protobuf:"bytes,7,rep"`
}

type RenewFileReq struct {
	Uin    int64  `protobuf:"varint,1,opt"`
	Uuid   []byte `protobuf:"bytes,2,opt"`
	AddTtl int32  `protobuf:"varint,3,opt"`
}

type RenewFileRsp struct {
	RetCode int32  `protobuf:"varint,1,opt"`
	RetMsg  string `protobuf:"bytes,2,opt"`
}

type C346ReqBody struct {
	Cmd                  int32                 `protobuf:"varint,1,opt"`
	Seq                  int32                 `protobuf:"varint,2,opt"`
	RecvListQueryReq     *RecvListQueryReq     `protobuf:"bytes,3,opt"`
	SendListQueryReq     *SendListQueryReq     `protobuf:"bytes,4,opt"`
	RenewFileReq         *RenewFileReq         `protobuf:"bytes,5,opt"`
	RecallFileReq        *RecallFileReq        `protobuf:"bytes,6,opt"`
	ApplyUploadReq       *ApplyUploadReq       `protobuf:"bytes,7,opt"`
	ApplyUploadHitReq    *ApplyUploadHitReq    `protobuf:"bytes,8,opt"`
	ApplyForwardFileReq  *ApplyForwardFileReq  `protobuf:"bytes,9,opt"`
	UploadSuccReq        *UploadSuccReq        `protobuf:"bytes,10,opt"`
	DeleteFileReq        *DeleteFileReq        `protobuf:"bytes,11,opt"`
	DownloadSuccReq      *DownloadSuccReq      `protobuf:"bytes,12,opt"`
	ApplyDownloadAbsReq  *ApplyDownloadAbsReq  `protobuf:"bytes,13,opt"`
	ApplyDownloadReq     *ApplyDownloadReq     `protobuf:"bytes,14,opt"`
	ApplyListDownloadReq *ApplyListDownloadReq `protobuf:"bytes,15,opt"`
	FileQueryReq         *FileQueryReq         `protobuf:"bytes,16,opt"`
	ApplyCopyFromReq     *ApplyCopyFromReq     `protobuf:"bytes,17,opt"`
	ApplyUploadReqV2     *ApplyUploadReqV2     `protobuf:"bytes,18,opt"`
	ApplyUploadReqV3     *ApplyUploadReqV3     `protobuf:"bytes,19,opt"`
	ApplyUploadHitReqV2  *ApplyUploadHitReqV2  `protobuf:"bytes,20,opt"`
	ApplyUploadHitReqV3  *ApplyUploadHitReqV3  `protobuf:"bytes,21,opt"`
	BusinessId           int32                 `protobuf:"varint,101,opt"`
	ClientType           int32                 `protobuf:"varint,102,opt"`
	ApplyCopyToReq       *ApplyCopyToReq       `protobuf:"bytes,90000,opt"`
	//ApplyCleanTrafficReq applyCleanTrafficReq = 90001; empty message
	ApplyGetTrafficReq *ApplyGetTrafficReq `protobuf:"bytes,90002,opt"`
	ExtensionReq       *ExtensionReq       `protobuf:"bytes,99999,opt"`
}

type C346RspBody struct {
	Cmd                  int32                 `protobuf:"varint,1,opt"`
	Seq                  int32                 `protobuf:"varint,2,opt"`
	RecvListQueryRsp     *RecvListQueryRsp     `protobuf:"bytes,3,opt"`
	SendListQueryRsp     *SendListQueryRsp     `protobuf:"bytes,4,opt"`
	RenewFileRsp         *RenewFileRsp         `protobuf:"bytes,5,opt"`
	RecallFileRsp        *RecallFileRsp        `protobuf:"bytes,6,opt"`
	ApplyUploadRsp       *ApplyUploadRsp       `protobuf:"bytes,7,opt"`
	ApplyUploadHitRsp    *ApplyUploadHitRsp    `protobuf:"bytes,8,opt"`
	ApplyForwardFileRsp  *ApplyForwardFileRsp  `protobuf:"bytes,9,opt"`
	UploadSuccRsp        *UploadSuccRsp        `protobuf:"bytes,10,opt"`
	DeleteFileRsp        *DeleteFileRsp        `protobuf:"bytes,11,opt"`
	DownloadSuccRsp      *DownloadSuccRsp      `protobuf:"bytes,12,opt"`
	ApplyDownloadAbsRsp  *ApplyDownloadAbsRsp  `protobuf:"bytes,13,opt"`
	ApplyDownloadRsp     *ApplyDownloadRsp     `protobuf:"bytes,14,opt"`
	ApplyListDownloadRsp *ApplyListDownloadRsp `protobuf:"bytes,15,opt"`
	FileQueryRsp         *FileQueryRsp         `protobuf:"bytes,16,opt"`
	ApplyCopyFromRsp     *ApplyCopyFromRsp     `protobuf:"bytes,17,opt"`
	ApplyUploadRspV2     *ApplyUploadRspV2     `protobuf:"bytes,18,opt"`
	ApplyUploadRspV3     *ApplyUploadRspV3     `protobuf:"bytes,19,opt"`
	ApplyUploadHitRspV2  *ApplyUploadHitRspV2  `protobuf:"bytes,20,opt"`
	ApplyUploadHitRspV3  *ApplyUploadHitRspV3  `protobuf:"bytes,21,opt"`
	BusinessId           int32                 `protobuf:"varint,101,opt"`
	ClientType           int32                 `protobuf:"varint,102,opt"`
	ApplyCopyToRsp       *ApplyCopyToRsp       `protobuf:"bytes,90000,opt"`
	ApplyCleanTrafficRsp *ApplyCleanTrafficRsp `protobuf:"bytes,90001,opt"`
	ApplyGetTrafficRsp   *ApplyGetTrafficRsp   `protobuf:"bytes,90002,opt"`
	ExtensionRsp         *ExtensionRsp         `protobuf:"bytes,99999,opt"`
}

type SendListQueryReq struct {
	Uin        int64 `protobuf:"varint,1,opt"`
	BeginIndex int32 `protobuf:"varint,2,opt"`
	ReqCount   int32 `protobuf:"varint,3,opt"`
}

type SendListQueryRsp struct {
	RetCode      int32       `protobuf:"varint,1,opt"`
	RetMsg       string      `protobuf:"bytes,2,opt"`
	FileTotCount int32       `protobuf:"varint,3,opt"`
	BeginIndex   int32       `protobuf:"varint,4,opt"`
	RspFileCount int32       `protobuf:"varint,5,opt"`
	IsEnd        int32       `protobuf:"varint,6,opt"`
	TotLimit     int64       `protobuf:"varint,7,opt"`
	UsedLimit    int64       `protobuf:"varint,8,opt"`
	FileList     []*FileInfo `protobuf:"bytes,9,rep"`
}

type UploadSuccReq struct {
	SenderUin int64  `protobuf:"varint,10,opt"`
	RecverUin int64  `protobuf:"varint,20,opt"`
	Uuid      []byte `protobuf:"bytes,30,opt"`
}

type UploadSuccRsp struct {
	RetCode  int32     `protobuf:"varint,10,opt"`
	RetMsg   string    `protobuf:"bytes,20,opt"`
	FileInfo *FileInfo `protobuf:"bytes,30,opt"`
}
