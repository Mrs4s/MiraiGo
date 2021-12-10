package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T16A(arr []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x16A)
		w.WriteBytesShort(arr)
	})
}
