package binary

import (
	"math/rand"
	"testing"
)

func BenchmarkNewWriterF128(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewWriterF(func(w *Writer) {
				w.Write(test)
			})
		}
	})
}

func BenchmarkNewWriterF128_3(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewWriterF(func(w *Writer) {
				w.Write(test)
				w.Write(test)
				w.Write(test)
			})
		}
	})
}

func BenchmarkNewWriterF128_5(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewWriterF(func(w *Writer) {
				w.Write(test)
				w.Write(test)
				w.Write(test)
				w.Write(test)
				w.Write(test)
			})
		}
	})
}

func BenchmarkOpenWriterF128(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, close := OpenWriterF(func(w *Writer) {
				w.Write(test)
			})
			close()
		}
	})
}

func BenchmarkOpenWriterF128_3(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, close := OpenWriterF(func(w *Writer) {
				w.Write(test)
				w.Write(test)
				w.Write(test)
			})
			close()
		}
	})
}

func BenchmarkOpenWriterF128_5(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, close := OpenWriterF(func(w *Writer) {
				w.Write(test)
				w.Write(test)
				w.Write(test)
				w.Write(test)
				w.Write(test)
			})
			close()
		}
	})
}
