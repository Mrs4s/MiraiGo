package client

type (
	HonorType int

	GroupHonorInfo struct {
		GroupCode        string            `json:"gc"`
		Uin              string            `json:"uin"`
		Type             HonorType         `json:"type"`
		TalkativeList    []HonorMemberInfo `json:"talkativeList"`
		CurrentTalkative CurrentTalkative  `json:"currentTalkative"`
		ActorList        []HonorMemberInfo `json:"actorList"`
		LegendList       []HonorMemberInfo `json:"legendList"`
		StrongNewbieList []HonorMemberInfo `json:"strongnewbieList"`
		EmotionList      []HonorMemberInfo `json:"emotionList"`
	}

	HonorMemberInfo struct {
		Uin    int64  `json:"uin"`
		Avatar string `json:"avatar"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
	}

	CurrentTalkative struct {
		Uin      int64  `json:"uin"`
		DayCount int32  `json:"day_count"`
		Avatar   string `json:"avatar"`
		Name     string `json:"nick"`
	}
)

const (
	Talkative    HonorType = 1 // 龙王
	Performer    HonorType = 2 // 群聊之火
	Legend       HonorType = 3 // 群聊炙焰
	StrongNewbie HonorType = 5 // 冒尖小春笋
	Emotion      HonorType = 6 // 快乐源泉
)
