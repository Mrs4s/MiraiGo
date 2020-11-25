package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func (c *QQClient) buildTranslatePacket(src, dst, text string) (uint16, []byte) {
	seq := c.nextSeq()
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
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x990", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
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
func decodeTranslateResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
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
