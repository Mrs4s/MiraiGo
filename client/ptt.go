package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"google.golang.org/protobuf/proto"
)

// PttStore.GroupPttUp
func (c *QQClient) buildGroupPttStorePacket(groupCode int64, md5 []byte, size, codec, voiceLength int32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &pb.D388ReqBody{
		NetType: 3,
		Subcmd:  3,
		MsgTryUpPttReq: []*pb.TryUpPttReq{
			{
				GroupCode:     groupCode,
				SrcUin:        c.Uin,
				FileMd5:       md5,
				FileSize:      int64(size),
				FileName:      md5,
				SrcTerm:       5,
				PlatformType:  9,
				BuType:        4,
				InnerIp:       0,
				BuildVer:      "6.5.5.663",
				VoiceLength:   voiceLength,
				Codec:         codec,
				VoiceType:     1,
				BoolNewUpChan: true,
			},
		},
		Extension: EmptyBytes,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "PttStore.GroupPttUp", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// PttStore.GroupPttUp
func decodeGroupPttStoreResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkt := pb.D388RespBody{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, err
	}
	rsp := pkt.MsgTryUpPttRsp[0]
	if rsp.Result != 0 {
		return pttUploadResponse{
			ResultCode: rsp.Result,
			Message:    rsp.FailMsg,
		}, nil
	}
	if rsp.BoolFileExit {
		return pttUploadResponse{IsExists: true}, nil
	}
	return pttUploadResponse{
		UploadKey:  rsp.UpUkey,
		UploadIp:   rsp.Uint32UpIp,
		UploadPort: rsp.Uint32UpPort,
		FileKey:    rsp.FileKey,
	}, nil
}
