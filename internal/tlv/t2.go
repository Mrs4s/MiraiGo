package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T2(result string, sign []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x02)
		pos := w.AllocUInt16Head()
		w.WriteUInt16(0)
		w.WriteStringShort(result)
		w.WriteBytesShort(sign)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
