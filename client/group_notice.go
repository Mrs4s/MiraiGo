package client

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/utils"
)

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
	body, err := ioutil.ReadAll(resp.Body)
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
	err = json.UnmarshalFromString(html.UnescapeString(res.ID), &ret)
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
