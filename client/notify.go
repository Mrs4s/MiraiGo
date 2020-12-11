package client

import (
	"fmt"
	"strconv"

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

	// FriendPokeNotifyEvent 好友戳一戳提示事件
	FriendPokeNotifyEvent struct {
		Sender   int64
		Receiver int64
	}
)

// grayTipProcessor 提取出来专门用于处理群内 notify tips
func (c *QQClient) grayTipProcessor(groupId int64, tipInfo *notify.GeneralGrayTipInfo) {
	switch tipInfo.TemplId {
	case 10043, 1136, 1132: // 戳一戳
		var sender int64 = 0
		receiver := c.Uin
		for _, templ := range tipInfo.MsgTemplParam {
			if templ.Name == "uin_str1" {
				sender, _ = strconv.ParseInt(templ.Value, 10, 64)
			}
			if templ.Name == "uin_str2" {
				receiver, _ = strconv.ParseInt(templ.Value, 10, 64)
			}
		}
		c.dispatchGroupNotifyEvent(&GroupPokeNotifyEvent{
			GroupCode: groupId,
			Sender:    sender,
			Receiver:  receiver,
		})
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
		c.dispatchGroupNotifyEvent(&MemberHonorChangedNotifyEvent{
			GroupCode: groupId,
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
