package binary

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	binary2 "encoding/binary"
	"encoding/hex"
	"net"

	"github.com/Mrs4s/MiraiGo/utils"
)

type GzipWriter struct {
	w   *gzip.Writer
	buf *bytes.Buffer
}

func (w *GzipWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *GzipWriter) Close() error {
	return w.w.Close()
}

func (w *GzipWriter) Bytes() []byte {
	return w.buf.Bytes()
}

func ZlibUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	defer r.Close()
	_, _ = out.ReadFrom(r)
	return out.Bytes()
}

func ZlibCompress(data []byte) []byte {
	zw := acquireZlibWriter()
	_, _ = zw.w.Write(data)
	_ = zw.w.Close()
	ret := make([]byte, len(zw.buf.Bytes()))
	copy(ret, zw.buf.Bytes())
	releaseZlibWriter(zw)
	return ret
}

func GZipCompress(data []byte) []byte {
	gw := AcquireGzipWriter()
	_, _ = gw.Write(data)
	_ = gw.Close()
	ret := make([]byte, len(gw.buf.Bytes()))
	copy(ret, gw.buf.Bytes())
	ReleaseGzipWriter(gw)
	return ret
}

func GZipUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := gzip.NewReader(b)
	defer r.Close()
	_, _ = out.ReadFrom(r)
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

func ToBytes(i any) []byte {
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
