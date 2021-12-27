package client

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

type EventHandler struct {
	PrivateMessageHandler               func(*QQClient, *message.PrivateMessage)
	TempMessageHandler                  func(*QQClient, *TempMessageEvent)
	GroupMessageHandler                 func(*QQClient, *message.GroupMessage)
	SelfPrivateMessageHandler           func(*QQClient, *message.PrivateMessage)
	SelfGroupMessageHandler             func(*QQClient, *message.GroupMessage)
	GuildChannelMessageHandler          func(*QQClient, *message.GuildChannelMessage)
	GuildMessageReactionsUpdatedHandler func(*QQClient, *GuildMessageReactionsUpdatedEvent)
	GuildMessageRecalledHandler         func(*QQClient, *GuildMessageRecalledEvent)
	GuildChannelUpdatedHandler          func(*QQClient, *GuildChannelUpdatedEvent)
	GuildChannelCreatedHandler          func(*QQClient, *GuildChannelOperationEvent)
	GuildChannelDestroyedHandles        func(*QQClient, *GuildChannelOperationEvent)
	MemberJoinedGuildHandler            func(*QQClient, *MemberJoinGuildEvent)
	GroupMuteEventHandler               func(*QQClient, *GroupMuteEvent)
	GroupRecalledHandler                func(*QQClient, *GroupMessageRecalledEvent)
	FriendRecalledHandler               func(*QQClient, *FriendMessageRecalledEvent)
	JoinGroupHandler                    func(*QQClient, *GroupInfo)
	LeaveGroupHandler                   func(*QQClient, *GroupLeaveEvent)
	MemberJoinedHandler                 func(*QQClient, *MemberJoinGroupEvent)
	MemberLeavedHandler                 func(*QQClient, *MemberLeaveGroupEvent)
	MemberCardUpdatedHandler            func(*QQClient, *MemberCardUpdatedEvent)
	GroupNameUpdatedHandler             func(*QQClient, *GroupNameUpdatedEvent)
	PermissionChangedHandler            func(*QQClient, *MemberPermissionChangedEvent)
	GroupInvitedHandler                 func(*QQClient, *GroupInvitedRequest)
	JoinRequestHandler                  func(*QQClient, *UserJoinGroupRequest)
	FriendRequestHandler                func(*QQClient, *NewFriendRequest)
	NewFriendHandler                    func(*QQClient, *NewFriendEvent)
	DisconnectHandler                   func(*QQClient, *ClientDisconnectedEvent)
	//OfflineHandler                      func(*QQClient, *ClientOfflineEvent)
	LogHandler                      func(*QQClient, *LogEvent)
	ServerUpdatedHandler            func(*QQClient, *ServerUpdatedEvent) bool
	GroupNotifyHandler              func(*QQClient, INotifyEvent)
	FriendNotifyHandler             func(*QQClient, INotifyEvent)
	MemberTitleUpdatedHandler       func(*QQClient, *MemberSpecialTitleUpdatedEvent)
	OfflineFileHandler              func(*QQClient, *OfflineFileEvent)
	OtherClientStatusChangedHandler func(*QQClient, *OtherClientStatusChangedEvent)
	GroupDigestHandler              func(*QQClient, *GroupDigestEvent)
	//TokenUpdatedHandler                 func(*QQClient)
}

var defaultHandlers = EventHandler{
	PrivateMessageHandler: func(client *QQClient, privateMessage *message.PrivateMessage) {
		client.dispatchPrivateMessage(privateMessage)
	},
	TempMessageHandler: func(client *QQClient, tempMessageEvent *TempMessageEvent) {
		client.dispatchTempMessage(tempMessageEvent)
	},
	GroupMessageHandler: func(client *QQClient, groupMessage *message.GroupMessage) {
		client.dispatchGroupMessage(groupMessage)
	},
	SelfPrivateMessageHandler: func(client *QQClient, privateMessage *message.PrivateMessage) {
		client.dispatchPrivateMessageSelf(privateMessage)
	},
	SelfGroupMessageHandler: func(client *QQClient, groupMessage *message.GroupMessage) {
		client.dispatchGroupMessageSelf(groupMessage)
	},
	GuildChannelMessageHandler: func(client *QQClient, guildChannelMessage *message.GuildChannelMessage) {
		client.dispatchGuildChannelMessage(guildChannelMessage)
	},
	GuildMessageReactionsUpdatedHandler: func(client *QQClient, guildMessageReactionsUpdatedEvent *GuildMessageReactionsUpdatedEvent) {
		client.dispatchGuildMessageReactionsUpdatedEvent(guildMessageReactionsUpdatedEvent)
	},
	GuildMessageRecalledHandler: func(client *QQClient, guildMessageRecalledEvent *GuildMessageRecalledEvent) {
		client.dispatchGuildMessageRecalledEvent(guildMessageRecalledEvent)
	},
	GuildChannelUpdatedHandler: func(client *QQClient, guildChannelUpdatedEvent *GuildChannelUpdatedEvent) {
		client.dispatchGuildChannelUpdatedEvent(guildChannelUpdatedEvent)
	},
	GuildChannelCreatedHandler: func(client *QQClient, guildChannelOperationEvent *GuildChannelOperationEvent) {
		client.dispatchGuildChannelCreatedEvent(guildChannelOperationEvent)
	},
	GuildChannelDestroyedHandles: func(client *QQClient, guildChannelOperationEvent *GuildChannelOperationEvent) {
		client.dispatchGuildChannelDestroyedEvent(guildChannelOperationEvent)
	},
	MemberJoinedGuildHandler: func(client *QQClient, memberJoinedGuildEvent *MemberJoinGuildEvent) {
		client.dispatchMemberJoinedGuildEvent(memberJoinedGuildEvent)
	},
	GroupMuteEventHandler: func(client *QQClient, groupMuteEvent *GroupMuteEvent) {
		client.dispatchGroupMuteEvent(groupMuteEvent)
	},
	GroupRecalledHandler: func(client *QQClient, groupRecalledEvent *GroupMessageRecalledEvent) {
		client.dispatchGroupMessageRecalledEvent(groupRecalledEvent)
	},
	FriendRecalledHandler: func(client *QQClient, friendRecalledEvent *FriendMessageRecalledEvent) {
		client.dispatchFriendMessageRecalledEvent(friendRecalledEvent)
	},
	JoinGroupHandler: func(client *QQClient, joinGroupEvent *GroupInfo) {
		client.dispatchJoinGroupEvent(joinGroupEvent)
	},
	LeaveGroupHandler: func(client *QQClient, leaveGroupEvent *GroupLeaveEvent) {
		client.dispatchLeaveGroupEvent(leaveGroupEvent)
	},
	MemberJoinedHandler: func(client *QQClient, memberJoinedEvent *MemberJoinGroupEvent) {
		client.dispatchNewMemberEvent(memberJoinedEvent)
	},
	MemberLeavedHandler: func(client *QQClient, memberLeavedEvent *MemberLeaveGroupEvent) {
		client.dispatchMemberLeaveEvent(memberLeavedEvent)
	},
	MemberCardUpdatedHandler: func(client *QQClient, memberCardUpdatedEvent *MemberCardUpdatedEvent) {
		client.dispatchMemberCardUpdatedEvent(memberCardUpdatedEvent)
	},
	GroupNameUpdatedHandler: func(client *QQClient, groupNameUpdatedEvent *GroupNameUpdatedEvent) {
		client.dispatchGroupNameUpdatedEvent(groupNameUpdatedEvent)
	},
	PermissionChangedHandler: func(client *QQClient, permissionChangedEvent *MemberPermissionChangedEvent) {
		client.dispatchPermissionChanged(permissionChangedEvent)
	},
	GroupInvitedHandler: func(client *QQClient, groupInvitedEvent *GroupInvitedRequest) {
		client.dispatchGroupInvitedEvent(groupInvitedEvent)
	},
	JoinRequestHandler: func(client *QQClient, joinRequestEvent *UserJoinGroupRequest) {
		client.dispatchJoinGroupRequest(joinRequestEvent)
	},
	FriendRequestHandler: func(client *QQClient, friendRequestEvent *NewFriendRequest) {
		client.dispatchNewFriendRequest(friendRequestEvent)
	},
	NewFriendHandler: func(client *QQClient, newFriendEvent *NewFriendEvent) {
		client.dispatchNewFriendEvent(newFriendEvent)
	},
	DisconnectHandler: func(client *QQClient, disconnectEvent *ClientDisconnectedEvent) {
		client.dispatchDisconnectEvent(disconnectEvent)
	},
	LogHandler: func(client *QQClient, logEvent *LogEvent) {
		client.dispatchLogEvent(logEvent)
	},
	ServerUpdatedHandler: func(client *QQClient, serverUpdatedEvent *ServerUpdatedEvent) bool {
		return client.dispatchServerUpdatedEvent(serverUpdatedEvent)
	},
	GroupNotifyHandler: func(client *QQClient, groupNotifyEvent INotifyEvent) {
		client.dispatchGroupNotifyEvent(groupNotifyEvent)
	},
	FriendNotifyHandler: func(client *QQClient, friendNotifyEvent INotifyEvent) {
		client.dispatchFriendNotifyEvent(friendNotifyEvent)
	},
	MemberTitleUpdatedHandler: func(client *QQClient, memberTitleUpdatedEvent *MemberSpecialTitleUpdatedEvent) {
		client.dispatchMemberSpecialTitleUpdateEvent(memberTitleUpdatedEvent)
	},
	OfflineFileHandler: func(client *QQClient, offlineFileEvent *OfflineFileEvent) {
		client.dispatchOfflineFileEvent(offlineFileEvent)
	},
	OtherClientStatusChangedHandler: func(client *QQClient, otherClientStatusChangedEvent *OtherClientStatusChangedEvent) {
		client.dispatchOtherClientStatusChangedEvent(otherClientStatusChangedEvent)
	},
	GroupDigestHandler: func(client *QQClient, groupDigestEvent *GroupDigestEvent) {
		client.dispatchGroupDigestEvent(groupDigestEvent)
	},
}

func (c *QQClient) dispatchServerUpdatedEvent(e *ServerUpdatedEvent) (f bool) {
	f = true
	if c.serverUpdatedHandlers == nil {
		return
	}
	for _, handler := range c.serverUpdatedHandlers {
		utils.CoverError(func() {
			if !handler(c, e) {
				f = false
			}
		})
		if !f {
			return
		}
	}
	return
}

func (c *QQClient) OnServerUpdated(f func(*QQClient, *ServerUpdatedEvent) bool) {
	c.serverUpdatedHandlers = append(c.serverUpdatedHandlers, f)
}
