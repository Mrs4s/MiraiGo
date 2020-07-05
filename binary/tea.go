package binary

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"
)

const (
	delta   = uint32(0x9E3779B9)
	fillnor = 0xF8
)

type teaCipher struct {
	keys     [4]uint32
	value    []byte
	byte8    [8]byte
	ubyte32  [2]uint32
	xor      [8]byte //xor
	fxor     [8]byte //first xor
	lxor     [8]byte //last xor
	nxor     [8]byte //new xor Decrypt add
	balebuff *bytes.Buffer
	seedrand *rand.Rand
}

func NewTeaCipher(key []byte) *teaCipher {
	if len(key) != 16 {
		return nil
	}
	cipher := &teaCipher{
		balebuff: bytes.NewBuffer(nil),
	}
	for i := 0; i < 4; i++ {
		cipher.keys[i] = binary.BigEndian.Uint32(key[i*4:])
	}
	cipher.seedrand = rand.New(rand.NewSource(time.Now().UnixNano()))
	return cipher
}

func (c *teaCipher) Encrypt(value []byte) []byte {
	c.balebuff.Reset()
	vl := len(value)
	filln := (8 - (vl + 2)) % 8
	if filln < 0 {
		filln += 2 + 8
	} else {
		filln += 2
	}
	bindex := filln + 1
	if bindex <= 0 {
		return nil
	}
	rands := make([]byte, bindex)
	for i := 1; i < bindex; i++ {
		rands[i] = byte((c.seedrand.Intn(236) + 1))
	}
	rands[0] = byte((filln - 2) | fillnor)
	c.balebuff.Write(rands)
	c.balebuff.Write(value)
	c.balebuff.Write([]byte{00, 00, 00, 00, 00, 00, 00})
	vl = c.balebuff.Len()
	c.value = c.balebuff.Bytes()
	c.balebuff.Reset()
	for i := 0; i < vl; i += 8 {
		c.xor = xor(c.value[i:i+8], c.fxor[0:8])
		c.ubyte32[0] = binary.BigEndian.Uint32(c.xor[0:4])
		c.ubyte32[1] = binary.BigEndian.Uint32(c.xor[4:8])
		c.encipher()
		c.fxor = xor(c.byte8[0:8], c.lxor[0:8])
		c.balebuff.Write(c.fxor[0:8])
		c.lxor = c.xor

	}
	return c.balebuff.Bytes()
}

func (c *teaCipher) Decrypt(value []byte) []byte {
	vl := len(value)
	if vl <= 0 || (vl%8) != 0 {
		return nil
	}
	c.balebuff.Reset()
	c.ubyte32[0] = binary.BigEndian.Uint32(value[0:4])
	c.ubyte32[1] = binary.BigEndian.Uint32(value[4:8])
	c.decipher()
	copy(c.lxor[0:8], value[0:8])
	c.fxor = c.byte8
	pos := int((c.byte8[0] & 0x7) + 2)
	c.balebuff.Write(c.byte8[0:8])
	for i := 8; i < vl; i += 8 {
		c.xor = xor(value[i:i+8], c.fxor[0:8])
		c.ubyte32[0] = binary.BigEndian.Uint32(c.xor[0:4])
		c.ubyte32[1] = binary.BigEndian.Uint32(c.xor[4:8])
		c.decipher()
		c.nxor = xor(c.byte8[0:8], c.lxor[0:8])
		c.balebuff.Write(c.nxor[0:8])
		c.fxor = xor(c.nxor[0:8], c.lxor[0:8])
		copy(c.lxor[0:8], value[i:i+8])
	}
	pos++
	c.value = c.balebuff.Bytes()
	nl := c.balebuff.Len()
	if pos >= c.balebuff.Len() || (nl-7) <= pos {
		return nil
	}
	return c.value[pos : nl-7]
}

func (c *teaCipher) encipher() {
	sum := delta
	for i := 0x10; i > 0; i-- {
		c.ubyte32[0] += ((c.ubyte32[1] << 4 & 0xFFFFFFF0) + c.keys[0]) ^ (c.ubyte32[1] + sum) ^ ((c.ubyte32[1] >> 5 & 0x07ffffff) + c.keys[1])
		c.ubyte32[1] += ((c.ubyte32[0] << 4 & 0xFFFFFFF0) + c.keys[2]) ^ (c.ubyte32[0] + sum) ^ ((c.ubyte32[0] >> 5 & 0x07ffffff) + c.keys[3])
		sum += delta
	}
	binary.BigEndian.PutUint32(c.byte8[0:4], c.ubyte32[0])
	binary.BigEndian.PutUint32(c.byte8[4:8], c.ubyte32[1])
}

func (c *teaCipher) decipher() {
	sum := delta
	sum = (sum << 4) & 0xffffffff

	for i := 0x10; i > 0; i-- {
		c.ubyte32[1] -= (((c.ubyte32[0] << 4 & 0xFFFFFFF0) + c.keys[2]) ^ (c.ubyte32[0] + sum) ^ ((c.ubyte32[0] >> 5 & 0x07ffffff) + c.keys[3]))
		c.ubyte32[1] &= 0xffffffff
		c.ubyte32[0] -= (((c.ubyte32[1] << 4 & 0xFFFFFFF0) + c.keys[0]) ^ (c.ubyte32[1] + sum) ^ ((c.ubyte32[1] >> 5 & 0x07ffffff) + c.keys[1]))
		c.ubyte32[0] &= 0xffffffff
		sum -= delta
	}
	binary.BigEndian.PutUint32(c.byte8[0:4], c.ubyte32[0])
	binary.BigEndian.PutUint32(c.byte8[4:8], c.ubyte32[1])
}

func xor(a, b []byte) (bts [8]byte) {
	l := len(a)
	for i := 0; i < l; i += 4 {
		binary.BigEndian.PutUint32(bts[i:i+4], binary.BigEndian.Uint32(a[i:i+4])^binary.BigEndian.Uint32(b[i:i+4]))
	}
	return bts
}
