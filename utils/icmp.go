package utils

import (
	"math/rand"
	"net"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPPingResult struct {
	PacketsSent int
	PacketsRecv int
	PacketsLoss int
	Rtts        []int64
}

func RunICMPPingLoop(addr *net.IPAddr, count int) *ICMPPingResult {
	if count <= 0 {
		return nil
	}
	r := &ICMPPingResult{
		PacketsSent: count,
		Rtts:        make([]int64, count),
	}
	for i := 1; i <= count; i++ {
		rtt, err := SendICMPRequest(addr, i)
		if err != nil {
			r.PacketsLoss++
			r.Rtts[i-1] = 9999
			continue
		}
		r.PacketsRecv++
		r.Rtts[i-1] = rtt
		time.Sleep(time.Millisecond * 100)
	}
	return r
}

func SendICMPRequest(addr *net.IPAddr, seq int) (int64, error) {
	data := make([]byte, 32)
	rand.Read(data)
	body := &icmp.Echo{
		ID:   0,
		Seq:  seq,
		Data: data,
	}
	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}
	msgBytes, _ := msg.Marshal(nil)
	conn, err := net.DialIP("ip4:icmp", nil, addr)
	if err != nil {
		return 0, errors.Wrap(err, "dial icmp conn error")
	}
	defer func() { _ = conn.Close() }()
	if _, err = conn.Write(msgBytes); err != nil {
		return 0, errors.Wrap(err, "write icmp packet error")
	}
	start := time.Now()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	buff := make([]byte, 1024)
	_, err = conn.Read(buff)
	if err != nil {
		return 0, errors.Wrap(err, "read icmp conn error")
	}
	duration := time.Since(start).Milliseconds()
	return duration, nil
}
