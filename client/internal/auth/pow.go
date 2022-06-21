package auth

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
)

func CalcPow(data []byte) []byte {
	r := binary.NewReader(data)
	a := r.ReadByte()
	typ := r.ReadByte()
	c := r.ReadByte()
	ok := r.ReadByte() != 0
	e := r.ReadUInt16()
	f := r.ReadUInt16()
	src := r.ReadBytesShort()
	tgt := r.ReadBytesShort()
	cpy := r.ReadBytesShort()

	var dst []byte
	var elp, cnt uint32
	if typ == 2 && len(tgt) == 32 {
		start := time.Now()
		tmp := new(big.Int).SetBytes(src)
		hash := sha256.Sum256(tmp.Bytes())
		one := big.NewInt(1)
		for !bytes.Equal(hash[:], tgt) {
			tmp = tmp.Add(tmp, one)
			hash = sha256.Sum256(tmp.Bytes())
			cnt++
		}
		ok = true
		dst = tmp.Bytes()
		elp = uint32(time.Now().Sub(start).Milliseconds())
	}

	w := binary.SelectWriter()
	w.WriteByte(a)
	w.WriteByte(typ)
	w.WriteByte(c)
	w.WriteBool(ok)
	w.WriteUInt16(e)
	w.WriteUInt16(f)
	w.WriteBytesShort(src)
	w.WriteBytesShort(tgt)
	w.WriteBytesShort(cpy)
	if ok {
		w.WriteBytesShort(dst)
		w.WriteUInt32(elp)
		w.WriteUInt32(cnt)
	}
	return w.Bytes()
}
