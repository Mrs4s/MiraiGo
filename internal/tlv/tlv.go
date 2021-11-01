package tlv

func GuidFlag() uint32 {
	var flag uint32
	flag |= 1 << 24 & 0xFF000000
	flag |= 0 << 8 & 0xFF00
	return flag
}
