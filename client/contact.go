package client

import "github.com/Mrs4s/MiraiGo/message"

type Contact[T any] interface {
	Name() string
	AvatarUrl() string
	Message(*message.SendingMessage) (MessageReceipt, bool)
	// Delete means quit the group or remove the friend
	Delete()
}
