package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T536(loginExtraData []byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x536)
		pos := w.FillUInt16()
		w.Write(loginExtraData)
		w.WriteShortBufLenExcludeSelfAfterPos(pos)
	})
}
