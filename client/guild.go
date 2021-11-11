package client

import (
	"fmt"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/internal/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type (
	// GuildService 频道模块内自身的信息
	GuildService struct {
		TinyId     uint64
		Nickname   string
		AvatarUrl  string
		GuildCount uint32
		// Guilds 由服务器推送的频道列表
		Guilds []*GuildInfo

		c *QQClient
	}

	// GuildInfo 频道信息
	GuildInfo struct {
		GuildId   uint64
		GuildCode uint64
		GuildName string
		CoverUrl  string
		AvatarUrl string
		Channels  []*ChannelInfo
		Bots      []*GuildMemberInfo
		Members   []*GuildMemberInfo
		Admins    []*GuildMemberInfo
	}

	// GuildMeta 频道数据
	GuildMeta struct {
		GuildId        uint64
		GuildName      string
		GuildProfile   string
		MaxMemberCount int64
		MemberCount    int64
		CreateTime     int64
		MaxRobotCount  int32
		MaxAdminCount  int32
		OwnerId        uint64
	}

	GuildMemberInfo struct {
		TinyId        uint64
		Title         string
		Nickname      string
		LastSpeakTime int64
		Role          int32 // 0 = member 1 = admin 2 = owner ?
	}

	// GuildUserProfile 频道系统用户资料
	GuildUserProfile struct {
		TinyId    uint64
		Nickname  string
		AvatarUrl string
		JoinTime  int64 // 只有 GetGuildMemberProfileInfo 函数才会有
	}

	// ChannelInfo 子频道信息
	ChannelInfo struct {
		ChannelId   uint64
		ChannelName string
		Time        uint64
		EventTime   uint32
		NotifyType  uint32
		ChannelType ChannelType
		AtAllSeq    uint64
		Meta        *ChannelMeta

		fetchTime int64
	}

	ChannelMeta struct {
		CreatorUin           int64
		CreatorTinyId        uint64
		CreateTime           int64
		GuildId              uint64
		VisibleType          int32
		TopMessageSeq        uint64
		TopMessageTime       int64
		TopMessageOperatorId uint64
		CurrentSlowMode      int32
		TalkPermission       int32
		SlowModes            []*ChannelSlowModeInfo
	}

	ChannelSlowModeInfo struct {
		SlowModeKey    int32
		SpeakFrequency int32
		SlowModeCircle int32
		SlowModeText   string
	}

	ChannelType int32
)

const (
	ChannelTypeText  ChannelType = 1
	ChannelTypeVoice ChannelType = 2
	ChannelTypeLive  ChannelType = 5
)

func init() {
	decoders["trpc.group_pro.synclogic.SyncLogic.PushFirstView"] = decodeGuildPushFirstView
}

func (s *GuildService) FindGuild(guildId uint64) *GuildInfo {
	for _, i := range s.Guilds {
		if i.GuildId == guildId {
			return i
		}
	}
	return nil
}

func (g *GuildInfo) FindMember(tinyId uint64) *GuildMemberInfo {
	for i := 0; i < len(g.Members); i++ {
		if g.Members[i].TinyId == tinyId {
			return g.Members[i]
		}
	}
	for i := 0; i < len(g.Admins); i++ {
		if g.Admins[i].TinyId == tinyId {
			return g.Admins[i]
		}
	}
	for i := 0; i < len(g.Bots); i++ {
		if g.Bots[i].TinyId == tinyId {
			return g.Bots[i]
		}
	}
	return nil
}

func (g *GuildInfo) FindChannel(channelId uint64) *ChannelInfo {
	for _, c := range g.Channels {
		if c.ChannelId == channelId {
			return c
		}
	}
	return nil
}

func (s *GuildService) GetUserProfile(tinyId uint64) (*GuildUserProfile, error) {
	seq := s.c.nextSeq()
	flags := binary.DynamicProtoMessage{}
	for i := 3; i <= 29; i++ {
		flags[uint64(i)] = uint32(1)
	}
	flags[99] = uint32(1)
	flags[100] = uint32(1)
	payload := s.c.packOIDBPackageDynamically(3976, 1, binary.DynamicProtoMessage{
		1: flags,
		3: tinyId,
		4: uint32(0),
	})
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0xfc9_1", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key, payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xfc9Rsp)
	if err = s.c.unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	// todo: 解析个性档案
	return &GuildUserProfile{
		TinyId:    tinyId,
		Nickname:  body.Profile.GetNickname(),
		AvatarUrl: body.Profile.GetAvatarUrl(),
		JoinTime:  body.Profile.GetJoinTime(),
	}, nil
}

func (s *GuildService) GetGuildMembers(guildId uint64) (bots []*GuildMemberInfo, members []*GuildMemberInfo, admins []*GuildMemberInfo, err error) {
	seq := s.c.nextSeq()
	u1 := uint32(1)
	// todo: 这个包实际上是 fetchMemberListWithRole , 可以按channel, role等规则获取成员列表, 还需要修改
	payload := s.c.packOIDBPackageDynamically(3931, 1, binary.DynamicProtoMessage{ // todo: 可能还需要处理翻页的情况?
		1: guildId, // guild id
		2: uint32(3),
		3: uint32(0),
		4: binary.DynamicProtoMessage{ // unknown param, looks like flags
			1: u1, 2: u1, 3: u1, 4: u1, 5: u1, 6: u1, 7: u1, 8: u1, 20: u1,
		},
		6:  uint32(0),
		8:  uint32(500), // count
		14: uint32(2),
	})
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0xf5b_1", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key, payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf5BRsp)
	if err = s.c.unpackOIDBPackage(rsp, body); err != nil {
		return nil, nil, nil, errors.Wrap(err, "decode packet error")
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
	for _, mem := range body.Bots {
		bots = append(bots, protoToMemberInfo(mem))
	}
	for _, mem := range body.Members {
		members = append(members, protoToMemberInfo(mem))
	}
	for _, mem := range body.AdminInfo.Admins {
		admins = append(admins, protoToMemberInfo(mem))
	}
	return
}

func (s *GuildService) GetGuildMemberProfileInfo(guildId, tinyId uint64) (*GuildUserProfile, error) {
	seq := s.c.nextSeq()
	flags := binary.DynamicProtoMessage{}
	for i := 3; i <= 29; i++ {
		flags[uint64(i)] = uint32(1)
	}
	flags[99] = uint32(1)
	flags[100] = uint32(1)
	payload := s.c.packOIDBPackageDynamically(3976, 1, binary.DynamicProtoMessage{
		1: flags,
		3: tinyId,
		4: guildId,
	})
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0xf88_1", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key, payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf88Rsp)
	if err = s.c.unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	// todo: 解析个性档案
	return &GuildUserProfile{
		TinyId:    tinyId,
		Nickname:  body.Profile.GetNickname(),
		AvatarUrl: body.Profile.GetAvatarUrl(),
		JoinTime:  body.Profile.GetJoinTime(),
	}, nil
}

func (s *GuildService) FetchGuestGuild(guildId uint64) (*GuildMeta, error) {
	seq := s.c.nextSeq()
	u1 := uint32(1)
	payload := s.c.packOIDBPackageDynamically(3927, 9, binary.DynamicProtoMessage{
		1: binary.DynamicProtoMessage{
			1: binary.DynamicProtoMessage{
				2: u1, 4: u1, 5: u1, 6: u1, 7: u1, 8: u1, 11: u1, 12: u1, 13: u1, 14: u1, 45: u1,
				18: u1, 19: u1, 20: u1, 22: u1, 23: u1, 5002: u1, 5003: u1, 5004: u1, 5005: u1, 10007: u1,
			},
			2: binary.DynamicProtoMessage{
				3: u1, 4: u1, 6: u1, 11: u1, 14: u1, 15: u1, 16: u1, 17: u1,
			},
		},
		2: binary.DynamicProtoMessage{
			1: guildId,
		},
	})
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0xf57_9", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key, payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf57Rsp)
	if err = s.c.unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	return &GuildMeta{
		GuildName:      body.Rsp.Meta.GetName(),
		GuildProfile:   body.Rsp.Meta.GetProfile(),
		MaxMemberCount: body.Rsp.Meta.GetMaxMemberCount(),
		MemberCount:    body.Rsp.Meta.GetMemberCount(),
		CreateTime:     body.Rsp.Meta.GetCreateTime(),
		MaxRobotCount:  body.Rsp.Meta.GetRobotMaxNum(),
		MaxAdminCount:  body.Rsp.Meta.GetAdminMaxNum(),
		OwnerId:        body.Rsp.Meta.GetOwnerId(),
	}, nil
}

func (s *GuildService) FetchChannelList(guildId uint64) (r []*ChannelInfo, e error) {
	seq := s.c.nextSeq()
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0xf5d_1", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key,
		s.c.packOIDBPackageDynamically(3933, 1, binary.DynamicProtoMessage{1: guildId, 3: binary.DynamicProtoMessage{1: uint32(1)}}))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf5DRsp)
	if err = s.c.unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	for _, info := range body.Rsp.Channels {
		r = append(r, convertChannelInfo(info))
	}
	return
}

func (s *GuildService) FetchChannelInfo(guildId, channelId uint64) (*ChannelInfo, error) {
	seq := s.c.nextSeq()
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0xf55_1", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key,
		s.c.packOIDBPackageDynamically(3925, 1, binary.DynamicProtoMessage{1: guildId, 2: channelId}))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf55Rsp)
	if err = s.c.unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	return convertChannelInfo(body.Info), nil
}

/* need analysis
func (s *GuildService) fetchChannelListState(guildId uint64, channels []*ChannelInfo) {
	seq := s.c.nextSeq()
	var ids []uint64
	for _, info := range channels {
		ids = append(ids, info.ChannelId)
	}
	payload := s.c.packOIDBPackageDynamically(4104, 1, binary.DynamicProtoMessage{
		1: binary.DynamicProtoMessage{
			1: guildId,
			2: ids,
		},
	})
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0x1008_1", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key, payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return
	}
	pkg := new(oidb.OIDBSSOPkg)
	if err = proto.Unmarshal(rsp, pkg); err != nil {
		return //nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
}
*/

func convertChannelInfo(info *channel.GuildChannelInfo) *ChannelInfo {
	meta := &ChannelMeta{
		CreatorUin:      info.GetCreatorUin(),
		CreatorTinyId:   info.GetCreatorTinyId(),
		CreateTime:      info.GetCreateTime(),
		GuildId:         info.GetGuildId(),
		VisibleType:     info.GetVisibleType(),
		CurrentSlowMode: info.GetCurrentSlowModeKey(),
		TalkPermission:  info.GetTalkPermission(),
	}
	if info.TopMsg != nil {
		meta.TopMessageSeq = info.TopMsg.GetTopMsgSeq()
		meta.TopMessageTime = info.TopMsg.GetTopMsgTime()
		meta.TopMessageOperatorId = info.TopMsg.GetTopMsgOperatorTinyId()
	}
	for _, slow := range info.SlowModeInfos {
		meta.SlowModes = append(meta.SlowModes, &ChannelSlowModeInfo{
			SlowModeKey:    slow.GetSlowModeKey(),
			SpeakFrequency: slow.GetSpeakFrequency(),
			SlowModeCircle: slow.GetSlowModeCircle(),
			SlowModeText:   slow.GetSlowModeText(),
		})
	}
	return &ChannelInfo{
		ChannelId:   info.GetChannelId(),
		ChannelName: info.GetChannelName(),
		NotifyType:  uint32(info.GetFinalNotifyType()),
		ChannelType: ChannelType(info.GetChannelType()),
		Meta:        meta,
		fetchTime:   time.Now().Unix(),
	}
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
	c.GuildService.TinyId = firstViewRsp.GetSelfTinyid()
	c.GuildService.GuildCount = firstViewRsp.GetGuildCount()
	if self, err := c.GuildService.GetUserProfile(c.GuildService.TinyId); err == nil {
		c.GuildService.Nickname = self.Nickname
		c.GuildService.AvatarUrl = self.AvatarUrl
	} else {
		c.Error("get self guild profile error: %v", err)
	}
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

func decodeGuildPushFirstView(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	firstViewMsg := new(channel.FirstViewMsg)
	if err := proto.Unmarshal(payload, firstViewMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(firstViewMsg.GuildNodes) > 0 {
		c.GuildService.Guilds = []*GuildInfo{}
		for _, guild := range firstViewMsg.GuildNodes {
			info := &GuildInfo{
				GuildId:   guild.GetGuildId(),
				GuildCode: guild.GetGuildCode(),
				GuildName: utils.B2S(guild.GuildName),
				CoverUrl:  fmt.Sprintf("https://groupprocover-76483.picgzc.qpic.cn/%v", guild.GetGuildId()),
				AvatarUrl: fmt.Sprintf("https://groupprohead-76292.picgzc.qpic.cn/%v", guild.GetGuildId()),
			}
			channels, err := c.GuildService.FetchChannelList(info.GuildId)
			if err != nil {
				c.Warning("waring: fetch guild %v channel error %v. will use sync node to fill channel list field", guild.GuildId, err)
				for _, node := range guild.ChannelNodes {
					meta := new(channel.ChannelMsgMeta)
					_ = proto.Unmarshal(node.Meta, meta)
					info.Channels = append(info.Channels, &ChannelInfo{
						ChannelId:   node.GetChannelId(),
						ChannelName: utils.B2S(node.ChannelName),
						Time:        node.GetTime(),
						EventTime:   node.GetEventTime(),
						NotifyType:  node.GetNotifyType(),
						ChannelType: ChannelType(node.GetChannelType()),
						AtAllSeq:    meta.GetAtAllSeq(),
					})
				}
			} else {
				info.Channels = channels
			}
			info.Bots, info.Members, info.Admins, _ = c.GuildService.GetGuildMembers(info.GuildId)
			c.GuildService.Guilds = append(c.GuildService.Guilds, info)
		}
	}
	if len(firstViewMsg.ChannelMsgs) > 0 { // sync msg

	}
	return nil, nil
}
