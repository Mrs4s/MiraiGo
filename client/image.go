package client

import (
	"bytes"
	"encoding/hex"
	"image"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/highway"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x388"
	highway2 "github.com/Mrs4s/MiraiGo/client/pb/highway"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

func init() {
	decoders["ImgStore.GroupPicUp"] = decodeGroupImageStoreResponse
	decoders["ImgStore.GroupPicDown"] = decodeGroupImageDownloadResponse
	decoders["OidbSvc.0xe07_0"] = decodeImageOcrResponse
}

var imgWaiter = utils.NewUploadWaiter()

type imageUploadResponse struct {
	UploadKey  []byte
	UploadIp   []uint32
	UploadPort []uint32
	ResourceId string
	Message    string
	FileId     int64
	Width      int32
	Height     int32
	ResultCode int32
	IsExists   bool
}

func (c *QQClient) UploadGroupImage(groupCode int64, img io.ReadSeeker) (*message.GroupImageElement, error) {
	_, _ = img.Seek(0, io.SeekStart) // safe
	fh, length := utils.ComputeMd5AndLength(img)
	_, _ = img.Seek(0, io.SeekStart)

	key := string(fh)
	imgWaiter.Wait(key)
	defer imgWaiter.Done(key)

	seq, pkt := c.buildGroupImageStorePacket(groupCode, fh, int32(length))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(*imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		goto ok
	}
	if c.highwaySession.AddrLength() == 0 {
		for i, addr := range rsp.UploadIp {
			c.highwaySession.AppendAddr(addr, rsp.UploadPort[i])
		}
	}
	if _, err = c.highwaySession.UploadBDH(highway.BdhInput{
		CommandID: 2,
		Body:      img,
		Ticket:    rsp.UploadKey,
		Ext:       EmptyBytes,
		Encrypt:   false,
	}); err == nil {
		goto ok
	}
	return nil, errors.Wrap(err, "upload failed")
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
	return message.NewGroupImage(binary.CalculateImageResourceId(fh), fh, rsp.FileId, int32(length), int32(i.Width), int32(i.Height), imageType), nil
}

func (c *QQClient) UploadGroupImageByFile(groupCode int64, path string) (*message.GroupImageElement, error) {
	img, err := os.OpenFile(path, os.O_RDONLY, 0o666)
	if err != nil {
		return nil, err
	}
	defer func() { _ = img.Close() }()
	fh, length := utils.ComputeMd5AndLength(img)

	key := hex.EncodeToString(fh)
	imgWaiter.Wait(key)
	defer imgWaiter.Done(key)

	seq, pkt := c.buildGroupImageStorePacket(groupCode, fh, int32(length))
	r, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return nil, err
	}
	rsp := r.(*imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if rsp.IsExists {
		goto ok
	}
	if c.highwaySession.AddrLength() == 0 {
		for i, addr := range rsp.UploadIp {
			c.highwaySession.AppendAddr(addr, rsp.UploadPort[i])
		}
	}

	if _, err = c.highwaySession.UploadBDHMultiThread(highway.BdhInput{
		CommandID: 2,
		File:      path,
		Ticket:    rsp.UploadKey,
		Ext:       EmptyBytes,
		Encrypt:   false,
	}, 1); err == nil {
		goto ok
	}
	return nil, errors.Wrap(err, "upload failed")
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
	return message.NewGroupImage(binary.CalculateImageResourceId(fh), fh, rsp.FileId, int32(length), int32(i.Width), int32(i.Height), imageType), nil
}

func (c *QQClient) UploadPrivateImage(target int64, img io.ReadSeeker) (*message.FriendImageElement, error) {
	return c.uploadPrivateImage(target, img, 0)
}

func (c *QQClient) GetGroupImageDownloadUrl(fileId, groupCode int64, fileMd5 []byte) (string, error) {
	i, err := c.sendAndWait(c.buildGroupImageDownloadPacket(fileId, groupCode, fileMd5))
	if err != nil {
		return "", err
	}
	return i.(string), nil
}

func (c *QQClient) uploadPrivateImage(target int64, img io.ReadSeeker, count int) (*message.FriendImageElement, error) {
	_, _ = img.Seek(0, io.SeekStart)
	count++
	fh, length := utils.ComputeMd5AndLength(img)
	_, _ = img.Seek(0, io.SeekStart)
	e, err := c.QueryFriendImage(target, fh, int32(length))
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
	url := ""
	switch e := img.(type) {
	case *message.GroupImageElement:
		url = e.Url
		if b, err := utils.HTTPGetReadCloser(e.Url, ""); err == nil {
			if url, err = c.uploadOcrImage(b); err != nil {
				url = e.Url
			}
			_ = b.Close()
		}
		rsp, err := c.sendAndWait(c.buildImageOcrRequestPacket(url, strings.ToUpper(hex.EncodeToString(e.Md5)), e.Size, e.Width, e.Height))
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
	rsp := r.(*imageUploadResponse)
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
	rsp := i.(*imageUploadResponse)
	if rsp.ResultCode != 0 {
		return nil, errors.New(rsp.Message)
	}
	if !rsp.IsExists {
		return &message.FriendImageElement{
			ImageId: rsp.ResourceId,
			Md5:     hash,
			Url:     "https://c2cpicdw.qpic.cn/offpic_new/0/" + rsp.ResourceId + "/0?term=2",
		}, errors.WithStack(ErrNotExists)
	}
	return &message.FriendImageElement{
		ImageId: rsp.ResourceId,
		Md5:     hash,
		Url:     "https://c2cpicdw.qpic.cn/offpic_new/0/" + rsp.ResourceId + "/0?term=2",
	}, nil
}

// ImgStore.GroupPicUp
func (c *QQClient) buildGroupImageStorePacket(groupCode int64, md5 []byte, size int32) (uint16, []byte) {
	name := utils.RandomString(16) + ".gif"
	req := &cmd0x388.D388ReqBody{
		NetType: proto.Uint32(3),
		Subcmd:  proto.Uint32(1),
		TryupImgReq: []*cmd0x388.TryUpImgReq{
			{
				GroupCode:    proto.Uint64(uint64(groupCode)),
				SrcUin:       proto.Uint64(uint64(c.Uin)),
				FileMd5:      md5,
				FileSize:     proto.Uint64(uint64(size)),
				FileName:     utils.S2B(name),
				SrcTerm:      proto.Uint32(5),
				PlatformType: proto.Uint32(9),
				BuType:       proto.Uint32(1),
				PicType:      proto.Uint32(1000),
				BuildVer:     utils.S2B("8.2.7.4410"),
				AppPicType:   proto.Uint32(1006),
				FileIndex:    EmptyBytes,
				TransferUrl:  EmptyBytes,
			},
		},
		Extension: EmptyBytes,
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("ImgStore.GroupPicUp", payload)
}

func (c *QQClient) buildGroupImageDownloadPacket(fileId, groupCode int64, fileMd5 []byte) (uint16, []byte) {
	req := &cmd0x388.D388ReqBody{
		NetType: proto.Uint32(3),
		Subcmd:  proto.Uint32(2),
		GetimgUrlReq: []*cmd0x388.GetImgUrlReq{
			{
				FileId:          proto.Uint64(0), // index
				DstUin:          proto.Uint64(uint64(c.Uin)),
				GroupCode:       proto.Uint64(uint64(groupCode)),
				FileMd5:         fileMd5,
				PicUpTimestamp:  proto.Uint32(uint32(time.Now().Unix())),
				Fileid:          proto.Uint64(uint64(fileId)),
				UrlFlag:         proto.Uint32(8),
				UrlType:         proto.Uint32(3),
				ReqPlatformType: proto.Uint32(9),
				ReqTerm:         proto.Uint32(5),
				InnerIp:         proto.Uint32(0),
			},
		},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("ImgStore.GroupPicDown", payload)
}

func (c *QQClient) uploadOcrImage(img io.Reader) (string, error) {
	r := make([]byte, 16)
	rand.Read(r)
	ext, _ := proto.Marshal(&highway2.CommFileExtReq{
		ActionType: proto.Uint32(0),
		Uuid:       binary.GenUUID(r),
	})

	buf, _ := io.ReadAll(img)
	rsp, err := c.highwaySession.UploadBDH(highway.BdhInput{
		CommandID: 76,
		Body:      bytes.NewReader(buf),
		Ticket:    c.highwaySession.SigSession,
		Ext:       ext,
		Encrypt:   false,
	})
	if err != nil {
		return "", errors.Wrap(err, "upload ocr image error")
	}
	rspExt := highway2.CommFileExtRsp{}
	if err = proto.Unmarshal(rsp, &rspExt); err != nil {
		return "", errors.Wrap(err, "error unmarshal highway resp")
	}
	return string(rspExt.GetDownloadUrl()), nil
}

// OidbSvc.0xe07_0
func (c *QQClient) buildImageOcrRequestPacket(url, md5 string, size, weight, height int32) (uint16, []byte) {
	body := &oidb.DE07ReqBody{
		Version:  1,
		Entrance: 3,
		OcrReqBody: &oidb.OCRReqBody{
			ImageUrl:              url,
			OriginMd5:             md5,
			AfterCompressMd5:      md5,
			AfterCompressFileSize: size,
			AfterCompressWeight:   weight,
			AfterCompressHeight:   height,
			IsCut:                 false,
		},
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(3591, 0, b)
	return c.uniPacket("OidbSvc.0xe07_0", payload)
}

// ImgStore.GroupPicUp
func decodeGroupImageStoreResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
	pkt := cmd0x388.D388RspBody{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	rsp := pkt.TryupImgRsp[0]
	if rsp.GetResult() != 0 {
		return &imageUploadResponse{
			ResultCode: int32(rsp.GetResult()),
			Message:    utils.B2S(rsp.GetFailMsg()),
		}, nil
	}
	if rsp.GetFileExit() {
		if rsp.GetImgInfo() != nil {
			return &imageUploadResponse{IsExists: true, FileId: int64(rsp.GetFileid()), Width: int32(rsp.ImgInfo.GetFileWidth()), Height: int32(rsp.ImgInfo.GetFileHeight())}, nil
		}
		return &imageUploadResponse{IsExists: true, FileId: int64(rsp.GetFileid())}, nil
	}
	return &imageUploadResponse{
		FileId:     int64(rsp.GetFileid()),
		UploadKey:  rsp.UpUkey,
		UploadIp:   rsp.GetUpIp(),
		UploadPort: rsp.GetUpPort(),
	}, nil
}

func decodeGroupImageDownloadResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
	pkt := cmd0x388.D388RspBody{}
	if err := proto.Unmarshal(payload, &pkt); err != nil {
		return nil, errors.Wrap(err, "unmarshal protobuf message error")
	}
	if len(pkt.GetimgUrlRsp) == 0 {
		return nil, errors.New("response not found")
	}
	if len(pkt.GetimgUrlRsp[0].FailMsg) != 0 {
		return nil, errors.New(utils.B2S(pkt.GetimgUrlRsp[0].FailMsg))
	}
	return "https://" + utils.B2S(pkt.GetimgUrlRsp[0].DownDomain) + utils.B2S(pkt.GetimgUrlRsp[0].BigDownPara), nil
}

// OidbSvc.0xe07_0
func decodeImageOcrResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.DE07RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.Wording != "" {
		if strings.Contains(rsp.Wording, "服务忙") {
			return nil, errors.New("未识别到文本")
		}
		return nil, errors.New(rsp.Wording)
	}
	if rsp.RetCode != 0 {
		return nil, errors.Errorf("server error, code: %v msg: %v", rsp.RetCode, rsp.ErrMsg)
	}
	texts := make([]*TextDetection, 0, len(rsp.OcrRspBody.TextDetections))
	for _, text := range rsp.OcrRspBody.TextDetections {
		points := make([]*Coordinate, 0, len(text.Polygon.Coordinates))
		for _, c := range text.Polygon.Coordinates {
			points = append(points, &Coordinate{
				X: c.X,
				Y: c.Y,
			})
		}
		texts = append(texts, &TextDetection{
			Text:        text.DetectedText,
			Confidence:  text.Confidence,
			Coordinates: points,
		})
	}
	return &OcrResponse{
		Texts:    texts,
		Language: rsp.OcrRspBody.Language,
	}, nil
}
