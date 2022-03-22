package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/Mrs4s/MiraiGo/utils"
)

type (
	ModelVariant struct {
		NeedPay   bool
		ModelShow string
	}

	ModelGet struct {
		Req ModelReq `json:"13030"`
	}

	ModelSet struct {
		Req ModelReq `json:"13031"`
	}

	ModelReq struct {
		Req ModelReqData `json:"req"`
	}

	ModelReqData struct {
		Uin            int64  `json:"lUin"`
		Model          string `json:"sModel"`
		AppType        int32  `json:"iAppType"`
		IMei           string `json:"sIMei"`
		ShowInfo       bool   `json:"bShowInfo"`
		ModelShow      string `json:"sModelShow"`
		RecoverDefault bool   `json:"bRecoverDefault"`
	}
)

func (c *QQClient) getGtk(domain string) int {
	if psKey, ok := c.sig.PsKeyMap[domain]; ok {
		accu := 5381
		for _, b := range psKey {
			accu = accu + (accu << 5) + int(b)
		}
		return 2147483647 & accu
	}
	return 0
}

func (c *QQClient) GetModelShow(modelName string) ([]*ModelVariant, error) {
	req := ModelGet{
		Req: ModelReq{
			Req: ModelReqData{
				Uin:            c.Uin,
				Model:          strings.ReplaceAll(url.QueryEscape(modelName), "+", "%20"),
				AppType:        0,
				IMei:           c.deviceInfo.IMEI,
				ShowInfo:       true,
				ModelShow:      "",
				RecoverDefault: false,
			},
		},
	}

	ts := time.Now().UnixNano() / 1e6
	g_tk := c.getGtk("vip.qq.com")
	data, _ := json.Marshal(req)
	b, err := utils.HttpGetBytes(
		fmt.Sprintf("https://proxy.vip.qq.com/cgi-bin/srfentry.fcgi?ts=%d&daid=18&g_tk=%d&pt4_token=&data=%s", ts, g_tk, url.QueryEscape(string(data))),
		c.getCookiesWithDomain("vip.qq.com"),
	)
	if err != nil {
		return nil, err
	}

	variants := make([]*ModelVariant, 0)
	gjson.ParseBytes(b).Get("13030.data.rsp.vItemList").ForEach(func(_, value gjson.Result) bool {
		variants = append(variants, &ModelVariant{
			ModelShow: value.Get("sModelShow").String(),
			NeedPay:   value.Get("bNeedPay").Bool(),
		})
		return true
	})
	return variants, nil
}

func (c *QQClient) SetModelShow(modelName string, modelShow string) error {
	req := ModelSet{
		Req: ModelReq{
			Req: ModelReqData{
				Uin:            c.Uin,
				Model:          strings.ReplaceAll(url.QueryEscape(modelName), "+", "%20"),
				AppType:        0,
				IMei:           c.deviceInfo.IMEI,
				ShowInfo:       true,
				ModelShow:      strings.ReplaceAll(url.QueryEscape(modelShow), "+", "%20"),
				RecoverDefault: modelShow == "",
			},
		},
	}

	ts := time.Now().UnixNano() / 1e6
	g_tk := c.getGtk("vip.qq.com")
	data, _ := json.Marshal(req)
	_, err := utils.HttpGetBytes(
		fmt.Sprintf("https://proxy.vip.qq.com/cgi-bin/srfentry.fcgi?ts=%d&daid=18&g_tk=%d&pt4_token=&data=%s", ts, g_tk, url.QueryEscape(string(data))),
		c.getCookiesWithDomain("vip.qq.com"),
	)
	if err != nil {
		return err
	}
	return nil
}
