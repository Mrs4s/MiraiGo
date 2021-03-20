package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T1F(isRoot bool, osName, osVersion, simOperatorName, apn []byte, networkType uint16) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x1F)
		w.WriteBytesShort(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteByte(func() byte {
				if isRoot {
					return 1
				} else {
					return 0
				}
			}())
			w.WriteBytesShort(osName)
			w.WriteBytesShort(osVersion)
			w.WriteUInt16(networkType)
			w.WriteBytesShort(simOperatorName)
			w.WriteBytesShort([]byte{})
			w.WriteBytesShort(apn)
		}))
	})
}
