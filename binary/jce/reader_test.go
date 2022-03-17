package jce

import (
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestJceReader_ReadSlice(t *testing.T) {
	s := make([][]byte, 50)
	for i := range s {
		b := make([]byte, 64)
		_, _ = rand.Read(b)
		s[i] = b
	}
	w := NewJceWriter()
	w.WriteBytesSlice(s, 1)
	r := NewJceReader(w.Bytes())
	result := r.ReadByteArrArr(1)
	assert.Equal(t, s, result)
}

var test []*BigDataIPInfo

func BenchmarkJceReader_ReadSlice(b *testing.B) {
	for i := 0; i <= 500; i++ {
		test = append(test, &BigDataIPInfo{
			Type:   1,
			Server: "test1",
			Port:   8080,
		})
	}
	w := NewJceWriter()
	w.WriteObject(test, 1)
	src := w.Bytes()
	b.SetBytes(int64(len(src)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		r := NewJceReader(src)
		_ = r.ReadBigDataIPInfos(1)
	}
}

var req = RequestDataVersion2{
	Map: map[string]map[string][]byte{
		"1": {
			"123": []byte(`123`),
		},
		"2": {
			"123": []byte(`123`),
		},
		"3": {
			"123": []byte(`123`),
		},
		"4": {
			"123": []byte(`123`),
		},
		"5": {
			"123": []byte(`123`),
		},
	},
}

func TestRequestDataVersion2_ReadFrom(t *testing.T) {
	// todo(wdv): fuzz test
	w := NewJceWriter()
	w.writeMapStrMapStrBytes(req.Map, 0)
	src := w.Bytes()
	result := RequestDataVersion2{}
	result.ReadFrom(NewJceReader(src))
	assert.Equal(t, req, result)
}

func BenchmarkRequestDataVersion2_ReadFrom(b *testing.B) {
	w := NewJceWriter()
	w.writeMapStrMapStrBytes(req.Map, 0)
	src := w.Bytes()
	b.SetBytes(int64(len(src)))
	result := &RequestDataVersion2{}
	for i := 0; i < b.N; i++ {
		result.ReadFrom(NewJceReader(src))
	}
}

func TestJceReader_ReadBytes(t *testing.T) {
	b := make([]byte, 1024)
	rand.Read(b)

	w := NewJceWriter()
	w.WriteBytes(b, 0)
	r := NewJceReader(w.Bytes())
	rb := r.ReadBytes(0)

	assert.Equal(t, b, rb)
}

func (w *JceWriter) WriteObject(i any, tag byte) {
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
func (w *JceWriter) WriteJceStructRaw(s any) {
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

func (w *JceWriter) WriteSlice(i any, tag byte) {
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

func (w *JceWriter) WriteMap(m any, tag byte) {
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

type value struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
	flag uintptr
}

func pointerOf(v reflect.Value) unsafe.Pointer {
	return (*value)(unsafe.Pointer(&v)).data
}
