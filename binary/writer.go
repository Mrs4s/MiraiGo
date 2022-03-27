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
	b := make([]byte, len(w.Bytes()))
	copy(b, w.Bytes())
	w.put()
	return b
}

// OpenWriterF must call func cl to close
func OpenWriterF(f func(*Writer)) (b []byte, cl func()) {
	w := SelectWriter()
	f(w)
	return w.Bytes(), w.put
}

func (w *Writer) FillUInt16() (pos int) {
	pos = w.Len()
	(*bytes.Buffer)(w).Write([]byte{0, 0})
	return
}

func (w *Writer) WriteUInt16At(pos int, v uint16) {
	newdata := (*bytes.Buffer)(w).Bytes()[pos:]
	binary.BigEndian.PutUint16(newdata, v)
}

func (w *Writer) FillUInt32() (pos int) {
	pos = w.Len()
	(*bytes.Buffer)(w).Write([]byte{0, 0, 0, 0})
	return
}

func (w *Writer) WriteUInt32At(pos int, v uint32) {
	newdata := (*bytes.Buffer)(w).Bytes()[pos:]
	binary.BigEndian.PutUint32(newdata, v)
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
	pos := w.FillUInt32()
	f(w)
	w.WriteUInt32At(pos, uint32(w.Len()+offset-pos-4))
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

func (w *Writer) Len() int {
	return (*bytes.Buffer)(w).Len()
}

func (w *Writer) Bytes() []byte {
	return (*bytes.Buffer)(w).Bytes()
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
