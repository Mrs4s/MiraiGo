package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T194(imsiMd5 []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x194)
		pos := w.AllocUInt16Head()
		w.Write(imsiMd5)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
