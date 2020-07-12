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
	case left >= 310 && left <= 499:
		left += 3800 - 310
	}
	return left*1000000 + groupCode%1000000
}

func ToGroupCode(groupUin int64) int64 {
	left := groupUin / 1000000
	switch {
	case left >= 0+202 && left <= 10+202:
		left -= 202
	case left >= 11+480-11 && left <= 19+480-11:
		left -= 480 - 11
	case left >= 20+2100-20 && left <= 66+2100-20:
		left -= 2100 - 20
	case left >= 67+2010-67 && left <= 156+2010-67:
		left -= 2010 - 67
	case left >= 157+2147-157 && left <= 209+2147-157:
		left -= 2147 - 157
	case left >= 210+4100-210 && left <= 309+4100-210:
		left -= 4100 - 210
	case left >= 310+3800-310 && left <= 499+3800-310:
		left -= 3800 - 310
	}
	return left*1000000 + groupUin%1000000
}
