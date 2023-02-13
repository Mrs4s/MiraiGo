package tlv

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

// Record represents a Tag-Length-Value record.
type Record struct {
	Tag    int
	Length int
	Value  []byte
}

type RecordMap map[int][]byte

func (rm RecordMap) Exists(key int) bool {
	_, ok := rm[key]
	return ok
}

var ErrMessageTooShort = errors.New("tlv: message too short")

// Decoder is a configurable TLV decoder.
type Decoder struct {
	tagSize  uint8
	lenSize  uint8
	headSize uint8
}

func NewDecoder(tagSize, lenSize uint8) *Decoder {
	check := func(t string, s uint8) {
		switch s {
		case 1, 2, 4:
			// ok
		default:
			panic("invalid " + t)
		}
	}
	check("tag size", tagSize)
	check("len size", lenSize)

	return &Decoder{tagSize: tagSize, lenSize: lenSize, headSize: tagSize + lenSize}
}

func (d *Decoder) decodeRecord(data []byte) (r Record, err error) {
	tagSize := d.tagSize
	lenSize := d.lenSize
	headSize := int(tagSize + lenSize)
	if len(data) < headSize {
		err = ErrMessageTooShort
		return
	}

	r.Tag = d.read(tagSize, data)
	r.Length = d.read(lenSize, data[tagSize:])

	if len(data) < headSize+r.Length {
		err = ErrMessageTooShort
		return
	}
	r.Value = data[headSize : headSize+r.Length : headSize+r.Length]
	return
}

func (d *Decoder) read(size uint8, data []byte) int {
	switch size {
	case 1:
		return int(data[0])
	case 2:
		return int(binary.BigEndian.Uint16(data))
	case 4:
		return int(binary.BigEndian.Uint32(data))
	default:
		panic("invalid size")
	}
}

func (d *Decoder) Decode(data []byte) ([]Record, error) {
	var records []Record
	for len(data) > 0 {
		r, err := d.decodeRecord(data)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
		data = data[int(d.headSize)+r.Length:]
	}
	return records, nil
}

func (d *Decoder) DecodeRecordMap(data []byte) (RecordMap, error) {
	records, err := d.Decode(data)
	if err != nil {
		return nil, err
	}

	rm := make(RecordMap, len(records))
	for _, record := range records {
		rm[record.Tag] = record.Value
	}

	return rm, nil
}
