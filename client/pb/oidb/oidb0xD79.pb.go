// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/oidb/oidb0xD79.proto

package oidb

type D79ReqBody struct {
	Seq          uint64 `protobuf:"varint,1,opt"`
	Uin          uint64 `protobuf:"varint,2,opt"`
	CompressFlag uint32 `protobuf:"varint,3,opt"`
	Content      []byte `protobuf:"bytes,4,opt"`
	SenderUin    uint64 `protobuf:"varint,5,opt"`
	Qua          []byte `protobuf:"bytes,6,opt"`
	WordExt      []byte `protobuf:"bytes,7,opt"`
}

func (x *D79ReqBody) GetSeq() uint64 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *D79ReqBody) GetUin() uint64 {
	if x != nil {
		return x.Uin
	}
	return 0
}

func (x *D79ReqBody) GetCompressFlag() uint32 {
	if x != nil {
		return x.CompressFlag
	}
	return 0
}

func (x *D79ReqBody) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *D79ReqBody) GetSenderUin() uint64 {
	if x != nil {
		return x.SenderUin
	}
	return 0
}

func (x *D79ReqBody) GetQua() []byte {
	if x != nil {
		return x.Qua
	}
	return nil
}

func (x *D79ReqBody) GetWordExt() []byte {
	if x != nil {
		return x.WordExt
	}
	return nil
}

type D79RspBody struct {
	Ret          uint32      `protobuf:"varint,1,opt"`
	Seq          uint64      `protobuf:"varint,2,opt"`
	Uin          uint64      `protobuf:"varint,3,opt"`
	CompressFlag uint32      `protobuf:"varint,4,opt"`
	Content      *D79Content `protobuf:"bytes,5,opt"`
}

func (x *D79RspBody) GetRet() uint32 {
	if x != nil {
		return x.Ret
	}
	return 0
}

func (x *D79RspBody) GetSeq() uint64 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *D79RspBody) GetUin() uint64 {
	if x != nil {
		return x.Uin
	}
	return 0
}

func (x *D79RspBody) GetCompressFlag() uint32 {
	if x != nil {
		return x.CompressFlag
	}
	return 0
}

func (x *D79RspBody) GetContent() *D79Content {
	if x != nil {
		return x.Content
	}
	return nil
}

type D79Content struct {
	SliceContent [][]byte `protobuf:"bytes,1,rep"`
}

func (x *D79Content) GetSliceContent() [][]byte {
	if x != nil {
		return x.SliceContent
	}
	return nil
}
