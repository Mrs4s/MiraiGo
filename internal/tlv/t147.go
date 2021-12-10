package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T147(appId uint32, apkVersionName, apkSignatureMd5 []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x147)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt32(appId)
			w.WriteTlvLimitedSize(apkVersionName, 32)
			w.WriteTlvLimitedSize(apkSignatureMd5, 32)
		}))
	})
}
