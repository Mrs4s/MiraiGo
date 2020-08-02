package crypto

import (
	"crypto/md5"
	"crypto/rand"
	"github.com/Mrs4s/MiraiGo/binary"
	"math/big"
)

type EncryptECDH struct {
	InitialShareKey []byte
	PublicKey       []byte
}

var ECDH = &EncryptECDH{}

var tenKeyX = new(big.Int).SetBytes([]byte{ // pubkey[1:24]
	0x92, 0x8d, 0x88, 0x50, 0x67, 0x30, 0x88, 0xb3, 0x43, 0x26, 0x4e, 0x0c,
	0x6b, 0xac, 0xb8, 0x49, 0x6d, 0x69, 0x77, 0x99, 0xf3, 0x72, 0x11, 0xde,
})

var tenKeyY = new(big.Int).SetBytes([]byte{ // pubkey[25:48]
	0xb2, 0x5b, 0xb7, 0x39, 0x06, 0xcb, 0x08, 0x9f, 0xea, 0x96, 0x39, 0xb4,
	0xe0, 0x26, 0x04, 0x98, 0xb5, 0x1a, 0x99, 0x2d, 0x50, 0x81, 0x3d, 0xa8,
})

func init() {
	key, sx, sy, err := secp192k1.GenerateKey(rand.Reader)
	if err != nil {
		panic("Can't Create ECDH key pair")
	}
	x, _ := secp192k1.ScalarMult(tenKeyX, tenKeyY, key)
	hash := md5.Sum(x.Bytes())
	ECDH.InitialShareKey = hash[:]
	ECDH.PublicKey = make([]byte, 49)[:0]
	ECDH.PublicKey = append(ECDH.PublicKey, 0x04)
	ECDH.PublicKey = append(ECDH.PublicKey, sx.Bytes()...)
	ECDH.PublicKey = append(ECDH.PublicKey, sy.Bytes()...)

	//ECDH.InitialShareKey, _ = hex.DecodeString("41d0d17c506a5256d0d08d7aac133c70")
	//ECDH.PublicKey, _ = hex.DecodeString("049fb03421ba7ab5fc91c2d94a7657fff7ba8fe09f08a22951a24865212cbc45aff1b5125188fa8f0e30473bc55d54edc2")
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
