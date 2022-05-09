package utils

func ToGroupUin(groupCode int64) int64 {
	left := groupCode / 1000000
	switch {
	case left >= 0 && left <= 10:
		left += 202
	case left >= 11 && left <= 19:
		left += 480 - 11
	case left >= 20 && left <= 66:
		left += 2100 - 20
	case left >= 67 && left <= 156:
		left += 2010 - 67
	case left >= 157 && left <= 209:
		left += 2147 - 157
	case left >= 210 && left <= 309:
		left += 4100 - 210
	case left >= 310 && left <= 335:
		left += 3800 - 310
	case left >= 336 && left <= 386:
		left += 2265
	case left >= 387 && left <= 499:
		left += 3490
	}
	return left*1000000 + groupCode%1000000
}

func ToGroupCode(groupUin int64) int64 {
	left := groupUin / 1000000
	switch {
	case left >= 202 && left <= 212:
		left -= 202
	case left >= 480 && left <= 488:
		left -= 480 - 11
	case left >= 2100 && left <= 2146:
		left -= 2100 - 20
	case left >= 2010 && left <= 2099:
		left -= 2010 - 67
	case left >= 2147 && left <= 2199:
		left -= 2147 - 157
	case left >= 2600 && left <= 2651:
		left -= 2265
	case left >= 4100 && left <= 4199:
		left -= 4100 - 210
	case left >= 3800 && left <= 3989:
		left -= 3800 - 310
	}
	return left*1000000 + groupUin%1000000
}
