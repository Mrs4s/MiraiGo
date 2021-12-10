package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T525(t536 []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x525)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt16(1)
			w.Write(t536)
		}))
	})
}
