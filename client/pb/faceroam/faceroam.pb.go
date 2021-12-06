// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/faceroam/faceroam.proto

package faceroam

type PlatInfo struct {
	Implat *int64  `protobuf:"varint,1,opt"`
	Osver  *string `protobuf:"bytes,2,opt"`
	Mqqver *string `protobuf:"bytes,3,opt"`
}

func (x *PlatInfo) GetImplat() int64 {
	if x != nil && x.Implat != nil {
		return *x.Implat
	}
	return 0
}

func (x *PlatInfo) GetOsver() string {
	if x != nil && x.Osver != nil {
		return *x.Osver
	}
	return ""
}

func (x *PlatInfo) GetMqqver() string {
	if x != nil && x.Mqqver != nil {
		return *x.Mqqver
	}
	return ""
}

type FaceroamReqBody struct {
	Comm          *PlatInfo      `protobuf:"bytes,1,opt"`
	Uin           *uint64        `protobuf:"varint,2,opt"`
	SubCmd        *uint32        `protobuf:"varint,3,opt"`
	ReqUserInfo   *ReqUserInfo   `protobuf:"bytes,4,opt"`
	ReqDeleteItem *ReqDeleteItem `protobuf:"bytes,5,opt"`
}

func (x *FaceroamReqBody) GetComm() *PlatInfo {
	if x != nil {
		return x.Comm
	}
	return nil
}

func (x *FaceroamReqBody) GetUin() uint64 {
	if x != nil && x.Uin != nil {
		return *x.Uin
	}
	return 0
}

func (x *FaceroamReqBody) GetSubCmd() uint32 {
	if x != nil && x.SubCmd != nil {
		return *x.SubCmd
	}
	return 0
}

func (x *FaceroamReqBody) GetReqUserInfo() *ReqUserInfo {
	if x != nil {
		return x.ReqUserInfo
	}
	return nil
}

func (x *FaceroamReqBody) GetReqDeleteItem() *ReqDeleteItem {
	if x != nil {
		return x.ReqDeleteItem
	}
	return nil
}

type ReqDeleteItem struct {
	Filename []string `protobuf:"bytes,1,rep"`
}

func (x *ReqDeleteItem) GetFilename() []string {
	if x != nil {
		return x.Filename
	}
	return nil
}

type ReqUserInfo struct {
}

type FaceroamRspBody struct {
	Ret           *int64         `protobuf:"varint,1,opt"`
	Errmsg        *string        `protobuf:"bytes,2,opt"`
	SubCmd        *uint32        `protobuf:"varint,3,opt"`
	RspUserInfo   *RspUserInfo   `protobuf:"bytes,4,opt"`
	RspDeleteItem *RspDeleteItem `protobuf:"bytes,5,opt"`
}

func (x *FaceroamRspBody) GetRet() int64 {
	if x != nil && x.Ret != nil {
		return *x.Ret
	}
	return 0
}

func (x *FaceroamRspBody) GetErrmsg() string {
	if x != nil && x.Errmsg != nil {
		return *x.Errmsg
	}
	return ""
}

func (x *FaceroamRspBody) GetSubCmd() uint32 {
	if x != nil && x.SubCmd != nil {
		return *x.SubCmd
	}
	return 0
}

func (x *FaceroamRspBody) GetRspUserInfo() *RspUserInfo {
	if x != nil {
		return x.RspUserInfo
	}
	return nil
}

func (x *FaceroamRspBody) GetRspDeleteItem() *RspDeleteItem {
	if x != nil {
		return x.RspDeleteItem
	}
	return nil
}

type RspDeleteItem struct {
	Filename []string `protobuf:"bytes,1,rep"`
	Ret      []int64  `protobuf:"varint,2,rep"`
}

func (x *RspDeleteItem) GetFilename() []string {
	if x != nil {
		return x.Filename
	}
	return nil
}

func (x *RspDeleteItem) GetRet() []int64 {
	if x != nil {
		return x.Ret
	}
	return nil
}

type RspUserInfo struct {
	Filename    []string `protobuf:"bytes,1,rep"`
	DeleteFile  []string `protobuf:"bytes,2,rep"`
	Bid         *string  `protobuf:"bytes,3,opt"`
	MaxRoamSize *uint32  `protobuf:"varint,4,opt"`
	EmojiType   []uint32 `protobuf:"varint,5,rep"`
}

func (x *RspUserInfo) GetFilename() []string {
	if x != nil {
		return x.Filename
	}
	return nil
}

func (x *RspUserInfo) GetDeleteFile() []string {
	if x != nil {
		return x.DeleteFile
	}
	return nil
}

func (x *RspUserInfo) GetBid() string {
	if x != nil && x.Bid != nil {
		return *x.Bid
	}
	return ""
}

func (x *RspUserInfo) GetMaxRoamSize() uint32 {
	if x != nil && x.MaxRoamSize != nil {
		return *x.MaxRoamSize
	}
	return 0
}

func (x *RspUserInfo) GetEmojiType() []uint32 {
	if x != nil {
		return x.EmojiType
	}
	return nil
}
