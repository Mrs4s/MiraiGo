package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T521(i uint32) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x521)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt32(i)
			w.WriteUInt16(0)
		}))
	})
}
