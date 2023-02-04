package jce

import (
	"bytes"
	goBinary "encoding/binary"
	"math"
)

type Writer struct {
	buf *bytes.Buffer
}

func NewWriter() *Writer {
	return &Writer{buf: new(bytes.Buffer)}
}

func (w *Writer) writeHead(t, tag byte) {
	if tag < 0xF {
		w.buf.WriteByte(tag<<4 | t)
	} else {
		w.buf.WriteByte(0xF0 | t)
		w.buf.WriteByte(tag)
	}
}

func (w *Writer) WriteByte(b, tag byte) *Writer {
	if b == 0 {
		w.writeHead(12, tag)
	} else {
		w.writeHead(0, tag)
		w.buf.WriteByte(b)
	}
	return w
}

func (w *Writer) WriteBool(b bool, tag byte) {
	var by byte
	if b {
		by = 1
	}
	w.WriteByte(by, tag)
}

func (w *Writer) WriteInt16(n int16, tag byte) {
	switch {
	case n >= -128 && n <= 127:
		w.WriteByte(byte(n), tag)
	default:
		w.putInt16(n, tag)
	}
}

//go:nosplit
func (w *Writer) putInt16(n int16, tag byte) {
	w.writeHead(1, tag)
	var buf [2]byte
	goBinary.BigEndian.PutUint16(buf[:], uint16(n))
	w.buf.Write(buf[:])
}

func (w *Writer) WriteInt32(n int32, tag byte) *Writer {
	switch {
	case n >= -128 && n <= 127:
		w.WriteByte(byte(n), tag)
	case n >= -32768 && n <= 32767:
		w.putInt16(int16(n), tag)
	default:
		w.putInt32(n, tag)
	}
	return w
}

//go:nosplit
func (w *Writer) putInt32(n int32, tag byte) {
	w.writeHead(2, tag)
	var buf [4]byte
	goBinary.BigEndian.PutUint32(buf[:], uint32(n))
	w.buf.Write(buf[:])
}

func (w *Writer) WriteInt64(n int64, tag byte) *Writer {
	switch {
	case n >= -128 && n <= 127:
		w.WriteByte(byte(n), tag)
	case n >= -32768 && n <= 32767:
		w.putInt16(int16(n), tag)
	case n >= -2147483648 && n <= 2147483647:
		w.putInt32(int32(n), tag)
	default:
		w.putInt64(n, tag)
	}
	return w
}

//go:nosplit
func (w *Writer) putInt64(n int64, tag byte) {
	w.writeHead(3, tag)
	var buf [8]byte
	goBinary.BigEndian.PutUint64(buf[:], uint64(n))
	w.buf.Write(buf[:])
}

//go:nosplit
func (w *Writer) WriteFloat32(n float32, tag byte) {
	w.writeHead(4, tag)
	var buf [4]byte
	goBinary.BigEndian.PutUint32(buf[:], math.Float32bits(n))
	w.buf.Write(buf[:])
}

//go:nosplit
func (w *Writer) WriteFloat64(n float64, tag byte) {
	w.writeHead(5, tag)
	var buf [8]byte
	goBinary.BigEndian.PutUint64(buf[:], math.Float64bits(n))
	w.buf.Write(buf[:])
}

func (w *Writer) WriteString(s string, tag byte) *Writer {
	if len(s) > 255 {
		w.writeHead(7, tag)
		var buf [4]byte
		goBinary.BigEndian.PutUint32(buf[:], uint32(len(s)))
		w.buf.Write(buf[:])
		w.buf.WriteString(s)
		return w
	}
	w.writeHead(6, tag)
	w.buf.WriteByte(byte(len(s)))
	w.buf.WriteString(s)
	return w
}

func (w *Writer) WriteBytes(l []byte, tag byte) *Writer {
	w.writeHead(13, tag)
	w.buf.WriteByte(0) // w.writeHead(0, 0)
	w.WriteInt32(int32(len(l)), 0)
	w.buf.Write(l)
	return w
}

func (w *Writer) WriteInt64Slice(l []int64, tag byte) {
	w.writeHead(9, tag)
	if len(l) == 0 {
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(len(l)), 0)
	for _, v := range l {
		w.WriteInt64(v, 0)
	}
}

func (w *Writer) WriteBytesSlice(l [][]byte, tag byte) {
	w.writeHead(9, tag)
	if len(l) == 0 {
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(len(l)), 0)
	for _, v := range l {
		w.WriteBytes(v, 0)
	}
}

func (w *Writer) writeMapStrStr(m map[string]string, tag byte) {
	if m == nil {
		w.writeHead(8, tag)
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	w.writeHead(8, tag)
	w.WriteInt32(int32(len(m)), 0)
	for k, v := range m {
		w.WriteString(k, 0)
		w.WriteString(v, 1)
	}
}

func (w *Writer) writeMapStrBytes(m map[string][]byte, tag byte) {
	if m == nil {
		w.writeHead(8, tag)
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	w.writeHead(8, tag)
	w.WriteInt32(int32(len(m)), 0)
	for k, v := range m {
		w.WriteString(k, 0)
		w.WriteBytes(v, 1)
	}
}

func (w *Writer) writeMapStrMapStrBytes(m map[string]map[string][]byte, tag byte) {
	if m == nil {
		w.writeHead(8, tag)
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	w.writeHead(8, tag)
	w.WriteInt32(int32(len(m)), 0)
	for k, v := range m {
		w.WriteString(k, 0)
		w.writeMapStrBytes(v, 1)
	}
}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}
