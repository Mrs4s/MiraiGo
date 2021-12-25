package client

import (
	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x346"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

func (c *QQClient) buildOfflineFileDownloadRequestPacket(uuid []byte) *network.Request {
	seq := c.nextSeq()
	req := &cmd0x346.C346ReqBody{
		Cmd:        1200,
		Seq:        int32(seq),
		BusinessId: 3,
		ClientType: 104,
		ApplyDownloadReq: &cmd0x346.ApplyDownloadReq{
			Uin:       c.Uin,
			Uuid:      uuid,
			OwnerType: 2,
		},
		ExtensionReq: &cmd0x346.ExtensionReq{
			DownloadUrlType: 1,
		},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacketWithSeq(seq, "OfflineFilleHandleSvr.pb_ftn_CMD_REQ_APPLY_DOWNLOAD-1200", payload, decodeOfflineFileDownloadResponse)
}

func decodeOfflineFileDownloadResponse(c *QQClient, resp *network.Response) (interface{}, error) {
	rsp := cmd0x346.C346RspBody{}
	if err := proto.Unmarshal(resp.Body, &rsp); err != nil {
		c.Error("unmarshal cmd0x346 rsp body error: %v", err)
		return nil, errors.Wrap(err, "unmarshal cmd0x346 rsp body error")
	}
	if rsp.ApplyDownloadRsp == nil {
		c.Error("decode apply download 1200 error: apply rsp is nil.")
		return nil, errors.New("apply rsp is nil")
	}
	if rsp.ApplyDownloadRsp.RetCode != 0 {
		c.Error("decode apply download 1200 error: %v", rsp.ApplyDownloadRsp.RetCode)
		return nil, errors.Errorf("apply download rsp error: %d", rsp.ApplyDownloadRsp.RetCode)
	}
	return "http://" + rsp.ApplyDownloadRsp.DownloadInfo.DownloadDomain + rsp.ApplyDownloadRsp.DownloadInfo.DownloadUrl, nil
}
