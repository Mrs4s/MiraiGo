package packets

import (
	"strconv"

	"github.com/Mrs4s/MiraiGo/binary"
)

func BuildLoginPacket(uin int64, bodyType byte, key, body, extraData []byte) []byte {
	w := binary.NewWriter()
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
	return w.Bytes()
}

func BuildUniPacket(uin int64, seq uint16, commandName string, encryptType byte, sessionId, extraData, key, body []byte) []byte {
	w := binary.NewWriter()
	w.WriteIntLvPacket(4, func(w *binary.Writer) {
		w.WriteUInt32(0x0B)
		w.WriteByte(encryptType)
		w.WriteUInt32(uint32(seq))
		w.WriteByte(0)
		w.WriteString(strconv.FormatInt(uin, 10))
		w.EncryptAndWrite(key, binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUniPacket(commandName, sessionId, extraData, body)
		}))
	})
	return w.Bytes()
}
