package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/pb/web"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

type UnidirectionalFriendInfo struct {
	Uin      int64
	Nickname string
	Age      int32
	Source   string
}

func (c *QQClient) GetUnidirectionalFriendList() (ret []*UnidirectionalFriendInfo, err error) {
	webRsp := &struct {
		BlockList []struct {
			Uin         int64  `json:"uint64_uin"`
			NickBytes   string `json:"bytes_nick"`
			Age         int32  `json:"uint32_age"`
			Sex         int32  `json:"uint32_sex"`
			SourceBytes string `json:"bytes_source"`
		} `json:"rpt_block_list"`
		ErrorCode int32 `json:"ErrorCode"`
	}{}
	rsp, err := c.webSsoRequest("ti.qq.com", "OidbSvc.0xe17_0", fmt.Sprintf(`{"uint64_uin":%v,"uint64_top":0,"uint32_req_num":99,"bytes_cookies":""}`, c.Uin))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(utils.S2B(rsp), webRsp); err != nil {
		return nil, errors.Wrap(err, "unmarshal json error")
	}
	if webRsp.ErrorCode != 0 {
		return nil, errors.Errorf("web sso request error: %v", webRsp.ErrorCode)
	}
	for _, block := range webRsp.BlockList {
		decodeBase64String := func(str string) string {
			b, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				return ""
			}
			return utils.B2S(b)
		}
		ret = append(ret, &UnidirectionalFriendInfo{
			Uin:      block.Uin,
			Nickname: decodeBase64String(block.NickBytes),
			Age:      block.Age,
			Source:   decodeBase64String(block.SourceBytes),
		})
	}
	return
}

func (c *QQClient) DeleteUnidirectionalFriend(uin int64) error {
	webRsp := &struct {
		ErrorCode int32 `json:"ErrorCode"`
	}{}
	rsp, err := c.webSsoRequest("ti.qq.com", "OidbSvc.0x5d4_0", fmt.Sprintf(`{"uin_list":[%v]}`, uin))
	if err != nil {
		return err
	}
	if err = json.Unmarshal(utils.S2B(rsp), webRsp); err != nil {
		return errors.Wrap(err, "unmarshal json error")
	}
	if webRsp.ErrorCode != 0 {
		return errors.Errorf("web sso request error: %v", webRsp.ErrorCode)
	}
	return nil
}

func (c *QQClient) webSsoRequest(host, webCmd, data string) (string, error) {
	s := strings.Split(host, `.`)
	sub := ""
	for i := len(s) - 1; i >= 0; i-- {
		sub += s[i]
		if i != 0 {
			sub += "_"
		}
	}
	cmd := "MQUpdateSvc_" + sub + ".web." + webCmd
	req, _ := proto.Marshal(&web.WebSsoRequestBody{
		Type: proto.Uint32(0),
		Data: &data,
	})
	rspData, err := c.sendAndWaitDynamic(c.uniPacket(cmd, req))
	if err != nil {
		return "", errors.Wrap(err, "send web sso request error")
	}
	rsp := &web.WebSsoResponseBody{}
	if err = proto.Unmarshal(rspData, rsp); err != nil {
		return "", errors.Wrap(err, "unmarshal response error")
	}
	return rsp.GetData(), nil
}
