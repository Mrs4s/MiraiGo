package client

import (
	"go.uber.org/atomic"
)

type Statistics struct {
	PacketReceived  atomic.Uint64
	PacketSent      atomic.Uint64
	PacketLost      atomic.Uint64
	MessageReceived atomic.Uint64
	MessageSent     atomic.Uint64
	LastMessageTime atomic.Int64
	DisconnectTimes atomic.Uint32
	LostTimes       atomic.Uint32
}

func (c *QQClient) GetStatistics() *Statistics {
	return &c.stat
}
