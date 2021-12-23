package tlv

import (
	"github.com/Mrs4s/MiraiGo/binary"
)

func T511(domains []string) []byte {
	nonnildomains := domains
	// 目前的所有调用均无 ""
	/*
		hasnil := false
		for _, d := range domains {
			if d == "" {
				hasnil = true
				break
			}
		}
		if hasnil {
			nonnildomains = nonnildomains[:0]
			for _, d := range domains {
				if d != "" {
					nonnildomains = append(nonnildomains, d)
				}
			}
		}
	*/
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt16(0x511)
		pos := w.FillUInt16()
		w.WriteUInt16(uint16(len(nonnildomains)))
		for _, d := range nonnildomains {
			// 目前的所有调用均不会出现 ()
			// indexOf := strings.Index(d, "(")
			// indexOf2 := strings.Index(d, ")")
			// if indexOf != 0 || indexOf2 <= 0 {
			w.WriteByte(0x01)
			w.WriteStringShort(d)
			/* } else {
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
					w.WriteBytesShort([]byte(d[indexOf2+1:]))
				}
			}*/
		}
		w.WriteUInt16At(pos, uint16(w.Len()-4))
	})
}
