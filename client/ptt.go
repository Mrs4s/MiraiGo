package client

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x346"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/pttcenter"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
)

func init() {
	decoders["PttCenterSvr.ShortVideoDownReq"] = decodePttShortVideoDownResponse
	decoders["PttCenterSvr.GroupShortVideoUpReq"] = decodeGroupShortVideoUploadResponse
}

// UploadGroupPtt 将语音数据使用群语音通道上传到服务器, 返回 message.GroupVoiceElement 可直接发送
func (c *QQClient) UploadGroupPtt(groupCode int64, voice io.ReadSeeker) (*message.GroupVoiceElement, error) {
	h := md5.New()
	length, _ := io.Copy(h, voice)
	fh := h.Sum(nil)
	_, _ = voice.Seek(0, io.SeekStart)
	ext := c.buildGroupPttStoreBDHExt(groupCode, fh[:], int32(length), 0, int32(length))
	rsp, err := c.highwayUploadByBDH(voice, length, 29, c.bigDataSession.SigSession, fh, ext, false)
	if err != nil {
		return nil, err
	}
	if len(rsp) == 0 {
		return nil, errors.New("miss rsp")
	}
	pkt := pb.D388RespBody{}
	if err = proto.Unmarshal(rsp, &pkt); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(pkt.MsgTryUpPttRsp) == 0 {
		return nil, errors.New("miss try up rsp")
	}
	return &message.GroupVoiceElement{
		Ptt: &msg.Ptt{
			FileType:     proto.Int32(4),
			SrcUin:       &c.Uin,
			FileMd5:      fh[:],
			FileName:     proto.String(hex.EncodeToString(fh[:]) + ".amr"),
			FileSize:     proto.Int32(int32(length)),
			GroupFileKey: pkt.MsgTryUpPttRsp[0].FileKey,
			BoolValid:    proto.Bool(true),
			PbReserve:    []byte{8, 0, 40, 0, 56, 0},
		},
	}, nil
}

// UploadPrivatePtt 将语音数据使用好友语音通道上传到服务器, 返回 message.PrivateVoiceElement 可直接发送
func (c *QQClient) UploadPrivatePtt(target int64, voice io.ReadSeeker) (*message.PrivateVoiceElement, error) {
	h := md5.New()
	length, _ := io.Copy(h, voice)
	fh := h.Sum(nil)
	_, _ = voice.Seek(0, io.SeekStart)
	ext := c.buildC2CPttStoreBDHExt(target, fh[:], int32(length), int32(length))
	rsp, err := c.highwayUploadByBDH(voice, length, 26, c.bigDataSession.SigSession, fh, ext, false)
	if err != nil {
		return nil, err
	}
	if len(rsp) == 0 {
		return nil, errors.New("miss rsp")
	}
	pkt := cmd0x346.C346RspBody{}
	if err = proto.Unmarshal(rsp, &pkt); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if pkt.ApplyUploadRsp == nil {
		return nil, errors.New("miss apply upload rsp")
	}
	return &message.PrivateVoiceElement{
		Ptt: &msg.Ptt{
			FileType: proto.Int32(4),
			SrcUin:   &c.Uin,
			FileUuid: pkt.ApplyUploadRsp.Uuid,
			FileMd5:  fh[:],
			FileName: proto.String(hex.EncodeToString(fh[:]) + ".amr"),
			FileSize: proto.Int32(int32(length)),
			// Reserve:   constructPTTExtraInfo(1, int32(len(voice))), // todo length
			BoolValid: proto.Bool(true),
		},
	}, nil
}

// UploadGroupShortVideo 将视频和封面上传到服务器, 返回 message.ShortVideoElement 可直接发送
// combinedCache 本地文件缓存, 设置后可多线程上传
func (c *QQClient) UploadGroupShortVideo(groupCode int64, video, thumb io.ReadSeeker, combinedCache ...string) (*message.ShortVideoElement, error) {
	videoHash, videoLen := utils.ComputeMd5AndLength(video)
	thumbHash, thumbLen := utils.ComputeMd5AndLength(thumb)
	cache := ""
	if len(combinedCache) > 0 {
		cache = combinedCache[0]
	}
	i, err := c.sendAndWait(c.buildPttGroupShortVideoUploadReqPacket(videoHash, thumbHash, groupCode, videoLen, thumbLen))
	if err != nil {
		return nil, errors.Wrap(err, "upload req error")
	}
	rsp := i.(*pttcenter.ShortVideoUploadRsp)
	if rsp.FileExists == 1 {
		return &message.ShortVideoElement{
			Uuid:      []byte(rsp.FileId),
			Size:      int32(videoLen),
			ThumbSize: int32(thumbLen),
			Md5:       videoHash,
			ThumbMd5:  thumbHash,
		}, nil
	}
	ext, _ := proto.Marshal(c.buildPttGroupShortVideoProto(videoHash, thumbHash, groupCode, videoLen, thumbLen).PttShortVideoUploadReq)
	var hwRsp []byte
	multi := utils.MultiReadSeeker(thumb, video)
	h := md5.New()
	length, _ := io.Copy(h, multi)
	fh := h.Sum(nil)
	_, _ = multi.Seek(0, io.SeekStart)
	if cache != "" {
		var file *os.File
		file, err = os.OpenFile(cache, os.O_WRONLY|os.O_CREATE, 0o666)
		cp := func() error {
			_, err := io.Copy(file, utils.MultiReadSeeker(thumb, video))
			return err
		}
		if err != nil || cp() != nil {
			hwRsp, err = c.highwayUploadByBDH(multi, length, 25, c.bigDataSession.SigSession, fh, ext, true)
		} else {
			_ = file.Close()
			hwRsp, err = c.highwayUploadFileMultiThreadingByBDH(cache, 25, 8, c.bigDataSession.SigSession, ext, true)
			_ = os.Remove(cache)
		}
	} else {
		hwRsp, err = c.highwayUploadByBDH(multi, length, 25, c.bigDataSession.SigSession, fh, ext, true)
	}
	if err != nil {
		return nil, errors.Wrap(err, "upload video file error")
	}
	if len(hwRsp) == 0 {
		return nil, errors.New("resp is empty")
	}
	rsp = &pttcenter.ShortVideoUploadRsp{}
	if err = proto.Unmarshal(hwRsp, rsp); err != nil {
		return nil, errors.Wrap(err, "decode error")
	}
	return &message.ShortVideoElement{
		Uuid:      []byte(rsp.FileId),
		Size:      int32(videoLen),
		ThumbSize: int32(thumbLen),
		Md5:       videoHash,
		ThumbMd5:  thumbHash,
	}, nil
}

func (c *QQClient) GetShortVideoUrl(uuid, md5 []byte) string {
	i, err := c.sendAndWait(c.buildPttShortVideoDownReqPacket(uuid, md5))
	if err != nil {
		return ""
	}
	return i.(string)
}

// PttStore.GroupPttUp
func (c *QQClient) buildGroupPttStorePacket(groupCode int64, md5 []byte, size, codec, voiceLength int32) (uint16, []byte) {
	seq := c.nextSeq()
	packet := packets.BuildUniPacket(c.Uin, seq, "PttStore.GroupPttUp", 1, c.OutGoingPacketSessionId,
		EmptyBytes, c.sigInfo.d2Key, c.buildGroupPttStoreBDHExt(groupCode, md5, size, codec, voiceLength))
	return seq, packet
}

func (c *QQClient) buildGroupPttStoreBDHExt(groupCode int64, md5 []byte, size, codec, voiceLength int32) []byte {
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
	}
	payload, _ := proto.Marshal(req)
	return payload
}

// PttCenterSvr.ShortVideoDownReq
func (c *QQClient) buildPttShortVideoDownReqPacket(uuid, md5 []byte) (uint16, []byte) {
	seq := c.nextSeq()
	body := &pttcenter.ShortVideoReqBody{
		Cmd: 400,
		Seq: int32(seq),
		PttShortVideoDownloadReq: &pttcenter.ShortVideoDownloadReq{
			FromUin:      c.Uin,
			ToUin:        c.Uin,
			ChatType:     1,
			ClientType:   7,
			FileId:       string(uuid),
			GroupCode:    1,
			FileMd5:      md5,
			BusinessType: 1,
			FileType:     2,
			DownType:     2,
			SceneType:    2,
		},
	}
	payload, _ := proto.Marshal(body)
	packet := packets.BuildUniPacket(c.Uin, seq, "PttCenterSvr.ShortVideoDownReq", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildPttGroupShortVideoProto(videoHash, thumbHash []byte, toUin, videoSize, thumbSize int64) *pttcenter.ShortVideoReqBody {
	seq := c.nextSeq()
	return &pttcenter.ShortVideoReqBody{
		Cmd: 300,
		Seq: int32(seq),
		PttShortVideoUploadReq: &pttcenter.ShortVideoUploadReq{
			FromUin:    c.Uin,
			ToUin:      toUin,
			ChatType:   1,
			ClientType: 2,
			Info: &pttcenter.ShortVideoFileInfo{
				FileName:      hex.EncodeToString(videoHash) + ".mp4",
				FileMd5:       videoHash,
				ThumbFileMd5:  thumbHash,
				FileSize:      videoSize,
				FileResLength: 1280,
				FileResWidth:  720,
				FileFormat:    3,
				FileTime:      120,
				ThumbFileSize: thumbSize,
			},
			GroupCode:        toUin,
			SupportLargeSize: 1,
		},
		ExtensionReq: []*pttcenter.ShortVideoExtensionReq{
			{
				SubBusiType: 0,
				UserCnt:     1,
			},
		},
	}
}

// PttCenterSvr.GroupShortVideoUpReq
func (c *QQClient) buildPttGroupShortVideoUploadReqPacket(videoHash, thumbHash []byte, toUin, videoSize, thumbSize int64) (uint16, []byte) {
	seq := c.nextSeq()
	payload, _ := proto.Marshal(c.buildPttGroupShortVideoProto(videoHash, thumbHash, toUin, videoSize, thumbSize))
	packet := packets.BuildUniPacket(c.Uin, seq, "PttCenterSvr.GroupShortVideoUpReq", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// PttStore.GroupPttUp
func decodeGroupPttStoreResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
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
	ip := make([]string, 0, len(rsp.Uint32UpIp))
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
func (c *QQClient) buildC2CPttStoreBDHExt(target int64, md5 []byte, size, voiceLength int32) []byte {
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
	return payload
}

// PttCenterSvr.ShortVideoDownReq
func decodePttShortVideoDownResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	rsp := pttcenter.ShortVideoRspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.PttShortVideoDownloadRsp == nil || rsp.PttShortVideoDownloadRsp.DownloadAddr == nil {
		return nil, errors.New("resp error")
	}
	return rsp.PttShortVideoDownloadRsp.DownloadAddr.Host[0] + rsp.PttShortVideoDownloadRsp.DownloadAddr.UrlArgs, nil
}

// PttCenterSvr.GroupShortVideoUpReq
func decodeGroupShortVideoUploadResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	rsp := pttcenter.ShortVideoRspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.PttShortVideoUploadRsp == nil {
		return nil, errors.New("resp error")
	}
	if rsp.PttShortVideoUploadRsp.RetCode != 0 {
		return nil, errors.Errorf("ret code error: %v", rsp.PttShortVideoUploadRsp.RetCode)
	}
	return rsp.PttShortVideoUploadRsp, nil
}

func constructPTTExtraInfo(codec, length int32) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		w.WriteByte(3)
		w.WriteByte(8)
		w.WriteUInt16(4)
		w.WriteUInt32(uint32(codec))
		w.WriteByte(9)
		w.WriteUInt16(4)
		w.WriteUInt32(uint32(14)) // length 时间
		w.WriteByte(10)
		info := []byte{0x08, 0x00, 0x28, 0x00, 0x38, 0x00} // todo
		w.WriteUInt16(uint16(len(info)))
		w.Write(info)
	})
}
