package utils

import (
	"crypto/md5"
	"errors"
	"io"
)

type multiReadSeeker struct {
	readers     []io.ReadSeeker
	multiReader io.Reader
}

func ComputeMd5AndLength(r io.Reader) ([]byte, int64) {
	h := md5.New()
	length, _ := io.Copy(h, r)
	fh := h.Sum(nil)
	return fh, length
}

func (r *multiReadSeeker) Read(p []byte) (int, error) {
	if r.multiReader == nil {
		readers := make([]io.Reader, len(r.readers))
		for i := range r.readers {
			_, _ = r.readers[i].Seek(0, io.SeekStart)
			readers[i] = r.readers[i]
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

// Select 如果A为nil 将会返回 B 否则返回A
// 对应 ?? 语法
func Select(a, b []byte) []byte {
	if a == nil {
		return b
	}
	return a
}
