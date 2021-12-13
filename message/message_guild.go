package message

type (
	GuildChannelMessage struct {
		Id         uint64
		InternalId uint64
		GuildId    uint64
		ChannelId  uint64
		Time       int64
		Sender     *GuildSender
		Elements   []IMessageElement
		Reactions  []*GuildMessageEmojiReaction // only available for pulling messages
	}

	GuildMessageEmojiReaction struct {
		EmojiId   string
		EmojiType uint64
		Face      *FaceElement
		Count     int32
		Clicked   bool
	}

	GuildSender struct {
		TinyId   uint64
		Nickname string
	}
)
