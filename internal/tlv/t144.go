package tlv

import (
	"github.com/Mrs4s/MiraiGo/binary"
)

func T144(
	imei, devInfo, osType, osVersion, simInfo, apn []byte,
	isGuidFromFileNull, isGuidAvailable, isGuidChanged bool,
	guidFlag uint32,
	buildModel, guid, buildBrand, tgtgtKey []byte,
) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x144)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			w.EncryptAndWrite(tgtgtKey, binary.NewWriterF(func(w *binary.Writer) {
				w.WriteUInt16(5)
				w.WriteAndClose(T109(imei))
				w.WriteAndClose(T52D(devInfo))
				w.WriteAndClose(T124(osType, osVersion, simInfo, apn))
				w.WriteAndClose(T128(isGuidFromFileNull, isGuidAvailable, isGuidChanged, guidFlag, buildModel, guid, buildBrand))
				w.WriteAndClose(T16E(buildModel))
			}))
		}))
	})
}
