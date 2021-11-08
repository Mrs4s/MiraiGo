package client

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"google.golang.org/protobuf/proto"
	"time"
)

func init() {
	decoders["MsgPush.PushGroupProMsg"] = decodeGuildMessagePushPacket
}

func (c *QQClient) parseGuildChannelMessage(msg *channel.ChannelMsgContent) *message.GuildChannelMessage {
	guild := c.GuildService.FindGuild(msg.Head.RoutingHead.GetGuildId())
	if guild == nil {
		return nil // todo: sync guild info
	}
	// mem := guild.FindMember(msg.Head.RoutingHead.GetFromTinyid())
	return &message.GuildChannelMessage{
		Id:         msg.Head.ContentHead.GetSeq(),
		InternalId: msg.Body.RichText.Attr.GetRandom(),
		GuildId:    msg.Head.RoutingHead.GetGuildId(),
		ChannelId:  msg.Head.RoutingHead.GetChannelId(),
		Time:       int64(msg.Head.ContentHead.GetTime()),
		Sender: &message.GuildSender{
			TinyId:   msg.Head.RoutingHead.GetFromTinyid(),
			Nickname: string(msg.ExtInfo.GetFromNick()),
		},
		Elements: message.ParseMessageElems(msg.Body.RichText.Elems),
	}
}

func decodeGuildMessagePushPacket(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	push := new(channel.MsgOnlinePush)
	if err := proto.Unmarshal(payload, push); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	for _, m := range push.Msgs {
		if m.Head.ContentHead.GetType() == 3841 {
			type tipsPushInfo struct {
				TinyId                 uint64
				TargetMessageSenderUin int64
				GuildId                uint64
				ChannelId              uint64
			}
			// todo: 回头 event flow 的处理移出去重构下逻辑, 先暂时这样方便改
			var common *msg.CommonElem
			if m.Body != nil {
				for _, e := range m.Body.RichText.Elems {
					if e.CommonElem != nil {
						common = e.CommonElem
						break
					}
				}
			}
			if m.Head.ContentHead.GetSubType() == 2 { // todo: tips?
				if common == nil { // empty tips

				}
				tipsInfo := &tipsPushInfo{
					TinyId:    m.Head.RoutingHead.GetFromTinyid(),
					GuildId:   m.Head.RoutingHead.GetGuildId(),
					ChannelId: m.Head.RoutingHead.GetChannelId(),
				}
				if len(m.CtrlHead.IncludeUin) > 0 {
					tipsInfo.TargetMessageSenderUin = int64(m.CtrlHead.IncludeUin[0])
				}
				return tipsInfo, nil
			}
			if common == nil || common.GetServiceType() != 500 {
				continue
			}
			eventBody := new(channel.EventBody)
			if err := proto.Unmarshal(common.PbElem, eventBody); err != nil {
				c.Error("failed to unmarshal guild channel event body: %v", err)
				continue
			}
			if eventBody.UpdateMsg != nil {
				if eventBody.UpdateMsg.GetEventType() == 1 || eventBody.UpdateMsg.GetEventType() == 2 { // todo: 撤回消息
					continue
				}
				if eventBody.UpdateMsg.GetEventType() == 4 { // 消息贴表情更新 (包含添加或删除)
					t, err := c.GuildService.pullRoamMsgByEventFlow(m.Head.RoutingHead.GetGuildId(), m.Head.RoutingHead.GetChannelId(), eventBody.UpdateMsg.GetMsgSeq(), eventBody.UpdateMsg.GetMsgSeq(), eventBody.UpdateMsg.GetEventVersion()-1)
					if err != nil || len(t) == 0 {
						c.Error("process guild event flow error: pull eventMsg message error: %v", err)
						continue
					}
					// 自己的消息被贴表情会单独推送一个tips, 这里不需要解析
					if t[0].Head.RoutingHead.GetFromTinyid() == c.GuildService.TinyId {
						continue
					}
					updatedEvent := &GuildMessageReactionsUpdatedEvent{
						GuildId:          m.Head.RoutingHead.GetGuildId(),
						ChannelId:        m.Head.RoutingHead.GetChannelId(),
						MessageId:        t[0].Head.ContentHead.GetSeq(),
						CurrentReactions: decodeGuildMessageEmojiReactions(t[0]),
					}
					tipsInfo, err := c.waitPacketTimeoutSyncF("MsgPush.PushGroupProMsg", time.Second, func(i interface{}) bool {
						if i == nil {
							return false
						}
						_, ok := i.(*tipsPushInfo)
						return ok
					})
					if err == nil {
						updatedEvent.OperatorId = tipsInfo.(*tipsPushInfo).TinyId
						updatedEvent.MessageSenderUin = tipsInfo.(*tipsPushInfo).TargetMessageSenderUin
					}
					c.dispatchGuildMessageReactionsUpdatedEvent(updatedEvent)
				}
			}
			continue
		}
		if cm := c.parseGuildChannelMessage(m); cm != nil {
			c.dispatchGuildChannelMessage(cm)
		}
	}
	return nil, nil
}

func decodeGuildMessageEmojiReactions(content *channel.ChannelMsgContent) (r []*message.GuildMessageEmojiReaction) {
	r = []*message.GuildMessageEmojiReaction{}
	var common *msg.CommonElem
	for _, elem := range content.Body.RichText.Elems {
		if elem.CommonElem != nil && elem.CommonElem.GetServiceType() == 38 {
			common = elem.CommonElem
			break
		}
	}
	if common == nil {
		return
	}
	serv38 := new(msg.MsgElemInfoServtype38)
	_ = proto.Unmarshal(common.PbElem, serv38)
	if len(serv38.ReactData) > 0 {
		cnt := new(channel.MsgCnt)
		_ = proto.Unmarshal(serv38.ReactData, cnt)
		if len(cnt.EmojiReaction) == 0 {
			return
		}
		for _, e := range cnt.EmojiReaction {
			reaction := &message.GuildMessageEmojiReaction{
				EmojiId:   e.GetEmojiId(),
				EmojiType: e.GetEmojiType(),
				Count:     int32(e.GetCnt()),
				Clicked:   e.GetIsClicked(),
			}
			if index, err := strconv.ParseInt(e.GetEmojiId(), 10, 32); err == nil {
				reaction.Face = message.NewFace(int32(index))
			}
			r = append(r, reaction)
		}
	}
	return
}
