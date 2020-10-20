package crypto

// Copyright 2010 The Go Authors. All rights reserved.
// Copyright 2011 ThePiachu. All rights reserved.
// Copyright 2020 LXY1226. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bitelliptic implements several Koblitz elliptic curves over prime
// fields.

// Origin File at:
// https://github.com/ThePiachu/Split-Vanity-Miner-Golang/blob/03677bc96ff4f5c2771e528562360ccbc513db8d/src/pkg/bitelliptic/bitelliptic.go

// This package operates, internally, on Jacobian coordinates. For a given
// (x, y) position on the curve, the Jacobian coordinates are (x1, y1, z1)
// where x = x1/z1² and y = y1/z1³. The greatest speedups come when the whole
// calculation can be performed within the transform (as in ScalarMult and
// ScalarBaseMult). But even for Add and Double, it's faster to apply and
// reverse the transform than to operate in affine coordinates.

import (
	"io"
	"math/big"
)

// A BitCurve represents a Koblitz Curve with a=0.
// See http://www.hyperelliptic.org/EFD/g1p/auto-shortw.html
type BitCurve struct {
	P       *big.Int // the order of the underlying field
	N       *big.Int // the order of the base point
	B       *big.Int // the constant of the BitCurve equation
	Gx, Gy  *big.Int // (x,y) of the base point
	BitSize int      // the size of the underlying field
}

// See FIPS 186-3, section D.2.2.1
// And http://www.secg.org/sec2-v2.pdf section 2.2.1
var secp192k1 = &BitCurve{
	P: new(big.Int).SetBytes([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xEE, 0x37,
	}), N: new(big.Int).SetBytes([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
		0x26, 0xF2, 0xFC, 0x17, 0x0F, 0x69, 0x46, 0x6A, 0x74, 0xDE, 0xFD, 0x8D,
	}), B: new(big.Int).SetBytes([]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
	}), Gx: new(big.Int).SetBytes([]byte{
		0xDB, 0x4F, 0xF1, 0x0E, 0xC0, 0x57, 0xE9, 0xAE, 0x26, 0xB0, 0x7D, 0x02,
		0x80, 0xB7, 0xF4, 0x34, 0x1D, 0xA5, 0xD1, 0xB1, 0xEA, 0xE0, 0x6C, 0x7D,
	}), Gy: new(big.Int).SetBytes([]byte{
		0x9B, 0x2F, 0x2F, 0x6D, 0x9C, 0x56, 0x28, 0xA7, 0x84, 0x41, 0x63, 0xD0,
		0x15, 0xBE, 0x86, 0x34, 0x40, 0x82, 0xAA, 0x88, 0xD9, 0x5E, 0x2F, 0x9D,
	}),
	BitSize: 192,
}

//TODO: double check if the function is okay
// affineFromJacobian reverses the Jacobian transform. See the comment at the
// top of the file.
func (BitCurve *BitCurve) affineFromJacobian(x, y, z *big.Int) (xOut, yOut *big.Int) {
	zinv := new(big.Int).ModInverse(z, BitCurve.P)
	zinvsq := new(big.Int).Mul(zinv, zinv)

	xOut = new(big.Int).Mul(x, zinvsq)
	xOut.Mod(xOut, BitCurve.P)
	zinvsq.Mul(zinvsq, zinv)
	yOut = new(big.Int).Mul(y, zinvsq)
	yOut.Mod(yOut, BitCurve.P)
	return
}

// addJacobian takes two points in Jacobian coordinates, (x1, y1, z1) and
// (x2, y2, z2) and returns their sum, also in Jacobian form.
func (BitCurve *BitCurve) addJacobian(x1, y1, z1, x2, y2, z2 *big.Int) (*big.Int, *big.Int, *big.Int) {
	// See http://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html#addition-add-2007-bl
	z1z1 := new(big.Int).Mul(z1, z1)
	z1z1.Mod(z1z1, BitCurve.P)
	z2z2 := new(big.Int).Mul(z2, z2)
	z2z2.Mod(z2z2, BitCurve.P)

	u1 := new(big.Int).Mul(x1, z2z2)
	u1.Mod(u1, BitCurve.P)
	u2 := new(big.Int).Mul(x2, z1z1)
	u2.Mod(u2, BitCurve.P)
	h := new(big.Int).Sub(u2, u1)
	if h.Sign() == -1 {
		h.Add(h, BitCurve.P)
	}
	i := new(big.Int).Lsh(h, 1)
	i.Mul(i, i)
	j := new(big.Int).Mul(h, i)

	s1 := new(big.Int).Mul(y1, z2)
	s1.Mul(s1, z2z2)
	s1.Mod(s1, BitCurve.P)
	s2 := new(big.Int).Mul(y2, z1)
	s2.Mul(s2, z1z1)
	s2.Mod(s2, BitCurve.P)
	r := new(big.Int).Sub(s2, s1)
	if r.Sign() == -1 {
		r.Add(r, BitCurve.P)
	}
	r.Lsh(r, 1)
	v := new(big.Int).Mul(u1, i)

	x3 := new(big.Int).Set(r)
	x3.Mul(x3, x3)
	x3.Sub(x3, j)
	x3.Sub(x3, v)
	x3.Sub(x3, v)
	x3.Mod(x3, BitCurve.P)

	y3 := new(big.Int).Set(r)
	v.Sub(v, x3)
	y3.Mul(y3, v)
	s1.Mul(s1, j)
	s1.Lsh(s1, 1)
	y3.Sub(y3, s1)
	y3.Mod(y3, BitCurve.P)

	z3 := new(big.Int).Add(z1, z2)
	z3.Mul(z3, z3)
	z3.Sub(z3, z1z1)
	if z3.Sign() == -1 {
		z3.Add(z3, BitCurve.P)
	}
	z3.Sub(z3, z2z2)
	if z3.Sign() == -1 {
		z3.Add(z3, BitCurve.P)
	}
	z3.Mul(z3, h)
	z3.Mod(z3, BitCurve.P)

	return x3, y3, z3
}

// doubleJacobian takes a point in Jacobian coordinates, (x, y, z), and
// returns its double, also in Jacobian form.
func (BitCurve *BitCurve) doubleJacobian(x, y, z *big.Int) (*big.Int, *big.Int, *big.Int) {
	// See http://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html#doubling-dbl-2009-l

	a := new(big.Int).Mul(x, x) //X1²
	b := new(big.Int).Mul(y, y) //Y1²
	c := new(big.Int).Mul(b, b) //B²

	d := new(big.Int).Add(x, b) //X1+B
	d.Mul(d, d)                 //(X1+B)²
	d.Sub(d, a)                 //(X1+B)²-A
	d.Sub(d, c)                 //(X1+B)²-A-C
	d.Mul(d, big.NewInt(2))     //2*((X1+B)²-A-C)

	e := new(big.Int).Mul(big.NewInt(3), a) //3*A
	f := new(big.Int).Mul(e, e)             //E²

	x3 := new(big.Int).Mul(big.NewInt(2), d) //2*D
	x3.Sub(f, x3)                            //F-2*D
	x3.Mod(x3, BitCurve.P)

	y3 := new(big.Int).Sub(d, x3)                  //D-X3
	y3.Mul(e, y3)                                  //E*(D-X3)
	y3.Sub(y3, new(big.Int).Mul(big.NewInt(8), c)) //E*(D-X3)-8*C
	y3.Mod(y3, BitCurve.P)

	z3 := new(big.Int).Mul(y, z) //Y1*Z1
	z3.Mul(big.NewInt(2), z3)    //3*Y1*Z1
	z3.Mod(z3, BitCurve.P)

	return x3, y3, z3
}

//TODO: double check if it is okay
// ScalarMult returns k*(Bx,By) where k is a number in big-endian form.
func (BitCurve *BitCurve) ScalarMult(Bx, By *big.Int, k []byte) (*big.Int, *big.Int) {
	// We have a slight problem in that the identity of the group (the
	// point at infinity) cannot be represented in (x, y) form on a finite
	// machine. Thus the standard add/double algorithm has to be tweaked
	// slightly: our initial state is not the identity, but x, and we
	// ignore the first true bit in |k|.  If we don't find any true bits in
	// |k|, then we return nil, nil, because we cannot return the identity
	// element.

	Bz := new(big.Int).SetInt64(1)
	x := Bx
	y := By
	z := Bz

	seenFirstTrue := false
	for _, b := range k {
		for bitNum := 0; bitNum < 8; bitNum++ {
			if seenFirstTrue {
				x, y, z = BitCurve.doubleJacobian(x, y, z)
			}
			if b&0x80 == 0x80 {
				if !seenFirstTrue {
					seenFirstTrue = true
				} else {
					x, y, z = BitCurve.addJacobian(Bx, By, Bz, x, y, z)
				}
			}
			b <<= 1
		}
	}

	if !seenFirstTrue {
		return nil, nil
	}

	return BitCurve.affineFromJacobian(x, y, z)
}

// ScalarBaseMult returns k*G, where G is the base point of the group and k is
// an integer in big-endian form.
func (BitCurve *BitCurve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return BitCurve.ScalarMult(BitCurve.Gx, BitCurve.Gy, k)
}

var mask = []byte{0xff, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f}

//TODO: double check if it is okay
// GenerateKey returns a public/private key pair. The private key is generated
// using the given reader, which must return random data.
func (BitCurve *BitCurve) GenerateKey(rand io.Reader) (priv []byte, x, y *big.Int, err error) {
	byteLen := (BitCurve.BitSize + 7) >> 3
	priv = make([]byte, byteLen)

	for x == nil {
		_, err = rand.Read(priv)
		if err != nil {
			return
		}
		// We have to mask off any excess bits in the case that the size of the
		// underlying field is not a whole number of bytes.
		priv[0] &= mask[BitCurve.BitSize%8]
		// This is because, in tests, rand will return all zeros and we don't
		// want to get the point at infinity and loop forever.
		priv[1] ^= 0x42
		x, y = BitCurve.ScalarBaseMult(priv)
	}
	return
}

/*
$ openssl asn1parse -in 1.cer  -inform DER -dump
    0:d=0  hl=2 l=  70 cons: SEQUENCE
    2:d=1  hl=2 l=  16 cons: SEQUENCE
    4:d=2  hl=2 l=   7 prim: OBJECT            :id-ecPublicKey
   13:d=2  hl=2 l=   5 prim: OBJECT            :secp192k1
   20:d=1  hl=2 l=  50 prim: BIT STRING
      0000 - 00 04 92 8d 88 50 67 30-88 b3 43 26 4e 0c 6b ac   .....Pg0..C&N.k.
      0010 - b8 49 6d 69 77 99 f3 72-11 de b2 5b b7 39 06 cb   .Imiw..r...[.9..
      0020 - 08 9f ea 96 39 b4 e0 26-04 98 b5 1a 99 2d 50 81   ....9..&.....-P.
      0030 - 3d a8                                             =.
*/
