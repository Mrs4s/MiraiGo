package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T193(ticket string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x193)
		pos := w.FillUInt16()
		w.WriteString(ticket)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
