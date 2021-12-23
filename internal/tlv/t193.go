package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T193(ticket string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x193)
		pos := w.AllocHead16()
		w.Write([]byte(ticket))
		w.WriteHead16ExcludeSelf(pos)
	})
}
