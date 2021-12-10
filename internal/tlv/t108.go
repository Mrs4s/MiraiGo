package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T108(ksid []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x108)
		w.WriteBytesShort(ksid)
	})
}
