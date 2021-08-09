package client

import (
	"bytes"
	"strconv"
	"sync/atomic"
)

type Statistics struct {
	PacketReceived  uint64
	PacketSent      uint64
	PacketLost      uint64
	MessageReceived uint64
	MessageSent     uint64
	LastMessageTime int64
	DisconnectTimes uint32
	LostTimes       uint32
}

func (s *Statistics) MarshalJSON() ([]byte, error) {
	var w bytes.Buffer
	w.Grow(256)
	w.WriteString(`{"packet_received":`)
	w.WriteString(strconv.FormatUint(atomic.LoadUint64(&s.PacketReceived), 10))
	w.WriteString(`,"packet_sent":`)
	w.WriteString(strconv.FormatUint(atomic.LoadUint64(&s.PacketSent), 10))
	w.WriteString(`,"packet_lost":`)
	w.WriteString(strconv.FormatUint(atomic.LoadUint64(&s.PacketLost), 10))
	w.WriteString(`,"message_received":`)
	w.WriteString(strconv.FormatUint(atomic.LoadUint64(&s.MessageReceived), 10))
	w.WriteString(`,"message_sent":`)
	w.WriteString(strconv.FormatUint(atomic.LoadUint64(&s.MessageSent), 10))
	w.WriteString(`,"disconnect_times":`)
	w.WriteString(strconv.FormatUint(uint64(atomic.LoadUint32(&s.DisconnectTimes)), 10))
	w.WriteString(`,"lost_times":`)
	w.WriteString(strconv.FormatUint(uint64(atomic.LoadUint32(&s.LostTimes)), 10))
	w.WriteString(`,"last_message_time":`)
	w.WriteString(strconv.FormatInt(atomic.LoadInt64(&s.LastMessageTime), 10))
	w.WriteByte('}')
	return w.Bytes(), nil
}

func (c *QQClient) GetStatistics() *Statistics {
	return &c.stat
}
