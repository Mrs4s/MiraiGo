package client

import (
	"encoding/hex"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x388"
	"github.com/Mrs4s/MiraiGo/internal/packets"
	"strconv"

	"github.com/pkg/errors"

	"time"

	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"google.golang.org/protobuf/proto"
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

func (s *GuildService) QueryImage(guildId, channelId uint64, hash []byte, size uint64) (*message.GuildImageElement, error) {
	seq := s.c.nextSeq()
	payload, _ := proto.Marshal(&cmd0x388.D388ReqBody{
		NetType: proto.Uint32(3),
		Subcmd:  proto.Uint32(1),
		TryupImgReq: []*cmd0x388.TryUpImgReq{
			{
				GroupCode:       &channelId,
				SrcUin:          proto.Uint64(uint64(s.c.Uin)),
				FileId:          proto.Uint64(0),
				FileMd5:         hash,
				FileSize:        &size,
				FileName:        []byte(hex.EncodeToString(hash) + ".jpg"),
				SrcTerm:         proto.Uint32(5),
				PlatformType:    proto.Uint32(9),
				BuType:          proto.Uint32(211),
				PicType:         proto.Uint32(1000),
				BuildVer:        []byte("8.8.38.2266"),
				AppPicType:      proto.Uint32(1052),
				SrvUpload:       proto.Uint32(0),
				QqmeetGuildId:   &guildId,
				QqmeetChannelId: &channelId,
			},
		},
		CommandId: proto.Uint32(83),
	})
	packet := packets.BuildUniPacket(s.c.Uin, seq, "ImgStore.QQMeetPicUp", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key, payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(cmd0x388.D388RspBody)
	if err = proto.Unmarshal(rsp, body); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(body.TryupImgRsp) == 0 {
		return nil, errors.New("response is empty")
	}
	if body.TryupImgRsp[0].GetFileExit() {
		return &message.GuildImageElement{
			FileId:        body.TryupImgRsp[0].GetFileid(),
			FilePath:      hex.EncodeToString(hash) + ".jpg",
			Size:          int32(size),
			DownloadIndex: string(body.TryupImgRsp[0].GetDownloadIndex()),
			Md5:           hash,
		}, nil
	}
	return nil, errors.New("image is not exists")
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
