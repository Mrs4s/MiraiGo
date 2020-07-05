package binary

import (
	"bytes"
	"encoding/binary"
)

type Writer struct {
	buf *bytes.Buffer
}

func NewWriter() *Writer {
	return &Writer{buf: new(bytes.Buffer)}
}

func NewWriterF(f func(writer *Writer)) []byte {
	w := NewWriter()
	f(w)
	return w.Bytes()
}

func (w *Writer) Write(b []byte) {
	w.buf.Write(b)
}

func (w *Writer) WriteByte(b byte) {
	w.buf.WriteByte(b)
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
	payload := []byte(v)
	w.WriteUInt32(uint32(len(payload) + 4))
	w.Write(payload)
}

func (w *Writer) WriteStringShort(v string) {
	w.WriteTlv([]byte(v))
}

func (w *Writer) WriteBool(b bool) {
	if b {
		w.WriteByte(0x01)
	} else {
		w.WriteByte(0x00)
	}
}

func (w *Writer) EncryptAndWrite(key []byte, data []byte) {
	tea := NewTeaCipher(key)
	ed := tea.Encrypt(data)
	w.Write(ed)
}

func (w *Writer) WriteIntLvPacket(offset int, f func(writer *Writer)) {
	t := NewWriter()
	f(t)
	data := t.Bytes()
	w.WriteUInt32(uint32(len(data) + offset))
	w.Write(data)
}

func (w *Writer) WriteUniPacket(commandName string, sessionId, extraData, body []byte) {
	w.WriteIntLvPacket(4, func(w *Writer) {
		w.WriteString(commandName)
		w.WriteUInt32(8)
		w.Write(sessionId)
		if len(extraData) == 0 {
			w.WriteUInt32(0x04)
		} else {
			w.WriteUInt32(uint32(len(extraData) + 4))
			w.Write(extraData)
		}
	})
	w.WriteIntLvPacket(4, func(w *Writer) {
		w.Write(body)
	})
}

func (w *Writer) WriteTlv(data []byte) {
	w.WriteUInt16(uint16(len(data)))
	w.Write(data)
}

func (w *Writer) WriteTlvLimitedSize(data []byte, limit int) {
	if len(data) <= limit {
		w.WriteTlv(data)
		return
	}
	w.WriteTlv(data[:limit])
}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}
