package tlv

import "github.com/Mrs4s/MiraiGo/binary"

// temporary solution

func T544(userId uint64, moduleId string, subCmd uint32, sdkVersion string, guid []byte, signer func(uint64, string, []byte) ([]byte, error)) []byte {
	salt := binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt64(userId)
		w.WriteBytesShort(guid)
		w.WriteBytesShort([]byte(sdkVersion))
		w.WriteUInt32(subCmd)
	})
	sign, err := signer(userId, moduleId, salt)
	if err != nil {
		return nil
	}
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x544)
		w.WriteBytesShort(sign)
	})
}

func T544Custom(userId uint64, moduleId string, salt []byte, signer func(uint64, string, []byte) ([]byte, error)) []byte {
	sign, err := signer(userId, moduleId, salt)
	if err != nil {
		return nil
	}
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x544)
		w.WriteBytesShort(sign)
	})
}
