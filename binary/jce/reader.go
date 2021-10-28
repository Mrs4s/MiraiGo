package jce

import (
	"bytes"
	goBinary "encoding/binary"
	"io"
	"math"
	"reflect"
)

type JceReader struct {
	buf *bytes.Reader
}

type HeadData struct {
	Type byte
	Tag  int
}

func NewJceReader(data []byte) *JceReader {
	buf := bytes.NewReader(data)
	return &JceReader{buf: buf}
}

func (r *JceReader) readHead() (hd HeadData, l int32) {
	b, _ := r.buf.ReadByte()
	hd.Type = b & 0xF
	hd.Tag = (int(b) & 0xF0) >> 4
	if hd.Tag == 15 {
		b, _ = r.buf.ReadByte()
		hd.Tag = int(b) & 0xFF
		return hd, 2
	}
	return hd, 1
}

func (r *JceReader) peakHead() (h HeadData, l int32) {
	h, l = r.readHead()
	_, _ = r.buf.Seek(int64(-l), io.SeekCurrent)
	return
}

func (r *JceReader) skip(l int) {
	r.readBytes(l)
}

func (r *JceReader) skipField(t byte) {
	switch t {
	case 0:
		r.skip(1)
	case 1:
		r.skip(2)
	case 2, 4:
		r.skip(4)
	case 3, 5:
		r.skip(8)
	case 6:
		b, _ := r.buf.ReadByte()
		r.skip(int(b))
	case 7:
		r.skip(int(r.readInt32()))
	case 8:
		s := r.ReadInt32(0)
		for i := 0; i < int(s)*2; i++ {
			r.skipNextField()
		}
	case 9:
		s := r.ReadInt32(0)
		for i := 0; i < int(s); i++ {
			r.skipNextField()
		}
	case 13:
		r.readHead()
		s := r.ReadInt32(0)
		r.skip(int(s))
	case 10:
		r.skipToStructEnd()
	}
}

func (r *JceReader) skipNextField() {
	hd, _ := r.readHead()
	r.skipField(hd.Type)
}

func (r *JceReader) SkipField(c int) {
	for i := 0; i < c; i++ {
		r.skipNextField()
	}
}

func (r *JceReader) readBytes(len int) []byte {
	if len == 0 {
		return nil
	}
	b := make([]byte, len)
	_, err := r.buf.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func (r *JceReader) readByte() byte {
	b, err := r.buf.ReadByte()
	if err != nil {
		panic(err)
	}
	return b
}

func (r *JceReader) readUInt16() uint16 {
	f, _ := r.buf.ReadByte()
	s, err := r.buf.ReadByte()
	if err != nil {
		panic(err)
	}
	return uint16((int32(f) << 8) + int32(s))
}

func (r *JceReader) readInt32() int32 {
	b := r.readBytes(4)
	return int32(goBinary.BigEndian.Uint32(b))
}

func (r *JceReader) readInt64() int64 {
	b := r.readBytes(8)
	return int64(goBinary.BigEndian.Uint64(b))
}

func (r *JceReader) readFloat32() float32 {
	b := r.readInt32()
	return math.Float32frombits(uint32(b))
}

func (r *JceReader) readFloat64() float64 {
	b := r.readInt64()
	return math.Float64frombits(uint64(b))
}

func (r *JceReader) skipToTag(tag int) bool {
	for {
		hd, l := r.peakHead()
		if tag <= hd.Tag || hd.Type == 11 {
			return tag == hd.Tag
		}
		r.skip(int(l))
		r.skipField(hd.Type)
	}
}

func (r *JceReader) skipToStructEnd() {
	for {
		hd, _ := r.readHead()
		r.skipField(hd.Type)
		if hd.Type == 11 {
			return
		}
	}
}

func (r *JceReader) ReadByte(tag int) byte {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return r.readByte()
	default:
		return 0
	}
}

func (r *JceReader) ReadBool(tag int) bool {
	return r.ReadByte(tag) != 0
}

func (r *JceReader) ReadInt16(tag int) int16 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return int16(r.readByte())
	case 1:
		return int16(r.readUInt16())
	default:
		return 0
	}
}

func (r *JceReader) ReadInt32(tag int) int32 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return int32(r.readByte())
	case 1:
		return int32(r.readUInt16())
	case 2:
		return r.readInt32()
	default:
		return 0
	}
}

func (r *JceReader) ReadInt64(tag int) int64 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return int64(r.readByte())
	case 1:
		return int64(int16(r.readUInt16()))
	case 2:
		return int64(r.readInt32())
	case 3:
		return r.readInt64()
	default:
		return 0
	}
}

func (r *JceReader) ReadFloat32(tag int) float32 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 4:
		return r.readFloat32()
	default:
		return 0
	}
}

func (r *JceReader) ReadFloat64(tag int) float64 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 4:
		return float64(r.readFloat32())
	case 5:
		return r.readFloat64()
	default:
		return 0
	}
}

func (r *JceReader) ReadString(tag int) string {
	if !r.skipToTag(tag) {
		return ""
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 6:
		return string(r.readBytes(int(r.readByte())))
	case 7:
		return string(r.readBytes(int(r.readInt32())))
	default:
		return ""
	}
}

func (r *JceReader) ReadBytes(tag int) []byte {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		b := make([]byte, s)
		for i := 0; i < int(s); i++ {
			b[i] = r.ReadByte(0)
		}
		return b
	case 13:
		r.readHead()
		return r.readBytes(int(r.ReadInt32(0)))
	default:
		return nil
	}
}

// ReadAny Read any type via tag, unsupported JceStruct
func (r *JceReader) ReadAny(tag int) interface{} {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 0:
		return r.readByte()
	case 1:
		return r.readUInt16()
	case 2:
		return r.readInt32()
	case 3:
		return r.readInt64()
	case 4:
		return r.readFloat32()
	case 5:
		return r.readFloat64()
	case 6:
		return string(r.readBytes(int(r.readByte())))
	case 7:
		return string(r.readBytes(int(r.readInt32())))
	case 8:
		s := r.ReadInt32(0)
		m := make(map[interface{}]interface{})
		for i := 0; i < int(s); i++ {
			m[r.ReadAny(0)] = r.ReadAny(1)
		}
		return m
	case 9:
		var sl []interface{}
		s := r.ReadInt32(0)
		for i := 0; i < int(s); i++ {
			sl = append(sl, r.ReadAny(0))
		}
		return sl
	case 12:
		return 0
	case 13:
		r.readHead()
		return r.readBytes(int(r.ReadInt32(0)))
	default:
		return nil
	}
}

func (r *JceReader) ReadJceStruct(obj IJceStruct, tag int) {
	if !r.skipToTag(tag) {
		return
	}
	hd, _ := r.readHead()
	if hd.Type != 10 {
		return
	}
	obj.ReadFrom(r)
	r.skipToStructEnd()
}

func (r *JceReader) ReadMap(i interface{}, tag int) {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Map {
		return
	}
	if !r.skipToTag(tag) {
		return
	}
	t := v.Type()

	kt := t.Key()
	kv := reflect.New(kt)

	vt := t.Elem()
	vv := reflect.New(vt)

	r.readHead()
	s := r.ReadInt32(0)
	for i := 0; i < int(s); i++ {
		r.ReadObject(kv.Interface(), 0)
		r.ReadObject(vv.Interface(), 1)
		v.SetMapIndex(kv.Elem(), vv.Elem())
	}
}

func (r *JceReader) ReadSlice(i interface{}, tag int) {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i).Elem()
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		return
	}
	t = t.Elem()
	if !r.skipToTag(tag) {
		return
	}
	hd, _ := r.readHead()
	if hd.Type == 9 {
		s := r.ReadInt32(0)
		sv := reflect.MakeSlice(t, int(s), int(s))
		t = t.Elem()
		val := reflect.New(t)
		for i := 0; i < int(s); i++ {
			r.ReadObject(val.Interface(), 0)
			sv.Index(i).Set(val.Elem())
		}
		v.Set(sv)
	}
	if hd.Type == 13 && t.Elem().Kind() == reflect.Uint8 {
		r.readHead()
		arr := r.readBytes(int(r.ReadInt32(0)))
		v.SetBytes(arr)
	}
}

func (r *JceReader) ReadObject(i interface{}, tag int) {
	va := reflect.ValueOf(i)
	if va.Kind() != reflect.Ptr || va.IsNil() {
		return
	}
	if ve := va.Elem(); ve.Kind() == reflect.Map {
		ve.Set(reflect.MakeMap(ve.Type()))
		r.ReadMap(ve.Interface(), tag)
		return
	}
	switch o := i.(type) {
	case *byte:
		*o = r.ReadByte(tag)
	case *bool:
		*o = r.ReadBool(tag)
	case *int16:
		*o = r.ReadInt16(tag)
	case *int:
		*o = int(r.ReadInt32(tag))
	case *int32:
		*o = r.ReadInt32(tag)
	case *int64:
		*o = r.ReadInt64(tag)
	case *float32:
		*o = r.ReadFloat32(tag)
	case *float64:
		*o = r.ReadFloat64(tag)
	case *string:
		*o = r.ReadString(tag)
	case *[]byte:
		r.ReadSlice(o, tag)
	case IJceStruct:
		r.readHead()
		o.ReadFrom(r)
		r.skipToStructEnd()
	}
}

func (r *JceReader) ReadAvailable() []byte {
	return r.readBytes(r.buf.Len())
}
