package binary

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(Writer)
	},
}

// NewWriter 从池中取出一个 Writer
func NewWriter() *Writer {
	w := bufferPool.Get().(*Writer)
	if w == nil {
		return new(Writer)
	}
	return w
}

// PutBuffer 将 Writer 放回池中
func PutBuffer(w *Writer) {
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16
	if w.Cap() < maxSize { // 对于大Buffer直接丢弃
		w.Reset()
		bufferPool.Put(w)
	}
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		buf := new(bytes.Buffer)
		w := gzip.NewWriter(buf)
		return &GzipWriter{
			w:   w,
			buf: buf,
		}
	},
}

func AcquireGzipWriter() *GzipWriter {
	ret := gzipPool.Get().(*GzipWriter)
	ret.buf.Reset()
	ret.w.Reset(ret.buf)
	return ret
}

func ReleaseGzipWriter(w *GzipWriter) {
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16
	if w.buf.Cap() < maxSize {
		w.buf.Reset()
		gzipPool.Put(w)
	}
}

type zlibWriter struct {
	w   *zlib.Writer
	buf *bytes.Buffer
}

var zlibPool = sync.Pool{
	New: func() interface{} {
		buf := new(bytes.Buffer)
		w := zlib.NewWriter(buf)
		return &zlibWriter{
			w:   w,
			buf: buf,
		}
	},
}

func acquireZlibWriter() *zlibWriter {
	ret := zlibPool.Get().(*zlibWriter)
	ret.buf.Reset()
	ret.w.Reset(ret.buf)
	return ret
}

func releaseZlibWriter(w *zlibWriter) {
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16
	if w.buf.Cap() < maxSize {
		w.buf.Reset()
		zlibPool.Put(w)
	}
}

const size256k = 256 * 1024

var b256kPool = sync.Pool{
	New: func() interface{} {
		return make128kSlicePointer()
	},
}

// Get256KBytes 获取一个128k大小 []byte
func Get256KBytes() *[]byte {
	buf := b256kPool.Get().(*[]byte)
	if buf == nil {
		return make128kSlicePointer()
	}
	if cap(*buf) < size256k {
		return make128kSlicePointer()
	}
	*buf = (*buf)[:size256k]
	return buf
}

// Put256KBytes 放回一个128k大小 []byte
func Put256KBytes(b *[]byte) {
	if cap(*b) < size256k || cap(*b) > 2*size256k { // 太大或太小的 []byte 不要放入
		return
	}
	*b = (*b)[:cap(*b)]
	b256kPool.Put(b)
}

func make128kSlicePointer() *[]byte {
	data := make([]byte, size256k)
	return &data
}
