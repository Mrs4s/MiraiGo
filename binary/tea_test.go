package binary

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/Mrs4s/MiraiGo/utils"
)

var testTEA = NewTeaCipher([]byte("0123456789ABCDEF"))

const (
	KEY = iota
	DAT
	ENC
)

var sampleData = func() [][3]string {
	out := [][3]string{
		{"0123456789ABCDEF", "MiraiGO Here", "b7b2e52af7f5b1fbf37fc3d5546ac7569aecd01bbacf09bf"},
		{"0123456789ABCDEF", "LXY Testing~", "9d0ab85aa14f5434ee83cd2a6b28bf306263cdf88e01264c"},

		{"0123456789ABCDEF", "s", "528e8b5c48300b548e94262736ebb8b7"},
		{"0123456789ABCDEF", "long long long long long long long", "95715fab6efbd0fd4b76dbc80bd633ebe805849dbc242053b06557f87e748effd9f613f782749fb9fdfa3f45c0c26161"},

		{"LXY1226    Mrs4s", "LXY Testing~", "ab20caa63f3a6503a84f3cb28f9e26b6c18c051e995d1721"},
	}
	for i, _ := range out {
		c, _ := hex.DecodeString(out[i][ENC])
		out[i][ENC] = utils.B2S(c)
	}
	return out
}()

func TestTEA(t *testing.T) {
	// Self Testing
	for _, sample := range sampleData {
		tea := NewTeaCipher(utils.S2B(sample[KEY]))
		dat := utils.B2S(tea.Decrypt(utils.S2B(sample[ENC])))
		if dat != sample[DAT] {
			t.Fatalf("error decrypt %v %x", sample, dat)
		}
		enc := utils.B2S(tea.Encrypt(utils.S2B(sample[DAT])))
		dat = utils.B2S(tea.Decrypt(utils.S2B(enc)))
		if dat != sample[DAT] {
			t.Fatal("error self test", sample)
		}
	}

	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	// Random data testing
	for i := 1; i < 0xFF; i++ {
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		tea := NewTeaCipher(key)

		dat := make([]byte, i)
		_, err = rand.Read(dat)
		if err != nil {
			panic(err)
		}
		enc := tea.Encrypt(dat)
		dec := tea.Decrypt(enc)
		if !bytes.Equal(dat, dec) {
			t.Fatalf("error in %d, %x %x %x", i, key, dat, enc)
		}
	}
}

func BenchmarkTEAen16(b *testing.B) {
	data := make([]byte, 16)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTEA.Encrypt(data)
	}
}

func BenchmarkTEAde16(b *testing.B) {
	data := make([]byte, 16)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	data = testTEA.Encrypt(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTEA.Decrypt(data)
	}
}

func BenchmarkTEAen256(b *testing.B) {
	data := make([]byte, 256)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTEA.Encrypt(data)
	}
}

func BenchmarkTEAde256(b *testing.B) {
	data := make([]byte, 256)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	data = testTEA.Encrypt(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTEA.Decrypt(data)
	}
}
