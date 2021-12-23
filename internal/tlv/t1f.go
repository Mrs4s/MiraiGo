package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T1F(isRoot bool, osName, osVersion, simOperatorName, apn []byte, networkType uint16) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x1F)
		pos := w.AllocHead16()
		w.WriteBool(isRoot)
		w.WriteBytesShort(osName)
		w.WriteBytesShort(osVersion)
		w.WriteUInt16(networkType)
		w.WriteBytesShort(simOperatorName)
		w.WriteBytesShort([]byte{})
		w.WriteBytesShort(apn)
		w.WriteHead16ExcludeSelf(pos)
	})
}
