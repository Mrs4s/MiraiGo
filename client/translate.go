package client

import (
	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

func (c *QQClient) buildTranslatePacket(src, dst, text string) (uint16, []byte) {
	body := &oidb.TranslateReqBody{
		BatchTranslateReq: &oidb.BatchTranslateReq{
			SrcLanguage: src,
			DstLanguage: dst,
			SrcTextList: []string{text},
		},
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:     2448,
		ServiceType: 2,
		Bodybuffer:  b,
	}
	payload, _ := proto.Marshal(req)
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
func decodeTranslateResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.TranslateRspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp.BatchTranslateRsp, nil
}
