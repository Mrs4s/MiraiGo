package packets

import (
	"strconv"

	"github.com/Mrs4s/MiraiGo/binary"
)

func BuildLoginPacket(uin int64, bodyType byte, key, body, extraData []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteIntLvPacket(4, func(w *binary.Writer) {
			w.WriteUInt32(0x00_00_00_0A)
			w.WriteByte(bodyType)
			w.WriteIntLvPacket(4, func(w *binary.Writer) {
				w.Write(extraData)
			})
			w.WriteByte(0x00)
			w.WriteString(strconv.FormatInt(uin, 10))
			if len(key) == 0 {
				w.Write(body)
			} else {
				w.EncryptAndWrite(key, body)
			}
		})
	})
}
