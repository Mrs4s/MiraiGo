package message

type SourceType byte

// MessageSourceType 常量
const (
	SourcePrivate SourceType = 1 << iota
	SourceGroup
	SourceGuildChannel
	SourceGuildDirect
)

// Source 消息来源
type Source struct {
	SourceType  SourceType
	PrimaryID   int64 // 群号/QQ号/guild_id
	SecondaryID int64 // channel_id
}
