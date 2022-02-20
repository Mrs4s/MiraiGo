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

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func IsChanClosed(ch interface{}) bool {
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		panic("object is not a channel.")
	}
	return *(*uint32)(
		add(*(*unsafe.Pointer)(add(unsafe.Pointer(&ch), unsafe.Sizeof(uintptr(0)))),
			unsafe.Sizeof(uint(0))*2+unsafe.Sizeof(uintptr(0))+unsafe.Sizeof(uint16(0))),
	) > 0
}

func ComputeMd5AndLength(r io.Reader) ([]byte, int64) {
	h := md5.New()
	length, _ := io.Copy(h, r)
	fh := h.Sum(nil)
	return fh, length
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

type multiReadAt struct {
	first      io.ReadSeeker
	second     io.ReadSeeker
	firstSize  int64
	secondSize int64
}

func (m *multiReadAt) ReadAt(p []byte, off int64) (n int, err error) {
	if m.second == nil { // quick path
		_, _ = m.first.Seek(off, io.SeekStart)
		return m.first.Read(p)
	}
	if off < m.firstSize && off+int64(len(p)) < m.firstSize {
		_, err = m.first.Seek(off, io.SeekStart)
		if err != nil {
			return
		}
		return m.first.Read(p)
	} else if off < m.firstSize && off+int64(len(p)) >= m.firstSize {
		_, _ = m.first.Seek(off, io.SeekStart)
		_, _ = m.second.Seek(0, io.SeekStart)
		n, err = m.first.Read(p[:m.firstSize-off])
		if err != nil {
			return
		}
		n2, err := m.second.Read(p[m.firstSize-off:])
		return n + n2, err
	}
	_, err = m.second.Seek(off-m.firstSize, io.SeekStart)
	if err != nil {
		return
	}
	return m.second.Read(p)
}

func ReaderAtFrom2ReadSeeker(first, second io.ReadSeeker) io.ReaderAt {
	firstSize, _ := first.Seek(0, io.SeekEnd)
	if second == nil {
		return &multiReadAt{
			first:      first,
			firstSize:  firstSize,
			secondSize: 0,
		}
	}
	secondSize, _ := second.Seek(0, io.SeekEnd)
	return &multiReadAt{
		first:      first,
		second:     second,
		firstSize:  firstSize,
		secondSize: secondSize,
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
