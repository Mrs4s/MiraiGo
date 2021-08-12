package jce

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJceReader_ReadSlice(t *testing.T) {
	s := make([]int64, 50)
	for i := range s {
		s[i] = rand.Int63()
	}
	w := NewJceWriter()
	w.WriteObject(s, 1)
	r := NewJceReader(w.Bytes())
	var result []int64
	r.ReadSlice(&result, 1)
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
	var result = []BigDataIPInfo{}
	for i := 0; i < b.N; i++ {
		r := NewJceReader(src)
		r.ReadSlice(&result, 1)
	}
}
