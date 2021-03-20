package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T141(simInfo, apn []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x141)
		w.WriteBytesShort(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt16(1)
			w.WriteBytesShort(simInfo)
			w.WriteUInt16(2) // network type wifi
			w.WriteBytesShort(apn)
		}))
	})
}
