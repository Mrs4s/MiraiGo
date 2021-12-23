package binary

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

// Writer 写入
type Writer bytes.Buffer

func NewWriterF(f func(writer *Writer)) []byte {
	w := SelectWriter()
	f(w)
	b := append([]byte(nil), w.Bytes()...)
	w.put()
	return b
}

// OpenWriterF must call func cl to close
func OpenWriterF(f func(*Writer)) (b []byte, cl func()) {
	w := SelectWriter()
	f(w)
	return w.Bytes(), w.put
}

func (w *Writer) AllocHead16() (pos int) {
	pos = (*bytes.Buffer)(w).Len()
	(*bytes.Buffer)(w).Write([]byte{0, 0})
	return
}

func (w *Writer) WriteHead16(pos int) {
	newdata := (*bytes.Buffer)(w).Bytes()[pos:]
	binary.BigEndian.PutUint16(newdata, uint16(len(newdata)))
}

func (w *Writer) WriteHead16ExcludeSelf(pos int) {
	newdata := (*bytes.Buffer)(w).Bytes()[pos:]
	binary.BigEndian.PutUint16(newdata, uint16(len(newdata)-2))
}

func (w *Writer) AllocHead32() (pos int) {
	pos = (*bytes.Buffer)(w).Len()
	(*bytes.Buffer)(w).Write([]byte{0, 0, 0, 0})
	return
}

func (w *Writer) WriteHead32(pos int) {
	newdata := (*bytes.Buffer)(w).Bytes()[pos:]
	binary.BigEndian.PutUint32(newdata, uint32(len(newdata)))
}

func (w *Writer) Write(b []byte) {
	(*bytes.Buffer)(w).Write(b)
}

func (w *Writer) WriteHex(h string) {
	b, _ := hex.DecodeString(h)
	w.Write(b)
}

func (w *Writer) WriteByte(b byte) {
	(*bytes.Buffer)(w).WriteByte(b)
}

func (w *Writer) WriteUInt16(v uint16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	w.Write(b)
}

func (w *Writer) WriteUInt32(v uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	w.Write(b)
}

func (w *Writer) WriteUInt64(v uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	w.Write(b)
}

func (w *Writer) WriteString(v string) {
	w.WriteUInt32(uint32(len(v) + 4))
	(*bytes.Buffer)(w).WriteString(v)
}

func (w *Writer) WriteStringShort(v string) {
	w.WriteUInt16(uint16(len(v)))
	(*bytes.Buffer)(w).WriteString(v)
}

func (w *Writer) WriteBool(b bool) {
	if b {
		w.WriteByte(0x01)
	} else {
		w.WriteByte(0x00)
	}
}

func (w *Writer) EncryptAndWrite(key []byte, data []byte) {
	w.Write(NewTeaCipher(key).Encrypt(data))
}

func (w *Writer) WriteIntLvPacket(offset int, f func(*Writer)) {
	data, cl := OpenWriterF(f)
	w.WriteUInt32(uint32(len(data) + offset))
	w.Write(data)
	cl()
}

func (w *Writer) WriteUniPacket(commandName string, sessionId, extraData, body []byte) {
	pos := w.AllocHead32()
	// vvv WriteIntLvPacket vvv
	w.WriteString(commandName)
	w.WriteUInt32(8)
	w.Write(sessionId)
	if len(extraData) == 0 {
		w.WriteUInt32(0x04)
	} else {
		w.WriteUInt32(uint32(len(extraData) + 4))
		w.Write(extraData)
	}
	// ^^^ WriteIntLvPacket ^^^
	w.WriteHead32(pos)

	w.WriteUInt32(uint32(len(body) + 4)) // WriteIntLvPacket
	w.Write(body)
}

func (w *Writer) WriteBytesShort(data []byte) {
	w.WriteUInt16(uint16(len(data)))
	w.Write(data)
}

func (w *Writer) WriteTlvLimitedSize(data []byte, limit int) {
	if len(data) <= limit {
		w.WriteBytesShort(data)
		return
	}
	w.WriteBytesShort(data[:limit])
}

func (w *Writer) Bytes() []byte {
	return (*bytes.Buffer)(w).Bytes()
}

func (w *Writer) Cap() int {
	return (*bytes.Buffer)(w).Cap()
}

func (w *Writer) Reset() {
	(*bytes.Buffer)(w).Reset()
}

func (w *Writer) Grow(n int) {
	(*bytes.Buffer)(w).Grow(n)
}

func (w *Writer) put() {
	PutWriter(w)
}
