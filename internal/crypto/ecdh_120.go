//go:build go1.20

package crypto

import (
	"crypto/ecdh"
	"crypto/md5"
	"crypto/rand"
)

func (e *ECDH) init(svrPubKey []byte) {
	p256 := ecdh.P256()
	local, _ := p256.GenerateKey(rand.Reader)
	remote, _ := p256.NewPublicKey(svrPubKey)
	share, _ := p256.ECDH(local, remote)

	hash := md5.New()
	hash.Write(share[:16])
	e.ShareKey = hash.Sum(nil)
	e.PublicKey = local.PublicKey().Bytes()
}
