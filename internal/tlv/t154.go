package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T154(seq uint16) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x154)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt32(uint32(seq))
		}))
	})
}
