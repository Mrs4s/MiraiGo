package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T177(buildTime uint32, sdkVersion string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x177)
		pos := w.FillUInt16()
		w.WriteByte(0x01)
		w.WriteUInt32(buildTime)
		w.WriteStringShort(sdkVersion)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
