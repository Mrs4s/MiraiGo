package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T16E(buildModel []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x16e)
		pos := w.AllocHead16()
		w.Write(buildModel)
		w.WriteHead16ExcludeSelf(pos)
	})
}
