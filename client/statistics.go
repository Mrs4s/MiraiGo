package client

import (
	"bytes"
	"strconv"
	"sync/atomic"
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

// MarshalJSON encodes the wrapped statistics into JSON.
func (m *Statistics) MarshalJSON() ([]byte, error) {
	var w bytes.Buffer
	w.Grow(256)
	w.WriteString(`{"packet_received":`)
	w.WriteString(strconv.FormatUint(m.PacketReceived.Load(), 10))
	w.WriteString(`,"packet_sent":`)
	w.WriteString(strconv.FormatUint(m.PacketSent.Load(), 10))
	w.WriteString(`,"packet_lost":`)
	w.WriteString(strconv.FormatUint(m.PacketLost.Load(), 10))
	w.WriteString(`,"message_received":`)
	w.WriteString(strconv.FormatUint(m.MessageReceived.Load(), 10))
	w.WriteString(`,"message_sent":`)
	w.WriteString(strconv.FormatUint(m.MessageSent.Load(), 10))
	w.WriteString(`,"disconnect_times":`)
	w.WriteString(strconv.FormatUint(uint64(m.DisconnectTimes.Load()), 10))
	w.WriteString(`,"lost_times":`)
	w.WriteString(strconv.FormatUint(uint64(m.LostTimes.Load()), 10))
	w.WriteString(`,"last_message_time":`)
	w.WriteString(strconv.FormatInt(m.LastMessageTime.Load(), 10))
	w.WriteByte('}')
	return w.Bytes(), nil
}
