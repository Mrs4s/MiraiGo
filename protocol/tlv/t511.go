package tlv

import (
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/binary"
)

func T511(domains []string) []byte {
	var arr2 []string
	for _, d := range domains {
		if d != "" {
			arr2 = append(arr2, d)
		}
	}
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x511)
		w.WriteTlv(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteUInt16(uint16(len(arr2)))
			for _, d := range arr2 {
				indexOf := strings.Index(d, "(")
				indexOf2 := strings.Index(d, ")")
				if indexOf != 0 || indexOf2 <= 0 {
					w.WriteByte(0x01)
					w.WriteTlv([]byte(d))
				} else {
					var b byte
					var z bool
					i, err := strconv.Atoi(d[indexOf+1 : indexOf2])
					if err == nil {
						z2 := (1048576 & i) > 0
						if (i & 134217728) > 0 {
							z = true
						} else {
							z = false
						}
						if z2 {
							b = 1
						} else {
							b = 0
						}
						if z {
							b |= 2
						}
						w.WriteByte(b)
						w.WriteTlv([]byte(d[indexOf2+1:]))
					}
				}
			}
		}))
	})
}
