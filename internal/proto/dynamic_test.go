package proto

import (
	"bytes"
	"math"
	"testing"
)

func benchEncoderUvarint(b *testing.B, v uint64) {
	e := encoder{}
	for i := 0; i < b.N; i++ {
		e.buf = e.buf[:0]
		e.uvarint(v)
	}
}

func benchEncoderSvarint(b *testing.B, v int64) {
	e := encoder{}
	for i := 0; i < b.N; i++ {
		e.buf = e.buf[:0]
		e.svarint(v)
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

func Benchmark_encoder_svarint(b *testing.B) {
	b.Run("short", func(b *testing.B) {
		benchEncoderSvarint(b, int64(1))
	})
	b.Run("medium", func(b *testing.B) {
		benchEncoderSvarint(b, int64(114514))
	})
	b.Run("large", func(b *testing.B) {
		benchEncoderSvarint(b, math.MaxInt64)
	})
}

func TestDynamicMessage_Encode(t *testing.T) {
	input := DynamicMessage{
		1: 2,
		3: 4,
	}
	got := input.Encode()
	expected := []byte{1 << 3, 2, 3 << 3, 4}
	if !bytes.Equal(got, expected) {
		t.Fatalf("expected %v but got %v", expected, got)
	}
}
