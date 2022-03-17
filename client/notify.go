package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/utils"

	"github.com/Mrs4s/MiraiGo/client/pb/notify"
)

type (
	// GroupPokeNotifyEvent 群内戳一戳提示事件
	GroupPokeNotifyEvent struct {
		GroupCode int64
		Sender    int64
		Receiver  int64
	}

	// GroupRedBagLuckyKingNotifyEvent 群内抢红包运气王提示事件
	GroupRedBagLuckyKingNotifyEvent struct {
		GroupCode int64
		Sender    int64
		LuckyKing int64
	}

	// MemberHonorChangedNotifyEvent 群成员荣誉变更提示事件
	MemberHonorChangedNotifyEvent struct {
		GroupCode int64
		Honor     HonorType
		Uin       int64
		Nick      string
	}

	// MemberSpecialTitleUpdatedEvent 群成员头衔更新事件
	MemberSpecialTitleUpdatedEvent struct {
		GroupCode int64
		Uin       int64
		NewTitle  string
	}

	// FriendPokeNotifyEvent 好友戳一戳提示事件
	FriendPokeNotifyEvent struct {
		Sender   int64
		Receiver int64
	}
)

// grayTipProcessor 提取出来专门用于处理群内 notify tips
func (c *QQClient) grayTipProcessor(groupCode int64, tipInfo *notify.GeneralGrayTipInfo) {
	if tipInfo.BusiType == 12 && tipInfo.BusiId == 1061 {
		sender := int64(0)
		receiver := c.Uin
		for _, templ := range tipInfo.MsgTemplParam {
			if templ.Name == "uin_str1" {
				sender, _ = strconv.ParseInt(templ.Value, 10, 64)
			}
			if templ.Name == "uin_str2" {
				receiver, _ = strconv.ParseInt(templ.Value, 10, 64)
			}
		}
		if sender != 0 {
			c.GroupNotifyEvent.dispatch(c, &GroupPokeNotifyEvent{
				GroupCode: groupCode,
				Sender:    sender,
				Receiver:  receiver,
			})
		}
	}
	switch tipInfo.TemplId {
	case 1052, 1053, 1054, 1067: // 群荣誉
		var nick string
		var uin int64
		for _, templ := range tipInfo.MsgTemplParam {
			if templ.Name == "nick" {
				nick = templ.Value
			}
			if templ.Name == "uin" {
				uin, _ = strconv.ParseInt(templ.Value, 10, 64)
			}
		}
		c.GroupNotifyEvent.dispatch(c, &MemberHonorChangedNotifyEvent{
			GroupCode: groupCode,
			Honor: func() HonorType {
				switch tipInfo.TemplId {
				case 1052:
					return Performer
				case 1053, 1054:
					return Talkative
				case 1067:
					return Emotion
				default:
					return 0
				}
			}(),
			Uin:  uin,
			Nick: nick,
		})
	}
}

// msgGrayTipProcessor 用于处理群内 aio notify tips
func (c *QQClient) msgGrayTipProcessor(groupCode int64, tipInfo *notify.AIOGrayTipsInfo) {
	if len(tipInfo.Content) == 0 {
		return
	}
	type tipCommand struct {
		Command int    `json:"cmd"`
		Data    string `json:"data"`
		Text    string `json:"text"`
	}
	content := utils.B2S(tipInfo.Content)
	var tipCmds []*tipCommand
	start := -1
	for i := 0; i < len(content); i++ {
		if content[i] == '<' && len(content) > i+1 && content[i+1] == '{' {
			start = i + 1
		}
		if content[i] == '>' && content[i-1] == '}' && start != -1 {
			tip := &tipCommand{}
			if err := json.Unmarshal(utils.S2B(content[start:i]), tip); err == nil {
				tipCmds = append(tipCmds, tip)
			}
			start = -1
		}
	}
	// 好像只能这么判断
	switch {
	case strings.Contains(content, "头衔"):
		event := &MemberSpecialTitleUpdatedEvent{GroupCode: groupCode}
		for _, cmd := range tipCmds {
			if cmd.Command == 5 {
				event.Uin, _ = strconv.ParseInt(cmd.Data, 10, 64)
			}
			if cmd.Command == 1 {
				event.NewTitle = cmd.Text
			}
		}
		if event.Uin == 0 {
			c.error("process special title updated tips error: missing cmd")
			return
		}
		if mem := c.FindGroup(groupCode).FindMember(event.Uin); mem != nil {
			mem.SpecialTitle = event.NewTitle
		}
		c.MemberSpecialTitleUpdatedEvent.dispatch(c, event)
	}
}

func (e *GroupPokeNotifyEvent) From() int64 {
	return e.GroupCode
}

func (e *GroupPokeNotifyEvent) Content() string {
	return fmt.Sprintf("%d戳了戳%d", e.Sender, e.Receiver)
}

func (e *FriendPokeNotifyEvent) From() int64 {
	return e.Sender
}

func (e *FriendPokeNotifyEvent) Content() string {
	return fmt.Sprintf("%d戳了戳%d", e.Sender, e.Receiver)
}

func (e *GroupRedBagLuckyKingNotifyEvent) From() int64 {
	return e.GroupCode
}

func (e *GroupRedBagLuckyKingNotifyEvent) Content() string {
	return fmt.Sprintf("%d发的红包被领完, %d是运气王", e.Sender, e.LuckyKing)
}

func (e *MemberHonorChangedNotifyEvent) From() int64 {
	return e.GroupCode
}

func (e *MemberHonorChangedNotifyEvent) Content() string {
	switch e.Honor {
	case Talkative:
		return fmt.Sprintf("昨日 %s(%d) 在群 %d 内发言最积极, 获得 龙王 标识。", e.Nick, e.Uin, e.GroupCode)
	case Performer:
		return fmt.Sprintf("%s(%d) 在群 %d 里连续发消息超过7天, 获得 群聊之火 标识。", e.Nick, e.Uin, e.GroupCode)
	case Emotion:
		return fmt.Sprintf("%s(%d) 在群聊 %d 中连续发表情包超过3天，且累计数量超过20条，获得 快乐源泉 标识。", e.Nick, e.Uin, e.GroupCode)
	}
	return "ERROR"
}
