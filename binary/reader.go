package binary

import (
	"bytes"
	"io"
	"net"
)

type Reader struct {
	buf *bytes.Reader
}

type NetworkReader struct {
	conn net.Conn
}

type TlvMap map[uint16][]byte

// --- ByteStream reader ---

func NewReader(data []byte) *Reader {
	buf := bytes.NewReader(data)
	return &Reader{
		buf: buf,
	}
}

func (r *Reader) ReadByte() byte {
	b, err := r.buf.ReadByte()
	if err != nil {
		panic(err)
	}
	return b
}

func (r *Reader) ReadBytes(len int) []byte {
	b := make([]byte, len)
	if len > 0 {
		_, err := r.buf.Read(b)
		if err != nil {
			panic(err)
		}
	}
	return b
}

func (r *Reader) ReadBytesShort() []byte {
	return r.ReadBytes(int(r.ReadUInt16()))
}

func (r *Reader) ReadUInt16() uint16 {
	f, _ := r.buf.ReadByte()
	s, err := r.buf.ReadByte()
	if err != nil {
		panic(err)
	}
	return uint16((int32(f) << 8) + int32(s))
}

func (r *Reader) ReadInt32() int32 {
	b := r.ReadBytes(4)
	return (int32(b[0]) << 24) | (int32(b[1]) << 16) | (int32(b[2]) << 8) | int32(b[3])
}

func (r *Reader) ReadString() string {
	data := r.ReadBytes(int(r.ReadInt32() - 4))
	return string(data)
}

func (r *Reader) ReadStringShort() string {
	data := r.ReadBytes(int(r.ReadUInt16()))
	return string(data)
}

func (r *Reader) ReadStringLimit(limit int) string {
	data := r.ReadBytes(limit)
	return string(data)
}

func (r *Reader) ReadAvailable() []byte {
	return r.ReadBytes(r.buf.Len())
}

func (r *Reader) ReadTlvMap(tagSize int) (m TlvMap) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: error
		}
	}()
	m = make(map[uint16][]byte)
	for {
		if r.Len() < tagSize {
			return m
		}
		var k uint16
		if tagSize == 1 {
			k = uint16(r.ReadByte())
		} else if tagSize == 2 {
			k = r.ReadUInt16()
		} else if tagSize == 4 {
			k = uint16(r.ReadInt32())
		}
		if k == 255 {
			return m
		}
		m[k] = r.ReadBytes(int(r.ReadUInt16()))
	}
}

func (r *Reader) Len() int {
	return r.buf.Len()
}

func (tlv TlvMap) Exists(key uint16) bool {
	if _, ok := tlv[key]; ok {
		return true
	}
	return false
}

// --- Network reader ---

func NewNetworkReader(conn net.Conn) *NetworkReader {
	return &NetworkReader{conn: conn}
}

func (r *NetworkReader) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	n, err := r.conn.Read(buf)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return r.ReadByte()
	}
	return buf[0], nil
}

func (r *NetworkReader) ReadBytes(len int) ([]byte, error) {
	buf := make([]byte, len)
	_, err := io.ReadFull(r.conn, buf)
	//for i := 0; i < len; i++ {
	//	b, err := r.ReadByte()
	//	if err != nil {
	//		return nil, err
	//	}
	//	buf[i] = b
	//}
	return buf, err
}

func (r *NetworkReader) ReadInt32() (int32, error) {
	b, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return (int32(b[0]) << 24) | (int32(b[1]) << 16) | (int32(b[2]) << 8) | int32(b[3]), nil
}
