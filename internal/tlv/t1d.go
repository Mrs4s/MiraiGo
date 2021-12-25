package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T1D(miscBitmap uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x1D)
		pos := w.FillUInt16()
		w.WriteByte(1)
		w.WriteUInt32(miscBitmap)
		w.WriteUInt32(0)
		w.WriteByte(0)
		w.WriteUInt32(0)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
