package tlv

import (
	"crypto/md5"
	binary2 "encoding/binary"
	"github.com/Mrs4s/MiraiGo/binary"
	"math/rand"
	"strconv"
	"time"
)

func T106(uin, salt, appId, ssoVer uint32, passwordMd5 [16]byte, guidAvailable bool, guid, tgtgtKey []byte, wtf uint32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x106)
		body := binary.NewWriterF(func(w *binary.Writer) {
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
			w.WriteUInt32(uint32(time.Now().UnixNano() / 1e6))
			w.Write([]byte{0x00, 0x00, 0x00, 0x00}) // fake ip
			w.WriteByte(0x01)
			w.Write(passwordMd5[:])
			w.Write(tgtgtKey)
			w.WriteUInt32(wtf)
			w.WriteBool(guidAvailable)
			if len(guid) == 0 {
				for i := 0; i < 4; i++ {
					w.WriteUInt32(rand.Uint32())
				}
			} else {
				w.Write(guid)
			}
			w.WriteUInt32(appId)
			w.WriteUInt32(1) // password login
			w.WriteTlv([]byte(strconv.FormatInt(int64(uin), 10)))
			w.WriteUInt16(0)
		})
		w.WriteTlv(binary.NewWriterF(func(w *binary.Writer) {
			b := make([]byte, 4)
			if salt != 0 {
				binary2.BigEndian.PutUint32(b, salt)
			} else {
				binary2.BigEndian.PutUint32(b, uin)
			}
			key := md5.Sum(append(append(passwordMd5[:], []byte{0x00, 0x00, 0x00, 0x00}...), b...))
			w.EncryptAndWrite(key[:], body)
		}))
	})
}
