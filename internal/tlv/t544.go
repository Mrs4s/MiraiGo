package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T544(userId uint64, moduleId string, subCmd uint32, sdkVersion string, guid []byte, signer func(string, []byte) []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x544)
		salt := binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt64(userId)
			w.WriteBytesShort(guid)
			w.WriteBytesShort([]byte(sdkVersion))
			w.WriteUInt32(subCmd)
		})
		w.WriteBytesShort(signer(moduleId, salt)) // temporary solution
	})
}

func T544Custom(moduleId string, salt []byte, signer func(string, []byte) []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x544)
		w.WriteBytesShort(signer(moduleId, salt))
	})
}
