package utils

import (
	"encoding/xml"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func RandomString(len int) string {
	return RandomStringRange(len, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
}

func RandomStringRange(length int, str string) string {
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(str[rand.Intn(len(str))])
	}
	return sb.String()
}

func ChunkString(s string, chunkSize int) []string {
	runes := []rune(s)
	if len(runes) == 0 || len(runes) <= chunkSize {
		return []string{s}
	}

	chunkLen := len(runes) / chunkSize
	if len(runes)%chunkSize != 0 {
		chunkLen++
	}

	chunks := make([]string, 0, chunkLen)
	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

func ConvertSubVersionToInt(str string) int32 {
	i, _ := strconv.ParseInt(strings.Join(strings.Split(str, "."), ""), 10, 64)
	return int32(i) * 10
}

// B2S converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
// from github.com/savsgio/gotils/strconv
func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// S2B converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func S2B(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return
}

// XmlEscape xml escape string
func XmlEscape(c string) string {
	buf := new(strings.Builder)
	_ = xml.EscapeText(buf, []byte(c))
	return buf.String()
}
