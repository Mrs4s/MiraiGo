package tlv

import (
	"math/rand"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T1(uin uint32, ip []byte) []byte {
	if len(ip) != 4 {
		panic("invalid ip")
	}
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x01)
		pos := w.FillUInt16()
		w.WriteUInt16(1)
		w.WriteUInt32(rand.Uint32())
		w.WriteUInt32(uin)
		w.WriteUInt32(uint32(time.Now().Unix()))
		w.Write(ip)
		w.WriteUInt16(0)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
