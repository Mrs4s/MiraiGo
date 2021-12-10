package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T191(k byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x191)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteByte(k)
		}))
	})
}
