package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T16(ssoVersion, appId, subAppId uint32, guid, apkId, apkVersionName, apkSign []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x16)
		w.WriteBytesShort(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt32(ssoVersion)
			w.WriteUInt32(appId)
			w.WriteUInt32(subAppId)
			w.Write(guid)
			w.WriteBytesShort(apkId)
			w.WriteBytesShort(apkVersionName)
			w.WriteBytesShort(apkSign)
		}))
	})
}
