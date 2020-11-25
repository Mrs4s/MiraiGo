package client

import (
	"math/rand"
	"time"

	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type RichClientInfo struct {
	Platform    uint32
	SdkVersion  string
	PackageName string
	Signature   string
}

func (c *QQClient) SendGroupRichMessage(target, appId int64, appType, msgStyle uint32, client RichClientInfo, msg *message.RichMessage) (*message.GroupMessage, error) {
	ch := make(chan *message.GroupMessage)
	eid := utils.RandomString(6)
	c.onGroupMessageReceipt(eid, func(c *QQClient, e *groupMessageReceiptEvent) {
		for _, elem := range e.Msg.Elements {
			if elem.Type() == message.LightApp || elem.Type() == message.Service {
				ch <- e.Msg
			}
		}
	})
	defer c.onGroupMessageReceipt(eid)
	_, _ = c.sendAndWait(c.buildRichMsgSendingPacket(target, appId, appType, msgStyle, 1, client, msg)) // rsp is empty chunk
	select {
	case ret := <-ch:
		return ret, nil
	case <-time.After(time.Second * 5):
		return nil, errors.New("timeout")
	}
}

func (c *QQClient) SendFriendRichMessage(target, appId int64, appType, msgStyle uint32, client RichClientInfo, msg *message.RichMessage) {
	_, _ = c.sendAndWait(c.buildRichMsgSendingPacket(target, appId, appType, msgStyle, 0, client, msg))
}

// OidbSvc.0xb77_9
func (c *QQClient) buildRichMsgSendingPacket(target, appId int64, appType, msgStyle, sendType uint32, client RichClientInfo, msg *message.RichMessage) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.DB77ReqBody{
		AppId:    uint64(appId),
		AppType:  appType,
		MsgStyle: msgStyle,
		ClientInfo: &oidb.DB77ClientInfo{
			Platform:           client.Platform,
			SdkVersion:         client.SdkVersion,
			AndroidPackageName: client.PackageName,
			AndroidSignature:   client.Signature,
		},
		ExtInfo:  &oidb.DB77ExtInfo{MsgSeq: rand.Uint64()},
		SendType: sendType,
		RecvUin:  uint64(target),
		RichMsgBody: &oidb.DB77RichMsgBody{
			Title:      msg.Title,
			Summary:    msg.Summary,
			Brief:      msg.Brief,
			Url:        msg.Url,
			PictureUrl: msg.PictureUrl,
			MusicUrl:   msg.MusicUrl,
		},
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:       2935,
		ServiceType:   9,
		Bodybuffer:    b,
		ClientVersion: "android 8.4.8",
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0xb77_9", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}
