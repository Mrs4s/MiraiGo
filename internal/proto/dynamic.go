package proto

import (
	"encoding/binary"
	"math"
	"sort"
)

type DynamicMessage map[uint64]any

// zigzag encoding types
type (
	SInt   int
	SInt32 int32
	SInt64 int64
)

type encoder struct {
	buf []byte
}

func (msg DynamicMessage) Encode() []byte {
	en := encoder{}

	// sort all items
	type pair struct {
		key   uint64
		value any
	}
	all := make([]pair, len(msg))
	for k, v := range msg {
		all = append(all, pair{key: k, value: v})
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].key < all[j].key
	})

	//nolint:staticcheck
	for _, item := range all {
		key := item.key << 3
		switch v := item.value.(type) {
		case bool:
			en.uvarint(key | 0)
			vi := uint64(0)
			if v {
				vi = 1
			}
			en.uvarint(vi)
		case int:
			en.uvarint(key | 0)
			en.uvarint(uint64(v))
		case uint:
			en.uvarint(key | 0)
			en.uvarint(uint64(v))
		case int32:
			en.uvarint(key | 0)
			en.uvarint(uint64(v))
		case int64:
			en.uvarint(key | 0)
			en.uvarint(uint64(v))
		case uint32:
			en.uvarint(key | 0)
			en.uvarint(uint64(v))
		case uint64:
			en.uvarint(key | 0)
			en.uvarint(v)
		case SInt:
			en.uvarint(key | 0)
			en.svarint(int64(v))
		case SInt32:
			en.uvarint(key | 0)
			en.svarint(int64(v))
		case SInt64:
			en.uvarint(key | 0)
			en.svarint(int64(v))
		case float32:
			en.uvarint(key | 5)
			en.u32(math.Float32bits(v))
		case float64:
			en.uvarint(key | 1)
			en.u64(math.Float64bits(v))
		case string:
			en.uvarint(key | 2)
			en.uvarint(uint64(len(v)))
			en.buf = append(en.buf, v...)
		case []uint64:
			for i := 0; i < len(v); i++ {
				en.uvarint(key | 0)
				en.uvarint(v[i])
			}
		case []byte:
			en.uvarint(key | 2)
			en.uvarint(uint64(len(v)))
			en.buf = append(en.buf, v...)
		case DynamicMessage:
			en.uvarint(key | 2)
			b := v.Encode()
			en.uvarint(uint64(len(b)))
			en.buf = append(en.buf, b...)
		}
	}
	return en.buf
}

func (en *encoder) uvarint(v uint64) {
	en.buf = binary.AppendUvarint(en.buf, v)
}

func (en *encoder) svarint(v int64) {
	en.buf = binary.AppendVarint(en.buf, v)
}

func (en *encoder) u32(v uint32) {
	en.buf = binary.LittleEndian.AppendUint32(en.buf, v)
}

func (en *encoder) u64(v uint64) {
	en.buf = binary.LittleEndian.AppendUint64(en.buf, v)
}
