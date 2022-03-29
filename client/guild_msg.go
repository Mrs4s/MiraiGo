package client

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x388"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
)

func init() {
	decoders["ImgStore.QQMeetPicUp"] = decodeGuildImageStoreResponse
}

func (s *GuildService) SendGuildChannelMessage(guildId, channelId uint64, m *message.SendingMessage) (*message.GuildChannelMessage, error) {
	mr := rand.Uint32() // 客户端似乎是生成的 u32 虽然类型是u64
	for _, elem := range m.Elements {
		if elem.Type() == message.At {
			at := elem.(*message.AtElement)
			if at.SubType == message.AtTypeGroupMember {
				at.SubType = message.AtTypeGuildMember
			}
		}
	}
	req := &channel.DF62ReqBody{Msg: &channel.ChannelMsgContent{
		Head: &channel.ChannelMsgHead{
			RoutingHead: &channel.ChannelRoutingHead{
				GuildId:   &guildId,
				ChannelId: &channelId,
				FromUin:   proto.Uint64(uint64(s.c.Uin)),
			},
			ContentHead: &channel.ChannelContentHead{
				Type:   proto.Uint64(3840), // const
				Random: proto.Uint64(uint64(mr)),
			},
		},
		Body: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m.Elements, true),
			},
		},
	}}
	payload, _ := proto.Marshal(req)
	seq, packet := s.c.uniPacket("MsgProxy.SendMsg", payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.DF62RspBody)
	if err = proto.Unmarshal(rsp, body); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if body.GetResult() != 0 {
		return nil, errors.Errorf("send channel message error: server response %v", body.GetResult())
	}
	elements := m.Elements
	if body.Body != nil && body.Body.RichText != nil {
		elements = message.ParseMessageElems(body.Body.RichText.Elems)
	}
	return &message.GuildChannelMessage{
		Id:         body.Head.ContentHead.GetSeq(),
		InternalId: body.Head.ContentHead.GetRandom(),
		GuildId:    guildId,
		ChannelId:  channelId,
		Time:       int64(body.GetSendTime()),
		Sender: &message.GuildSender{
			TinyId:   body.Head.RoutingHead.GetFromTinyid(),
			Nickname: s.Nickname,
		},
		Elements: elements,
	}, nil
}

func (s *GuildService) QueryImage(guildId, channelId uint64, hash []byte, size uint64) (*message.GuildImageElement, error) {
	rsp, err := s.c.sendAndWait(s.c.buildGuildImageStorePacket(guildId, channelId, hash, size))
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := rsp.(*imageUploadResponse)
	if body.IsExists {
		return &message.GuildImageElement{
			FileId:        body.FileId,
			FilePath:      fmt.Sprintf("%x.jpg", hash),
			Size:          int32(size),
			DownloadIndex: body.DownloadIndex,
			Width:         body.Width,
			Height:        body.Height,
			Md5:           hash,
		}, nil
	}
	return nil, errors.New("image is not exists")
}

// Deprecated: use QQClient.UploadImage instead
func (s *GuildService) UploadGuildImage(guildId, channelId uint64, img io.ReadSeeker) (*message.GuildImageElement, error) {
	source := message.Source{
		SourceType:  message.SourceGuildChannel,
		PrimaryID:   int64(guildId),
		SecondaryID: int64(channelId),
	}
	image, err := s.c.uploadGroupOrGuildImage(source, img)
	if err != nil {
		return nil, err
	}
	return image.(*message.GuildImageElement), nil
}

func (s *GuildService) PullGuildChannelMessage(guildId, channelId, beginSeq, endSeq uint64) (r []*message.GuildChannelMessage, e error) {
	contents, err := s.pullChannelMessages(guildId, channelId, beginSeq, endSeq, 0, false)
	if err != nil {
		return nil, errors.Wrap(err, "pull channel message error")
	}
	for _, c := range contents {
		if cm := s.parseGuildChannelMessage(c); cm != nil {
			cm.Reactions = decodeGuildMessageEmojiReactions(c)
			r = append(r, cm)
		}
	}
	if len(r) == 0 {
		return nil, errors.New("message not found")
	}
	return
}

func (s *GuildService) pullChannelMessages(guildId, channelId, beginSeq, endSeq, eventVersion uint64, direct bool) ([]*channel.ChannelMsgContent, error) {
	param := &channel.ChannelParam{
		GuildId:   &guildId,
		ChannelId: &channelId,
		BeginSeq:  &beginSeq,
		EndSeq:    &endSeq,
	}
	if eventVersion != 0 {
		param.Version = []uint64{eventVersion}
	}

	withVersionFlag := uint32(0)
	if eventVersion != 0 {
		withVersionFlag = 1
	}
	directFlag := uint32(0)
	if direct {
		directFlag = 1
	}
	payload, _ := proto.Marshal(&channel.ChannelMsgReq{
		ChannelParam:      param,
		WithVersionFlag:   &withVersionFlag,
		DirectMessageFlag: &directFlag,
	})
	seq, packet := s.c.uniPacket("trpc.group_pro.synclogic.SyncLogic.GetChannelMsg", payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	msgRsp := new(channel.ChannelMsgRsp)
	if err = proto.Unmarshal(rsp, msgRsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return msgRsp.ChannelMsg.Msgs, nil
}

func (c *QQClient) buildGuildImageStorePacket(guildId, channelId uint64, hash []byte, size uint64) (uint16, []byte) {
	payload, _ := proto.Marshal(&cmd0x388.D388ReqBody{
		NetType: proto.Uint32(3),
		Subcmd:  proto.Uint32(1),
		TryupImgReq: []*cmd0x388.TryUpImgReq{
			{
				GroupCode:       &channelId,
				SrcUin:          proto.Uint64(uint64(c.Uin)),
				FileId:          proto.Uint64(0),
				FileMd5:         hash,
				FileSize:        &size,
				FileName:        []byte(fmt.Sprintf("%x.jpg", hash)),
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
	return c.uniPacket("ImgStore.QQMeetPicUp", payload)
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

func decodeGuildImageStoreResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	body := new(cmd0x388.D388RspBody)
	if err := proto.Unmarshal(payload, body); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(body.TryupImgRsp) == 0 {
		return nil, errors.New("response is empty")
	}
	rsp := body.TryupImgRsp[0]
	if rsp.GetResult() != 0 {
		return &imageUploadResponse{
			ResultCode: int32(rsp.GetResult()),
			Message:    string(rsp.FailMsg),
		}, nil
	}
	if rsp.GetFileExit() {
		resp := &imageUploadResponse{
			IsExists:      true,
			FileId:        int64(rsp.GetFileid()),
			DownloadIndex: string(rsp.DownloadIndex),
		}
		if rsp.ImgInfo != nil {
			resp.Width = int32(rsp.ImgInfo.GetFileWidth())
			resp.Height = int32(rsp.ImgInfo.GetFileHeight())
		}
		return rsp, nil
	}
	return &imageUploadResponse{
		FileId:        int64(rsp.GetFileid()),
		UploadKey:     rsp.UpUkey,
		UploadIp:      rsp.UpIp,
		UploadPort:    rsp.UpPort,
		DownloadIndex: string(rsp.DownloadIndex),
	}, nil
}

func (s *GuildService) parseGuildChannelMessage(msg *channel.ChannelMsgContent) *message.GuildChannelMessage {
	guild := s.FindGuild(msg.Head.RoutingHead.GetGuildId())
	if guild == nil {
		return nil // todo: sync guild info
	}
	if msg.Body == nil || msg.Body.RichText == nil {
		return nil
	}
	// mem := guild.FindMember(msg.Head.RoutingHead.GetFromTinyid())
	memberName := msg.ExtInfo.MemberName
	if memberName == nil {
		memberName = msg.ExtInfo.FromNick
	}
	return &message.GuildChannelMessage{
		Id:         msg.Head.ContentHead.GetSeq(),
		InternalId: msg.Head.ContentHead.GetRandom(),
		GuildId:    msg.Head.RoutingHead.GetGuildId(),
		ChannelId:  msg.Head.RoutingHead.GetChannelId(),
		Time:       int64(msg.Head.ContentHead.GetTime()),
		Sender: &message.GuildSender{
			TinyId:   msg.Head.RoutingHead.GetFromTinyid(),
			Nickname: string(memberName),
		},
		Elements: message.ParseMessageElems(msg.Body.RichText.Elems),
	}
}
