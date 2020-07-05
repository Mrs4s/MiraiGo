package crypto

import (
	"encoding/hex"
	"github.com/Mrs4s/MiraiGo/binary"
)

type EncryptECDH struct {
	InitialShareKey []byte
	PublicKey       []byte
}

var ECDH = &EncryptECDH{}

func init() {
	//TODO: Keygen
	ECDH.InitialShareKey, _ = hex.DecodeString("41d0d17c506a5256d0d08d7aac133c70")
	ECDH.PublicKey, _ = hex.DecodeString("049fb03421ba7ab5fc91c2d94a7657fff7ba8fe09f08a22951a24865212cbc45aff1b5125188fa8f0e30473bc55d54edc2")
}

func (e *EncryptECDH) DoEncrypt(d, k []byte) []byte {
	w := binary.NewWriter()
	w.WriteByte(0x01)
	w.WriteByte(0x01)
	w.Write(k)
	w.WriteUInt16(258)
	w.WriteUInt16(uint16(len(ECDH.PublicKey)))
	w.Write(ECDH.PublicKey)
	w.EncryptAndWrite(ECDH.InitialShareKey, d)
	return w.Bytes()
}

func (e *EncryptECDH) Id() byte {
	return 7
}
