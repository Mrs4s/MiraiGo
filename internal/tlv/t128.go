package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T128(isGuidFromFileNull, isGuidAvailable, isGuidChanged bool, guidFlag uint32, buildModel, guid, buildBrand []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x128)
		pos := w.FillUInt16()
		w.WriteUInt16(0)
		w.WriteBool(isGuidFromFileNull)
		w.WriteBool(isGuidAvailable)
		w.WriteBool(isGuidChanged)
		w.WriteUInt32(guidFlag)
		w.WriteTlvLimitedSize(buildModel, 32)
		w.WriteTlvLimitedSize(guid, 16)
		w.WriteTlvLimitedSize(buildBrand, 16)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
