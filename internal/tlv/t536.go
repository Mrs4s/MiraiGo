package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T536(loginExtraData []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x536)
		pos := w.AllocUInt16Head()
		w.Write(loginExtraData)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
