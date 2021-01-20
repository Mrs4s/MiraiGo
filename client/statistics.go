package client

import "sync"

type Statistics struct {
	PacketReceived  uint64 `json:"packet_received"`
	PacketSent      uint64 `json:"packet_sent"`
	PacketLost      uint32 `json:"packet_lost"`
	MessageReceived uint64 `json:"message_received"`
	MessageSent     uint64 `json:"message_sent"`
	DisconnectTimes uint32 `json:"disconnect_times"`
	LostTimes       uint32 `json:"lost_times"`
	LastMessageTime int64  `json:"last_message_time"`

	once sync.Once
}

func (c *QQClient) GetStatistics() *Statistics {
	return c.stat
}
