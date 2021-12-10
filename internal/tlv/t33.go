package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T33(guid []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x33)
		w.WriteBytesShort(guid)
	})
}
