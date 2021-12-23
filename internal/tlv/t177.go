package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T177(buildTime uint32, sdkVersion string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x177)
		pos := w.AllocUInt16Head()
		w.WriteByte(0x01)
		w.WriteUInt32(buildTime)
		w.WriteStringShort(sdkVersion)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
