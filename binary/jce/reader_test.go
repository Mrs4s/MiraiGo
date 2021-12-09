package jce

import (
	"math/rand"
	"reflect"
	"testing"

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
	}}

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

func (w *JceWriter) WriteObject(i interface{}, tag byte) {
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
