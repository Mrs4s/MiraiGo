package topic

import (
	"strconv"
)

type (
	TextElement struct {
		Content string
	}

	EmojiElement struct {
		Index int32
		Id    string
		Name  string
	}

	AtElement struct {
		Id       string
		TinyId   uint64
		Nickname string
	}

	ChannelQuoteElement struct {
		GuildId     uint64
		ChannelId   uint64
		DisplayText string
	}

	UrlQuoteElement struct {
		Url         string
		DisplayText string
	}
)

func selectContent(b bool, c1, c2 content) content {
	if b {
		return c1
	}
	return c2
}

func (e *TextElement) pack(patternId string, isPatternData bool) content {
	return selectContent(isPatternData,
		content{
			"type":     1,
			"style":    "n",
			"text":     e.Content,
			"children": make([]int, 0),
		},
		content{
			"type": 1,
			"text_content": content{
				"text": e.Content,
			},
		})
}

func (e *EmojiElement) pack(patternId string, isPatternData bool) content {
	return selectContent(isPatternData,
		content{
			"type":      2,
			"id":        patternId,
			"emojiType": "1",
			"emojiId":   e.Id,
		},
		content{
			"type":       4,
			"pattern_id": patternId,
			"emoji_content": content{
				"type": "1",
				"id":   e.Id,
			},
		})
}

func (e *AtElement) pack(patternId string, isPatternData bool) content {
	return selectContent(isPatternData,
		content{
			"type": 3,
			"id":   patternId,
			"user": content{
				"id":   strconv.FormatUint(e.TinyId, 10),
				"nick": e.Nickname,
			},
		},
		content{
			"type":       2,
			"pattern_id": patternId,
			"at_content": content{
				"type": 1,
				"user": content{
					"id":   e.Id,
					"nick": e.Nickname,
				},
			},
		})
}

func (e *ChannelQuoteElement) pack(patternId string, isPatternData bool) content {
	return selectContent(isPatternData,
		content{
			"type": 4,
			"id":   patternId,
			"guild_info": content{
				"channel_id": strconv.FormatUint(e.ChannelId, 10),
				"name":       e.DisplayText,
			},
		},
		content{
			"type":       5,
			"pattern_id": patternId,
			"channel_content": content{
				"channel_info": content{
					"name": e.DisplayText,
					"sign": content{
						"guild_id":   strconv.FormatUint(e.GuildId, 10),
						"channel_id": strconv.FormatUint(e.ChannelId, 10),
					},
				},
			},
		})
}

func (e *UrlQuoteElement) pack(patternId string, isPatternData bool) content {
	return selectContent(isPatternData,
		content{
			"type": 5,
			"desc": e.DisplayText,
			"href": e.Url,
			"id":   patternId,
		},
		content{
			"type":       3,
			"pattern_id": patternId,
			"url_content": content{
				"url":         e.Url,
				"displayText": e.DisplayText,
			},
		})
}
