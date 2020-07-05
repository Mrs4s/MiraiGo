package client

import "github.com/Mrs4s/MiraiGo/message"

func (c *QQClient) OnPrivateMessage(f func(*QQClient, *message.PrivateMessage)) {
	c.privateMessageHandlers = append(c.privateMessageHandlers, f)
}

func (c *QQClient) OnPrivateMessageF(filter func(*message.PrivateMessage) bool, f func(*QQClient, *message.PrivateMessage)) {
	c.privateMessageHandlers = append(c.privateMessageHandlers, func(client *QQClient, msg *message.PrivateMessage) {
		if filter(msg) {
			f(client, msg)
		}
	})
}

func (c *QQClient) OnGroupMessage(f func(*QQClient, *message.GroupMessage)) {
	c.groupMessageHandlers = append(c.groupMessageHandlers, f)
}

func NewUinFilterPrivate(uin int64) func(*message.PrivateMessage) bool {
	return func(msg *message.PrivateMessage) bool {
		return msg.Sender.Uin == uin
	}
}

func (c *QQClient) dispatchFriendMessage(msg *message.PrivateMessage) {
	if msg == nil {
		return
	}
	for _, f := range c.privateMessageHandlers {
		func() {
			defer func() {
				if pan := recover(); pan != nil {
					//
				}
			}()
			f(c, msg)
		}()
	}
}

func (c *QQClient) dispatchGroupMessage(msg *message.GroupMessage) {
	if msg == nil {
		return
	}
	for _, f := range c.groupMessageHandlers {
		func() {
			defer func() {
				if pan := recover(); pan != nil {
					// TODO: logger
				}
			}()
			f(c, msg)
		}()
	}
}
