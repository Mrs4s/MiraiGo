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

func BuildUniPacket(uin int64, seq uint16, commandName string, encryptType byte, sessionID, extraData, key, body []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w2 := binary.NewWriter()
		{ // w.WriteIntLvPacket
			w2.WriteUInt32(0x0B)
			w2.WriteByte(encryptType)
			w2.WriteUInt32(uint32(seq))
			w2.WriteByte(0)
			w2.WriteString(strconv.FormatInt(uin, 10))

			// inline NewWriterF
			w3 := binary.NewWriter()
			w3.WriteUniPacket(commandName, sessionID, extraData, body)
			w2.EncryptAndWrite(key, w3.Bytes())
			binary.PutBuffer(w3)
		}
		data := w2.Bytes()
		w.WriteUInt32(uint32(len(data) + 4))
		w.Write(data)
		binary.PutBuffer(w2)
	})
}
