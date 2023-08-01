package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/rand"
	"time"

	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/tidwall/gjson"
)

const (
	secret = "ZdJqM15EeO2zWc08"
	rsaKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDEIxgwoutfwoJxcGQeedgP7FG9
qaIuS0qzfR8gWkrkTZKM2iWHn2ajQpBRZjMSoSf6+KJGvar2ORhBfpDXyVtZCKpq
LQ+FLkpncClKVIrBwv6PHyUvuCb0rIarmgDnzkfQAqVufEtR64iazGDKatvJ9y6B
9NMbHddGSAUmRTCrHQIDAQAB
-----END PUBLIC KEY-----`
)

func (info *Device) RequestQImei() {
	if info.Protocol.Version().AppKey == "" {
		return
	}

	// init params
	payload, _ := json.Marshal(genRandomPayloadByDevice(info))
	cryptKey := utils.RandomStringRange(16, "abcdef1234567890")
	ts := time.Now().Unix() * 1000
	nonce := utils.RandomStringRange(16, "abcdef1234567890")

	// init rsa key and aes key
	publicKey := initPublicKey()
	encryptedAesKey, _ := rsa.EncryptPKCS1v15(rand.New(rand.NewSource(time.Now().UnixNano())), publicKey, []byte(cryptKey))

	encryptedPayload := aesEncrypt(payload, []byte(cryptKey))

	key := base64.StdEncoding.EncodeToString(encryptedAesKey)
	params := base64.StdEncoding.EncodeToString(encryptedPayload)

	postData, _ := json.Marshal(map[string]any{
		"key":    key,
		"params": params,
		"time":   ts,
		"nonce":  nonce,
		"sign":   sign(key, params, fmt.Sprint(ts), nonce),
		"extra":  "",
	})

	resp, _ := utils.HttpPostBytesWithCookie("https://snowflake.qq.com/ola/android", postData, "", "application/json")
	if gjson.GetBytes(resp, "code").Int() != 0 {
		return
	}
	encryptedResponse, _ := base64.StdEncoding.DecodeString(gjson.GetBytes(resp, "data").String())
	if len(encryptedResponse) == 0 {
		return
	}
	decryptedResponse := aesDecrypt(encryptedResponse, []byte(cryptKey))
	info.QImei16 = gjson.GetBytes(decryptedResponse, "q16").String()
	info.QImei36 = gjson.GetBytes(decryptedResponse, "q36").String()
}

func initPublicKey() *rsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(rsaKey))
	pub, _ := x509.ParsePKIXPublicKey(blockPub.Bytes)
	return pub.(*rsa.PublicKey)
}

func sign(key, params, ts, nonce string) string {
	h := md5.Sum([]byte(key + params + ts + nonce + secret))
	return hex.EncodeToString(h[:])
}

func aesEncrypt(src []byte, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	ecb := cipher.NewCBCEncrypter(block, key)
	content := src
	content = pkcs5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	return crypted
}

func aesDecrypt(crypt []byte, key []byte) []byte {
	block, _ := aes.NewCipher(key)
	ecb := cipher.NewCBCDecrypter(block, key)
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)
	return pkcs5Trimming(decrypted)
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func genRandomPayloadByDevice(info *Device) map[string]any {
	now := time.Now()
	seed := int64(0x6F4)
	for _, b := range info.Guid {
		seed += int64(b)
	}
	fixedRand := rand.New(rand.NewSource(seed))

	reserved := map[string]string{
		"harmony":    "0",
		"clone":      "0",
		"containe":   "",
		"oz":         "UhYmelwouA+V2nPWbOvLTgN2/m8jwGB+yUB5v9tysQg=",
		"oo":         "Xecjt+9S1+f8Pz2VLSxgpw==",
		"kelong":     "0",
		"uptimes":    time.Unix(now.Unix()-fixedRand.Int63n(14400), 0).Format(time.DateTime),
		"multiUser":  "0",
		"bod":        string(info.Board),
		"brd":        string(info.Brand),
		"dv":         string(info.Device),
		"firstLevel": "",
		"manufact":   string(info.Brand),
		"name":       string(info.Model),
		"host":       "se.infra",
		"kernel":     string(info.ProcVersion),
	}
	reservedBytes, _ := json.Marshal(reserved)
	deviceType := "Phone"
	if info.Protocol == AndroidPad {
		deviceType = "Pad"
	}
	beaconId := ""
	timeMonth := time.Now().Format("2006-01-") + "01"
	rand1 := fixedRand.Intn(899999) + 100000
	rand2 := fixedRand.Intn(899999999) + 100000000
	for i := 1; i <= 40; i++ {
		switch i {
		case 1, 2, 13, 14, 17, 18, 21, 22, 25, 26, 29, 30, 33, 34, 37, 38:
			beaconId += fmt.Sprintf("k%v:%v%v.%v", i, timeMonth, rand1, rand2)
		case 3:
			beaconId += "k3:0000000000000000"
		case 4:
			beaconId += "k4:" + utils.RandomStringRange(16, "123456789abcdef")
		default:
			beaconId += fmt.Sprintf("k%v:%v", i, fixedRand.Intn(10000))
		}
		beaconId += ";"
	}
	return map[string]any{
		"androidId":   string(info.AndroidId),
		"platformId":  1,
		"appKey":      info.Protocol.Version().AppKey,
		"appVersion":  info.Protocol.Version().SortVersionName,
		"beaconIdSrc": beaconId,
		"brand":       string(info.Brand),
		"channelId":   "2017",
		"cid":         "",
		"imei":        info.IMEI,
		"imsi":        "",
		"mac":         "",
		"model":       string(info.Model),
		"networkType": "unknown",
		"oaid":        "",
		"osVersion":   "Android " + string(info.Version.Release) + ",level " + fmt.Sprint(info.Version.SDK),
		"qimei":       "",
		"qimei36":     "",
		"sdkVersion":  "1.2.13.6",
		"audit":       "",
		"userId":      "{}",
		"packageId":   info.Protocol.Version().ApkId,
		"deviceType":  deviceType,
		"sdkName":     "",
		"reserved":    string(reservedBytes),
	}
}
