package client

import "fmt"

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
)

func (e *GroupPokeNotifyEvent) From() int64 {
	return e.GroupCode
}

func (e *GroupPokeNotifyEvent) Name() string {
	return "戳一戳"
}

func (e *GroupPokeNotifyEvent) Content() string {
	return fmt.Sprintf("%d戳了戳%d", e.Sender, e.Receiver)
}

func (e *GroupRedBagLuckyKingNotifyEvent) From() int64 {
	return e.GroupCode
}

func (e *GroupRedBagLuckyKingNotifyEvent) Name() string {
	return "运气王"
}

func (e *GroupRedBagLuckyKingNotifyEvent) Content() string {
	return fmt.Sprintf("%d发的红包被领完, %d是运气王", e.Sender, e.LuckyKing)
}
