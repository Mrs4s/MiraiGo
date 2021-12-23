package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T525(t536 []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x525)
		pos := w.AllocUInt16Head()
		w.WriteUInt16(1)
		w.Write(t536)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
