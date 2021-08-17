package jce

import "testing"

var globalBytes []byte

func BenchmarkJceWriter_WriteMap(b *testing.B) {
	var x = globalBytes
	for i := 0; i < b.N; i++ {
		w := NewJceWriter()
		w.WriteMap(req.Map, 0)
		x = w.Bytes()
	}
	globalBytes = x
	b.SetBytes(int64(len(globalBytes)))
}
