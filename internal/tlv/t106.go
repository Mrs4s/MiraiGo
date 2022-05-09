package tlv

import (
	"crypto/md5"
	"math/rand"
	"strconv"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T106(uin, salt, appId, ssoVer uint32, passwordMd5 [16]byte, guidAvailable bool, guid, tgtgtKey []byte, wtf uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x106)
		pos := w.FillUInt16()
		keydata, kcl := binary.OpenWriterF(func(w *binary.Writer) {
			w.Write(passwordMd5[:])
			w.WriteUInt32(0) // []byte{0x00, 0x00, 0x00, 0x00}...
			if salt != 0 {
				w.WriteUInt32(salt)
			} else {
				w.WriteUInt32(uin)
			}
		})
		key := md5.Sum(keydata)
		kcl()
		body, cl := binary.OpenWriterF(func(w *binary.Writer) {
			w.WriteUInt16(4)
			w.WriteUInt32(rand.Uint32())
			w.WriteUInt32(ssoVer)
			w.WriteUInt32(16) // appId
			w.WriteUInt32(0)  // app client version
			if uin == 0 {
				w.WriteUInt64(uint64(salt))
			} else {
				w.WriteUInt64(uint64(uin))
			}
			w.WriteUInt32(uint32(time.Now().Unix()))
			w.WriteUInt32(0) // fake ip w.Write([]byte{0x00, 0x00, 0x00, 0x00})
			w.WriteByte(0x01)
			w.Write(passwordMd5[:])
			w.Write(tgtgtKey)
			w.WriteUInt32(wtf)
			w.WriteBool(guidAvailable)
			if len(guid) == 0 {
				w.WriteUInt32(rand.Uint32())
				w.WriteUInt32(rand.Uint32())
				w.WriteUInt32(rand.Uint32())
				w.WriteUInt32(rand.Uint32())
			} else {
				w.Write(guid)
			}
			w.WriteUInt32(appId)
			w.WriteUInt32(1) // password login
			w.WriteStringShort(strconv.FormatInt(int64(uin), 10))
			w.WriteUInt16(0)
		})
		w.EncryptAndWrite(key[:], body)
		w.WriteUInt16At(pos, uint16(w.Len()-4))
		cl()
	})
}
