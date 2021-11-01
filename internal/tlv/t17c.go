package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T17C(code string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x17c)
		w.WriteBytesShort(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteStringShort(code)
		}))
	})
}
