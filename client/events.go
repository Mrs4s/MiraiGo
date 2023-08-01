package client

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"

	"github.com/Mrs4s/MiraiGo/message"
)

// protected all EventHandle, since write is very rare, use
// only one lock to save memory
var eventMu sync.RWMutex

type EventHandle[T any] struct {
	// QQClient?
	handlers []func(client *QQClient, event T)
}

func (handle *EventHandle[T]) Subscribe(handler func(client *QQClient, event T)) {
	eventMu.Lock()
	defer eventMu.Unlock()
	// shrink the slice
	newHandlers := make([]func(client *QQClient, event T), len(handle.handlers)+1)
	copy(newHandlers, handle.handlers)
	newHandlers[len(handle.handlers)] = handler
	handle.handlers = newHandlers
}

func (handle *EventHandle[T]) dispatch(client *QQClient, event T) {
	eventMu.RLock()
	defer func() {
		eventMu.RUnlock()
		if pan := recover(); pan != nil {
			fmt.Printf("event error: %v\n%s", pan, debug.Stack())
		}
	}()
	for _, handler := range handle.handlers {
		handler(client, event)
	}
	if len(client.eventHandlers.subscribedEventHandlers) > 0 {
		for _, h := range client.eventHandlers.subscribedEventHandlers {
			ht := reflect.TypeOf(h)
			for i := 0; i < ht.NumMethod(); i++ {
				method := ht.Method(i)
				if method.Type.NumIn() != 3 {
					continue
				}
				if method.Type.In(1) != reflect.TypeOf(client) || method.Type.In(2) != reflect.TypeOf(event) {
					continue
				}
				method.Func.Call([]reflect.Value{reflect.ValueOf(h), reflect.ValueOf(client), reflect.ValueOf(event)})
			}
		}
	}
}

type eventHandlers struct {
	// todo: move to event handle
	guildChannelMessageHandlers          []func(*QQClient, *message.GuildChannelMessage)
	guildMessageReactionsUpdatedHandlers []func(*QQClient, *GuildMessageReactionsUpdatedEvent)
	guildMessageRecalledHandlers         []func(*QQClient, *GuildMessageRecalledEvent)
	guildChannelUpdatedHandlers          []func(*QQClient, *GuildChannelUpdatedEvent)
	guildChannelCreatedHandlers          []func(*QQClient, *GuildChannelOperationEvent)
	guildChannelDestroyedHandlers        []func(*QQClient, *GuildChannelOperationEvent)
	memberJoinedGuildHandlers            []func(*QQClient, *MemberJoinGuildEvent)

	serverUpdatedHandlers       []func(*QQClient, *ServerUpdatedEvent) bool
	subscribedEventHandlers     []any
	groupMessageReceiptHandlers sync.Map
}

func (c *QQClient) SubscribeEventHandler(handler any) {
	c.eventHandlers.subscribedEventHandlers = append(c.eventHandlers.subscribedEventHandlers, handler)
}

func (s *GuildService) OnGuildChannelMessage(f func(*QQClient, *message.GuildChannelMessage)) {
	s.c.eventHandlers.guildChannelMessageHandlers = append(s.c.eventHandlers.guildChannelMessageHandlers, f)
}

func (s *GuildService) OnGuildMessageReactionsUpdated(f func(*QQClient, *GuildMessageReactionsUpdatedEvent)) {
	s.c.eventHandlers.guildMessageReactionsUpdatedHandlers = append(s.c.eventHandlers.guildMessageReactionsUpdatedHandlers, f)
}

func (s *GuildService) OnGuildMessageRecalled(f func(*QQClient, *GuildMessageRecalledEvent)) {
	s.c.eventHandlers.guildMessageRecalledHandlers = append(s.c.eventHandlers.guildMessageRecalledHandlers, f)
}

func (s *GuildService) OnGuildChannelUpdated(f func(*QQClient, *GuildChannelUpdatedEvent)) {
	s.c.eventHandlers.guildChannelUpdatedHandlers = append(s.c.eventHandlers.guildChannelUpdatedHandlers, f)
}

func (s *GuildService) OnGuildChannelCreated(f func(*QQClient, *GuildChannelOperationEvent)) {
	s.c.eventHandlers.guildChannelCreatedHandlers = append(s.c.eventHandlers.guildChannelCreatedHandlers, f)
}

func (s *GuildService) OnGuildChannelDestroyed(f func(*QQClient, *GuildChannelOperationEvent)) {
	s.c.eventHandlers.guildChannelDestroyedHandlers = append(s.c.eventHandlers.guildChannelDestroyedHandlers, f)
}

func (s *GuildService) OnMemberJoinedGuild(f func(*QQClient, *MemberJoinGuildEvent)) {
	s.c.eventHandlers.memberJoinedGuildHandlers = append(s.c.eventHandlers.memberJoinedGuildHandlers, f)
}

func (c *QQClient) OnServerUpdated(f func(*QQClient, *ServerUpdatedEvent) bool) {
	c.eventHandlers.serverUpdatedHandlers = append(c.eventHandlers.serverUpdatedHandlers, f)
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

func (c *QQClient) dispatchGuildChannelMessage(msg *message.GuildChannelMessage) {
	if msg == nil {
		return
	}
	for _, f := range c.eventHandlers.guildChannelMessageHandlers {
		cover(func() {
			f(c, msg)
		})
	}
}

func (c *QQClient) dispatchGuildMessageReactionsUpdatedEvent(e *GuildMessageReactionsUpdatedEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.guildMessageReactionsUpdatedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGuildMessageRecalledEvent(e *GuildMessageRecalledEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.guildMessageRecalledHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGuildChannelUpdatedEvent(e *GuildChannelUpdatedEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.guildChannelUpdatedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGuildChannelCreatedEvent(e *GuildChannelOperationEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.guildChannelCreatedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGuildChannelDestroyedEvent(e *GuildChannelOperationEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.guildChannelDestroyedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchMemberJoinedGuildEvent(e *MemberJoinGuildEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.memberJoinedGuildHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGroupMessageReceiptEvent(e *groupMessageReceiptEvent) {
	c.eventHandlers.groupMessageReceiptHandlers.Range(func(_, f any) bool {
		go f.(func(*QQClient, *groupMessageReceiptEvent))(c, e)
		return true
	})
}

func cover(f func()) {
	defer func() {
		if pan := recover(); pan != nil {
			fmt.Printf("event error: %v\n%s", pan, debug.Stack())
		}
	}()
	f()
}
