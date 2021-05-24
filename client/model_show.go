package client

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

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
	if psKey, ok := c.sigInfo.psKeyMap[domain]; ok {
		accu := 5381
		for _, b := range psKey {
			accu = accu + (accu << 5) + int(b)
		}
		return 2147483647 & accu
	} else {
		return 0
	}
}

func (c *QQClient) GetModelShow(modelName string) ([]*ModelVariant, error) {
	req := ModelGet{
		Req: ModelReq{
			Req: ModelReqData{
				Uin:            c.Uin,
				Model:          strings.ReplaceAll(url.QueryEscape(modelName), "+", "%20"),
				AppType:        0,
				IMei:           SystemDeviceInfo.IMEI,
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

	items := jsoniter.Get(b, "13030", "data", "rsp", "vItemList")
	size := items.Size()
	variants := make([]*ModelVariant, size)
	for i := 0; i < size; i++ {
		item := items.Get(i)
		variants[i] = &ModelVariant{
			ModelShow: item.Get("sModelShow").ToString(),
			NeedPay:   item.Get("bNeedPay").ToBool(),
		}
	}
	return variants, nil
}

func (c *QQClient) SetModelShow(modelName string, modelShow string) error {
	req := ModelSet{
		Req: ModelReq{
			Req: ModelReqData{
				Uin:            c.Uin,
				Model:          strings.ReplaceAll(url.QueryEscape(modelName), "+", "%20"),
				AppType:        0,
				IMei:           SystemDeviceInfo.IMEI,
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
