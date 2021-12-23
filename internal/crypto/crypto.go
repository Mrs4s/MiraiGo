package crypto

import (
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
)

type ECDH struct {
	SvrPublicKeyVer uint16
	PublicKey       []byte
	ShareKey        []byte
}

type EncryptSession struct {
	T133 []byte
}

const serverPublicKey = "04EBCA94D733E399B2DB96EACDD3F69A8BB0F74224E2B44E3357812211D2E62EFBC91BB553098E25E33A799ADC7F76FEB208DA7C6522CDB0719A305180CC54A82E"

func NewECDH() *ECDH {
	e := &ECDH{
		SvrPublicKeyVer: 1,
	}
	e.generateKey(serverPublicKey)
	return e
}

func (e *ECDH) generateKey(sPubKey string) {
	pub, _ := hex.DecodeString(sPubKey)
	p256 := elliptic.P256()
	key, sx, sy, _ := elliptic.GenerateKey(p256, rand.Reader)
	tx, ty := elliptic.Unmarshal(p256, pub)
	x, _ := p256.ScalarMult(tx, ty, key)
	hash := md5.Sum(x.Bytes()[:16])
	e.ShareKey = hash[:]
	e.PublicKey = elliptic.Marshal(p256, sx, sy)
}

type pubKeyResp struct {
	Meta struct {
		PubKeyVer uint16 `json:"KeyVer"`
		PubKey    string `json:"PubKey"`
	} `json:"PubKeyMeta"`
}

// FetchPubKey 从服务器获取PubKey
func (e *ECDH) FetchPubKey(uin int64) {
	resp, err := http.Get("https://keyrotate.qq.com/rotate_key?cipher_suite_ver=305&uin=" + strconv.FormatInt(uin, 10))
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	pubKey := pubKeyResp{}
	err = json.NewDecoder(resp.Body).Decode(&pubKey)
	if err != nil {
		return
	}
	e.SvrPublicKeyVer = pubKey.Meta.PubKeyVer
	e.generateKey(pubKey.Meta.PubKey) // todo check key sign
}
