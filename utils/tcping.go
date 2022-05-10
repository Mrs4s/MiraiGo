package utils

import (
	"net"
	"time"
)

type ICMPPingResult struct {
	PacketsSent int
	PacketsLoss int
	AvgTimeMill int64
}

// RunTCPPingLoop 使用 tcp 进行 ping
func RunTCPPingLoop(ipport string, count int) (r ICMPPingResult) {
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
		d, err := tcping(ipport)
		if err == nil {
			r.PacketsLoss--
			durs = append(durs, d)
		}
		time.Sleep(time.Millisecond * 100)
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

func tcping(ipport string) (int64, error) {
	t := time.Now().UnixMilli()
	conn, err := net.DialTimeout("tcp", ipport, time.Second*10)
	if err != nil {
		return 9999, err
	}
	_ = conn.Close()
	return time.Now().UnixMilli() - t, nil
}
