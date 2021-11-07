package client

import (
	"github.com/pkg/errors"
	protobuf "github.com/segmentio/encoding/proto"

	"github.com/Mrs4s/MiraiGo/internal/packets"
	"github.com/Mrs4s/MiraiGo/internal/protobuf/data/oidb"
	"github.com/Mrs4s/MiraiGo/internal/protobuf/data/oidb/oidb0x990"
	"github.com/Mrs4s/MiraiGo/utils"
)

func (c *QQClient) buildTranslatePacket(src, dst, text string) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb0x990.ReqBody{
		BatchTranslateReq: &oidb0x990.BatchTranslateReq{
			SrcLanguage:      &src,
			DstLanguage:      &dst,
			SrcBytesTextList: [][]byte{utils.S2B(text)},
		},
	}
	payload := c.packOIDBPackageProto2(2448, 2, body)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x990", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) Translate(src, dst, text string) (string, error) {
	rsp, err := c.sendAndWait(c.buildTranslatePacket(src, dst, text))
	if err != nil {
		return "", err
	}
	if data, ok := rsp.(*oidb0x990.BatchTranslateRsp); ok {
		if data.GetErrorCode() != 0 {
			return "", errors.New(string(data.ErrorMsg))
		}
		return string(data.DstBytesTextList[0]), nil
	}
	return "", errors.New("decode error")
}

// OidbSvc.0x990
func decodeTranslateResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb0x990.RspBody{}
	if err := protobuf.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := protobuf.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp.BatchTranslateRsp, nil
}
