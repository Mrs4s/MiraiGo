package utils

import (
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
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

const (
	escQuot = "&#34;" // shorter than "&quot;"
	escApos = "&#39;" // shorter than "&apos;"
	escAmp  = "&amp;"
	escLT   = "&lt;"
	escGT   = "&gt;"
	escTab  = "&#x9;"
	escNL   = "&#xA;"
	escCR   = "&#xD;"
	escFFFD = "\uFFFD" // Unicode replacement character
)

func isInCharacterRange(r rune) (inrange bool) {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xD7FF ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}

// XmlEscape xml escape string
func XmlEscape(s string) string {
	var esc string
	var sb strings.Builder
	sb.Grow(len(s))
	last := 0
	for i, r := range s {
		width := utf8.RuneLen(r)
		switch r {
		case '"':
			esc = escQuot
		case '\'':
			esc = escApos
		case '&':
			esc = escAmp
		case '<':
			esc = escLT
		case '>':
			esc = escGT
		case '\t':
			esc = escTab
		case '\n':
			esc = escNL
		case '\r':
			esc = escCR
		default:
			if !isInCharacterRange(r) || (r == 0xFFFD && width == 1) {
				esc = escFFFD
				break
			}
			continue
		}
		sb.WriteString(s[last:i])
		sb.WriteString(esc)
		last = i + width
	}
	sb.WriteString(s[last:])
	return sb.String()
}
