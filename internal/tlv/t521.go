package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T521(i uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x521)
		w.WriteUInt16(4 + 2)
		w.WriteUInt32(i)
		w.WriteUInt16(0)
	})
}
