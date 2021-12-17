package codec

import "github.com/Mrs4s/MiraiGo/binary"

type Encryptor interface {
	Encrypt([]byte, []byte) []byte
	ID() byte
}

type OICQ struct {
	Uin           uint32
	Command       uint16
	EncryptMethod Encryptor
	Key           []byte
	Body          []byte
}

func (m *OICQ) Encode() []byte {
	body := m.EncryptMethod.Encrypt(m.Body, m.Key)

	p := binary.SelectWriter()
	defer binary.PutWriter(p)

	p.WriteByte(0x02)
	p.WriteUInt16(27 + 2 + uint16(len(body)))
	p.WriteUInt16(8001)
	p.WriteUInt16(m.Command)
	p.WriteUInt16(1)
	p.WriteUInt32(m.Uin)
	p.WriteByte(3)
	p.WriteByte(m.EncryptMethod.ID())
	p.WriteByte(0)
	p.WriteUInt32(2)
	p.WriteUInt32(0)
	p.WriteUInt32(0)
	p.Write(body)
	p.WriteByte(0x03)

	return append([]byte(nil), p.Bytes()...)
}
