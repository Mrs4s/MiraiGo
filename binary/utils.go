package binary

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	binary2 "encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"sync"

	"github.com/Mrs4s/MiraiGo/utils"
)

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
	const maxSize = 1 << 15
	if w.buf.Cap() < maxSize {
		w.buf.Reset()
		gzipPool.Put(w)
	}
}

func ZlibUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	defer r.Close()
	io.Copy(&out, r)
	return out.Bytes()
}

func ZlibCompress(data []byte) []byte {
	buf := new(bytes.Buffer)
	w := zlib.NewWriter(buf)
	_, _ = w.Write(data)
	w.Close()
	return buf.Bytes()
}

func GZipCompress(data []byte) []byte {
	gw := acquireGzipWriter()
	_, _ = gw.w.Write(data)
	_ = gw.w.Close()
	ret := append([]byte(nil), gw.buf.Bytes()...)
	releaseGzipWriter(gw)
	return ret
}

func GZipUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := gzip.NewReader(b)
	defer r.Close()
	_, _ = io.Copy(&out, r)
	return out.Bytes()
}

func CalculateImageResourceId(md5 []byte) string {
	id := make([]byte, 36+6)[:0]
	id = append(id, '{')
	AppendUUID(id[1:], md5)
	id = id[:37]
	id = append(id, "}.png"...)
	return utils.B2S(bytes.ToUpper(id))

}

func GenUUID(uuid []byte) []byte {
	return AppendUUID(nil, uuid)
}

func AppendUUID(dst []byte, uuid []byte) []byte {
	_ = uuid[15]
	if cap(dst) > 36 {
		dst = dst[:36]
		dst[8] = '-'
		dst[13] = '-'
		dst[18] = '-'
		dst[23] = '-'
	} else { // Need Grow
		dst = append(dst, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"...)
	}
	hex.Encode(dst[0:], uuid[0:4])
	hex.Encode(dst[9:], uuid[4:6])
	hex.Encode(dst[14:], uuid[6:8])
	hex.Encode(dst[19:], uuid[8:10])
	hex.Encode(dst[24:], uuid[10:16])
	return dst
}

func ToIPV4Address(arr []byte) string {
	ip := (net.IP)(arr)
	return ip.String()
}

func UInt32ToIPV4Address(i uint32) string {
	ip := net.IP{0, 0, 0, 0}
	binary2.LittleEndian.PutUint32(ip, i)
	return ip.String()
}

func ToChunkedBytesF(b []byte, size int, f func([]byte)) {
	r := NewReader(b)
	for r.Len() >= size {
		f(r.ReadBytes(size))
	}
	if r.Len() > 0 {
		f(r.ReadAvailable())
	}
}

func ToBytes(i interface{}) []byte {
	return NewWriterF(func(w *Writer) {
		// TODO: more types
		switch t := i.(type) {
		case int16:
			w.WriteUInt16(uint16(t))
		case int32:
			w.WriteUInt32(uint32(t))
		}
	})
}
