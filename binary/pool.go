package binary

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(Writer)
	},
}

// SelectWriter 从池中取出一个 Writer
func SelectWriter() *Writer {
	// 因为 bufferPool 定义有 New 函数
	// 所以 bufferPool.Get() 永不为 nil
	// 不用判空
	return bufferPool.Get().(*Writer)
}

// PutWriter 将 Writer 放回池中
func PutWriter(w *Writer) {
	// See https://golang.org/issue/23199
	const maxSize = 32 * 1024
	if (*bytes.Buffer)(w).Cap() < maxSize { // 对于大Buffer直接丢弃
		w.Reset()
		bufferPool.Put(w)
	}
}

var gzipPool = sync.Pool{
	New: func() any {
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
	New: func() any {
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
