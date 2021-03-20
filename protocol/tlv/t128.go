package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T128(isGuidFromFileNull, isGuidAvailable, isGuidChanged bool, guidFlag uint32, buildModel, guid, buildBrand []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x128)
		w.WriteBytesShort(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt16(0)
			w.WriteBool(isGuidFromFileNull)
			w.WriteBool(isGuidAvailable)
			w.WriteBool(isGuidChanged)
			w.WriteUInt32(guidFlag)
			w.WriteTlvLimitedSize(buildModel, 32)
			w.WriteTlvLimitedSize(guid, 16)
			w.WriteTlvLimitedSize(buildBrand, 16)
		}))
	})
}
