package client

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/highway"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x346"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x388"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/pttcenter"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

func init() {
	decoders["PttCenterSvr.ShortVideoDownReq"] = decodePttShortVideoDownResponse
	decoders["PttCenterSvr.GroupShortVideoUpReq"] = decodeGroupShortVideoUploadResponse
}

var pttWaiter = utils.NewUploadWaiter()

var c2cPttExtraInfo = binary.NewWriterF(func(w *binary.Writer) {
	w.WriteByte(2) // tlv count

	w.WriteByte(8)
	w.WriteUInt16(4)
	w.WriteUInt32(1) // codec

	w.WriteByte(9)
	w.WriteUInt16(4)
	w.WriteUInt32(0) // 时长

	w.WriteByte(10)
	reserveInfo := []byte{0x08, 0x00, 0x28, 0x00, 0x38, 0x00} // todo
	w.WriteBytesShort(reserveInfo)
})

// UploadGroupPtt 将语音数据使用群语音通道上传到服务器, 返回 message.GroupVoiceElement 可直接发送
func (c *QQClient) UploadGroupPtt(groupCode int64, voice io.ReadSeeker) (*message.GroupVoiceElement, error) {
	fh, length := utils.ComputeMd5AndLength(voice)
	_, _ = voice.Seek(0, io.SeekStart)

	key := string(fh)
	pttWaiter.Wait(key)
	defer pttWaiter.Done(key)

	ext := c.buildGroupPttStoreBDHExt(groupCode, fh, int32(length), 0, int32(length))
	rsp, err := c.highwaySession.UploadBDH(highway.BdhInput{
		CommandID: 29,
		Body:      voice,
		Ticket:    c.highwaySession.SigSession,
		Ext:       ext,
		Encrypt:   false,
	})
	if err != nil {
		return nil, err
	}
	if len(rsp) == 0 {
		return nil, errors.New("miss rsp")
	}
	pkt := cmd0x388.D388RspBody{}
	if err = proto.Unmarshal(rsp, &pkt); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if len(pkt.TryupPttRsp) == 0 {
		return nil, errors.New("miss try up rsp")
	}
	return &message.GroupVoiceElement{
		Ptt: &msg.Ptt{
			FileType:     proto.Int32(4),
			SrcUin:       &c.Uin,
			FileMd5:      fh,
			FileName:     proto.String(hex.EncodeToString(fh) + ".amr"),
			FileSize:     proto.Int32(int32(length)),
			GroupFileKey: pkt.TryupPttRsp[0].FileKey,
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

	key := hex.EncodeToString(fh)
	pttWaiter.Wait(key)
	defer pttWaiter.Done(key)

	ext := c.buildC2CPttStoreBDHExt(target, fh, int32(length), int32(length))
	rsp, err := c.highwaySession.UploadBDH(highway.BdhInput{
		CommandID: 26,
		Body:      voice,
		Ticket:    c.highwaySession.SigSession,
		Ext:       ext,
		Encrypt:   false,
	})
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
			FileType:  proto.Int32(4),
			SrcUin:    &c.Uin,
			FileUuid:  pkt.ApplyUploadRsp.Uuid,
			FileMd5:   fh,
			FileName:  proto.String(hex.EncodeToString(fh) + ".amr"),
			FileSize:  proto.Int32(int32(length)),
			Reserve:   c2cPttExtraInfo,
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

	key := string(videoHash) + string(thumbHash)
	pttWaiter.Wait(key)
	defer pttWaiter.Done(key)

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
	ext, _ := proto.Marshal(c.buildPttGroupShortVideoProto(videoHash, thumbHash, groupCode, videoLen, thumbLen, 1).PttShortVideoUploadReq)

	var hwRsp []byte
	multi := utils.MultiReadSeeker(thumb, video)
	input := highway.BdhInput{
		CommandID: 25,
		File:      cache,
		Body:      multi,
		Ticket:    c.highwaySession.SigSession,
		Ext:       ext,
		Encrypt:   true,
	}
	if cache != "" {
		var file *os.File
		file, err = os.OpenFile(cache, os.O_WRONLY|os.O_CREATE, 0o666)
		cp := func() error {
			_, err := io.Copy(file, utils.MultiReadSeeker(thumb, video))
			return err
		}
		if err != nil || cp() != nil {
			hwRsp, err = c.highwaySession.UploadBDH(input)
		} else {
			_ = file.Close()
			hwRsp, err = c.highwaySession.UploadBDHMultiThread(input, 8)
			_ = os.Remove(cache)
		}
	} else {
		hwRsp, err = c.highwaySession.UploadBDH(input)
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

func (c *QQClient) buildGroupPttStoreBDHExt(groupCode int64, md5 []byte, size, codec, voiceLength int32) []byte {
	req := &cmd0x388.D388ReqBody{
		NetType: proto.Uint32(3),
		Subcmd:  proto.Uint32(3),
		TryupPttReq: []*cmd0x388.TryUpPttReq{
			{
				GroupCode:    proto.Uint64(uint64(groupCode)),
				SrcUin:       proto.Uint64(uint64(c.Uin)),
				FileMd5:      md5,
				FileSize:     proto.Uint64(uint64(size)),
				FileName:     md5,
				SrcTerm:      proto.Uint32(5),
				PlatformType: proto.Uint32(9),
				BuType:       proto.Uint32(4),
				InnerIp:      proto.Uint32(0),
				BuildVer:     utils.S2B("6.5.5.663"),
				VoiceLength:  proto.Uint32(uint32(voiceLength)),
				Codec:        proto.Uint32(uint32(codec)),
				VoiceType:    proto.Uint32(1),
				NewUpChan:    proto.Bool(true),
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
	packet := c.uniPacketWithSeq(seq, "PttCenterSvr.ShortVideoDownReq", payload)
	return seq, packet
}

func (c *QQClient) buildPttGroupShortVideoProto(videoHash, thumbHash []byte, toUin, videoSize, thumbSize int64, chattype int32) *pttcenter.ShortVideoReqBody {
	seq := c.nextSeq()
	return &pttcenter.ShortVideoReqBody{
		Cmd: 300,
		Seq: int32(seq),
		PttShortVideoUploadReq: &pttcenter.ShortVideoUploadReq{
			FromUin:    c.Uin,
			ToUin:      toUin,
			ChatType:   chattype,
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
	payload, _ := proto.Marshal(c.buildPttGroupShortVideoProto(videoHash, thumbHash, toUin, videoSize, thumbSize, 1))
	return c.uniPacket("PttCenterSvr.GroupShortVideoUpReq", payload)
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
func decodePttShortVideoDownResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
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
func decodeGroupShortVideoUploadResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
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
