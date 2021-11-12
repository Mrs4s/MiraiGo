package binary

import (
	"encoding/binary"
	"math/rand"
	"reflect"
	"unsafe"

	"github.com/segmentio/asm/bswap"

	"github.com/Mrs4s/MiraiGo/internal/cpu"
)

func xorQ(a, b []byte, c []byte) { // MAGIC
	*(*uint64)(unsafe.Pointer(&c[0])) =
		*(*uint64)(unsafe.Pointer(&a[0])) ^ *(*uint64)(unsafe.Pointer(&b[0]))
}

type TEA [4]uint32

// Encrypt tea 加密
// http://bbs.chinaunix.net/thread-583468-1-1.html
// 感谢xichen大佬对TEA的解释
func (t *TEA) Encrypt(src []byte) (dst []byte) {
	lens := len(src)
	fill := 10 - (lens+1)%8
	dst = make([]byte, fill+lens+7)
	_, _ = rand.Read(dst[0:fill])
	dst[0] = byte(fill-3) | 0xF8 // 存储pad长度
	copy(dst[fill:], src)
	if cpu.LittleEndian {
		bswap.Swap64(dst)
	}

	var iv1, iv2, holder int64
	var blocks []int64
	dstHeader := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	blocksHeader := (*reflect.SliceHeader)(unsafe.Pointer(&blocks))
	blocksHeader.Data = dstHeader.Data
	blocksHeader.Len = dstHeader.Len / 8
	blocksHeader.Cap = blocksHeader.Len
	for i, block := range blocks {
		holder = block ^ iv1
		iv1 = t.encode(holder)
		iv1 = iv1 ^ iv2
		iv2 = holder
		blocks[i] = iv1
	}
	if cpu.LittleEndian {
		bswap.Swap64(dst)
	}
	return dst
}

func (t *TEA) Decrypt(data []byte) []byte {
	if len(data) < 16 || len(data)%8 != 0 {
		return nil
	}
	dst := make([]byte, len(data))
	copy(dst, data)
	if cpu.LittleEndian {
		bswap.Swap64(dst)
	}

	var iv1, iv2, holder, tmp int64
	var blocks []int64
	dstHeader := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	blocksHeader := (*reflect.SliceHeader)(unsafe.Pointer(&blocks))
	blocksHeader.Data = dstHeader.Data
	blocksHeader.Len = dstHeader.Len / 8
	blocksHeader.Cap = blocksHeader.Len
	for i, block := range blocks {
		tmp = t.decode(block ^ iv2)
		iv2 = tmp
		holder = tmp ^ iv1
		iv1 = block
		blocks[i] = holder
	}

	if cpu.LittleEndian {
		bswap.Swap64(dst)
	}
	return dst[dst[0]&7+3 : len(data)-7]
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
func (t *TEA) encode(n int64) int64 {
	v0, v1 := uint32(n>>32), uint32(n)
	for i := 0; i < 0x10; i++ {
		v0 += ((v1 << 4) + t[0]) ^ (v1 + sumTable[i]) ^ ((v1 >> 5) + t[1])
		v1 += ((v0 << 4) + t[2]) ^ (v0 + sumTable[i]) ^ ((v0 >> 5) + t[3])
	}
	return int64(v0)<<32 | int64(v1)
}

// 每次8字节
//go:nosplit
func (t *TEA) decode(n int64) int64 {
	v0, v1 := uint32(n>>32), uint32(n)
	for i := 0xf; i >= 0; i-- {
		v1 -= ((v0 << 4) + t[2]) ^ (v0 + sumTable[i]) ^ ((v0 >> 5) + t[3])
		v0 -= ((v1 << 4) + t[0]) ^ (v1 + sumTable[i]) ^ ((v1 >> 5) + t[1])
	}
	return int64(v0)<<32 | int64(v1)
}

//go:nosplit
func NewTeaCipher(key []byte) *TEA {
	if len(key) != 16 {
		return nil
	}
	t := new(TEA)
	t[3] = binary.BigEndian.Uint32(key[12:])
	t[2] = binary.BigEndian.Uint32(key[8:])
	t[1] = binary.BigEndian.Uint32(key[4:])
	t[0] = binary.BigEndian.Uint32(key[0:])
	return t
}
