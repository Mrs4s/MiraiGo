package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T191(k byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x191)
		pos := w.AllocUInt16Head()
		w.WriteByte(k)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
