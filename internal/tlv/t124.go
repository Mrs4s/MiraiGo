package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T124(osType, osVersion, simInfo, apn []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x124)
		pos := w.FillUInt16()
		w.WriteTlvLimitedSize(osType, 16)
		w.WriteTlvLimitedSize(osVersion, 16)
		w.WriteUInt16(2) // Network type wifi
		w.WriteTlvLimitedSize(simInfo, 16)
		w.WriteTlvLimitedSize([]byte{}, 16)
		w.WriteTlvLimitedSize(apn, 16)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
