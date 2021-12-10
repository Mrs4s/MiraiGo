package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T194(imsiMd5 []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x194)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.Write(imsiMd5)
		}))
	})
}
