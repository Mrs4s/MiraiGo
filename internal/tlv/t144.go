package tlv

import (
	"github.com/Mrs4s/MiraiGo/binary"
)

func T144(
	imei, devInfo, osType, osVersion, simInfo, apn []byte,
	isGuidFromFileNull, isGuidAvailable, isGuidChanged bool,
	guidFlag uint32,
	buildModel, guid, buildBrand, tgtgtKey []byte,
) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x144)
		pos := w.FillUInt16()
		w.EncryptAndWrite(tgtgtKey, binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt16(5)
			w.Write(T109(imei))
			w.Write(T52D(devInfo))
			w.Write(T124(osType, osVersion, simInfo, apn))
			w.Write(T128(isGuidFromFileNull, isGuidAvailable, isGuidChanged, guidFlag, buildModel, guid, buildBrand))
			w.Write(T16E(buildModel))
		}))
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
