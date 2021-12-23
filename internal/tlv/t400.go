package tlv

import (
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T400(g []byte, uin int64, guid, dpwd []byte, j2, j3 int64, randSeed []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x400)
		pos := w.FillUInt16()
		w.EncryptAndWrite(g, binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt16(1) // version
			w.WriteUInt64(uint64(uin))
			w.Write(guid)
			w.Write(dpwd)
			w.WriteUInt32(uint32(j2))
			w.WriteUInt32(uint32(j3))
			w.WriteUInt32(uint32(time.Now().Unix()))
			w.Write(randSeed)
		}))
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
