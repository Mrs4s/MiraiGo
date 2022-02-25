package utils

import (
	"crypto/md5"
	"errors"
	"io"
)

type doubleReadSeeker struct {
	rs1, rs2       io.ReadSeeker
	rs1len, rs2len int64
	pos            int64
}

func (r *doubleReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var err error
	switch whence {
	case io.SeekStart:
		if offset < r.rs1len {
			r.pos, err = r.rs1.Seek(offset, io.SeekStart)
			return r.pos, err
		} else {
			r.pos, err = r.rs2.Seek(offset-r.rs1len, io.SeekStart)
			r.pos += r.rs1len
			return r.pos, err
		}
	case io.SeekEnd: // negative offset
		return r.Seek(r.rs1len+r.rs2len+offset-1, io.SeekStart)
	default: // io.SeekCurrent
		return r.Seek(r.pos+offset, io.SeekStart)
	}
}

func (r *doubleReadSeeker) Read(p []byte) (n int, err error) {
	switch {
	case r.pos >= r.rs1len: // read only from the second reader
		n, err := r.rs2.Read(p)
		r.pos += int64(n)
		return n, err
	case r.pos+int64(len(p)) <= r.rs1len: // read only from the first reader
		n, err := r.rs1.Read(p)
		r.pos += int64(n)
		return n, err
	default: // read on the border - end of first reader and start of second reader
		n1, err := r.rs1.Read(p)
		r.pos += int64(n1)
		if r.pos != r.rs1len || (err != nil && errors.Is(err, io.EOF)) {
			// Read() might not read all, return
			// If error (but not EOF), return
			return n1, err
		}
		_, err = r.rs2.Seek(0, io.SeekStart)
		if err != nil {
			return n1, err
		}
		n2, err := r.rs2.Read(p[n1:])
		r.pos += int64(n2)
		return n1 + n2, err
	}
}

func ComputeMd5AndLength(r io.Reader) ([]byte, int64) {
	h := md5.New()
	length, _ := io.Copy(h, r)
	fh := h.Sum(nil)
	return fh, length
}

// DoubleReadSeeker combines two io.ReadSeeker into one.
// input two io.ReadSeeker must be at the start.
func DoubleReadSeeker(first, second io.ReadSeeker) io.ReadSeeker {
	rs1Len, _ := first.Seek(0, io.SeekEnd)
	_, _ = first.Seek(0, io.SeekStart) // reset to start
	rs2Len, _ := second.Seek(0, io.SeekEnd)
	_, _ = second.Seek(0, io.SeekStart) // reset to start
	return &doubleReadSeeker{
		rs1:    first,
		rs2:    second,
		rs1len: rs1Len,
		rs2len: rs2Len,
		pos:    0,
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
