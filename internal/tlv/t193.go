package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T193(ticket string) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x193)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.Write([]byte(ticket))
		}))
	})
}
