package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T108(imei string) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x108)
		w.WriteTlv([]byte(imei))
	})
}
