package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T1B(micro, version, size, margin, dpi, ecLevel, hint uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x1B)
		pos := w.FillUInt16()
		w.WriteUInt32(micro)
		w.WriteUInt32(version)
		w.WriteUInt32(size)
		w.WriteUInt32(margin)
		w.WriteUInt32(dpi)
		w.WriteUInt32(ecLevel)
		w.WriteUInt32(hint)
		w.WriteUInt16(0)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
