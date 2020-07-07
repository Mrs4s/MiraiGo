package binary

import (
	"bytes"
	"compress/zlib"
	binary2 "encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

func ZlibUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	io.Copy(&out, r)
	return out.Bytes()
}

func CalculateImageResourceId(md5 []byte) string {
	return strings.ToUpper(fmt.Sprintf(
		"{%s-%s-%s-%s-%s}.png",
		hex.EncodeToString(md5[0:4]), hex.EncodeToString(md5[4:6]), hex.EncodeToString(md5[6:8]),
		hex.EncodeToString(md5[8:10]), hex.EncodeToString(md5[10:]),
	))
}

func ToIPV4Address(arr []byte) string {
	if len(arr) != 4 {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", arr[0], arr[1], arr[2], arr[3])
}

func UInt32ToIPV4Address(i uint32) string {
	addr := make([]byte, 4)
	binary2.LittleEndian.PutUint32(addr, i)
	return ToIPV4Address(addr)
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
