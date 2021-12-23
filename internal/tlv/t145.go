package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T145(guid []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x145)
		pos := w.AllocUInt16Head()
		w.Write(guid)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
