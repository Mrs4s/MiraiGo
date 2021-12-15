package jce

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var globalBytes []byte

func BenchmarkJceWriter_WriteMap(b *testing.B) {
	x := globalBytes
	for i := 0; i < b.N; i++ {
		w := NewJceWriter()
		w.writeMapStrMapStrBytes(req.Map, 0)
		x = w.Bytes()
	}
	globalBytes = x
	b.SetBytes(int64(len(globalBytes)))
}

var reqPacket1 = &RequestPacket{
	IVersion:     1,
	CPacketType:  114,
	IMessageType: 514,
	IRequestId:   1919,
	SServantName: "田所",
	SFuncName:    "浩二",
	SBuffer:      []byte{1, 1, 4, 5, 1, 4, 1, 9, 1, 9, 8, 1, 0},
	ITimeout:     810,
	Context: map[string]string{
		"114":  "514",
		"1919": "810",
	},
	Status: map[string]string{
		"野兽": "前辈",
		"田所": "浩二",
	},
}

func BenchmarkJceWriter_WriteJceStructRaw(b *testing.B) {
	x := globalBytes
	for i := 0; i < b.N; i++ {
		_ = reqPacket1.ToBytes()
	}
	globalBytes = x
	b.SetBytes(int64(len(globalBytes)))
}

func TestJceWriter_WriteJceStructRaw(t *testing.T) {
	r := NewJceReader(reqPacket1.ToBytes())
	var reqPacket2 RequestPacket
	reqPacket2.ReadFrom(r)
	assert.Equal(t, reqPacket1, &reqPacket2)
}
