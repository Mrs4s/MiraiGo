package tlv

import (
	"crypto/md5"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T187(macAddress []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x187)
		pos := w.AllocHead16()
		h := md5.Sum(macAddress)
		w.Write(h[:])
		w.WriteHead16ExcludeSelf(pos)
	})
}
