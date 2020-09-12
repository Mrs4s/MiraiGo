package client

import "fmt"

type (
	GroupPokeNotifyEvent struct {
		GroupCode int64
		Sender    int64
		Receiver  int64
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
