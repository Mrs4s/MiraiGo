package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T52D(devInfo []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x52d)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.Write(devInfo)
		}))
	})
}
