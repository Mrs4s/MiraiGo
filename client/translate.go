package client

import (
	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
)

func (c *QQClient) buildTranslatePacket(src, dst, text string) (uint16, []byte) {
	body := &oidb.TranslateReqBody{
		BatchTranslateReq: &oidb.BatchTranslateReq{
			SrcLanguage: src,
			DstLanguage: dst,
			SrcTextList: []string{text},
		},
	}
	payload := c.packOIDBPackageProto(2448, 2, body)
	return c.uniPacket("OidbSvc.0x990", payload)
}

func (c *QQClient) Translate(src, dst, text string) (string, error) {
	rsp, err := c.sendAndWait(c.buildTranslatePacket(src, dst, text))
	if err != nil {
		return "", err
	}
	if data, ok := rsp.(*oidb.BatchTranslateRsp); ok {
		if data.ErrorCode != 0 {
			return "", errors.New(string(data.ErrorMsg))
		}
		return data.DstTextList[0], nil
	}
	return "", errors.New("decode error")
}

// OidbSvc.0x990
func decodeTranslateResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.TranslateRspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	return rsp.BatchTranslateRsp, nil
}
