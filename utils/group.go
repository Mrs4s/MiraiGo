package utils

var (
	uinRanges = []int64{
		0, 0,
		11, 202,
		20, 480,
		67, 2100,
		157, 2010,
		210, 2147,
		310, 4100,
		500, 3800,
	}
	uinDivisor int64 = 1000000
)

func ToGroupUin(groupCode int64) int64 {
	left := groupCode / uinDivisor
	for i := 2; i < len(uinRanges); i += 2 {
		diff := uinRanges[i+1] - uinRanges[i-2]
		if left < uinRanges[i] {
			left += diff
			break
		}
	}
	return left*uinDivisor + groupCode%uinDivisor
}

func ToGroupCode(groupUin int64) int64 {
	left := groupUin / uinDivisor
	for i := 2; i < len(uinRanges); i += 2 {
		diff := uinRanges[i+1] - uinRanges[i-2]
		if uinRanges[i+1] <= left && left < uinRanges[i]+diff {
			left -= diff
			break
		}
	}
	return left*uinDivisor + groupUin%uinDivisor
}
