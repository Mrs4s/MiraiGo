package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T177() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x177)
		w.WriteTlv(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteByte(0x01)
			w.WriteUInt32(1571193922)
			w.WriteTlv([]byte("6.0.0.2413"))
		}))
	})
}
