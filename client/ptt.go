package client

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x346"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// 语音相关处理逻辑

// UploadGroupPtt 将语音数据使用群语音通道上传到服务器, 返回 message.GroupVoiceElement 可直接发送
func (c *QQClient) UploadGroupPtt(groupCode int64, voice []byte) (*message.GroupVoiceElement, error) {
	h := md5.Sum(voice)
	seq, pkt := c.buildGroupPttStorePacket(groupCode, h[:], int32(len(voice)), 0, int32(len(voice)))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(pttUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		goto ok
	}
	for i, ip := range rsp.UploadIp {
		err := c.uploadPtt(ip, rsp.UploadPort[i], rsp.UploadKey, rsp.FileKey, voice, h[:])
		if err != nil {
			continue
		}
		goto ok
	}
	return nil, errors.New("upload failed")
ok:
	return &message.GroupVoiceElement{
		Ptt: &msg.Ptt{
			FileType:     proto.Int32(4),
			SrcUin:       &c.Uin,
			FileMd5:      h[:],
			FileName:     proto.String(hex.EncodeToString(h[:]) + ".amr"),
			FileSize:     proto.Int32(int32(len(voice))),
			GroupFileKey: rsp.FileKey,
			BoolValid:    proto.Bool(true),
			PbReserve:    []byte{8, 0, 40, 0, 56, 0},
		}}, nil
}

// UploadPrivatePtt 将语音数据使用好友语音通道上传到服务器, 返回 message.PrivateVoiceElement 可直接发送
func (c *QQClient) UploadPrivatePtt(target int64, voice []byte) (*message.PrivateVoiceElement, error) {
	h := md5.Sum(voice)
	i, err := c.sendAndWait(c.buildPrivatePttStorePacket(target, h[:], int32(len(voice)), int32(len(voice))))
	if err != nil {
		return nil, err
	}
	rsp := i.(pttUploadResponse)
	if rsp.IsExists {
		goto ok
	}
	for i, ip := range rsp.UploadIp {
		err := c.uploadPtt(ip, rsp.UploadPort[i], rsp.UploadKey, rsp.FileKey, voice, h[:])
		if err != nil {
			continue
		}
		goto ok
	}
	return nil, errors.New("upload failed")
ok:
	return &message.PrivateVoiceElement{
		Ptt: &msg.Ptt{
			FileType:  proto.Int32(4),
			SrcUin:    &c.Uin,
			FileMd5:   h[:],
			FileName:  proto.String(hex.EncodeToString(h[:]) + ".amr"),
			FileSize:  proto.Int32(int32(len(voice))),
			FileKey:   rsp.FileKey,
			BoolValid: proto.Bool(true),
			PbReserve: []byte{8, 0, 40, 0, 56, 0},
		}}, nil
}

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
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
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
	var ip []string
	for _, i := range rsp.Uint32UpIp {
		ip = append(ip, binary.UInt32ToIPV4Address(uint32(i)))
	}
	return pttUploadResponse{
		UploadKey:  rsp.UpUkey,
		UploadIp:   ip,
		UploadPort: rsp.Uint32UpPort,
		FileKey:    rsp.FileKey,
		FileId:     rsp.FileId2,
	}, nil
}

// PttCenterSvr.pb_pttCenter_CMD_REQ_APPLY_UPLOAD-500
func (c *QQClient) buildPrivatePttStorePacket(target int64, md5 []byte, size, voiceLength int32) (uint16, []byte) {
	seq := c.nextSeq()
	req := &cmd0x346.C346ReqBody{
		Cmd: 500,
		Seq: int32(seq),
		ApplyUploadReq: &cmd0x346.ApplyUploadReq{
			SenderUin:    c.Uin,
			RecverUin:    target,
			FileType:     2,
			FileSize:     int64(size),
			FileName:     hex.EncodeToString(md5),
			Bytes_10MMd5: md5, // 超过10M可能会炸
		},
		BusinessId: 17,
		ClientType: 104,
		ExtensionReq: &cmd0x346.ExtensionReq{
			Id:        3,
			PttFormat: 1,
			NetType:   3,
			VoiceType: 2,
			PttTime:   voiceLength,
		},
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "PttCenterSvr.pb_pttCenter_CMD_REQ_APPLY_UPLOAD-500", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// PttCenterSvr.pb_pttCenter_CMD_REQ_APPLY_UPLOAD-500
func decodePrivatePttStoreResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := cmd0x346.C346RspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		c.Error("unmarshal cmd0x346 rsp body error: %v", err)
		return nil, errors.Wrap(err, "unmarshal cmd0x346 rsp body error")
	}
	if rsp.ApplyUploadRsp == nil {
		c.Error("decode apply upload 500 error: apply rsp is nil.")
		return nil, errors.New("apply rsp is nil")
	}
	if rsp.ApplyUploadRsp.RetCode != 0 {
		c.Error("decode apply upload 500 error: %v", rsp.ApplyUploadRsp.RetCode)
		return nil, errors.Errorf("apply upload rsp error: %d", rsp.ApplyUploadRsp.RetCode)
	}
	if rsp.ApplyUploadRsp.BoolFileExist {
		return pttUploadResponse{IsExists: true}, nil
	}
	var port []int32
	for range rsp.ApplyUploadRsp.UploadipList {
		port = append(port, rsp.ApplyUploadRsp.UploadPort)
	}
	return pttUploadResponse{
		UploadKey:  rsp.ApplyUploadRsp.UploadKey,
		UploadIp:   rsp.ApplyUploadRsp.UploadipList,
		UploadPort: port,
		FileKey:    rsp.ApplyUploadRsp.Uuid,
	}, nil
}
