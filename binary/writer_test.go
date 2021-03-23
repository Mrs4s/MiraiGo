package binary

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func NewWriterFOld(f func(writer *Writer)) []byte {
	w := new(bytes.Buffer)
	f((*Writer)(w))
	return w.Bytes()
}

func TestNewWriterF(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			test := make([]byte, 1024)
			rand.Read(test)
			b1 := NewWriterFOld(func(writer *Writer) {
				writer.Write(test)
				writer.Write(NewWriterFOld(func(writer *Writer) {
					writer.Write(test)
					writer.Write(NewWriterFOld(func(writer *Writer) {
						writer.Write(test)
					}))
				}))
			})

			b2 := NewWriterF(func(writer *Writer) {
				writer.Write(test)
				writer.Write(NewWriterF(func(writer *Writer) {
					writer.Write(test)
					writer.Write(NewWriterF(func(writer *Writer) {
						writer.Write(test)
					}))
				}))
			})

			if !bytes.Equal(b1, b2) {
				fmt.Println(hex.EncodeToString(b1))
				fmt.Println(hex.EncodeToString(b2))
				t.Error("Not equal!!!")
			}
		}()
	}
	wg.Wait()
}

func BenchmarkNewWriterFOld128(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewWriterFOld(func(w *Writer) {
				w.Write(test)
			})
		}
	})
}

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

func BenchmarkNewWriterFOld128_3(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewWriterFOld(func(w *Writer) {
				w.Write(test)
				w.Write(test)
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

func BenchmarkNewWriterFOld128_5(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewWriterFOld(func(w *Writer) {
				w.Write(test)
				w.Write(test)
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
