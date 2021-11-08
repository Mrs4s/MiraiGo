package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/internal/packets"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *GuildService) pullRoamMsgByEventFlow(guildId, channelId, beginSeq, endSeq, eventVersion uint64) ([]*channel.ChannelMsgContent, error) {
	payload, _ := proto.Marshal(&channel.ChannelMsgReq{
		ChannelParam: &channel.ChannelParam{
			GuildId:   &guildId,
			ChannelId: &channelId,
			BeginSeq:  &beginSeq,
			EndSeq:    &endSeq,
			Version:   []uint64{eventVersion},
		},
		WithVersionFlag:   proto.Uint32(1),
		DirectMessageFlag: proto.Uint32(0),
	})
	seq := s.c.nextSeq()
	packet := packets.BuildUniPacket(s.c.Uin, seq, "trpc.group_pro.synclogic.SyncLogic.GetChannelMsg", 1, s.c.OutGoingPacketSessionId, []byte{}, s.c.sigInfo.d2Key, payload)
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
