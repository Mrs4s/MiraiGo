package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

// SendGroupSign 发送群聊打卡消息
func (c *QQClient) SendGroupSign(target int64) {
	_, _ = c.sendAndWait(c.buildGroupSignPacket(target, 0))
}

func (c *QQClient) buildGroupSignPacket(groupId int64, scene uint32) (uint16, []byte) {
	body := &oidb.Oidb_0Xeb7_ReqBody{
		SignInStatusReq: &oidb.Oidb_0Xeb7_StSignInStatusReq{
			Uid:           proto.String(string(c.Uin)),
			GroupId:       proto.String(string(groupId)),
			Scene:         proto.Uint32(scene),
			ClientVersion: proto.String("8.5.0.5025"),
		},
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(3767, 0, b)
	return c.uniPacket("OidbSvc.0xeb7", payload)
}
