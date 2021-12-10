package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T516() ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x516)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt32(0)
		}))
	})
}
