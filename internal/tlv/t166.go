package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T166(imageType byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x166)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteByte(imageType)
		}))
	})
}
