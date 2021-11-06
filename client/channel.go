package client

import (
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type (
	// ChannelService 频道模块内自身的信息
	ChannelService struct {
		TinyId       uint64
		ChannelCount uint32
		// Guilds 由服务器推送的频道列表
		Guilds []*GuildInfo
	}

	// GuildInfo 频道信息
	GuildInfo struct {
		GuildId   uint64
		GuildCode uint64
		GuildName string
		Channels  []*ChannelInfo
		Bots      []*GuildMemberInfo
		Members   []*GuildMemberInfo
		Admins    []*GuildMemberInfo
	}

	GuildMemberInfo struct {
		TinyId        uint64
		Title         string
		Nickname      string
		LastSpeakTime int64
		Role          int32 // 0 = member 1 = admin 2 = owner ?
	}

	// ChannelInfo 子频道信息
	ChannelInfo struct {
		ChannelId   uint64
		ChannelName string
		Time        uint64
		EventTime   uint32
		NotifyType  uint32
		ChannelType uint32
		AtAllSeq    uint64
	}
)

func init() {
	decoders["trpc.group_pro.synclogic.SyncLogic.PushFirstView"] = decodeChannelPushFirstView
	decoders["MsgPush.PushGroupProMsg"] = decodeChannelMessagePushPacket
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
	c.ChannelService.TinyId = firstViewRsp.GetSelfTinyid()
	c.ChannelService.ChannelCount = firstViewRsp.GetGuildCount()
}

func (c *QQClient) GetGuildMembers(guildId uint64) (bots []*GuildMemberInfo, members []*GuildMemberInfo, admins []*GuildMemberInfo, err error) {
	seq := c.nextSeq()
	u1 := uint32(1)
	payload := c.packOIDBPackage(3931, 1, binary.EncodeDynamicProtoMessage(binary.DynamicProtoMessage{ // todo: 可能还需要处理翻页的情况?
		1: guildId, // guild id
		2: uint32(3),
		3: uint32(0),
		4: binary.DynamicProtoMessage{ // unknown param, looks like flags
			1: u1, 2: u1, 3: u1, 4: u1, 5: u1, 6: u1, 7: u1, 8: u1, 20: u1,
		},
		6:  uint32(0),
		8:  uint32(500), // max response?
		14: uint32(2),
	}))
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvcTrpcTcp.0xf5b_1", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, payload)
	rsp, err := c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "send packet error")
	}
	pkg := new(oidb.OIDBSSOPkg)
	oidbRsp := new(channel.ChannelOidb0Xf5BRsp)
	if err = proto.Unmarshal(rsp, pkg); err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err = proto.Unmarshal(pkg.Bodybuffer, oidbRsp); err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	protoToMemberInfo := func(mem *channel.GuildMemberInfo) *GuildMemberInfo {
		return &GuildMemberInfo{
			TinyId:        mem.GetTinyId(),
			Title:         mem.GetTitle(),
			Nickname:      mem.GetNickname(),
			LastSpeakTime: mem.GetLastSpeakTime(),
			Role:          mem.GetRole(),
		}
	}
	for _, mem := range oidbRsp.Bots {
		bots = append(bots, protoToMemberInfo(mem))
	}
	for _, mem := range oidbRsp.Members {
		members = append(members, protoToMemberInfo(mem))
	}
	for _, mem := range oidbRsp.AdminInfo.Admins {
		admins = append(admins, protoToMemberInfo(mem))
	}
	return
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

func decodeChannelMessagePushPacket(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	return nil, nil
}

func decodeChannelPushFirstView(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	firstViewMsg := new(channel.FirstViewMsg)
	if err := proto.Unmarshal(payload, firstViewMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(firstViewMsg.GuildNodes) > 0 {
		c.ChannelService.Guilds = []*GuildInfo{}
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
			info.Bots, info.Members, info.Admins, _ = c.GetGuildMembers(info.GuildId)
			c.ChannelService.Guilds = append(c.ChannelService.Guilds, info)
		}
	}
	if len(firstViewMsg.ChannelMsgs) > 0 { // sync msg

	}
	return nil, nil
}
