package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T174(data []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x174)
		w.WriteBytesShort(data)
	})
}
