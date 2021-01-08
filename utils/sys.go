package utils

import (
	"crypto/md5"
	"errors"
	"io"
	"reflect"
	"unsafe"
)

type multiReadSeeker struct {
	readers     []io.ReadSeeker
	multiReader io.Reader
}

func IsChanClosed(ch interface{}) bool {
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		panic("object is not a channel.")
	}
	ptr := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&ch)) + unsafe.Sizeof(uint(0))))
	ptr += unsafe.Sizeof(uint(0)) * 2
	ptr += unsafe.Sizeof(uintptr(0))
	ptr += unsafe.Sizeof(uint16(0))
	return *(*uint32)(unsafe.Pointer(ptr)) > 0
}

func ComputeMd5AndLength(r io.Reader) ([]byte, int64) {
	h := md5.New()
	length, _ := io.Copy(h, r)
	fh := h.Sum(nil)
	return fh[:], length
}

func (r *multiReadSeeker) Read(p []byte) (int, error) {
	if r.multiReader == nil {
		var readers []io.Reader
		for i := range r.readers {
			_, _ = r.readers[i].Seek(0, io.SeekStart)
			readers = append(readers, r.readers[i])
		}
		r.multiReader = io.MultiReader(readers...)
	}
	return r.multiReader.Read(p)
}

func (r *multiReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence != 0 || offset != 0 {
		return -1, errors.New("unsupported offset")
	}
	r.multiReader = nil
	return 0, nil
}

func MultiReadSeeker(r ...io.ReadSeeker) io.ReadSeeker {
	return &multiReadSeeker{
		readers: r,
	}
}
