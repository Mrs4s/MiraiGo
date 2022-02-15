// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/cmd0x3f6/cmd0x3f6.proto

package cmd0x3f6

type C3F6ReqBody struct {
	SubCmd                            *uint32                            `protobuf:"varint,1,opt"`
	CrmCommonHead                     *C3F6CRMMsgHead                    `protobuf:"bytes,2,opt"`
	SubcmdLoginProcessCompleteReqBody *QDUserLoginProcessCompleteReqBody `protobuf:"bytes,42,opt"`
}

func (x *C3F6ReqBody) GetSubCmd() uint32 {
	if x != nil && x.SubCmd != nil {
		return *x.SubCmd
	}
	return 0
}

type C3F6RspBody struct {
	SubCmd                            *uint32                            `protobuf:"varint,1,opt"`
	CrmCommonHead                     *C3F6CRMMsgHead                    `protobuf:"bytes,2,opt"`
	SubcmdLoginProcessCompleteRspBody *QDUserLoginProcessCompleteRspBody `protobuf:"bytes,42,opt"`
}

func (x *C3F6RspBody) GetSubCmd() uint32 {
	if x != nil && x.SubCmd != nil {
		return *x.SubCmd
	}
	return 0
}

type QDUserLoginProcessCompleteReqBody struct {
	Kfext        *uint64 `protobuf:"varint,1,opt"`
	Pubno        *uint32 `protobuf:"varint,2,opt"`
	Buildno      *uint32 `protobuf:"varint,3,opt"`
	TerminalType *uint32 `protobuf:"varint,4,opt"`
	Status       *uint32 `protobuf:"varint,5,opt"`
	LoginTime    *uint32 `protobuf:"varint,6,opt"`
	HardwareInfo *string `protobuf:"bytes,7,opt"`
	SoftwareInfo *string `protobuf:"bytes,8,opt"`
	Guid         []byte  `protobuf:"bytes,9,opt"`
	AppName      *string `protobuf:"bytes,10,opt"`
	SubAppId     *uint32 `protobuf:"varint,11,opt"`
}

func (x *QDUserLoginProcessCompleteReqBody) GetKfext() uint64 {
	if x != nil && x.Kfext != nil {
		return *x.Kfext
	}
	return 0
}

func (x *QDUserLoginProcessCompleteReqBody) GetPubno() uint32 {
	if x != nil && x.Pubno != nil {
		return *x.Pubno
	}
	return 0
}

func (x *QDUserLoginProcessCompleteReqBody) GetBuildno() uint32 {
	if x != nil && x.Buildno != nil {
		return *x.Buildno
	}
	return 0
}

func (x *QDUserLoginProcessCompleteReqBody) GetTerminalType() uint32 {
	if x != nil && x.TerminalType != nil {
		return *x.TerminalType
	}
	return 0
}

func (x *QDUserLoginProcessCompleteReqBody) GetStatus() uint32 {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return 0
}

func (x *QDUserLoginProcessCompleteReqBody) GetLoginTime() uint32 {
	if x != nil && x.LoginTime != nil {
		return *x.LoginTime
	}
	return 0
}

func (x *QDUserLoginProcessCompleteReqBody) GetHardwareInfo() string {
	if x != nil && x.HardwareInfo != nil {
		return *x.HardwareInfo
	}
	return ""
}

func (x *QDUserLoginProcessCompleteReqBody) GetSoftwareInfo() string {
	if x != nil && x.SoftwareInfo != nil {
		return *x.SoftwareInfo
	}
	return ""
}

func (x *QDUserLoginProcessCompleteReqBody) GetAppName() string {
	if x != nil && x.AppName != nil {
		return *x.AppName
	}
	return ""
}

func (x *QDUserLoginProcessCompleteReqBody) GetSubAppId() uint32 {
	if x != nil && x.SubAppId != nil {
		return *x.SubAppId
	}
	return 0
}

type QDUserLoginProcessCompleteRspBody struct {
	Ret                *RetInfo `protobuf:"bytes,1,opt"`
	Url                *string  `protobuf:"bytes,2,opt"`
	Mobile             *string  `protobuf:"bytes,3,opt"`
	ExternalMobile     *string  `protobuf:"bytes,4,opt"`
	DataAnalysisPriv   *bool    `protobuf:"varint,5,opt"`
	DeviceLock         *bool    `protobuf:"varint,6,opt"`
	ModulePrivilege    *uint64  `protobuf:"varint,7,opt"`
	ModuleSubPrivilege []uint32 `protobuf:"varint,8,rep"`
	MasterSet          *uint32  `protobuf:"varint,9,opt"`
	ExtSet             *uint32  `protobuf:"varint,10,opt"`
	CorpConfProperty   *uint64  `protobuf:"varint,11,opt"`
	Corpuin            *uint64  `protobuf:"varint,12,opt"`
	Kfaccount          *uint64  `protobuf:"varint,13,opt"`
	SecurityLevel      *uint32  `protobuf:"varint,14,opt"`
	MsgTitle           *string  `protobuf:"bytes,15,opt"`
	SuccNoticeMsg      *string  `protobuf:"bytes,16,opt"`
	NameAccount        *uint64  `protobuf:"varint,17,opt"`
	CrmMigrateFlag     *uint32  `protobuf:"varint,18,opt"`
	ExtuinName         *string  `protobuf:"bytes,19,opt"`
	OpenAccountTime    *uint32  `protobuf:"varint,20,opt"`
}

func (x *QDUserLoginProcessCompleteRspBody) GetUrl() string {
	if x != nil && x.Url != nil {
		return *x.Url
	}
	return ""
}

func (x *QDUserLoginProcessCompleteRspBody) GetMobile() string {
	if x != nil && x.Mobile != nil {
		return *x.Mobile
	}
	return ""
}

func (x *QDUserLoginProcessCompleteRspBody) GetExternalMobile() string {
	if x != nil && x.ExternalMobile != nil {
		return *x.ExternalMobile
	}
	return ""
}

func (x *QDUserLoginProcessCompleteRspBody) GetDataAnalysisPriv() bool {
	if x != nil && x.DataAnalysisPriv != nil {
		return *x.DataAnalysisPriv
	}
	return false
}

func (x *QDUserLoginProcessCompleteRspBody) GetDeviceLock() bool {
	if x != nil && x.DeviceLock != nil {
		return *x.DeviceLock
	}
	return false
}

func (x *QDUserLoginProcessCompleteRspBody) GetModulePrivilege() uint64 {
	if x != nil && x.ModulePrivilege != nil {
		return *x.ModulePrivilege
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetMasterSet() uint32 {
	if x != nil && x.MasterSet != nil {
		return *x.MasterSet
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetExtSet() uint32 {
	if x != nil && x.ExtSet != nil {
		return *x.ExtSet
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetCorpConfProperty() uint64 {
	if x != nil && x.CorpConfProperty != nil {
		return *x.CorpConfProperty
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetCorpuin() uint64 {
	if x != nil && x.Corpuin != nil {
		return *x.Corpuin
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetKfaccount() uint64 {
	if x != nil && x.Kfaccount != nil {
		return *x.Kfaccount
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetSecurityLevel() uint32 {
	if x != nil && x.SecurityLevel != nil {
		return *x.SecurityLevel
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetMsgTitle() string {
	if x != nil && x.MsgTitle != nil {
		return *x.MsgTitle
	}
	return ""
}

func (x *QDUserLoginProcessCompleteRspBody) GetSuccNoticeMsg() string {
	if x != nil && x.SuccNoticeMsg != nil {
		return *x.SuccNoticeMsg
	}
	return ""
}

func (x *QDUserLoginProcessCompleteRspBody) GetNameAccount() uint64 {
	if x != nil && x.NameAccount != nil {
		return *x.NameAccount
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetCrmMigrateFlag() uint32 {
	if x != nil && x.CrmMigrateFlag != nil {
		return *x.CrmMigrateFlag
	}
	return 0
}

func (x *QDUserLoginProcessCompleteRspBody) GetExtuinName() string {
	if x != nil && x.ExtuinName != nil {
		return *x.ExtuinName
	}
	return ""
}

func (x *QDUserLoginProcessCompleteRspBody) GetOpenAccountTime() uint32 {
	if x != nil && x.OpenAccountTime != nil {
		return *x.OpenAccountTime
	}
	return 0
}

type RetInfo struct {
	RetCode  *uint32 `protobuf:"varint,1,opt"`
	ErrorMsg *string `protobuf:"bytes,2,opt"`
}

func (x *RetInfo) GetRetCode() uint32 {
	if x != nil && x.RetCode != nil {
		return *x.RetCode
	}
	return 0
}

func (x *RetInfo) GetErrorMsg() string {
	if x != nil && x.ErrorMsg != nil {
		return *x.ErrorMsg
	}
	return ""
}

type C3F6CRMMsgHead struct {
	CrmSubCmd  *uint32 `protobuf:"varint,1,opt"`
	HeadLen    *uint32 `protobuf:"varint,2,opt"`
	VerNo      *uint32 `protobuf:"varint,3,opt"`
	KfUin      *uint64 `protobuf:"varint,4,opt"`
	Seq        *uint32 `protobuf:"varint,5,opt"`
	PackNum    *uint32 `protobuf:"varint,6,opt"`
	CurPack    *uint32 `protobuf:"varint,7,opt"`
	BufSig     *string `protobuf:"bytes,8,opt"`
	Clienttype *uint32 `protobuf:"varint,9,opt"`
	LaborUin   *uint64 `protobuf:"varint,10,opt"`
	LaborName  *string `protobuf:"bytes,11,opt"`
	Kfaccount  *uint64 `protobuf:"varint,12,opt"`
	TraceId    *string `protobuf:"bytes,13,opt"`
	AppId      *uint32 `protobuf:"varint,14,opt"`
}

func (x *C3F6CRMMsgHead) GetCrmSubCmd() uint32 {
	if x != nil && x.CrmSubCmd != nil {
		return *x.CrmSubCmd
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetHeadLen() uint32 {
	if x != nil && x.HeadLen != nil {
		return *x.HeadLen
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetVerNo() uint32 {
	if x != nil && x.VerNo != nil {
		return *x.VerNo
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetKfUin() uint64 {
	if x != nil && x.KfUin != nil {
		return *x.KfUin
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetSeq() uint32 {
	if x != nil && x.Seq != nil {
		return *x.Seq
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetPackNum() uint32 {
	if x != nil && x.PackNum != nil {
		return *x.PackNum
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetCurPack() uint32 {
	if x != nil && x.CurPack != nil {
		return *x.CurPack
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetBufSig() string {
	if x != nil && x.BufSig != nil {
		return *x.BufSig
	}
	return ""
}

func (x *C3F6CRMMsgHead) GetClienttype() uint32 {
	if x != nil && x.Clienttype != nil {
		return *x.Clienttype
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetLaborUin() uint64 {
	if x != nil && x.LaborUin != nil {
		return *x.LaborUin
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetLaborName() string {
	if x != nil && x.LaborName != nil {
		return *x.LaborName
	}
	return ""
}

func (x *C3F6CRMMsgHead) GetKfaccount() uint64 {
	if x != nil && x.Kfaccount != nil {
		return *x.Kfaccount
	}
	return 0
}

func (x *C3F6CRMMsgHead) GetTraceId() string {
	if x != nil && x.TraceId != nil {
		return *x.TraceId
	}
	return ""
}

func (x *C3F6CRMMsgHead) GetAppId() uint32 {
	if x != nil && x.AppId != nil {
		return *x.AppId
	}
	return 0
}
