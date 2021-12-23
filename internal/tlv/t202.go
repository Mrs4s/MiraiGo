package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T202(wifiBSSID, wifiSSID []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x202)
		pos := w.FillUInt16()
		w.WriteTlvLimitedSize(wifiBSSID, 16)
		w.WriteTlvLimitedSize(wifiSSID, 32)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
