package client

import (
	"github.com/Mrs4s/MiraiGo/message"
	"sync"
)

type eventHandlers struct {
	privateMessageHandlers      []func(*QQClient, *message.PrivateMessage)
	groupMessageHandlers        []func(*QQClient, *message.GroupMessage)
	groupMuteEventHandlers      []func(*QQClient, *GroupMuteEvent)
	groupRecalledHandlers       []func(*QQClient, *GroupMessageRecalledEvent)
	joinGroupHandlers           []func(*QQClient, *GroupInfo)
	leaveGroupHandlers          []func(*QQClient, *GroupLeaveEvent)
	memberJoinedHandlers        []func(*QQClient, *GroupInfo, *GroupMemberInfo)
	groupMessageReceiptHandlers sync.Map
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

func (c *QQClient) OnJoinGroup(f func(*QQClient, *GroupInfo)) {
	c.eventHandlers.joinGroupHandlers = append(c.eventHandlers.joinGroupHandlers, f)
}

func (c *QQClient) OnLeaveGroup(f func(*QQClient, *GroupLeaveEvent)) {
	c.eventHandlers.leaveGroupHandlers = append(c.eventHandlers.leaveGroupHandlers, f)
}

func (c *QQClient) OnGroupMemberJoined(f func(*QQClient, *GroupInfo, *GroupMemberInfo)) {
	c.eventHandlers.memberJoinedHandlers = append(c.eventHandlers.memberJoinedHandlers, f)
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

func (c *QQClient) dispatchJoinGroupEvent(group *GroupInfo) {
	if group == nil {
		return
	}
	for _, f := range c.eventHandlers.joinGroupHandlers {
		cover(func() {
			f(c, group)
		})
	}
}

func (c *QQClient) dispatchLeaveGroupEvent(e *GroupLeaveEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.leaveGroupHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchNewMemberEvent(group *GroupInfo, mem *GroupMemberInfo) {
	if group == nil || mem == nil {
		return
	}
	for _, f := range c.eventHandlers.memberJoinedHandlers {
		cover(func() {
			f(c, group, mem)
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
