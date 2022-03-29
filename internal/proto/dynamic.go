package proto

import (
	"bytes"
	"encoding/binary"
	"math"
)

type DynamicMessage map[uint64]any

type encoder struct {
	bytes.Buffer
}

func (msg DynamicMessage) Encode() []byte {
	en := &encoder{}
	//nolint:staticcheck
	for id, value := range msg {
		key := id << 3
		switch v := value.(type) {
		case bool:
			en.uvarint(key | 0)
			vi := uint64(0)
			if v {
				vi = 1
			}
			en.uvarint(vi)
		case int:
			en.uvarint(key | 0)
			en.svarint(int64(v))
		case int32:
			en.uvarint(key | 0)
			en.svarint(int64(v))
		case int64:
			en.uvarint(key | 0)
			en.svarint(v)
		case uint32:
			en.uvarint(key | 0)
			en.uvarint(uint64(v))
		case uint64:
			en.uvarint(key | 0)
			en.uvarint(v)
		case float32:
			en.uvarint(key | 5)
			en.u32(math.Float32bits(v))
		case float64:
			en.uvarint(key | 1)
			en.u64(math.Float64bits(v))
		case string:
			en.uvarint(key | 2)
			b := []byte(v)
			en.uvarint(uint64(len(b)))
			_, _ = en.Write(b)
		case []uint64:
			for i := 0; i < len(v); i++ {
				en.uvarint(key | 0)
				en.uvarint(v[i])
			}
		case []byte:
			en.uvarint(key | 2)
			en.uvarint(uint64(len(v)))
			_, _ = en.Write(v)
		case DynamicMessage:
			en.uvarint(key | 2)
			b := v.Encode()
			en.uvarint(uint64(len(b)))
			_, _ = en.Write(b)
		}
	}
	return en.Bytes()
}

func (en *encoder) uvarint(v uint64) {
	var b [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(b[:], v)
	_, _ = en.Write(b[:n])
}

func (en *encoder) svarint(v int64) {
	en.uvarint(uint64(v)<<1 ^ uint64(v>>63))
}

func (en *encoder) u32(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	_, _ = en.Write(b[:])
}

func (en *encoder) u64(v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	_, _ = en.Write(b[:])
}
