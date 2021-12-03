package jce

import (
	"math/rand"
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
	w.WriteObject(s, 1)
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
	w.WriteObject(req.Map, 0)
	src := w.Bytes()
	result := RequestDataVersion2{}
	result.ReadFrom(NewJceReader(src))
	assert.Equal(t, req, result)
}

func BenchmarkRequestDataVersion2_ReadFrom(b *testing.B) {
	w := NewJceWriter()
	w.WriteObject(req.Map, 0)
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
