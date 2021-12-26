package utils

import (
	"testing"
)

func TestPing(t *testing.T) {
	r := RunICMPPingLoop("127.0.0.1", 4)
	if r.PacketsLoss == r.PacketsSent {
		t.Fatal(r)
	}
}
