package binary

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	"github.com/Mrs4s/MiraiGo/utils"
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
	b := make([]byte, 2)
	_, _ = r.buf.Read(b)
	return binary.BigEndian.Uint16(b)
}

func (r *Reader) ReadInt32() int32 {
	b := make([]byte, 4)
	_, _ = r.buf.Read(b)
	return int32(binary.BigEndian.Uint32(b))
}

func (r *Reader) ReadInt64() int64 {
	b := make([]byte, 8)
	_, _ = r.buf.Read(b)
	return int64(binary.BigEndian.Uint64(b))
}

func (r *Reader) ReadString() string {
	data := r.ReadBytes(int(r.ReadInt32() - 4))
	return utils.B2S(data)
}

func (r *Reader) ReadInt32Bytes() []byte {
	return r.ReadBytes(int(r.ReadInt32() - 4))
}

func (r *Reader) ReadStringShort() string {
	data := r.ReadBytes(int(r.ReadUInt16()))
	return utils.B2S(data)
}

func (r *Reader) ReadStringLimit(limit int) string {
	data := r.ReadBytes(limit)
	return utils.B2S(data)
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
		switch tagSize {
		case 1:
			k = uint16(r.ReadByte())
		case 2:
			k = r.ReadUInt16()
		case 4:
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

func (r *Reader) Index() int64 {
	return r.buf.Size()
}

func (tlv TlvMap) Exists(key uint16) bool {
	_, ok := tlv[key]
	return ok
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
	return buf, err
}

func (r *NetworkReader) ReadInt32() (int32, error) {
	b := make([]byte, 4)
	_, err := r.conn.Read(b)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}
