package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T35(productType uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x35)
		pos := w.FillUInt16()
		w.WriteUInt32(productType)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
