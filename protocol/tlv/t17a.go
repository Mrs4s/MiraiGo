package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T17A(value int32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x17a)
		w.WriteBytesShort(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt32(uint32(value))
		}))
	})
}
