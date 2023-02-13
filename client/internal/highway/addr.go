package highway

import (
	"fmt"
	"net"

	"github.com/Mrs4s/MiraiGo/binary"
)

type Addr struct {
	IP   uint32
	Port int
}

func (a Addr) AsNetIP() net.IP {
	return net.IPv4(byte(a.IP>>24), byte(a.IP>>16), byte(a.IP>>8), byte(a.IP))
}

func (a Addr) String() string {
	return fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(a.IP), a.Port)
}
