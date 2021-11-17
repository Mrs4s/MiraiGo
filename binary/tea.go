package binary

import (
	"encoding/binary"
	"math/rand"
)

type TEA [4]uint32

// Encrypt tea 加密
// http://bbs.chinaunix.net/thread-583468-1-1.html
// 感谢xichen大佬对TEA的解释
//go:nosplit
func (t TEA) Encrypt(src []byte) (dst []byte) {
	lens := len(src)
	fill := 10 - (lens+1)%8
	dst = make([]byte, fill+lens+7)
	_, _ = rand.Read(dst[0:fill])
	dst[0] = byte(fill-3) | 0xF8 // 存储pad长度
	copy(dst[fill:], src)

	var iv1, iv2, holder uint64
	var v0, v1 uint32
	for i := 0; i < len(dst); i += 8 {
		holder = binary.BigEndian.Uint64(dst[i:]) ^ iv1
		v0, v1 = uint32(holder>>32), uint32(holder)
		v0 += (v1 + 0x9e3779b9) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x9e3779b9) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x3c6ef372) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x3c6ef372) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0xdaa66d2b) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0xdaa66d2b) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x78dde6e4) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x78dde6e4) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x1715609d) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x1715609d) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0xb54cda56) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0xb54cda56) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x5384540f) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x5384540f) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0xf1bbcdc8) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0xf1bbcdc8) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x8ff34781) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x8ff34781) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x2e2ac13a) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x2e2ac13a) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0xcc623af3) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0xcc623af3) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x6a99b4ac) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x6a99b4ac) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x08d12e65) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x08d12e65) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0xa708a81e) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0xa708a81e) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0x454021d7) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0x454021d7) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 += (v1 + 0xe3779b90) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 += (v0 + 0xe3779b90) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		iv1 = (uint64(v0)<<32 | uint64(v1)) ^ iv2
		iv2 = holder
		binary.BigEndian.PutUint64(dst[i:], iv1)
	}

	return dst
}

//go:nosplit
func (t TEA) Decrypt(data []byte) []byte {
	if len(data) < 16 || len(data)%8 != 0 {
		return nil
	}
	dst := make([]byte, len(data))

	var iv1, iv2, holder uint64
	var v0, v1 uint32
	for i := 0; i < len(dst); i += 8 {
		holder = iv1
		iv1 = binary.BigEndian.Uint64(data[i:])
		iv2 ^= iv1
		v0, v1 = uint32(iv2>>32), uint32(iv2)
		v1 -= (v0 + 0xe3779b90) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0xe3779b90) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x454021d7) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x454021d7) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0xa708a81e) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0xa708a81e) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x08d12e65) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x08d12e65) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x6a99b4ac) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x6a99b4ac) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0xcc623af3) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0xcc623af3) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x2e2ac13a) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x2e2ac13a) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x8ff34781) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x8ff34781) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0xf1bbcdc8) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0xf1bbcdc8) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x5384540f) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x5384540f) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0xb54cda56) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0xb54cda56) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x1715609d) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x1715609d) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x78dde6e4) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x78dde6e4) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0xdaa66d2b) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0xdaa66d2b) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x3c6ef372) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x3c6ef372) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		v1 -= (v0 + 0x9e3779b9) ^ ((v0 << 4) + t[2]) ^ ((v0 >> 5) + t[3])
		v0 -= (v1 + 0x9e3779b9) ^ ((v1 << 4) + t[0]) ^ ((v1 >> 5) + t[1])
		iv2 = uint64(v0)<<32 | uint64(v1)
		binary.BigEndian.PutUint64(dst[i:], iv2^holder)
	}

	return dst[dst[0]&7+3 : len(data)-7]
}

//go:nosplit
func NewTeaCipher(key []byte) (t TEA) {
	if len(key) == 16 {
		t[3] = binary.BigEndian.Uint32(key[12:])
		t[2] = binary.BigEndian.Uint32(key[8:])
		t[1] = binary.BigEndian.Uint32(key[4:])
		t[0] = binary.BigEndian.Uint32(key[0:])
	}
	return
}
