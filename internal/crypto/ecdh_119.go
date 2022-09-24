//go:build !go1.20

package crypto

import (
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
)

func (e *ECDH) init(svrPubKey []byte) {
	p256 := elliptic.P256()
	key, sx, sy, _ := elliptic.GenerateKey(p256, rand.Reader)
	tx, ty := elliptic.Unmarshal(p256, svrPubKey)
	x, _ := p256.ScalarMult(tx, ty, key)
	hash := md5.Sum(x.Bytes()[:16])
	e.ShareKey = hash[:]
	e.PublicKey = elliptic.Marshal(p256, sx, sy)
}
