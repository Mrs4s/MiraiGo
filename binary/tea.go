package binary

import (
	"encoding/binary"
	"math/rand"
	"reflect"
	"unsafe"
)

func xorQ(a, b []byte, c []byte) { // MAGIC
	*(*uint64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&c)).Data)) =
		*(*uint64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&a)).Data)) ^
			*(*uint64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data))
}

func isZero(a []byte) bool { // MAGIC
	return *(*uint64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&a)).Data)) == 0
}

type TEA struct {
	key [4]uint32
}

// http://bbs.chinaunix.net/thread-583468-1-1.html
// 感谢xichen大佬对TEA的解释
func (t *TEA) Encrypt(src []byte) (dst []byte) {
	lens := len(src)
	fill := 10 - (lens+1)%8
	tmp1 := make([]byte, 8) // 非纯src的数据
	tmp2 := make([]byte, 8)
	dst = make([]byte, fill+lens+7)
	//for i := 0; i < fill; i++ {
	//	dst[i] = ' '
	//} // For test purpose
	_, _ = rand.Read(dst[0:fill])
	dst[0] = byte(fill-3) | 0xF8 // 存储pad长度
	in := 0                      // 位置
	// #1
	if fill < 8 {
		in = 8 - fill
		copy(dst[fill:8], src[:in])
	}
	copy(tmp2, dst[0:8])
	t.encode(dst[0:8], dst[0:8])
	out := 8 // 位置
	// #2
	if fill > 8 {
		copy(dst[fill:out+8], src[:16-fill])
		xorQ(dst[8:16], dst[0:8], dst[8:16]) // 与前一次结果xor
		copy(tmp1, dst[8:16])
		t.encode(dst[8:16], dst[8:16])
		xorQ(dst[8:16], tmp2, dst[8:16]) // 与前一次数据xor
		copy(tmp2, tmp1)
		in = 16 - fill
		out = 16
	}
	// #3+或#4+
	lens -= 8
	for in < lens {
		xorQ(src[in:in+8], dst[out-8:out], dst[out:out+8]) // 与前一次结果xor
		copy(tmp1, dst[out:out+8])
		t.encode(dst[out:out+8], dst[out:out+8])
		xorQ(dst[out:out+8], tmp2, dst[out:out+8]) // 与前一次数据xor
		copy(tmp2, tmp1)
		in += 8
		out += 8
	}
	tmp3 := make([]byte, 8)
	copy(tmp3, src[in:])
	xorQ(tmp3, dst[out-8:out], dst[out:out+8]) // 与前一次结果xor
	t.encode(dst[out:out+8], dst[out:out+8])
	xorQ(dst[out:out+8], tmp2, dst[out:out+8]) // 与前一次数据xor
	return dst
}

func (t *TEA) Decrypt(data []byte) []byte {
	if len(data) < 16 || len(data)%8 != 0 {
		return nil
	}
	dst := make([]byte, len(data))
	copy(dst, data)
	t.decode(dst[0:8], dst[0:8])
	tmp := make([]byte, 8)
	copy(tmp, dst[0:8])
	for in := 8; in < len(data); in += 8 {
		xorQ(dst[in:in+8], tmp, dst[in:in+8])
		t.decode(dst[in:in+8], dst[in:in+8])
		xorQ(dst[in:in+8], data[in-8:in], dst[in:in+8])
		xorQ(dst[in:in+8], data[in-8:in], tmp)
	}
	//if !isZero(dst[len(data)-7:]) {
	//	return nil
	//}
	return dst[dst[0]&7+3 : len(data)-7]
}

//go:nosplit
func unpack(data []byte) (v0, v1 uint32) {
	v1 = uint32(data[7]) | uint32(data[6])<<8 | uint32(data[5])<<16 | uint32(data[4])<<24
	v0 = uint32(data[3]) | uint32(data[2])<<8 | uint32(data[1])<<16 | uint32(data[0])<<24
	return v0, v1
}

//go:nosplit
func repack(data []byte, v0, v1 uint32) {
	_ = data[7] // early bounds check to guarantee safety of writes below
	data[0] = byte(v0 >> 24)
	data[1] = byte(v0 >> 16)
	data[2] = byte(v0 >> 8)
	data[3] = byte(v0)

	data[4] = byte(v1 >> 24)
	data[5] = byte(v1 >> 16)
	data[6] = byte(v1 >> 8)
	data[7] = byte(v1)
}

var sumTable = [0x10]uint32{
	0x9e3779b9,
	0x3c6ef372,
	0xdaa66d2b,
	0x78dde6e4,
	0x1715609d,
	0xb54cda56,
	0x5384540f,
	0xf1bbcdc8,
	0x8ff34781,
	0x2e2ac13a,
	0xcc623af3,
	0x6a99b4ac,
	0x08d12e65,
	0xa708a81e,
	0x454021d7,
	0xe3779b90,
}

//go:nosplit
func (t *TEA) encode(src, dst []byte) {
	v0, v1 := unpack(src)
	for i := 0; i < 0x10; i++ {
		v0 += ((v1 << 4) + t.key[0]) ^ (v1 + sumTable[i]) ^ ((v1 >> 5) + t.key[1])
		v1 += ((v0 << 4) + t.key[2]) ^ (v0 + sumTable[i]) ^ ((v0 >> 5) + t.key[3])
	}
	repack(dst, v0, v1)
}

// 每次8字节
//go:nosplit
func (t *TEA) decode(src, dst []byte) {
	v0, v1 := unpack(src)
	for i := 0xf; i >= 0; i-- {
		v1 -= ((v0 << 4) + t.key[2]) ^ (v0 + sumTable[i]) ^ ((v0 >> 5) + t.key[3])
		v0 -= ((v1 << 4) + t.key[0]) ^ (v1 + sumTable[i]) ^ ((v1 >> 5) + t.key[1])
	}
	repack(dst, v0, v1)
}

//go:nosplit
func NewTeaCipher(key []byte) *TEA {
	if len(key) != 16 {
		return nil
	}
	t := new(TEA)
	t.key[3] = binary.BigEndian.Uint32(key[12:])
	t.key[2] = binary.BigEndian.Uint32(key[8:])
	t.key[1] = binary.BigEndian.Uint32(key[4:])
	t.key[0] = binary.BigEndian.Uint32(key[0:])
	return t
}
