package client

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/faceroam"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

type CustomFace struct {
	ResId string
	Url   string
}

func init() {
	decoders["Faceroam.OpReq"] = decodeFaceroamResponse
}

func (c *QQClient) GetCustomFaces() ([]*CustomFace, error) {
	i, err := c.sendAndWait(c.buildFaceroamRequestPacket())
	if err != nil {
		return nil, errors.Wrap(err, "get faces error")
	}
	return i.([]*CustomFace), nil
}

func (c *QQClient) buildFaceroamRequestPacket() (uint16, []byte) {
	payload, _ := proto.Marshal(&faceroam.FaceroamReqBody{
		Comm: &faceroam.PlatInfo{
			Implat: proto.Int64(109),
			Osver:  proto.String(string(c.deviceInfo.Version.Release)),
			Mqqver: &c.version.SortVersionName,
		},
		Uin:         proto.Uint64(uint64(c.Uin)),
		SubCmd:      proto.Uint32(1),
		ReqUserInfo: &faceroam.ReqUserInfo{},
	})
	return c.uniPacket("Faceroam.OpReq", payload)
}

func decodeFaceroamResponse(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := faceroam.FaceroamRspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.RspUserInfo == nil {
		return nil, errors.New("user info not found")
	}
	res := make([]*CustomFace, len(rsp.RspUserInfo.Filename))
	for i := len(rsp.RspUserInfo.Filename) - 1; i >= 0; i-- {
		res[len(rsp.RspUserInfo.Filename)-1-i] = &CustomFace{
			ResId: rsp.RspUserInfo.Filename[i],
			Url:   fmt.Sprintf("https://p.qpic.cn/%s/%d/%s/0", rsp.RspUserInfo.GetBid(), c.Uin, rsp.RspUserInfo.Filename[i]),
		}
	}
	return res, nil
}
