package jce

import (
	"bytes"
	goBinary "encoding/binary"
	"math"
	"reflect"
	"strconv"
	"sync"
)

type JceWriter struct {
	buf *bytes.Buffer
}

func NewJceWriter() *JceWriter {
	return &JceWriter{buf: new(bytes.Buffer)}
}

func (w *JceWriter) writeHead(t, tag byte) {
	if tag < 0xF {
		w.buf.WriteByte(tag<<4 | t)
	} else {
		w.buf.WriteByte(0xF0 | t)
		w.buf.WriteByte(tag)
	}
}

func (w *JceWriter) WriteByte(b, tag byte) *JceWriter {
	if b == 0 {
		w.writeHead(12, tag)
	} else {
		w.writeHead(0, tag)
		w.buf.WriteByte(b)
	}
	return w
}

func (w *JceWriter) WriteBool(b bool, tag byte) {
	var by byte = 0
	if b {
		by = 1
	}
	w.WriteByte(by, tag)
}

func (w *JceWriter) WriteInt16(n int16, tag byte) {
	switch {
	case n >= -128 && n <= 127:
		w.WriteByte(byte(n), tag)
	default:
		w.putInt16(n, tag)
	}
}

//go:nosplit
func (w *JceWriter) putInt16(n int16, tag byte) {
	w.writeHead(1, tag)
	var buf [2]byte
	goBinary.BigEndian.PutUint16(buf[:], uint16(n))
	w.buf.Write(buf[:])
}

func (w *JceWriter) WriteInt32(n int32, tag byte) *JceWriter {
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
func (w *JceWriter) putInt32(n int32, tag byte) {
	w.writeHead(2, tag)
	var buf [4]byte
	goBinary.BigEndian.PutUint32(buf[:], uint32(n))
	w.buf.Write(buf[:])
}

func (w *JceWriter) WriteInt64(n int64, tag byte) *JceWriter {
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
func (w *JceWriter) putInt64(n int64, tag byte) {
	w.writeHead(3, tag)
	var buf [8]byte
	goBinary.BigEndian.PutUint64(buf[:], uint64(n))
	w.buf.Write(buf[:])
}

//go:nosplit
func (w *JceWriter) WriteFloat32(n float32, tag byte) {
	w.writeHead(4, tag)
	var buf [4]byte
	goBinary.BigEndian.PutUint32(buf[:], math.Float32bits(n))
	w.buf.Write(buf[:])
}

//go:nosplit
func (w *JceWriter) WriteFloat64(n float64, tag byte) {
	w.writeHead(5, tag)
	var buf [8]byte
	goBinary.BigEndian.PutUint64(buf[:], math.Float64bits(n))
	w.buf.Write(buf[:])
}

func (w *JceWriter) WriteString(s string, tag byte) *JceWriter {
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

func (w *JceWriter) WriteBytes(l []byte, tag byte) *JceWriter {
	w.writeHead(13, tag)
	w.buf.WriteByte(0) // w.writeHead(0, 0)
	w.WriteInt32(int32(len(l)), 0)
	w.buf.Write(l)
	return w
}

func (w *JceWriter) WriteInt64Slice(l []int64, tag byte) {
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

func (w *JceWriter) WriteBytesSlice(l [][]byte, tag byte) {
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

func (w *JceWriter) WriteSlice(i interface{}, tag byte) {
	va := reflect.ValueOf(i)
	if va.Kind() != reflect.Slice {
		panic("JceWriter.WriteSlice: not a slice")
	}
	w.writeSlice(va, tag)
}

func (w *JceWriter) writeSlice(slice reflect.Value, tag byte) {
	if slice.Kind() != reflect.Slice {
		return
	}
	w.writeHead(9, tag)
	if slice.Len() == 0 {
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(slice.Len()), 0)
	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		w.writeObject(v, 0)
	}
}

func (w *JceWriter) WriteJceStructSlice(l []IJceStruct, tag byte) {
	w.writeHead(9, tag)
	if len(l) == 0 {
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(len(l)), 0)
	for _, v := range l {
		w.WriteJceStruct(v, 0)
	}
}

func (w *JceWriter) WriteMap(m interface{}, tag byte) {
	va := reflect.ValueOf(m)
	if va.Kind() != reflect.Map {
		panic("JceWriter.WriteMap: not a map")
	}
	w.writeMap(va, tag)
}

func (w *JceWriter) writeMap(m reflect.Value, tag byte) {
	if m.IsNil() {
		w.writeHead(8, tag)
		w.writeHead(12, 0) // w.WriteInt32(0, 0)
		return
	}
	if m.Kind() != reflect.Map {
		return
	}
	w.writeHead(8, tag)
	w.WriteInt32(int32(m.Len()), 0)
	iter := m.MapRange()
	for iter.Next() {
		w.writeObject(iter.Key(), 0)
		w.writeObject(iter.Value(), 1)
	}
}

func (w *JceWriter) writeMapStrStr(m map[string]string, tag byte) {
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

func (w *JceWriter) writeMapStrBytes(m map[string][]byte, tag byte) {
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

func (w *JceWriter) writeMapStrMapStrBytes(m map[string]map[string][]byte, tag byte) {
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

func (w *JceWriter) writeObject(v reflect.Value, tag byte) {
	k := v.Kind()
	if k == reflect.Map {
		switch o := v.Interface().(type) {
		case map[string]string:
			w.writeMapStrStr(o, tag)
		case map[string][]byte:
			w.writeMapStrBytes(o, tag)
		case map[string]map[string][]byte:
			w.writeMapStrMapStrBytes(o, tag)
		default:
			w.writeMap(v, tag)
		}
		return
	}
	if k == reflect.Slice {
		switch o := v.Interface().(type) {
		case []byte:
			w.WriteBytes(o, tag)
		case []IJceStruct:
			w.WriteJceStructSlice(o, tag)
		default:
			w.writeSlice(v, tag)
		}
		return
	}
	switch k {
	case reflect.Uint8, reflect.Int8:
		w.WriteByte(*(*byte)(pointerOf(v)), tag)
	case reflect.Uint16, reflect.Int16:
		w.WriteInt16(*(*int16)(pointerOf(v)), tag)
	case reflect.Uint32, reflect.Int32:
		w.WriteInt32(*(*int32)(pointerOf(v)), tag)
	case reflect.Uint64, reflect.Int64:
		w.WriteInt64(*(*int64)(pointerOf(v)), tag)
	case reflect.String:
		w.WriteString(v.String(), tag)
	default:
		switch o := v.Interface().(type) {
		case IJceStruct:
			w.WriteJceStruct(o, tag)
		case float32:
			w.WriteFloat32(o, tag)
		case float64:
			w.WriteFloat64(o, tag)
		}
	}
}

type decoder struct {
	index int
	id    int
}

var decoderCache = sync.Map{}

// WriteJceStructRaw 写入 Jce 结构体
func (w *JceWriter) WriteJceStructRaw(s interface{}) {
	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Ptr {
		return
	}
	t = t.Elem()
	v := reflect.ValueOf(s).Elem()
	var jceDec []decoder
	dec, ok := decoderCache.Load(t)
	if ok { // 从缓存中加载
		jceDec = dec.([]decoder)
	} else { // 初次反射
		jceDec = make([]decoder, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			strId := field.Tag.Get("jceId")
			if strId == "" {
				continue
			}
			id, err := strconv.Atoi(strId)
			if err != nil {
				continue
			}
			jceDec = append(jceDec, decoder{
				index: i,
				id:    id,
			})
		}
		decoderCache.Store(t, jceDec) // 存入缓存
	}
	for _, dec := range jceDec {
		obj := v.Field(dec.index)
		w.writeObject(obj, byte(dec.id))
	}
}

func (w *JceWriter) WriteJceStruct(s IJceStruct, tag byte) {
	w.writeHead(10, tag)
	w.WriteJceStructRaw(s)
	w.writeHead(11, 0)
}

func (w *JceWriter) Bytes() []byte {
	return w.buf.Bytes()
}
