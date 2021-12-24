package packets

import (
	"github.com/Mrs4s/MiraiGo/binary"
)

type IncomingPacket struct {
	SequenceId  uint16
	Flag2       byte
	CommandName string
	SessionId   []byte
	Payload     []byte
}

func BuildCode2DRequestPacket(seq uint32, j uint64, cmd uint16, bodyFunc func(writer *binary.Writer)) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteByte(2)
		pos := w.FillUInt16()
		w.WriteUInt16(cmd)
		w.Write(make([]byte, 21))
		w.WriteByte(3)
		w.WriteUInt16(0)
		w.WriteUInt16(50) // version
		w.WriteUInt32(seq)
		w.WriteUInt64(j)
		bodyFunc(w)
		w.WriteByte(3)
		w.WriteUInt16At(pos, uint16(w.Len()))
	})
}
