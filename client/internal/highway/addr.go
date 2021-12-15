package highway

import (
	binary2 "encoding/binary"
	"fmt"
	"net"

	"github.com/Mrs4s/MiraiGo/binary"
)

type Addr struct {
	IP   uint32
	Port int
}

func (a Addr) asTcpAddr() *net.TCPAddr {
	addr := &net.TCPAddr{
		IP:   make([]byte, 4),
		Port: a.Port,
	}
	binary2.LittleEndian.PutUint32(addr.IP, a.IP)
	return addr
}

func (a Addr) AsNetIP() net.IP {
	return net.IPv4(byte(a.IP>>24), byte(a.IP>>16), byte(a.IP>>8), byte(a.IP))
}

func (a Addr) String() string {
	return fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(a.IP), a.Port)
}
