package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/richmedia"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

/* -------- VipInfo -------- */

type VipInfo struct {
	Uin            int64
	Name           string
	Level          int
	LevelSpeed     float64
	VipLevel       string
	VipGrowthSpeed int
	VipGrowthTotal int
}

func (c *QQClient) GetVipInfo(target int64) (*VipInfo, error) {
	b, err := utils.HttpGetBytes(fmt.Sprintf("https://h5.vip.qq.com/p/mc/cardv2/other?platform=1&qq=%d&adtag=geren&aid=mvip.pingtai.mobileqq.androidziliaoka.fromqita", target), c.getCookiesWithDomain("h5.vip.qq.com"))
	if err != nil {
		return nil, err
	}
	ret := VipInfo{Uin: target}
	b = b[bytes.Index(b, []byte(`<span class="ui-nowrap">`))+24:]
	t := b[:bytes.Index(b, []byte(`</span>`))]
	ret.Name = string(t)
	b = b[bytes.Index(b, []byte(`<small>LV</small>`))+17:]
	t = b[:bytes.Index(b, []byte(`</p>`))]
	ret.Level, _ = strconv.Atoi(string(t))
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	t = b[:bytes.Index(b, []byte(`<small>倍`))]
	ret.LevelSpeed, _ = strconv.ParseFloat(string(t), 64)
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	st := string(b[:bytes.Index(b, []byte(`</p>`))])
	st = strings.Replace(st, "<small>", "", 1)
	st = strings.Replace(st, "</small>", "", 1)
	ret.VipLevel = st
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	t = b[:bytes.Index(b, []byte(`</p>`))]
	ret.VipGrowthSpeed, _ = strconv.Atoi(string(t))
	b = b[bytes.Index(b, []byte(`<div class="pk-line pk-line-guest">`))+35:]
	b = b[bytes.Index(b, []byte(`<p>`))+3:]
	t = b[:bytes.Index(b, []byte(`</p>`))]
	ret.VipGrowthTotal, _ = strconv.Atoi(string(t))
	return &ret, nil
}

/* -------- GroupHonorInfo -------- */

type (
	HonorType int

	GroupHonorInfo struct {
		GroupCode        string            `json:"gc"`
		Uin              string            `json:"uin"`
		Type             HonorType         `json:"type"`
		TalkativeList    []HonorMemberInfo `json:"talkativeList"`
		CurrentTalkative CurrentTalkative  `json:"currentTalkative"`
		ActorList        []HonorMemberInfo `json:"actorList"`
		LegendList       []HonorMemberInfo `json:"legendList"`
		StrongNewbieList []HonorMemberInfo `json:"strongnewbieList"`
		EmotionList      []HonorMemberInfo `json:"emotionList"`
	}

	HonorMemberInfo struct {
		Uin    int64  `json:"uin"`
		Avatar string `json:"avatar"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
	}

	CurrentTalkative struct {
		Uin      int64  `json:"uin"`
		DayCount int32  `json:"day_count"`
		Avatar   string `json:"avatar"`
		Name     string `json:"nick"`
	}
)

const (
	Talkative    HonorType = 1 // 龙王
	Performer    HonorType = 2 // 群聊之火
	Legend       HonorType = 3 // 群聊炙焰
	StrongNewbie HonorType = 5 // 冒尖小春笋
	Emotion      HonorType = 6 // 快乐源泉
)

func (c *QQClient) GetGroupHonorInfo(groupCode int64, honorType HonorType) (*GroupHonorInfo, error) {
	b, err := utils.HttpGetBytes(fmt.Sprintf("https://qun.qq.com/interactive/honorlist?gc=%d&type=%d", groupCode, honorType), c.getCookiesWithDomain("qun.qq.com"))
	if err != nil {
		return nil, err
	}
	b = b[bytes.Index(b, []byte(`window.__INITIAL_STATE__=`))+25:]
	b = b[:bytes.Index(b, []byte("</script>"))]
	ret := GroupHonorInfo{}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

/* -------- TextToSpeech -------- */

func (c *QQClient) GetTts(text string) ([]byte, error) {
	apiUrl := "https://textts.qq.com/cgi-bin/tts"
	data := fmt.Sprintf(`{"appid": "201908021016","sendUin": %v,"text": %q}`, c.Uin, text)
	rsp, err := utils.HttpPostBytesWithCookie(apiUrl, []byte(data), c.getCookies())
	if err != nil {
		return nil, errors.Wrap(err, "failed to post to tts server")
	}
	ttsReader := binary.NewReader(rsp)
	ttsWriter := binary.SelectWriter()
	for {
		// 数据格式 69e(字符串)  十六进制   数据长度  0 为结尾
		// 0D 0A (分隔符) payload  0D 0A
		var dataLen []byte
		for b := ttsReader.ReadByte(); b != byte(0x0d); b = ttsReader.ReadByte() {
			dataLen = append(dataLen, b)
		}
		ttsReader.ReadByte()
		var length int
		_, _ = fmt.Sscan("0x"+string(dataLen), &length)
		if length == 0 {
			break
		}
		ttsRsp := &richmedia.TtsRspBody{}
		err := proto.Unmarshal(ttsReader.ReadBytes(length), ttsRsp)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
		}
		if ttsRsp.RetCode != 0 {
			return nil, errors.New("can't convert text to voice")
		}
		for _, voiceItem := range ttsRsp.VoiceData {
			ttsWriter.Write(voiceItem.Voice)
		}
		ttsReader.ReadBytes(2)
	}
	ret := ttsWriter.Bytes()
	ret[0] = '\x02'
	return ret, nil
}

/* -------- GroupNotice -------- */

type noticePicUpResponse struct {
	ErrorCode    int    `json:"ec"`
	ErrorMessage string `json:"em"`
	ID           string `json:"id"`
}

type noticeImage struct {
	Height string `json:"h"`
	Width  string `json:"w"`
	ID     string `json:"id"`
}

func (c *QQClient) uploadGroupNoticePic(img []byte) (*noticeImage, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	err := w.WriteField("bkn", strconv.Itoa(c.getCSRFToken()))
	if err != nil {
		return nil, errors.Wrap(err, "write multipart<bkn> failed")
	}
	err = w.WriteField("source", "troopNotice")
	if err != nil {
		return nil, errors.Wrap(err, "write multipart<source> failed")
	}
	err = w.WriteField("m", "0")
	if err != nil {
		return nil, errors.Wrap(err, "write multipart<m> failed")
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="pic_up"; filename="temp_uploadFile.png"`)
	h.Set("Content-Type", "image/png")
	fw, err := w.CreatePart(h)
	if err != nil {
		return nil, errors.Wrap(err, "create multipart field<pic_up> failed")
	}
	_, err = fw.Write(img)
	if err != nil {
		return nil, errors.Wrap(err, "write multipart<pic_up> failed")
	}
	err = w.Close()
	if err != nil {
		return nil, errors.Wrap(err, "close multipart failed")
	}
	req, err := http.NewRequest("POST", "https://web.qun.qq.com/cgi-bin/announce/upload_img", buf)
	if err != nil {
		return nil, errors.Wrap(err, "new request error")
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Cookie", c.getCookies())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "post error")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body error")
	}
	res := noticePicUpResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json")
	}
	if res.ErrorCode != 0 {
		return nil, errors.New(res.ErrorMessage)
	}
	ret := &noticeImage{}
	err = json.Unmarshal([]byte(html.UnescapeString(res.ID)), &ret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json")
	}
	return ret, nil
}

// AddGroupNoticeSimple 发群公告
func (c *QQClient) AddGroupNoticeSimple(groupCode int64, text string) error {
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}`, groupCode, c.getCSRFToken(), url.QueryEscape(text))
	_, err := utils.HttpPostBytesWithCookie("https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+fmt.Sprint(c.getCSRFToken()), []byte(body), c.getCookiesWithDomain("qun.qq.com"))
	if err != nil {
		return errors.Wrap(err, "request error")
	}
	return nil
}

// AddGroupNoticeWithPic 发群公告带图片
func (c *QQClient) AddGroupNoticeWithPic(groupCode int64, text string, pic []byte) error {
	img, err := c.uploadGroupNoticePic(pic)
	if err != nil {
		return err
	}
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}&pic=%v&imgWidth=%v&imgHeight=%v`, groupCode, c.getCSRFToken(), url.QueryEscape(text), img.ID, img.Width, img.Height)
	_, err = utils.HttpPostBytesWithCookie("https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+fmt.Sprint(c.getCSRFToken()), []byte(body), c.getCookiesWithDomain("qun.qq.com"))
	if err != nil {
		return errors.Wrap(err, "request error")
	}
	return nil
}
