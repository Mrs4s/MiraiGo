package client

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/Mrs4s/MiraiGo/message"
)

type eventHandlers struct {
	privateMessageHandlers      []func(*QQClient, *message.PrivateMessage)
	tempMessageHandlers         []func(*QQClient, *message.TempMessage)
	groupMessageHandlers        []func(*QQClient, *message.GroupMessage)
	groupMuteEventHandlers      []func(*QQClient, *GroupMuteEvent)
	groupRecalledHandlers       []func(*QQClient, *GroupMessageRecalledEvent)
	friendRecalledHandlers      []func(*QQClient, *FriendMessageRecalledEvent)
	joinGroupHandlers           []func(*QQClient, *GroupInfo)
	leaveGroupHandlers          []func(*QQClient, *GroupLeaveEvent)
	memberJoinedHandlers        []func(*QQClient, *MemberJoinGroupEvent)
	memberLeavedHandlers        []func(*QQClient, *MemberLeaveGroupEvent)
	memberCardUpdatedHandlers   []func(*QQClient, *MemberCardUpdatedEvent)
	permissionChangedHandlers   []func(*QQClient, *MemberPermissionChangedEvent)
	groupInvitedHandlers        []func(*QQClient, *GroupInvitedRequest)
	joinRequestHandlers         []func(*QQClient, *UserJoinGroupRequest)
	friendRequestHandlers       []func(*QQClient, *NewFriendRequest)
	newFriendHandlers           []func(*QQClient, *NewFriendEvent)
	disconnectHandlers          []func(*QQClient, *ClientDisconnectedEvent)
	logHandlers                 []func(*QQClient, *LogEvent)
	serverUpdatedHandlers       []func(*QQClient, *ServerUpdatedEvent) bool
	groupNotifyHandlers         []func(*QQClient, INotifyEvent)
	friendNotifyHandlers        []func(*QQClient, INotifyEvent)
	offlineFileHandlers         []func(*QQClient, *OfflineFileEvent)
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

func (c *QQClient) OnTempMessage(f func(*QQClient, *message.TempMessage)) {
	c.eventHandlers.tempMessageHandlers = append(c.eventHandlers.tempMessageHandlers, f)
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

func (c *QQClient) OnGroupMemberJoined(f func(*QQClient, *MemberJoinGroupEvent)) {
	c.eventHandlers.memberJoinedHandlers = append(c.eventHandlers.memberJoinedHandlers, f)
}

func (c *QQClient) OnGroupMemberLeaved(f func(*QQClient, *MemberLeaveGroupEvent)) {
	c.eventHandlers.memberLeavedHandlers = append(c.eventHandlers.memberLeavedHandlers, f)
}

func (c *QQClient) OnGroupMemberCardUpdated(f func(*QQClient, *MemberCardUpdatedEvent)) {
	c.eventHandlers.memberCardUpdatedHandlers = append(c.eventHandlers.memberCardUpdatedHandlers, f)
}

func (c *QQClient) OnGroupMemberPermissionChanged(f func(*QQClient, *MemberPermissionChangedEvent)) {
	c.eventHandlers.permissionChangedHandlers = append(c.eventHandlers.permissionChangedHandlers, f)
}

func (c *QQClient) OnGroupMessageRecalled(f func(*QQClient, *GroupMessageRecalledEvent)) {
	c.eventHandlers.groupRecalledHandlers = append(c.eventHandlers.groupRecalledHandlers, f)
}

func (c *QQClient) OnFriendMessageRecalled(f func(*QQClient, *FriendMessageRecalledEvent)) {
	c.eventHandlers.friendRecalledHandlers = append(c.eventHandlers.friendRecalledHandlers, f)
}

func (c *QQClient) OnGroupInvited(f func(*QQClient, *GroupInvitedRequest)) {
	c.eventHandlers.groupInvitedHandlers = append(c.eventHandlers.groupInvitedHandlers, f)
}

func (c *QQClient) OnUserWantJoinGroup(f func(*QQClient, *UserJoinGroupRequest)) {
	c.eventHandlers.joinRequestHandlers = append(c.eventHandlers.joinRequestHandlers, f)
}

func (c *QQClient) OnNewFriendRequest(f func(*QQClient, *NewFriendRequest)) {
	c.eventHandlers.friendRequestHandlers = append(c.eventHandlers.friendRequestHandlers, f)
}

func (c *QQClient) OnNewFriendAdded(f func(*QQClient, *NewFriendEvent)) {
	c.eventHandlers.newFriendHandlers = append(c.eventHandlers.newFriendHandlers, f)
}

func (c *QQClient) OnDisconnected(f func(*QQClient, *ClientDisconnectedEvent)) {
	c.eventHandlers.disconnectHandlers = append(c.eventHandlers.disconnectHandlers, f)
}

func (c *QQClient) OnServerUpdated(f func(*QQClient, *ServerUpdatedEvent) bool) {
	c.eventHandlers.serverUpdatedHandlers = append(c.eventHandlers.serverUpdatedHandlers, f)
}

func (c *QQClient) OnReceivedOfflineFile(f func(*QQClient, *OfflineFileEvent)) {
	c.eventHandlers.offlineFileHandlers = append(c.eventHandlers.offlineFileHandlers, f)
}

func (c *QQClient) OnLog(f func(*QQClient, *LogEvent)) {
	c.eventHandlers.logHandlers = append(c.eventHandlers.logHandlers, f)
}

func (c *QQClient) OnGroupNotify(f func(*QQClient, INotifyEvent)) {
	c.eventHandlers.groupNotifyHandlers = append(c.eventHandlers.groupNotifyHandlers, f)
}

func (c *QQClient) OnFriendNotify(f func(*QQClient, INotifyEvent)) {
	c.eventHandlers.friendNotifyHandlers = append(c.eventHandlers.friendNotifyHandlers, f)
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

func (c *QQClient) dispatchTempMessage(msg *message.TempMessage) {
	if msg == nil {
		return
	}
	for _, f := range c.eventHandlers.tempMessageHandlers {
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

func (c *QQClient) dispatchFriendMessageRecalledEvent(e *FriendMessageRecalledEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.friendRecalledHandlers {
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

func (c *QQClient) dispatchNewMemberEvent(e *MemberJoinGroupEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.memberJoinedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchMemberLeaveEvent(e *MemberLeaveGroupEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.memberLeavedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchMemberCardUpdatedEvent(e *MemberCardUpdatedEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.memberCardUpdatedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchPermissionChanged(e *MemberPermissionChangedEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.permissionChangedHandlers {
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

func (c *QQClient) dispatchGroupInvitedEvent(e *GroupInvitedRequest) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.groupInvitedHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchJoinGroupRequest(r *UserJoinGroupRequest) {
	if r == nil {
		return
	}
	for _, f := range c.eventHandlers.joinRequestHandlers {
		cover(func() {
			f(c, r)
		})
	}
}

func (c *QQClient) dispatchNewFriendRequest(r *NewFriendRequest) {
	if r == nil {
		return
	}
	for _, f := range c.eventHandlers.friendRequestHandlers {
		cover(func() {
			f(c, r)
		})
	}
}

func (c *QQClient) dispatchNewFriendEvent(e *NewFriendEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.newFriendHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchGroupNotifyEvent(e INotifyEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.groupNotifyHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchFriendNotifyEvent(e INotifyEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.friendNotifyHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchDisconnectEvent(e *ClientDisconnectedEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.disconnectHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchOfflineFileEvent(e *OfflineFileEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.offlineFileHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func (c *QQClient) dispatchLogEvent(e *LogEvent) {
	if e == nil {
		return
	}
	for _, f := range c.eventHandlers.logHandlers {
		cover(func() {
			f(c, e)
		})
	}
}

func cover(f func()) {
	defer func() {
		if pan := recover(); pan != nil {
			fmt.Printf("event error: %v\n%s", pan, debug.Stack())
		}
	}()
	f()
}
