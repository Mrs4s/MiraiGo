package client

import (
	"fmt"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb/richmedia"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func (c *QQClient) GetTts(text string) ([]byte, error) {
	url := "https://textts.qq.com/cgi-bin/tts"
	data := fmt.Sprintf("{\"appid\": \"201908021016\",\"sendUin\": %v,\"text\": \"%v\"}", c.Uin, text)
	rsp, err := utils.HttpPostBytesWithCookie(url, []byte(data), c.getCookies())
	if err != nil {
		return nil, errors.Wrap(err, "failed to post to tts server")
	}
	ttsReader := binary.NewReader(rsp)
	ttsWriter := binary.NewWriter()
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
