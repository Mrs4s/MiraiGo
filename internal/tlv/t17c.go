package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T17C(code string) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x17c)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteStringShort(code)
		}))
	})
}
