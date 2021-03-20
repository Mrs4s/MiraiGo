package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T35(productType uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x35)
		w.WriteBytesShort(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt32(productType)
		}))
	})
}
