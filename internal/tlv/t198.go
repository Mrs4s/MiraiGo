package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T198() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x198)
		w.Write([]byte{0, 1, 0}) // w.WriteBytesShort([]byte{0})
	})
}
