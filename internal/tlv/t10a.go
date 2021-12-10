package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T10A(arr []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x10A)
		w.WriteBytesShort(arr)
	})
}
