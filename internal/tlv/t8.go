package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T8(localId uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x8)
		pos := w.FillUInt16()
		w.WriteUInt16(0)
		w.WriteUInt32(localId)
		w.WriteUInt16(0)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
