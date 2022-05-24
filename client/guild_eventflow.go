package client

import (
	"sync"
	"time"

	"github.com/pierrec/lz4/v4"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

func init() {
	decoders["MsgPush.PushGroupProMsg"] = decodeGuildEventFlowPacket
}

var updateChanLock sync.Mutex

type tipsPushInfo struct {
	TinyId uint64
	// TargetMessageSenderUin int64
	GuildId   uint64
	ChannelId uint64
}

func decodeGuildEventFlowPacket(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	push := new(channel.MsgOnlinePush)
	if err := proto.Unmarshal(payload, push); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if push.CompressFlag.Unwrap() == 1 && len(push.CompressMsg) > 0 {
		press := new(channel.PressMsg)
		dst := make([]byte, len(push.CompressMsg)*2)
		i, err := lz4.UncompressBlock(push.CompressMsg, dst)
		for times := 0; err != nil && err.Error() == "lz4: invalid source or destination buffer too short" && times < 5; times++ {
			dst = append(dst, make([]byte, 1024)...)
			i, err = lz4.UncompressBlock(push.CompressMsg, dst)
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to decompress guild event packet")
		}
		if err = proto.Unmarshal(dst[:i], press); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
		}
		push.Msgs = press.Msgs
	}
	for _, m := range push.Msgs {
		if m.Head.ContentHead.Type.Unwrap() == 3841 {
			// todo: 回头 event flow 的处理移出去重构下逻辑, 先暂时这样方便改
			var common *msg.CommonElem
			if m.Body != nil && m.Body.RichText != nil {
				for _, e := range m.Body.RichText.Elems {
					if e.CommonElem != nil {
						common = e.CommonElem
						break
					}
				}
			}
			if m.Head.ContentHead.SubType.Unwrap() == 2 { // todo: tips?
				if common == nil { // empty tips
				}
				tipsInfo := &tipsPushInfo{
					TinyId:    m.Head.RoutingHead.FromTinyid.Unwrap(),
					GuildId:   m.Head.RoutingHead.GuildId.Unwrap(),
					ChannelId: m.Head.RoutingHead.ChannelId.Unwrap(),
				}
				/*
					if len(m.CtrlHead.IncludeUin) > 0 {
						tipsInfo.TargetMessageSenderUin = int64(m.CtrlHead.IncludeUin[0])
					}
				*/
				return tipsInfo, nil
			}
			if common == nil || common.ServiceType.Unwrap() != 500 {
				continue
			}
			eventBody := new(channel.EventBody)
			if err := proto.Unmarshal(common.PbElem, eventBody); err != nil {
				c.error("failed to unmarshal guild channel event body: %v", err)
				continue
			}
			c.processGuildEventBody(m, eventBody)
			continue
		}
		if m.Head.ContentHead.Type.Unwrap() == 3840 {
			if m.Head.RoutingHead.DirectMessageFlag.Unwrap() == 1 {
				// todo: direct message decode
				continue
			}
			if m.Head.RoutingHead.FromTinyid.Unwrap() == c.GuildService.TinyId {
				continue
			}
			if cm := c.GuildService.parseGuildChannelMessage(m); cm != nil {
				c.dispatchGuildChannelMessage(cm)
			}
		}
	}
	return nil, nil
}

func (c *QQClient) processGuildEventBody(m *channel.ChannelMsgContent, eventBody *channel.EventBody) {
	var guild *GuildInfo
	if m.Head.RoutingHead.GuildId.Unwrap() != 0 {
		if guild = c.GuildService.FindGuild(m.Head.RoutingHead.GuildId.Unwrap()); guild == nil {
			c.warning("process channel event error: guild not found.")
			return
		}
	}
	switch {
	case eventBody.CreateChan != nil:
		for _, chanId := range eventBody.CreateChan.CreateId {
			if guild.FindChannel(chanId.ChanId.Unwrap()) != nil {
				continue
			}
			channelInfo, err := c.GuildService.FetchChannelInfo(guild.GuildId, chanId.ChanId.Unwrap())
			if err != nil {
				c.warning("process create channel event error: fetch channel info error: %v", err)
				continue
			}
			guild.Channels = append(guild.Channels, channelInfo)
			c.dispatchGuildChannelCreatedEvent(&GuildChannelOperationEvent{
				OperatorId:  m.Head.RoutingHead.FromTinyid.Unwrap(),
				GuildId:     m.Head.RoutingHead.GuildId.Unwrap(),
				ChannelInfo: channelInfo,
			})
		}
	case eventBody.DestroyChan != nil:
		for _, chanId := range eventBody.DestroyChan.DeleteId {
			channelInfo := guild.FindChannel(chanId.ChanId.Unwrap())
			if channelInfo == nil {
				continue
			}
			guild.removeChannel(chanId.ChanId.Unwrap())
			c.dispatchGuildChannelDestroyedEvent(&GuildChannelOperationEvent{
				OperatorId:  m.Head.RoutingHead.FromTinyid.Unwrap(),
				GuildId:     guild.GuildId,
				ChannelInfo: channelInfo,
			})
		}
	case eventBody.ChangeChanInfo != nil:
		updateChanLock.Lock()
		defer updateChanLock.Unlock()
		oldInfo := guild.FindChannel(eventBody.ChangeChanInfo.ChanId.Unwrap())
		if oldInfo == nil {
			info, err := c.GuildService.FetchChannelInfo(m.Head.RoutingHead.GuildId.Unwrap(), eventBody.ChangeChanInfo.ChanId.Unwrap())
			if err != nil {
				c.error("error to decode channel info updated event: fetch channel info failed: %v", err)
				return
			}
			guild.Channels = append(guild.Channels, info)
			oldInfo = info
		}
		if time.Now().Unix()-oldInfo.fetchTime <= 2 {
			return
		}
		newInfo, err := c.GuildService.FetchChannelInfo(m.Head.RoutingHead.GuildId.Unwrap(), eventBody.ChangeChanInfo.ChanId.Unwrap())
		if err != nil {
			c.error("error to decode channel info updated event: fetch channel info failed: %v", err)
			return
		}
		for i := range guild.Channels {
			if guild.Channels[i].ChannelId == newInfo.ChannelId {
				guild.Channels[i] = newInfo
				break
			}
		}
		c.dispatchGuildChannelUpdatedEvent(&GuildChannelUpdatedEvent{
			OperatorId:     m.Head.RoutingHead.FromTinyid.Unwrap(),
			GuildId:        m.Head.RoutingHead.GuildId.Unwrap(),
			ChannelId:      eventBody.ChangeChanInfo.ChanId.Unwrap(),
			OldChannelInfo: oldInfo,
			NewChannelInfo: newInfo,
		})
	case eventBody.JoinGuild != nil:
		/* 应该不会重复推送把, 不会吧不会吧
		if mem := guild.FindMember(eventBody.JoinGuild.MemberTinyid.Unwrap()); mem != nil {
			c.info("ignore join guild event: member %v already exists", mem.TinyId)
			return
		}
		*/
		profile, err := c.GuildService.FetchGuildMemberProfileInfo(guild.GuildId, eventBody.JoinGuild.MemberTinyid.Unwrap())
		if err != nil {
			c.error("error to decode member join guild event: get member profile error: %v", err)
			return
		}
		info := &GuildMemberInfo{
			TinyId:   profile.TinyId,
			Nickname: profile.Nickname,
		}
		// guild.Members = append(guild.Members, info)
		c.dispatchMemberJoinedGuildEvent(&MemberJoinGuildEvent{
			Guild:  guild,
			Member: info,
		})
	case eventBody.UpdateMsg != nil:
		if eventBody.UpdateMsg.EventType.Unwrap() == 1 || eventBody.UpdateMsg.EventType.Unwrap() == 2 {
			c.dispatchGuildMessageRecalledEvent(&GuildMessageRecalledEvent{
				OperatorId: eventBody.UpdateMsg.OperatorTinyid.Unwrap(),
				GuildId:    m.Head.RoutingHead.GuildId.Unwrap(),
				ChannelId:  m.Head.RoutingHead.ChannelId.Unwrap(),
				MessageId:  eventBody.UpdateMsg.MsgSeq.Unwrap(),
				RecallTime: int64(m.Head.ContentHead.Time.Unwrap()),
			})
			return
		}
		if eventBody.UpdateMsg.EventType.Unwrap() == 4 { // 消息贴表情更新 (包含添加或删除)
			t, err := c.GuildService.pullChannelMessages(m.Head.RoutingHead.GuildId.Unwrap(), m.Head.RoutingHead.ChannelId.Unwrap(), eventBody.UpdateMsg.MsgSeq.Unwrap(), eventBody.UpdateMsg.MsgSeq.Unwrap(), eventBody.UpdateMsg.EventVersion.Unwrap()-1, false)
			if err != nil || len(t) == 0 {
				c.error("process guild event flow error: pull eventMsg message error: %v", err)
				return
			}
			// 自己的消息被贴表情会单独推送一个tips, 这里不需要解析
			if t[0].Head.RoutingHead.FromTinyid.Unwrap() == c.GuildService.TinyId {
				return
			}
			updatedEvent := &GuildMessageReactionsUpdatedEvent{
				GuildId:          m.Head.RoutingHead.GuildId.Unwrap(),
				ChannelId:        m.Head.RoutingHead.ChannelId.Unwrap(),
				MessageId:        t[0].Head.ContentHead.Seq.Unwrap(),
				CurrentReactions: decodeGuildMessageEmojiReactions(t[0]),
			}
			tipsInfo, err := c.waitPacketTimeoutSyncF("MsgPush.PushGroupProMsg", time.Second, func(i any) bool {
				if i == nil {
					return false
				}
				_, ok := i.(*tipsPushInfo)
				return ok
			})
			if err == nil {
				updatedEvent.OperatorId = tipsInfo.(*tipsPushInfo).TinyId
				// updatedEvent.MessageSenderUin = tipsInfo.(*tipsPushInfo).TargetMessageSenderUin
			}
			c.dispatchGuildMessageReactionsUpdatedEvent(updatedEvent)
		}
	}
}
