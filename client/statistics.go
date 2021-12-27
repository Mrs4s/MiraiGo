package client

import (
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"go.uber.org/atomic"
)

type Statistics struct {
	PacketReceived  atomic.Uint64
	PacketSent      atomic.Uint64
	PacketLost      atomic.Uint64
	DisconnectTimes atomic.Uint32
	LostTimes       atomic.Uint32
	// Deprecated
	MessageReceived atomic.Uint64
	// Deprecated
	MessageSent atomic.Uint64
	// Deprecated
	LastMessageTime atomic.Int64
}

// GetStatistics
// Deprecated use GetClientStatistics instead
func (c *QQClient) GetStatistics() *Statistics {
	return &c.stat
}

func (c *QQClient) GetClientStatistics() *network.Statistics {
	return c.transport.GetStatistics()
}
