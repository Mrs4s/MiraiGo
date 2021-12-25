package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T16E(buildModel []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x16e)
		pos := w.FillUInt16()
		w.Write(buildModel)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
