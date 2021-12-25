package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T17A(value int32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x17a)
		w.WriteUInt16(4) // len of uint32
		w.WriteUInt32(uint32(value))
	})
}
