package tlv

import (
	"crypto/md5"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T109(androidId []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x109)
		w.WriteTlv(binary.NewWriterF(func(w *binary.Writer) {
			h := md5.Sum(androidId)
			w.Write(h[:])
		}))
	})
}
