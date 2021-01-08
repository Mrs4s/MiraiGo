package crypto

import (
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"github.com/Mrs4s/MiraiGo/binary"
	"math/big"
)

type EncryptECDH struct {
	InitialShareKey []byte
	PublicKey       []byte
}

type EncryptSession struct {
	T133 []byte
}

var ECDH = &EncryptECDH{}

var tenKeyX = new(big.Int).SetBytes([]byte{
	0xEB, 0xCA, 0x94, 0xD7, 0x33, 0xE3, 0x99, 0xB2,
	0xDB, 0x96, 0xEA, 0xCD, 0xD3, 0xF6, 0x9A, 0x8B,
	0xB0, 0xF7, 0x42, 0x24, 0xE2, 0xB4, 0x4E, 0x33,
	0x57, 0x81, 0x22, 0x11, 0xD2, 0xE6, 0x2E, 0xFB,
})

var tenKeyY = new(big.Int).SetBytes([]byte{
	0xC9, 0x1B, 0xB5, 0x53, 0x09, 0x8E, 0x25, 0xE3,
	0x3A, 0x79, 0x9A, 0xDC, 0x7F, 0x76, 0xFE, 0xB2,
	0x08, 0xDA, 0x7C, 0x65, 0x22, 0xCD, 0xB0, 0x71,
	0x9A, 0x30, 0x51, 0x80, 0xCC, 0x54, 0xA8, 0x2E,
})

func init() {
	p256 := elliptic.P256()
	key, sx, sy, err := elliptic.GenerateKey(p256, rand.Reader)
	if err != nil {
		panic("Can't Create ECDH key pair")
	}
	x, _ := p256.ScalarMult(tenKeyX, tenKeyY, key)
	hash := md5.Sum(x.Bytes()[:16])
	ECDH.InitialShareKey = hash[:]
	ECDH.PublicKey = make([]byte, 65)[:0]
	ECDH.PublicKey = append(ECDH.PublicKey, 0x04)
	ECDH.PublicKey = append(ECDH.PublicKey, sx.Bytes()...)
	ECDH.PublicKey = append(ECDH.PublicKey, sy.Bytes()...)
}

func (e *EncryptECDH) DoEncrypt(d, k []byte) []byte {
	w := binary.NewWriter()
	w.WriteByte(0x02)
	w.WriteByte(0x01)
	w.Write(k)
	w.WriteUInt16(0x01_31)
	w.WriteUInt16(0x00_01)
	w.WriteUInt16(uint16(len(ECDH.PublicKey)))
	w.Write(ECDH.PublicKey)
	w.EncryptAndWrite(ECDH.InitialShareKey, d)
	return w.Bytes()
}

func (e *EncryptECDH) Id() byte {
	return 0x87
}

func NewEncryptSession(t133 []byte) *EncryptSession {
	return &EncryptSession{T133: t133}
}

func (e *EncryptSession) DoEncrypt(d, k []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		encrypt := binary.NewTeaCipher(k).Encrypt(d)
		w.WriteUInt16(uint16(len(e.T133)))
		w.Write(e.T133)
		w.Write(encrypt)
	})
}

func (e *EncryptSession) Id() byte {
	return 69
}
