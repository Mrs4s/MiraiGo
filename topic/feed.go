package topic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Mrs4s/MiraiGo/client/pb/channel"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

type (
	Feed struct {
		Id         string
		Title      string
		SubTitle   string
		CreateTime int64
		Poster     *FeedPoster
		GuildId    uint64
		ChannelId  uint64
		Images     []*FeedImageInfo
		Videos     []*FeedVideoInfo
		Contents   []IFeedRichContentElement
	}

	FeedPoster struct {
		TinyId    uint64
		TinyIdStr string
		Nickname  string
		IconUrl   string
	}

	FeedImageInfo struct {
		FileId    string
		PatternId string
		Url       string
		Width     uint32
		Height    uint32
	}

	FeedVideoInfo struct {
		FileId    string
		PatternId string
		Url       string
		Width     uint32
		Height    uint32
		// CoverImage FeedImageInfo
	}

	IFeedRichContentElement interface {
		pack(patternId string, isPatternData bool) content
	}

	content map[string]any
)

var globalBlockId int64

func genBlockId() string {
	id := atomic.AddInt64(&globalBlockId, 1)
	return fmt.Sprintf("%v_%v_%v", time.Now().UnixMilli(), utils.RandomStringRange(4, "0123456789"), id)
}

func (f *Feed) ToSendingPayload(selfUin int64) string {
	c := content{ // todo: support media
		"images": make([]int, 0),
		"videos": make([]int, 0),
		"poster": content{
			"id":   f.Poster.TinyIdStr,
			"nick": f.Poster.Nickname,
		},
		"channelInfo": content{
			"sign": content{
				"guild_id":   strconv.FormatUint(f.GuildId, 10),
				"channel_id": strconv.FormatUint(f.ChannelId, 10),
			},
		},
		"title": content{
			"contents": []content{
				(&TextElement{Content: f.Title}).pack("", false),
			},
		},
	}
	patternInfo := []content{
		{
			"id":   genBlockId(),
			"type": "blockParagraph",
			"data": []content{
				(&TextElement{Content: f.Title}).pack("", true),
			},
		},
	}
	patternData := make([]content, len(f.Contents))
	contents := make([]content, len(f.Contents))
	for i, c := range f.Contents {
		patternId := fmt.Sprintf("o%v_%v_%v", selfUin, time.Now().Format("2006_01_02_15_04_05"), strings.ToLower(utils.RandomStringRange(16, "0123456789abcdef"))) // readCookie("uin")_yyyy_MM_dd_hh_mm_ss_randomHex(16)
		contents[i] = c.pack(patternId, false)
		patternData[i] = c.pack(patternId, true)
	}
	c["contents"] = content{"contents": contents}
	patternInfo = append(patternInfo, content{
		"id":   genBlockId(),
		"type": "blockParagraph",
		"data": patternData,
	})
	packedPattern, _ := json.Marshal(patternInfo)
	c["patternInfo"] = utils.B2S(packedPattern)
	packedContent, _ := json.Marshal(c)
	return utils.B2S(packedContent)
}

func DecodeFeed(p *channel.StFeed) *Feed {
	f := &Feed{
		Id:         p.Id.Unwrap(),
		Title:      p.Title.Contents[0].TextContent.Text.Unwrap(),
		SubTitle:   "",
		CreateTime: int64(p.CreateTime.Unwrap()),
		GuildId:    p.ChannelInfo.Sign.GuildId.Unwrap(),
		ChannelId:  p.ChannelInfo.Sign.ChannelId.Unwrap(),
	}
	if p.Subtitle != nil && len(p.Subtitle.Contents) > 0 {
		f.SubTitle = p.Subtitle.Contents[0].TextContent.Text.Unwrap()
	}
	if p.Poster != nil {
		tinyId, _ := strconv.ParseUint(p.Poster.Id.Unwrap(), 10, 64)
		f.Poster = &FeedPoster{
			TinyId:    tinyId,
			TinyIdStr: p.Poster.Id.Unwrap(),
			Nickname:  p.Poster.Nick.Unwrap(),
		}
		if p.Poster.Icon != nil {
			f.Poster.IconUrl = p.Poster.Icon.IconUrl.Unwrap()
		}
	}
	for _, video := range p.Videos {
		f.Videos = append(f.Videos, &FeedVideoInfo{
			FileId:    video.FileId.Unwrap(),
			PatternId: video.PatternId.Unwrap(),
			Url:       video.PlayUrl.Unwrap(),
			Width:     video.Width.Unwrap(),
			Height:    video.Height.Unwrap(),
		})
	}
	for _, image := range p.Images {
		f.Images = append(f.Images, &FeedImageInfo{
			FileId:    image.PicId.Unwrap(),
			PatternId: image.PatternId.Unwrap(),
			Url:       image.PicUrl.Unwrap(),
			Width:     image.Width.Unwrap(),
			Height:    image.Height.Unwrap(),
		})
	}
	for _, c := range p.Contents.Contents {
		if c.TextContent != nil {
			f.Contents = append(f.Contents, &TextElement{Content: c.TextContent.Text.Unwrap()})
		}
		if c.EmojiContent != nil {
			id, _ := strconv.ParseInt(c.EmojiContent.Id.Unwrap(), 10, 64)
			f.Contents = append(f.Contents, &EmojiElement{
				Index: int32(id),
				Id:    c.EmojiContent.Id.Unwrap(),
				Name:  message.FaceNameById(int(id)),
			})
		}
		if c.ChannelContent != nil && c.ChannelContent.ChannelInfo != nil {
			f.Contents = append(f.Contents, &ChannelQuoteElement{
				GuildId:     c.ChannelContent.ChannelInfo.Sign.GuildId.Unwrap(),
				ChannelId:   c.ChannelContent.ChannelInfo.Sign.ChannelId.Unwrap(),
				DisplayText: c.ChannelContent.ChannelInfo.Name.Unwrap(),
			})
		}
		if c.AtContent != nil && c.AtContent.User != nil {
			tinyId, _ := strconv.ParseUint(c.AtContent.User.Id.Unwrap(), 10, 64)
			f.Contents = append(f.Contents, &AtElement{
				Id:       c.AtContent.User.Id.Unwrap(),
				TinyId:   tinyId,
				Nickname: c.AtContent.User.Nick.Unwrap(),
			})
		}
		if c.UrlContent != nil {
			f.Contents = append(f.Contents, &UrlQuoteElement{
				Url:         c.UrlContent.Url.Unwrap(),
				DisplayText: c.UrlContent.DisplayText.Unwrap(),
			})
		}
	}
	return f
}
