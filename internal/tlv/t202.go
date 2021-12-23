package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T202(wifiBSSID, wifiSSID []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x202)
		pos := w.AllocUInt16Head()
		w.WriteTlvLimitedSize(wifiBSSID, 16)
		w.WriteTlvLimitedSize(wifiSSID, 32)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
