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
	return binary.NewWriterF(func(w *binary.Writer) {
		body := m.EncryptMethod.Encrypt(m.Body, m.Key)
		w.WriteByte(0x02)                         // 1
		w.WriteUInt16(27 + 2 + uint16(len(body))) // 2
		w.WriteUInt16(8001)                       // 2
		w.WriteUInt16(m.Command)                  // 2
		w.WriteUInt16(1)                          // 2
		w.WriteUInt32(m.Uin)                      // 4
		w.WriteByte(3)                            // 1
		w.WriteByte(m.EncryptMethod.ID())         // 1
		w.WriteByte(0)                            // 1
		w.WriteUInt32(2)                          // 4
		w.WriteUInt32(0)                          // 4
		w.WriteUInt32(0)                          // 4
		w.Write(body)                             // len(body)
		w.WriteByte(0x03)                         // 1
	})
}
