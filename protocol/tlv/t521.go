package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T521() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x521)
		w.WriteTlv(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt32(0)
			w.WriteUInt16(0)
		}))
	})
}
