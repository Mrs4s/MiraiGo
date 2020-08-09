package utils

// GetAmrDuration - return .amr file duration form bytesData
func GetAmrDuration(bytesData []byte) int32 {
	duration := -1
	packedSize := [16]int{12, 13, 15, 17, 19, 20, 26, 31, 5, 0, 0, 0, 0, 0, 0, 0}

	length := len(bytesData)
	pos := 6
	var datas []byte
	frameCount := 0
	packedPos := -1

	for pos <= length {
		datas = bytesData[pos : pos+1]
		if len(datas) != 1 {
			if length > 0 {
				duration = (length - 6) / 650
			} else {
				duration = 0
			}
			break
		}
		packedPos = int((datas[0] >> 3) & 0x0F)
		pos += packedSize[packedPos] + 1
		frameCount++
	}

	duration += frameCount * 20
	return int32(duration/1000 + 1)
}
