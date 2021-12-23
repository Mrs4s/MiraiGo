package tlv

import (
	"crypto/md5"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T109(androidId []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x109)
		pos := w.FillUInt16()
		h := md5.Sum(androidId)
		w.Write(h[:])
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
