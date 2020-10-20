package binary

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	binary2 "encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
)

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
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	_, _ = w.Write(data)
	_ = w.Close()
	return buf.Bytes()
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
	return strings.ToUpper(fmt.Sprintf(
		"{%s}.png", GenUUID(md5),
	))
}

func GenUUID(uuid []byte) string {
	u := uuid[0:16]
	buf := make([]byte, 36)
	hex.Encode(buf[0:], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:16])
	return string(buf)
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
