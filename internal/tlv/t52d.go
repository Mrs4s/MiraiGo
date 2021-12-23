package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T52D(devInfo []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x52d)
		pos := w.FillUInt16()
		w.Write(devInfo)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
