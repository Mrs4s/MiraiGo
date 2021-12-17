package codec

import (
	"strconv"

	"github.com/Mrs4s/MiraiGo/binary"
)

type Uni struct {
	Uin         int64
	Seq         uint16
	CommandName string
	EncryptType byte
	SessionID   []byte
	ExtraData   []byte
	Key         []byte
	Body        []byte
}

func (u *Uni) Encode() []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w2 := binary.SelectWriter()
		{ // w.WriteIntLvPacket
			w2.WriteUInt32(0x0B)
			w2.WriteByte(u.EncryptType)
			w2.WriteUInt32(uint32(u.Seq))
			w2.WriteByte(0)
			w2.WriteString(strconv.FormatInt(u.Uin, 10))

			// inline NewWriterF
			w3 := binary.SelectWriter()
			w3.WriteUniPacket(u.CommandName, u.SessionID, u.ExtraData, u.Body)
			w2.EncryptAndWrite(u.Key, w3.Bytes())
			binary.PutWriter(w3)
		}
		data := w2.Bytes()
		w.WriteUInt32(uint32(len(data) + 4))
		w.Write(data)
		binary.PutWriter(w2)
	})
}
