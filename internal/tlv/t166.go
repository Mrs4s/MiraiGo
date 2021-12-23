package tlv

import "github.com/Mrs4s/MiraiGo/binary"

func T166(imageType byte) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x166)
		pos := w.FillUInt16()
		w.WriteByte(imageType)
		w.WriteShortBufLenExcludeSelfAfterPos(pos)
	})
}
