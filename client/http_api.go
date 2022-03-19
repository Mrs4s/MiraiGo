package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/richmedia"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

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

type groupNoticeRsp struct {
	Feeds []*struct {
		SenderId    uint32 `json:"u"`
		PublishTime uint64 `json:"pubt"`
		Message     struct {
			Text   string        `json:"text"`
			Images []noticeImage `json:"pics"`
		} `json:"msg"`
	} `json:"feeds"`
}

type GroupNoticeMessage struct {
	SenderId    uint32 `json:"sender_id"`
	PublishTime uint64 `json:"publish_time"`
	Message     struct {
		Text   string             `json:"text"`
		Images []GroupNoticeImage `json:"images"`
	} `json:"message"`
}

type GroupNoticeImage struct {
	Height string `json:"height"`
	Width  string `json:"width"`
	ID     string `json:"id"`
}

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

func (c *QQClient) GetGroupNotice(groupCode int64) (l []*GroupNoticeMessage, err error) {
	v := url.Values{}
	v.Set("bkn", strconv.Itoa(c.getCSRFToken()))
	v.Set("qid", strconv.FormatInt(groupCode, 10))
	v.Set("ft", "23")
	v.Set("ni", "1")
	v.Set("n", "1")
	v.Set("i", "1")
	v.Set("log_read", "1")
	v.Set("platform", "1")
	v.Set("s", "-1")
	v.Set("n", "20")

	req, _ := http.NewRequest(http.MethodGet, "https://web.qun.qq.com/cgi-bin/announce/get_t_list?"+v.Encode(), nil)
	req.Header.Set("Cookie", c.getCookies())
	rsp, err := utils.Client.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	r := groupNoticeRsp{}
	err = json.NewDecoder(rsp.Body).Decode(&r)
	if err != nil {
		return
	}

	return c.parseGroupNoticeJson(&r), nil
}

func (c *QQClient) parseGroupNoticeJson(s *groupNoticeRsp) []*GroupNoticeMessage {
	o := make([]*GroupNoticeMessage, 0, len(s.Feeds))
	for _, v := range s.Feeds {

		ims := make([]GroupNoticeImage, 0, len(v.Message.Images))
		for i := 0; i < len(v.Message.Images); i++ {
			ims = append(ims, GroupNoticeImage{
				Height: v.Message.Images[i].Height,
				Width:  v.Message.Images[i].Width,
				ID:     v.Message.Images[i].ID,
			})
		}

		o = append(o, &GroupNoticeMessage{
			SenderId:    v.SenderId,
			PublishTime: v.PublishTime,
			Message: struct {
				Text   string             `json:"text"`
				Images []GroupNoticeImage `json:"images"`
			}{
				Text:   v.Message.Text,
				Images: ims,
			},
		})
	}

	return o
}

func (c *QQClient) uploadGroupNoticePic(img []byte) (*noticeImage, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	_ = w.WriteField("bkn", strconv.Itoa(c.getCSRFToken()))
	_ = w.WriteField("source", "troopNotice")
	_ = w.WriteField("m", "0")
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="pic_up"; filename="temp_uploadFile.png"`)
	h.Set("Content-Type", "image/png")
	fw, _ := w.CreatePart(h)
	_, _ = fw.Write(img)
	_ = w.Close()
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
	var res noticePicUpResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
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
