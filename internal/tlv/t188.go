package tlv

import (
	"crypto/md5"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T188(androidId []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x188)
		pos := w.AllocHead16()
		h := md5.Sum(androidId)
		w.Write(h[:])
		w.WriteHead16ExcludeSelf(pos)
	})
}
