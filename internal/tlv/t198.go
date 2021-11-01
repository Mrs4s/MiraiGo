package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T198() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x198)
		w.WriteBytesShort([]byte{0})
	})
}
