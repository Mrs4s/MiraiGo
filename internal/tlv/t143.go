package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T143(arr []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x143)
		w.WriteBytesShort(arr)
	})
}
