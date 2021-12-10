package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T145(guid []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x145)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.Write(guid)
		}))
	})
}
