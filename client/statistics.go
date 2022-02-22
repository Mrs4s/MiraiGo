package client

import (
	"go.uber.org/atomic"
)

type Statistics struct {
	PacketReceived  atomic.Uint64 `json:"packet_received"`
	PacketSent      atomic.Uint64 `json:"packet_sent"`
	PacketLost      atomic.Uint64 `json:"packet_lost"`
	MessageReceived atomic.Uint64 `json:"message_received"`
	MessageSent     atomic.Uint64 `json:"message_sent"`
	LastMessageTime atomic.Int64  `json:"last_message_time"`
	DisconnectTimes atomic.Uint32 `json:"disconnect_times"`
	LostTimes       atomic.Uint32 `json:"lost_times"`
}

func (c *QQClient) GetStatistics() *Statistics {
	return &c.stat
}
