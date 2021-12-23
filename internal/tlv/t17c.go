package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T17C(code string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x17c)
		pos := w.AllocUInt16Head()
		w.WriteStringShort(code)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
