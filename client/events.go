package client

import (
	"errors"
	"github.com/Mrs4s/MiraiGo/message"
	"sync"
)

var ErrEventUndefined = errors.New("event undefined")

type eventHandlers struct {
	privateMessageHandlers      []func(*QQClient, *message.PrivateMessage)
	groupMessageHandlers        []func(*QQClient, *message.GroupMessage)
	groupMuteEventHandlers      []func(*QQClient, *GroupMuteEvent)
	groupRecalledHandlers       []func(*QQClient, *GroupMessageRecalledEvent)
	groupMessageReceiptHandlers sync.Map
}

func (c *QQClient) OnEvent(i interface{}) error {
	switch f := i.(type) {
	case func(*QQClient, *message.PrivateMessage):
		c.OnPrivateMessage(f)
	case func(*QQClient, *message.GroupMessage):
		c.OnGroupMessage(f)
	case func(*QQClient, *GroupMuteEvent):
		c.OnGroupMuted(f)
	case func(*QQClient, *GroupMessageRecalledEvent):
		c.OnGroupMessageRecalled(f)
	default:
		return ErrEventUndefined
	}
	return nil
}

func (c *QQClient) OnPrivateMessage(f func(*QQClient, *message.PrivateMessage)) {
	c.eventHandlers.privateMessageHandlers = append(c.eventHandlers.privateMessageHandlers, f)
}

func (c *QQClient) OnPrivateMessageF(filter func(*message.PrivateMessage) bool, f func(*QQClient, *message.PrivateMessage)) {
	c.OnPrivateMessage(func(client *QQClient, msg *message.PrivateMessage) {
		if filter(msg) {
			f(client, msg)
		}
	})
}

func (c *QQClient) OnGroupMessage(f func(*QQClient, *message.GroupMessage)) {
	c.eventHandlers.groupMessageHandlers = append(c.eventHandlers.groupMessageHandlers, f)
}

func (c *QQClient) OnGroupMuted(f func(*QQClient, *GroupMuteEvent)) {
	c.eventHandlers.groupMuteEventHandlers = append(c.eventHandlers.groupMuteEventHandlers, f)
}

func (c *QQClient) OnGroupMessageRecalled(f func(*QQClient, *GroupMessageRecalledEvent)) {
	c.eventHandlers.groupRecalledHandlers = append(c.eventHandlers.groupRecalledHandlers, f)
}

func NewUinFilterPrivate(uin int64) func(*message.PrivateMessage) bool {
	return func(msg *message.PrivateMessage) bool {
		return msg.Sender.Uin == uin
	}
}

func (c *QQClient) onGroupMessageReceipt(id string, f ...func(*QQClient, *groupMessageReceiptEvent)) {
	if len(f) == 0 {
		c.eventHandlers.groupMessageReceiptHandlers.Delete(id)
		return
	}
	c.eventHandlers.groupMessageReceiptHandlers.LoadOrStore(id, f[0])
}

func (c *QQClient) dispatchFriendMessage(msg *message.PrivateMessage) {
	if msg == nil {
		return
	}
	for _, f := range c.eventHandlers.privateMessageHandlers {
		cover(func() {
			f(c, msg)
		})
	}
}

func (c *QQClient) dispatchGroupMessage(msg *message.GroupMessage) {
	if msg == nil {
		return
	}
	for _, f := range c.eventHandlers.groupMessageHandlers {
		cover(func() {
			f(c, msg)
		})
	}
}

func (c *QQClient) dispatchGroupMuteEvent(e *GroupMuteEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.groupMuteEventHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGroupMessageRecalledEvent(e *GroupMessageRecalledEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.groupRecalledHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGroupMessageReceiptEvent(e *groupMessageReceiptEvent) {
	c.eventHandlers.groupMessageReceiptHandlers.Range(func(_, f interface{}) bool {
		go f.(func(*QQClient, *groupMessageReceiptEvent))(c, e)
		return true
	})
}

func cover(f func()) {
	defer func() {
		if pan := recover(); pan != nil {

		}
	}()
	f()
}
