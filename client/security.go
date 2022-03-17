package client

import (
	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

func init() {
	decoders["OidbSvc.0xbcb_0"] = decodeUrlCheckResponse
}

type UrlSecurityLevel int

const (
	Safe UrlSecurityLevel = iota + 1
	Unknown
	Danger
)

// CheckUrlSafely 通过TX服务器检查URL安全性
func (c *QQClient) CheckUrlSafely(url string) UrlSecurityLevel {
	i, err := c.sendAndWait(c.buildUrlCheckRequest(url))
	if err != nil {
		return Unknown
	}
	return i.(UrlSecurityLevel)
}

func (c *QQClient) buildUrlCheckRequest(url string) (uint16, []byte) {
	payload := c.packOIDBPackageProto(3019, 0, &oidb.DBCBReqBody{
		CheckUrlReq: &oidb.CheckUrlReq{
			Url:         []string{url},
			QqPfTo:      proto.String("mqq.group"),
			Type:        proto.Uint32(2),
			SendUin:     proto.Uint64(uint64(c.Uin)),
			ReqType:     proto.String("webview"),
			OriginalUrl: &url,
			IsArk:       proto.Bool(false),
			IsFinish:    proto.Bool(false),
			SrcUrls:     []string{url},
			SrcPlatform: proto.Uint32(1),
			Qua:         proto.String("AQQ_2013 4.6/2013 8.4.184945&NA_0/000000&ADR&null18&linux&2017&C2293D02BEE31158&7.1.2&V3"),
		},
	})
	return c.uniPacket("OidbSvc.0xbcb_0", payload)
}

func decodeUrlCheckResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := &oidb.DBCBRspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	if rsp.CheckUrlRsp == nil || len(rsp.CheckUrlRsp.Results) == 0 {
		return nil, errors.New("response is empty")
	}
	if rsp.CheckUrlRsp.Results[0].JumpUrl != nil {
		return Danger, nil
	}
	if rsp.CheckUrlRsp.Results[0].GetUmrtype() == 2 {
		return Safe, nil
	}
	return Unknown, nil
}
