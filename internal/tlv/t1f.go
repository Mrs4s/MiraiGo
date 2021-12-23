package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T1F(isRoot bool, osName, osVersion, simOperatorName, apn []byte, networkType uint16) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x1F)
		pos := w.AllocUInt16Head()
		w.WriteBool(isRoot)
		w.WriteBytesShort(osName)
		w.WriteBytesShort(osVersion)
		w.WriteUInt16(networkType)
		w.WriteBytesShort(simOperatorName)
		w.WriteUInt16(0) // w.WriteBytesShort([]byte{})
		w.WriteBytesShort(apn)
		w.WriteUInt16HeadExcludeSelfAt(pos)
	})
}
