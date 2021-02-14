package client

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"os"
	"strings"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func init() {
	decoders["ImgStore.GroupPicUp"] = decodeGroupImageStoreResponse
}

func (c *QQClient) UploadGroupImage(groupCode int64, img io.ReadSeeker) (*message.GroupImageElement, error) {
	_, _ = img.Seek(0, io.SeekStart) // safe
	fh, length := utils.ComputeMd5AndLength(img)
	_, _ = img.Seek(0, io.SeekStart)
	seq, pkt := c.buildGroupImageStorePacket(groupCode, fh[:], int32(length))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		goto ok
	}
	if len(c.srvSsoAddrs) == 0 {
		for i, addr := range rsp.UploadIp {
			c.srvSsoAddrs = append(c.srvSsoAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(uint32(addr)), rsp.UploadPort[i]))
		}
	}
	if _, err = c.highwayUploadByBDH(img, 2, rsp.UploadKey, EmptyBytes, false); err == nil {
		goto ok
	}
	return nil, errors.New("upload failed")
ok:
	_, _ = img.Seek(0, io.SeekStart)
	i, _, _ := image.DecodeConfig(img)
	var imageType int32 = 1000
	_, _ = img.Seek(0, io.SeekStart)
	tmp := make([]byte, 4)
	_, _ = img.Read(tmp)
	if bytes.Equal(tmp, []byte{0x47, 0x49, 0x46, 0x38}) {
		imageType = 2000
	}
	return message.NewGroupImage(binary.CalculateImageResourceId(fh[:]), fh[:], rsp.FileId, int32(length), int32(i.Width), int32(i.Height), imageType), nil
}

func (c *QQClient) UploadGroupImageByFile(groupCode int64, path string) (*message.GroupImageElement, error) {
	img, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer img.Close()
	fh, length := utils.ComputeMd5AndLength(img)
	seq, pkt := c.buildGroupImageStorePacket(groupCode, fh[:], int32(length))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		goto ok
	}
	if len(c.srvSsoAddrs) == 0 {
		for i, addr := range rsp.UploadIp {
			c.srvSsoAddrs = append(c.srvSsoAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(uint32(addr)), rsp.UploadPort[i]))
		}
	}
	if _, err = c.highwayUploadFileMultiThreadingByBDH(path, 2, 4, rsp.UploadKey, EmptyBytes, false); err == nil {
		goto ok
	}
	return nil, errors.New("upload failed")
ok:
	_, _ = img.Seek(0, io.SeekStart)
	i, _, _ := image.DecodeConfig(img)
	var imageType int32 = 1000
	_, _ = img.Seek(0, io.SeekStart)
	tmp := make([]byte, 4)
	_, _ = img.Read(tmp)
	if bytes.Equal(tmp, []byte{0x47, 0x49, 0x46, 0x38}) {
		imageType = 2000
	}
	return message.NewGroupImage(binary.CalculateImageResourceId(fh[:]), fh[:], rsp.FileId, int32(length), int32(i.Width), int32(i.Height), imageType), nil
}

func (c *QQClient) UploadPrivateImage(target int64, img io.ReadSeeker) (*message.FriendImageElement, error) {
	return c.uploadPrivateImage(target, img, 0)
}

func (c *QQClient) uploadPrivateImage(target int64, img io.ReadSeeker, count int) (*message.FriendImageElement, error) {
	_, _ = img.Seek(0, io.SeekStart)
	count++
	fh, length := utils.ComputeMd5AndLength(img)
	_, _ = img.Seek(0, io.SeekStart)
	e, err := c.QueryFriendImage(target, fh[:], int32(length))
	if errors.Is(err, ErrNotExists) {
		// use group highway upload and query again for image id.
		if _, err = c.UploadGroupImage(target, img); err != nil {
			return nil, err
		}
		if count >= 5 {
			return e, nil
		}
		return c.uploadPrivateImage(target, img, count)
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (c *QQClient) ImageOcr(img interface{}) (*OcrResponse, error) {
	switch e := img.(type) {
	case *message.GroupImageElement:
		rsp, err := c.sendAndWait(c.buildImageOcrRequestPacket(e.Url, strings.ToUpper(hex.EncodeToString(e.Md5)), e.Size, e.Width, e.Height))
		if err != nil {
			return nil, err
		}
		return rsp.(*OcrResponse), nil
	case *message.ImageElement:
		rsp, err := c.sendAndWait(c.buildImageOcrRequestPacket(e.Url, strings.ToUpper(hex.EncodeToString(e.Md5)), e.Size, e.Width, e.Height))
		if err != nil {
			return nil, err
		}
		return rsp.(*OcrResponse), nil
	}
	return nil, errors.New("image error")
}

func (c *QQClient) QueryGroupImage(groupCode int64, hash []byte, size int32) (*message.GroupImageElement, error) {
	r, err := c.sendAndWait(c.buildGroupImageStorePacket(groupCode, hash, size))
	if err != nil {
		return nil, err
	}
	rsp := r.(imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		return message.NewGroupImage(binary.CalculateImageResourceId(hash), hash, rsp.FileId, size, rsp.Width, rsp.Height, 1000), nil
	}
	return nil, errors.New("image does not exist")
}

func (c *QQClient) QueryFriendImage(target int64, hash []byte, size int32) (*message.FriendImageElement, error) {
	i, err := c.sendAndWait(c.buildOffPicUpPacket(target, hash, size))
	if err != nil {
		return nil, err
	}
	rsp := i.(imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if !rsp.IsExists {
		return &message.FriendImageElement{
			ImageId: rsp.ResourceId,
			Md5:     hash,
		}, errors.WithStack(ErrNotExists)
	}
	return &message.FriendImageElement{
		ImageId: rsp.ResourceId,
		Md5:     hash,
	}, nil
}

// ImgStore.GroupPicUp
func (c *QQClient) buildGroupImageStorePacket(groupCode int64, md5 []byte, size int32) (uint16, []byte) {
	seq := c.nextSeq()
	name := utils.RandomString(16) + ".gif"
	req := &pb.D388ReqBody{
		NetType: 3,
		Subcmd:  1,
		MsgTryUpImgReq: []*pb.TryUpImgReq{
			{
				GroupCode:    groupCode,
				SrcUin:       c.Uin,
				FileMd5:      md5,
				FileSize:     int64(size),
				FileName:     name,
				SrcTerm:      5,
				PlatformType: 9,
				BuType:       1,
				PicType:      1000,
				BuildVer:     "8.2.7.4410",
				AppPicType:   1006,
				FileIndex:    EmptyBytes,
				TransferUrl:  EmptyBytes,
			},
		},
		Extension: EmptyBytes,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ImgStore.GroupPicUp", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ImgStore.GroupPicUp
func decodeGroupImageStoreResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkt := pb.D388RespBody{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	rsp := pkt.MsgTryUpImgRsp[0]
	if rsp.Result != 0 {
		return imageUploadResponse{
			ResultCode: rsp.Result,
			Message:    rsp.FailMsg,
		}, nil
	}
	if rsp.BoolFileExit {
		if rsp.MsgImgInfo != nil {
			return imageUploadResponse{IsExists: true, FileId: rsp.Fid, Width: rsp.MsgImgInfo.FileWidth, Height: rsp.MsgImgInfo.FileHeight}, nil
		}
		return imageUploadResponse{IsExists: true, FileId: rsp.Fid}, nil
	}
	return imageUploadResponse{
		FileId:     rsp.Fid,
		UploadKey:  rsp.UpUkey,
		UploadIp:   rsp.Uint32UpIp,
		UploadPort: rsp.Uint32UpPort,
	}, nil
}
