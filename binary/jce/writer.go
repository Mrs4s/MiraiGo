package jce

import (
	"bytes"
	goBinary "encoding/binary"
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	"github.com/modern-go/reflect2"
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
	if n >= -128 && n <= 127 {
		w.WriteByte(byte(n), tag)
		return
	}
	w.writeHead(1, tag)
	_ = goBinary.Write(w.buf, goBinary.BigEndian, n)
}

func (w *JceWriter) WriteInt32(n int32, tag int) *JceWriter {
	if n >= -32768 && n <= 32767 { // ? if ((n >= 32768) && (n <= 32767))
		w.WriteInt16(int16(n), tag)
		return w
	}
	w.writeHead(2, tag)
	_ = goBinary.Write(w.buf, goBinary.BigEndian, n)
	return w
}

func (w *JceWriter) WriteInt64(n int64, tag int) *JceWriter {
	if n >= -2147483648 && n <= 2147483647 {
		return w.WriteInt32(int32(n), tag)
	}
	w.writeHead(3, tag)
	_ = goBinary.Write(w.buf, goBinary.BigEndian, n)
	return w
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
		return
	}
	w.writeHead(9, tag)
	if va.Len() == 0 {
		w.WriteInt32(0, 0)
		return
	}
	w.WriteInt32(int32(va.Len()), 0)
	for i := 0; i < va.Len(); i++ {
		v := va.Index(i)
		w.WriteObject(v.Interface(), 0)
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
	if m == nil {
		w.writeHead(8, tag)
		w.WriteInt32(0, 0)
		return
	}
	va := reflect.ValueOf(m)
	if va.Kind() != reflect.Map {
		return
	}
	w.writeHead(8, tag)
	w.WriteInt32(int32(len(va.MapKeys())), 0)
	for _, k := range va.MapKeys() {
		v := va.MapIndex(k)
		w.WriteObject(k.Interface(), 0)
		w.WriteObject(v.Interface(), 1)
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

type decoder []struct {
	ty     reflect2.Type
	offset uintptr
	id     int
}

var decoderCache = sync.Map{}

// WriteJceStructRaw 写入 Jce 结构体
func (w *JceWriter) WriteJceStructRaw(s IJceStruct) {
	var (
		ty2    = reflect2.TypeOf(s)
		jceDec decoder
	)
	dec, ok := decoderCache.Load(ty2)
	if ok { // 从缓存中加载
		jceDec = dec.(decoder)
	} else { // 初次反射
		jceDec = decoder{}
		t := reflect2.TypeOf(s).(reflect2.PtrType).Elem().(reflect2.StructType)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			strId := field.Tag().Get("jceId")
			if strId == "" {
				continue
			}
			id, err := strconv.Atoi(strId)
			if err != nil {
				continue
			}
			jceDec = append(jceDec, struct {
				ty     reflect2.Type
				offset uintptr
				id     int
			}{ty: field.Type(), offset: field.Offset(), id: id})
		}
		decoderCache.Store(ty2, jceDec) // 存入缓存
	}
	for _, dec := range jceDec {
		var obj = dec.ty.UnsafeIndirect(unsafe.Pointer(uintptr(reflect2.PtrOf(s)) + dec.offset)) // MAGIC!
		if obj != nil {
			w.WriteObject(obj, dec.id)
		}
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
