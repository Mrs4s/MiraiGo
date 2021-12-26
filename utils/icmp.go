package utils

import (
	"errors"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type ICMPPingResult struct {
	PacketsSent int
	PacketsLoss int
	AvgTimeMill int64
}

// RunICMPPingLoop unix 下的 ping
func RunICMPPingLoop(ip string, count int) (r ICMPPingResult) {
	r = ICMPPingResult{
		PacketsSent: count,
		PacketsLoss: count,
		AvgTimeMill: 9999,
	}
	if count <= 0 {
		return
	}
	durs := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		d, err := pingudp(ip)
		if err == nil {
			r.PacketsLoss--
			durs = append(durs, d)
		}
	}

	if len(durs) > 0 {
		r.AvgTimeMill = 0
		for _, d := range durs {
			r.AvgTimeMill += d
		}
		if len(durs) > 1 {
			r.AvgTimeMill /= int64(len(durs))
		}
	}

	return
}

func pingudp(ip string) (int64, error) {
	var buf [256]byte
	ch := make(chan error, 1)

	port := rand.Intn(10000) + 50000
	conn, err := net.Dial("udp", ip+":"+strconv.Itoa(port))
	if err != nil {
		return 9999, err
	}

	t := time.Now().UnixMilli()

	_, err = conn.Write([]byte("fill"))
	if err != nil {
		return 0, err
	}
	go func() {
		_, err := conn.Read(buf[:])
		ch <- err
	}()
	select {
	case <-time.NewTimer(time.Second * 4).C:
		err = errors.New("timeout")
	case err = <-ch:
	}

	if err != nil && err.Error() == "timeout" {
		return 9999, err
	}
	return time.Now().UnixMilli() - t, nil
}
