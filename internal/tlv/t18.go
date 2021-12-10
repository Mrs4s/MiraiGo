package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T18(appId uint32, uin uint32) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x18)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt16(1)
			w.WriteUInt32(1536)
			w.WriteUInt32(appId)
			w.WriteUInt32(0)
			w.WriteUInt32(uin)
			w.WriteUInt16(0)
			w.WriteUInt16(0)
		}))
	})
}
