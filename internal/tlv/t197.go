package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T197() ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x197)
		w.WriteBytesShort([]byte{0})
	})
}
