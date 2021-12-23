package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T147(appId uint32, apkVersionName, apkSignatureMd5 []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x147)
		pos := w.FillUInt16()
		w.WriteUInt32(appId)
		w.WriteTlvLimitedSize(apkVersionName, 32)
		w.WriteTlvLimitedSize(apkSignatureMd5, 32)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
