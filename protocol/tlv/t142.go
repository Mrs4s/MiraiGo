package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T142(apkId string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x142)
		w.WriteTlv(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt16(0)
			w.WriteTlvLimitedSize([]byte(apkId), 32)
		}))
	})
}
