package tlv

import "github.com/Mrs4s/MiraiGo/binary"

// temporary solution

func T544(userId uint64, moduleId string, subCmd uint32, sdkVersion string, guid []byte, appVersion string, signer func(uint64, string, string, []byte) ([]byte, error)) []byte {
	salt := binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt64(userId)
		w.WriteBytesShort(guid)
		w.WriteBytesShort([]byte(sdkVersion))
		w.WriteUInt32(subCmd)
	})
	return T544Custom(userId, moduleId, appVersion, salt, signer)
}

func T544v2(userId uint64, moduleId string, subCmd uint32, sdkVersion string, guid []byte, appVersion string, signer func(uint64, string, string, []byte) ([]byte, error)) []byte {
	salt := binary.NewWriterF(func(w *binary.Writer) {
		// w.Write(binary.NewWriterF(func(w *binary.Writer) { w.WriteUInt64(userId) })[:4])
		w.WriteUInt32(0)
		w.WriteBytesShort(guid)
		w.WriteBytesShort([]byte(sdkVersion))
		w.WriteUInt32(subCmd)
		w.WriteUInt32(0)
	})
	return T544Custom(userId, moduleId, appVersion, salt, signer)
}

func T544Custom(userId uint64, moduleId string, appVersion string, salt []byte, signer func(uint64, string, string, []byte) ([]byte, error)) []byte {
	sign, err := signer(userId, moduleId, appVersion, salt)
	if err != nil {
		return nil
	}
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x544)
		w.WriteBytesShort(sign)
	})
}
