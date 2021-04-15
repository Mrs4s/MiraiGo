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

type gzipWriter struct {
	w   *gzip.Writer
	buf *bytes.Buffer
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		buf := new(bytes.Buffer)
		w := gzip.NewWriter(buf)
		return &gzipWriter{
			w:   w,
			buf: buf,
		}
	},
}

func acquireGzipWriter() *gzipWriter {
	ret := gzipPool.Get().(*gzipWriter)
	ret.buf.Reset()
	ret.w.Reset(ret.buf)
	return ret
}

func releaseGzipWriter(w *gzipWriter) {
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
