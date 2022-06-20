package client

import (
	"strconv"

	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

// SendGroupSign 发送群聊打卡消息
func (c *QQClient) SendGroupSign(target int64) {
	_, pkt := c.buildGroupSignPacket(target)
	_ = c.sendPacket(pkt)
}

func (c *QQClient) buildGroupSignPacket(groupId int64) (uint16, []byte) {
	body := &oidb.DEB7ReqBody{
		SignInWriteReq: &oidb.StSignInWriteReq{
			Uid:           proto.Some(strconv.Itoa(int(c.Uin))),
			GroupId:       proto.Some(strconv.Itoa(int(groupId))),
			ClientVersion: proto.Some("8.5.0"),
		},
	}
	b, _ := proto.Marshal(body)
	payload := c.packOIDBPackage(3767, 1, b)
	return c.uniPacket("OidbSvc.0xeb7", payload)
}
