package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T107(picType uint16) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x107)
		pos := w.FillUInt16()
		w.WriteUInt16(picType)
		w.WriteByte(0x00)
		w.WriteUInt16(0)
		w.WriteByte(0x01)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
