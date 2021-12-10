package tlv

import (
	"crypto/md5"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T188(androidId []byte) ([]byte, func()) {
	return binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x188)
		w.WriteBytesShortAndClose(binary.OpenWriterF(func(w *binary.Writer) {
			h := md5.Sum(androidId)
			w.Write(h[:])
		}))
	})
}
