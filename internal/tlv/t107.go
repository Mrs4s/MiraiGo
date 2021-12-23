package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T107(picType uint16) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x107)
		pos := w.AllocHead16()
		w.WriteUInt16(picType)
		w.WriteByte(0x00)
		w.WriteUInt16(0)
		w.WriteByte(0x01)
		w.WriteHead16ExcludeSelf(pos)
	})
}
