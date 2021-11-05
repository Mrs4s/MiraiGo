package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/internal/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// ChannelSelfInfo 频道模块内自身的信息
type ChannelSelfInfo struct {
	TinyId       uint64
	ChannelCount uint32
	// Guilds 由服务器推送的频道列表
	Guilds []*GuildInfo
}

// GuildInfo 频道信息
type GuildInfo struct {
	GuildId   uint64
	GuildCode uint64
	GuildName string
	Channels  []*ChannelInfo
}

// ChannelInfo 子频道信息
type ChannelInfo struct {
	ChannelId   uint64
	ChannelName string
	Time        uint64
	EventTime   uint32
	NotifyType  uint32
	ChannelType uint32
	AtAllSeq    uint64
}

func init() {
	decoders["trpc.group_pro.synclogic.SyncLogic.PushFirstView"] = decodeChannelPushFirstView
}

func (c *QQClient) syncChannelFirstView() {
	rsp, err := c.sendAndWaitDynamic(c.buildSyncChannelFirstViewPacket())
	if err != nil {
		c.Error("sync channel error: %v", err)
		return
	}
	firstViewRsp := new(channel.FirstViewRsp)
	if err = proto.Unmarshal(rsp, firstViewRsp); err != nil {
		return
	}
	c.ChannelSelf.TinyId = firstViewRsp.GetSelfTinyid()
	c.ChannelSelf.ChannelCount = firstViewRsp.GetGuildCount()
}

func (c *QQClient) buildSyncChannelFirstViewPacket() (uint16, []byte) {
	seq := c.nextSeq()
	req := &channel.FirstViewReq{
		LastMsgTime:       proto.Uint64(0),
		Seq:               proto.Uint32(0),
		DirectMessageFlag: proto.Uint32(1),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "trpc.group_pro.synclogic.SyncLogic.SyncFirstView", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, payload)
	return seq, packet
}

func decodeChannelPushFirstView(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	firstViewMsg := new(channel.FirstViewMsg)
	if err := proto.Unmarshal(payload, firstViewMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(firstViewMsg.GuildNodes) > 0 {
		c.ChannelSelf.Guilds = []*GuildInfo{}
		for _, guild := range firstViewMsg.GuildNodes {
			info := &GuildInfo{
				GuildId:   guild.GetGuildId(),
				GuildCode: guild.GetGuildCode(),
				GuildName: utils.B2S(guild.GuildName),
			}
			for _, node := range guild.ChannelNodes {
				meta := new(channel.ChannelMsgMeta)
				_ = proto.Unmarshal(node.Meta, meta)
				info.Channels = append(info.Channels, &ChannelInfo{
					ChannelId:   node.GetChannelId(),
					ChannelName: utils.B2S(node.ChannelName),
					Time:        node.GetTime(),
					EventTime:   node.GetEventTime(),
					NotifyType:  node.GetNotifyType(),
					ChannelType: node.GetChannelType(),
					AtAllSeq:    meta.GetAtAllSeq(),
				})
			}
			c.ChannelSelf.Guilds = append(c.ChannelSelf.Guilds, info)
		}
	}
	if len(firstViewMsg.ChannelMsgs) > 0 { // sync msg

	}
	return nil, nil
}
