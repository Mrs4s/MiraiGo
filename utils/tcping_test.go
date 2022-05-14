package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	r := RunTCPPingLoop("127.0.0.1:23333", 4)
	assert.Equal(t, 4, r.PacketsLoss)
	assert.Equal(t, 4, r.PacketsSent)
	r = RunTCPPingLoop("114.114.114.114:53", 4)
	assert.Equal(t, 0, r.PacketsLoss)
	assert.Equal(t, 4, r.PacketsSent)
	t.Log(r.AvgTimeMill)
}
