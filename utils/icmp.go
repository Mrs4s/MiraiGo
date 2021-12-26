package utils

import (
	"net"
	"strings"
	"time"
)

type ICMPPingResult struct {
	PacketsSent int
	PacketsLoss int
	AvgTimeMill int64
}

// RunICMPPingLoop tcp 伪装的 icmp
func RunICMPPingLoop(ipport string, count int) (r ICMPPingResult) {
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
		d, err := pingtcp(ipport)
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

func pingtcp(ipport string) (int64, error) {
	t := time.Now().UnixMilli()
	conn, err := net.DialTimeout("tcp", ipport, time.Second*2)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			return 9999, err
		}
	} else {
		_ = conn.Close()
	}
	return time.Now().UnixMilli() - t, nil
}
