package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T141(simInfo, apn []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x141)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt16(1)
			w.WriteBytesShort(simInfo)
			w.WriteUInt16(2) // network type wifi
			w.WriteBytesShort(apn)
		}))
	})
}
