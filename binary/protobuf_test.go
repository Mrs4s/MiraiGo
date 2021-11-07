package binary

import (
	"math"
	"testing"
)

func benchEncoderUvarint(b *testing.B, v uint64) {
	e := encoder{}
	for i := 0; i < b.N; i++ {
		e.Reset()
		e.uvarint(v)
	}
}

func Benchmark_encoder_uvarint(b *testing.B) {
	b.Run("short", func(b *testing.B) {
		benchEncoderUvarint(b, uint64(1))
	})
	b.Run("medium", func(b *testing.B) {
		benchEncoderUvarint(b, uint64(114514))
	})
	b.Run("large", func(b *testing.B) {
		benchEncoderUvarint(b, math.MaxUint64)
	})
}
