package network

type RequestType uint32

const (
	RequestTypeLogin  = 0x0A
	RequestTypeSimple = 0x0B
)

type EncryptType uint32

const (
	EncryptTypeNoEncrypt EncryptType = iota // 0x00
	EncryptTypeD2Key                        // 0x01
	EncryptTypeEmptyKey                     // 0x02
)

type Request struct {
	Type        RequestType
	EncryptType EncryptType
	SequenceID  int32
	Uin         int64
	CommandName string
	Body        []byte
}
