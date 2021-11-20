package jce

import (
	"bytes"
	goBinary "encoding/binary"
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

func (w *JceWriter) writeHead(t byte, tag int) {
	if tag < 15 {
		b := byte(tag<<4) | t
		w.buf.WriteByte(b)
	} else if tag < 256 {
		b := 0xF0 | t
		w.buf.WriteByte(b)
		w.buf.WriteByte(byte(tag))
	}
}

func (w *JceWriter) WriteByte(b byte, tag int) *JceWriter {
	if b == 0 {
		w.writeHead(12, tag)
	} else {
		w.writeHead(0, tag)
		w.buf.WriteByte(b)
	}
	return w
}

func (w *JceWriter) WriteBool(b bool, tag int) {
	var by byte = 0
	if b {
		by = 1
	}
	w.WriteByte(by, tag)
}

func (w *JceWriter) WriteInt16(n int16, tag int) {
	switch {
	case n >= -128 && n <= 127:
		w.WriteByte(byte(n), tag)
	default:
		w.putInt16(n, tag)
	}
}

//go:nosplit
func (w *JceWriter) putInt16(n int16, tag int) {
	w.writeHead(1, tag)
	var buf [2]byte
	goBinary.BigEndian.PutUint16(buf[:], uint16(n))
	w.buf.Write(buf[:])
}

func (w *JceWriter) WriteInt32(n int32, tag int) *JceWriter {
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
func (w *JceWriter) putInt32(n int32, tag int) {
	w.writeHead(2, tag)
	var buf [4]byte
	goBinary.BigEndian.PutUint32(buf[:], uint32(n))
	w.buf.Write(buf[:])
}

func (w *JceWriter) WriteInt64(n int64, tag int) *JceWriter {
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
func (w *JceWriter) putInt64(n int64, tag int) {
	w.writeHead(3, tag)
	var buf [8]byte
	goBinary.BigEndian.PutUint64(buf[:], uint64(n))
	w.buf.Write(buf[:])
}

func (w *JceWriter) WriteFloat32(n float32, tag int) {
	w.writeHead(4, tag)
	_ = goBinary.Write(w.buf, goBinary.BigEndian, n)
}

func (w *JceWriter) WriteFloat64(n float64, tag int) {
	w.writeHead(5, tag)
	_ = goBinary.Write(w.buf, goBinary.BigEndian, n)
}

func (w *JceWriter) WriteString(s string, tag int) *JceWriter {
	by := []byte(s)
	if len(by) > 255 {
		w.writeHead(7, tag)
		_ = goBinary.Write(w.buf, goBinary.BigEndian, int32(len(by)))
		w.buf.Write(by)
		return w
	}
	w.writeHead(6, tag)
	w.buf.WriteByte(byte(len(by)))
	w.buf.Write(by)
	return w
}

func (w *JceWriter) WriteBytes(l []byte, tag int) *JceWriter {
	w.writeHead(13, tag)
	w.writeHead(0, 0)
	w.WriteInt32(int32(len(l)), 0)
	w.buf.Write(l)
	return w
}

func (w *JceWriter) WriteInt64Slice(l []int64, tag int) {
	w.writeHead(9, tag)
	if len(l) == 0 {
		w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(len(l)), 0)
	for _, v := range l {
		w.WriteInt64(v, 0)
	}
}

func (w *JceWriter) WriteSlice(i interface{}, tag int) {
	va := reflect.ValueOf(i)
	if va.Kind() != reflect.Slice {
		panic("JceWriter.WriteSlice: not a slice")
	}
	w.writeSlice(va, tag)
}

func (w *JceWriter) writeSlice(slice reflect.Value, tag int) {
	if slice.Kind() != reflect.Slice {
		return
	}
	w.writeHead(9, tag)
	if slice.Len() == 0 {
		w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(slice.Len()), 0)
	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		w.writeObject(v, 0)
	}
}

func (w *JceWriter) WriteJceStructSlice(l []IJceStruct, tag int) {
	w.writeHead(9, tag)
	if len(l) == 0 {
		w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(len(l)), 0)
	for _, v := range l {
		w.WriteJceStruct(v, 0)
	}
}

func (w *JceWriter) WriteMap(m interface{}, tag int) {
	va := reflect.ValueOf(m)
	if va.Kind() != reflect.Map {
		panic("JceWriter.WriteMap: not a map")
	}
	w.writeMap(va, tag)
}

func (w *JceWriter) writeMap(m reflect.Value, tag int) {
	if m.IsNil() {
		w.writeHead(8, tag)
		w.WriteInt32(0, 0)
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

func (w *JceWriter) WriteObject(i interface{}, tag int) {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Map {
		w.WriteMap(i, tag)
		return
	}
	if t.Kind() == reflect.Slice {
		if b, ok := i.([]byte); ok {
			w.WriteBytes(b, tag)
			return
		}
		w.WriteSlice(i, tag)
		return
	}
	switch o := i.(type) {
	case byte:
		w.WriteByte(o, tag)
	case bool:
		w.WriteBool(o, tag)
	case int16:
		w.WriteInt16(o, tag)
	case int32:
		w.WriteInt32(o, tag)
	case int64:
		w.WriteInt64(o, tag)
	case float32:
		w.WriteFloat32(o, tag)
	case float64:
		w.WriteFloat64(o, tag)
	case string:
		w.WriteString(o, tag)
	case IJceStruct:
		w.WriteJceStruct(o, tag)
	}
}

func (w *JceWriter) writeObject(v reflect.Value, tag int) {
	k := v.Kind()
	if k == reflect.Map {
		w.writeMap(v, tag)
		return
	}
	if k == reflect.Slice {
		if v.Type().Elem().Kind() == reflect.Uint8 {
			w.WriteBytes(v.Bytes(), tag)
			return
		}
		w.writeSlice(v, tag)
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
	reflect.ValueOf(s).Interface()
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
		w.writeObject(obj, dec.id)
	}
}

func (w *JceWriter) WriteJceStruct(s IJceStruct, tag int) {
	w.writeHead(10, tag)
	w.WriteJceStructRaw(s)
	w.writeHead(11, 0)
}

func (w *JceWriter) Bytes() []byte {
	return w.buf.Bytes()
}
