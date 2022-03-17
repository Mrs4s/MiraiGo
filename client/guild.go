package client

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/topic"

	"github.com/Mrs4s/MiraiGo/client/pb/qweb"
	"github.com/Mrs4s/MiraiGo/internal/proto"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/utils"
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
		// Bots      []*GuildMemberInfo
		// Members   []*GuildMemberInfo
		// Admins    []*GuildMemberInfo
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

	// GuildMemberInfo 频道成员信息, 仅通过频道成员列表API获取
	GuildMemberInfo struct {
		TinyId        uint64
		Title         string
		Nickname      string
		LastSpeakTime int64
		Role          uint64
		RoleName      string
	}

	// GuildUserProfile 频道系统用户资料
	GuildUserProfile struct {
		TinyId    uint64
		Nickname  string
		AvatarUrl string
		JoinTime  int64 // 只有 GetGuildMemberProfileInfo 函数才会有
		Roles     []*GuildRole
	}

	// GuildRole 频道身份组信息
	GuildRole struct {
		RoleId      uint64
		RoleName    string
		ArgbColor   uint32
		Independent bool
		Num         int32
		Owned       bool
		Disabled    bool
		MaxNum      int32
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

	FetchGuildMemberListWithRoleResult struct {
		Members        []*GuildMemberInfo
		NextIndex      uint32
		NextRoleId     uint64
		NextQueryParam string
		Finished       bool
	}

	ChannelType int32
)

const (
	ChannelTypeText  ChannelType = 1
	ChannelTypeVoice ChannelType = 2
	ChannelTypeLive  ChannelType = 5
	ChannelTypeTopic ChannelType = 7
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

func (g *GuildInfo) FindChannel(channelId uint64) *ChannelInfo {
	for _, c := range g.Channels {
		if c.ChannelId == channelId {
			return c
		}
	}
	return nil
}

func (g *GuildInfo) removeChannel(id uint64) {
	i := sort.Search(len(g.Channels), func(i int) bool {
		return g.Channels[i].ChannelId >= id
	})
	if i >= len(g.Channels) || g.Channels[i].ChannelId != id {
		return
	}
	g.Channels = append(g.Channels[:i], g.Channels[i+1:]...)
}

func (s *GuildService) GetUserProfile(tinyId uint64) (*GuildUserProfile, error) {
	flags := proto.DynamicMessage{}
	for i := 3; i <= 29; i++ {
		flags[uint64(i)] = uint32(1)
	}
	flags[99] = uint32(1)
	flags[100] = uint32(1)
	payload := s.c.packOIDBPackageDynamically(3976, 1, proto.DynamicMessage{
		1: flags,
		3: tinyId,
		4: uint32(0),
	})
	rsp, err := s.c.sendAndWaitDynamic(s.c.uniPacket("OidbSvcTrpcTcp.0xfc9_1", payload))
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xfc9Rsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
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

// FetchGuildMemberListWithRole 获取频道成员列表
// 第一次请求: startIndex = 0 , roleIdIndex = 2 param = ""
// 后续请求请根据上次请求的返回值进行设置
func (s *GuildService) FetchGuildMemberListWithRole(guildId, channelId uint64, startIndex uint32, roleIdIndex uint64, param string) (*FetchGuildMemberListWithRoleResult, error) {
	seq := s.c.nextSeq()
	u1 := uint32(1)
	m := proto.DynamicMessage{
		1: guildId, // guild id
		2: uint32(3),
		3: uint32(0),
		4: proto.DynamicMessage{ // unknown param, looks like flags
			1: u1, 2: u1, 3: u1, 4: u1, 5: u1, 6: u1, 7: u1, 8: u1, 20: u1,
		},
		6:  startIndex,
		8:  uint32(50), // count
		12: channelId,
	}
	if param != "" {
		m[13] = param
	}
	m[14] = roleIdIndex
	packet := s.c.uniPacketWithSeq(seq, "OidbSvcTrpcTcp.0xf5b_1", s.c.packOIDBPackageDynamically(3931, 1, m))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf5BRsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	var ret []*GuildMemberInfo
	for _, memberWithRole := range body.MemberWithRoles {
		for _, mem := range memberWithRole.Members {
			ret = append(ret, &GuildMemberInfo{
				TinyId:        mem.GetTinyId(),
				Title:         mem.GetTitle(),
				Nickname:      mem.GetNickname(),
				LastSpeakTime: mem.GetLastSpeakTime(),
				Role:          memberWithRole.GetRoleId(),
				RoleName:      memberWithRole.GetRoleName(),
			})
		}
	}
	for _, mem := range body.Members {
		ret = append(ret, &GuildMemberInfo{
			TinyId:        mem.GetTinyId(),
			Title:         mem.GetTitle(),
			Nickname:      mem.GetNickname(),
			LastSpeakTime: mem.GetLastSpeakTime(),
			Role:          1,
			RoleName:      "普通成员",
		})
	}
	return &FetchGuildMemberListWithRoleResult{
		Members:        ret,
		NextIndex:      body.GetNextIndex(),
		NextRoleId:     body.GetNextRoleIdIndex(),
		NextQueryParam: body.GetNextQueryParam(),
		Finished:       body.NextIndex == nil,
	}, nil
}

// FetchGuildMemberProfileInfo 获取单个频道成员资料
func (s *GuildService) FetchGuildMemberProfileInfo(guildId, tinyId uint64) (*GuildUserProfile, error) {
	seq := s.c.nextSeq()
	flags := proto.DynamicMessage{}
	for i := 3; i <= 29; i++ {
		flags[uint64(i)] = uint32(1)
	}
	flags[99] = uint32(1)
	flags[100] = uint32(1)
	payload := s.c.packOIDBPackageDynamically(3976, 1, proto.DynamicMessage{
		1: flags,
		3: tinyId,
		4: guildId,
	})
	packet := s.c.uniPacketWithSeq(seq, "OidbSvcTrpcTcp.0xf88_1", payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf88Rsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	roles, err := s.fetchMemberRoles(guildId, tinyId)
	if err != nil {
		return nil, errors.Wrap(err, "fetch roles error")
	}
	// todo: 解析个性档案
	return &GuildUserProfile{
		TinyId:    tinyId,
		Nickname:  body.Profile.GetNickname(),
		AvatarUrl: body.Profile.GetAvatarUrl(),
		JoinTime:  body.Profile.GetJoinTime(),
		Roles:     roles,
	}, nil
}

func (s *GuildService) GetGuildRoles(guildId uint64) ([]*GuildRole, error) {
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0x1019_1",
		s.c.packOIDBPackageDynamically(4121, 1, proto.DynamicMessage{1: guildId}))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0X1019Rsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	roles := make([]*GuildRole, 0, len(body.Roles))
	for _, role := range body.Roles {
		roles = append(roles, &GuildRole{
			RoleId:      role.GetRoleId(),
			RoleName:    role.GetName(),
			ArgbColor:   role.GetArgbColor(),
			Independent: role.GetIndependent() == 1,
			Num:         role.GetNum(),
			Owned:       role.GetOwned() == 1,
			Disabled:    role.GetDisabled() == 1,
			MaxNum:      role.GetMaxNum(),
		})
	}
	return roles, nil
}

func (s *GuildService) CreateGuildRole(guildId uint64, name string, color uint32, independent bool, initialUsers []uint64) (uint64, error) {
	u1 := uint32(1)
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0x1016_1", s.c.packOIDBPackageDynamically(4118, 1, proto.DynamicMessage{
		1: guildId,
		2: proto.DynamicMessage{ // todo: 未知参数
			1: u1,
			2: u1,
			3: u1,
		},
		3: proto.DynamicMessage{
			1: name,
			2: color,
			3: independent,
		},
		4: initialUsers,
	}))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return 0, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0X1016Rsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
		return 0, errors.Wrap(err, "decode packet error")
	}
	return body.GetRoleId(), nil
}

func (s *GuildService) DeleteGuildRole(guildId uint64, roleId uint64) error {
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0x100e_1", s.c.packOIDBPackageDynamically(4110, 1, proto.DynamicMessage{
		1: guildId,
		2: roleId,
	}))
	_, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return errors.Wrap(err, "send packet error")
	}
	return nil
}

func (s *GuildService) SetUserRoleInGuild(guildId uint64, set bool, roleId uint64, user []uint64) error { // remove => p2 = false
	setOrRemove := proto.DynamicMessage{
		1: roleId,
	}
	if set {
		setOrRemove[2] = user
	} else {
		setOrRemove[3] = user
	}
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0x101a_1", s.c.packOIDBPackageDynamically(4122, 1, proto.DynamicMessage{
		1: guildId,
		2: setOrRemove,
	}))
	_, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return errors.Wrap(err, "send packet error")
	}
	return nil
}

func (s *GuildService) ModifyRoleInGuild(guildId uint64, roleId uint64, name string, color uint32, indepedent bool) error {
	u1 := uint32(1)
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0x100d_1", s.c.packOIDBPackageDynamically(4109, 1, proto.DynamicMessage{
		1: guildId,
		2: roleId,
		3: proto.DynamicMessage{
			1: u1,
			2: u1,
			3: u1,
		},
		4: proto.DynamicMessage{
			1: name,
			2: color,
			3: indepedent,
		},
	}))
	_, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return errors.Wrap(err, "send packet error")
	}
	return nil
}

func (s *GuildService) FetchGuestGuild(guildId uint64) (*GuildMeta, error) {
	u1 := uint32(1)
	payload := s.c.packOIDBPackageDynamically(3927, 9, proto.DynamicMessage{
		1: proto.DynamicMessage{
			1: proto.DynamicMessage{
				2: u1, 4: u1, 5: u1, 6: u1, 7: u1, 8: u1, 11: u1, 12: u1, 13: u1, 14: u1, 45: u1,
				18: u1, 19: u1, 20: u1, 22: u1, 23: u1, 5002: u1, 5003: u1, 5004: u1, 5005: u1, 10007: u1,
			},
			2: proto.DynamicMessage{
				3: u1, 4: u1, 6: u1, 11: u1, 14: u1, 15: u1, 16: u1, 17: u1,
			},
		},
		2: proto.DynamicMessage{
			1: guildId,
		},
	})
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0xf57_9", payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf57Rsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
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
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0xf5d_1",
		s.c.packOIDBPackageDynamically(3933, 1,
			proto.DynamicMessage{
				1: guildId,
				3: proto.DynamicMessage{
					1: uint32(1),
				},
			}))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf5DRsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	for _, info := range body.Rsp.Channels {
		r = append(r, convertChannelInfo(info))
	}
	return
}

func (s *GuildService) FetchChannelInfo(guildId, channelId uint64) (*ChannelInfo, error) {
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0xf55_1", s.c.packOIDBPackageDynamically(3925, 1, proto.DynamicMessage{1: guildId, 2: channelId}))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0Xf55Rsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	return convertChannelInfo(body.Info), nil
}

func (s *GuildService) GetTopicChannelFeeds(guildId, channelId uint64) ([]*topic.Feed, error) {
	guild := s.FindGuild(guildId)
	if guild == nil {
		return nil, errors.New("guild not found")
	}
	channelInfo := guild.FindChannel(channelId)
	if channelInfo == nil {
		return nil, errors.New("channel not found")
	}
	if channelInfo.ChannelType != ChannelTypeTopic {
		return nil, errors.New("channel type error")
	}
	req, _ := proto.Marshal(&channel.StGetChannelFeedsReq{
		Count: proto.Uint32(12),
		From:  proto.Uint32(0),
		ChannelSign: &channel.StChannelSign{
			GuildId:   &guildId,
			ChannelId: &channelId,
		},
		FeedAttchInfo: proto.String(""), // isLoadMore
	})
	payload, _ := proto.Marshal(&qweb.QWebReq{
		Seq:        proto.Int64(s.c.nextQWebSeq()),
		Qua:        proto.String("V1_AND_SQ_8.8.50_2324_YYB_D"),
		DeviceInfo: proto.String(s.c.getWebDeviceInfo()),
		BusiBuff:   req,
		TraceId:    proto.String(fmt.Sprintf("%v_%v_%v", s.c.Uin, time.Now().Format("0102150405"), rand.Int63())),
		Extinfo: []*qweb.COMMEntry{
			{
				Key:   proto.String("fc-appid"),
				Value: proto.String("96"),
			},
			{
				Key:   proto.String("environment_id"),
				Value: proto.String("production"),
			},
			{
				Key:   proto.String("tiny_id"),
				Value: proto.String(fmt.Sprint(s.TinyId)),
			},
		},
	})
	seq, packet := s.c.uniPacket("QChannelSvr.trpc.qchannel.commreader.ComReader.GetChannelTimelineFeeds", payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.New("send packet error")
	}
	pkg := new(qweb.QWebRsp)
	body := new(channel.StGetChannelFeedsRsp)
	if err = proto.Unmarshal(rsp, pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err = proto.Unmarshal(pkg.BusiBuff, body); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	feeds := make([]*topic.Feed, 0, len(body.VecFeed))
	for _, f := range body.VecFeed {
		feeds = append(feeds, topic.DecodeFeed(f))
	}
	return feeds, nil
}

func (s *GuildService) PostTopicChannelFeed(guildId, channelId uint64, feed *topic.Feed) error {
	guild := s.FindGuild(guildId)
	if guild == nil {
		return errors.New("guild not found")
	}
	channelInfo := guild.FindChannel(channelId)
	if channelInfo == nil {
		return errors.New("channel not found")
	}
	if channelInfo.ChannelType != ChannelTypeTopic {
		return errors.New("channel type error")
	}
	feed.Poster = &topic.FeedPoster{
		TinyIdStr: strconv.FormatUint(s.TinyId, 10),
		Nickname:  s.Nickname,
	}
	feed.GuildId = guildId
	feed.ChannelId = channelId
	req, _ := proto.Marshal(&channel.StPublishFeedReq{
		ExtInfo: &channel.StCommonExt{
			MapInfo: []*channel.CommonEntry{
				{
					Key: proto.String("clientkey"), Value: proto.String("GuildMain" + utils.RandomStringRange(14, "0123456789")),
				},
			},
		},
		From:     proto.Int32(0),
		JsonFeed: proto.String(feed.ToSendingPayload(s.c.Uin)),
	})
	payload, _ := proto.Marshal(&qweb.QWebReq{
		Seq:        proto.Int64(s.c.nextQWebSeq()),
		Qua:        proto.String("V1_AND_SQ_8.8.50_2324_YYB_D"),
		DeviceInfo: proto.String(s.c.getWebDeviceInfo()),
		BusiBuff:   req,
		TraceId:    proto.String(fmt.Sprintf("%v_%v_%v", s.c.Uin, time.Now().Format("0102150405"), rand.Int63())),
		Extinfo: []*qweb.COMMEntry{
			{
				Key:   proto.String("fc-appid"),
				Value: proto.String("96"),
			},
			{
				Key:   proto.String("environment_id"),
				Value: proto.String("production"),
			},
			{
				Key:   proto.String("tiny_id"),
				Value: proto.String(fmt.Sprint(s.TinyId)),
			},
		},
	})
	seq, packet := s.c.uniPacket("QChannelSvr.trpc.qchannel.commwriter.ComWriter.PublishFeed", payload)
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return errors.New("send packet error")
	}
	pkg := new(qweb.QWebRsp)
	body := new(channel.StPublishFeedRsp)
	if err = proto.Unmarshal(rsp, pkg); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err = proto.Unmarshal(pkg.BusiBuff, body); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if body.Feed != nil && body.Feed.Id != nil {
		return nil
	}
	return errors.New("post feed error")
}

func (s *GuildService) fetchMemberRoles(guildId uint64, tinyId uint64) ([]*GuildRole, error) {
	u1 := uint32(1)
	seq, packet := s.c.uniPacket("OidbSvcTrpcTcp.0x1017_1", s.c.packOIDBPackageDynamically(4119, 1, proto.DynamicMessage{
		1: guildId,
		2: tinyId,
		4: proto.DynamicMessage{
			1: u1,
			2: u1,
			3: u1,
		},
	}))
	rsp, err := s.c.sendAndWaitDynamic(seq, packet)
	if err != nil {
		return nil, errors.Wrap(err, "send packet error")
	}
	body := new(channel.ChannelOidb0X1017Rsp)
	if err = unpackOIDBPackage(rsp, body); err != nil {
		return nil, errors.Wrap(err, "decode packet error")
	}
	p1 := body.P1
	if p1 == nil {
		return nil, errors.New("packet OidbSvcTrpcTcp.0x1017_1: decode p1 error")
	}
	roles := make([]*GuildRole, 0, len(p1.Roles))
	for _, role := range p1.Roles {
		roles = append(roles, &GuildRole{
			RoleId:    role.GetRoleId(),
			RoleName:  role.GetName(),
			ArgbColor: role.GetArgbColor(),
		})
	}
	return roles, nil
}

/* need analysis
func (s *GuildService) fetchChannelListState(guildId uint64, channels []*ChannelInfo) {
	seq := s.c.nextSeq()
	var ids []uint64
	for _, info := range channels {
		ids = append(ids, info.ChannelId)
	}
	payload := s.c.packOIDBPackageDynamically(4104, 1, binary.DynamicMessage{
		1: binary.DynamicMessage{
			1: guildId,
			2: ids,
		},
	})
	packet := packets.BuildUniPacket(s.c.Uin, seq, "OidbSvcTrpcTcp.0x1008_1", 1, s.c.SessionId, []byte{}, s.c.sigInfo.d2Key, payload)
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
		c.error("sync channel error: %v", err)
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
		c.error("get self guild profile error: %v", err)
	}
}

func (c *QQClient) buildSyncChannelFirstViewPacket() (uint16, []byte) {
	req := &channel.FirstViewReq{
		LastMsgTime:       proto.Uint64(0),
		Seq:               proto.Uint32(0),
		DirectMessageFlag: proto.Uint32(1),
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("trpc.group_pro.synclogic.SyncLogic.SyncFirstView", payload)
}

func decodeGuildPushFirstView(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
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
				c.warning("waring: fetch guild %v channel error %v. will use sync node to fill channel list field", guild.GuildId, err)
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
			// info.Bots, info.Members, info.Admins, _ = c.GuildService.GetGuildMembers(info.GuildId)
			c.GuildService.Guilds = append(c.GuildService.Guilds, info)
		}
	}
	if len(firstViewMsg.ChannelMsgs) > 0 { // sync msg
	}
	return nil, nil
}
