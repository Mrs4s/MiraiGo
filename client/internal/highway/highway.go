package highway

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

// see com/tencent/mobileqq/highway/utils/BaseConstants.java#L120-L121
const (
	_REQ_CMD_DATA        = "PicUp.DataUp"
	_REQ_CMD_HEART_BREAK = "PicUp.Echo"
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

type Session struct {
	Uin        string
	AppID      int32
	SigSession []byte
	SessionKey []byte
	SsoAddr    []Addr

	seq int32
	/*
		idleMu    sync.Mutex
		idleCount int
		idle      *idle
	*/
}

const highwayMaxResponseSize int32 = 1024 * 100 // 100k

func (s *Session) AddrLength() int {
	return len(s.SsoAddr)
}

func (s *Session) AppendAddr(ip, port uint32) {
	addr := Addr{
		IP:   ip,
		Port: int(port),
	}
	s.SsoAddr = append(s.SsoAddr, addr)
}

func (s *Session) nextSeq() int32 {
	return atomic.AddInt32(&s.seq, 2)
}

func (s *Session) sendHeartbreak(conn net.Conn) error {
	head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
		MsgBasehead: &pb.DataHighwayHead{
			Version:   1,
			Uin:       s.Uin,
			Command:   _REQ_CMD_HEART_BREAK,
			Seq:       s.nextSeq(),
			Appid:     s.AppID,
			Dataflag:  4096,
			CommandId: 0,
			LocaleId:  2052,
		},
	})
	buffers := frame(head, nil)
	_, err := buffers.WriteTo(conn)
	return err
}

func (s *Session) sendEcho(conn net.Conn) error {
	err := s.sendHeartbreak(conn)
	if err != nil {
		return errors.Wrap(err, "echo error")
	}
	if _, err = readResponse(binary.NewNetworkReader(conn)); err != nil {
		return errors.Wrap(err, "echo error")
	}
	return nil
}

func readResponse(r *binary.NetworkReader) (*pb.RspDataHighwayHead, error) {
	_, err := r.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read byte")
	}
	hl, _ := r.ReadInt32()
	a2, _ := r.ReadInt32()
	if hl > highwayMaxResponseSize || a2 > highwayMaxResponseSize {
		return nil, errors.Errorf("highway response invild. head size: %v body size: %v", hl, a2)
	}
	head, _ := r.ReadBytes(int(hl))
	_, _ = r.ReadBytes(int(a2)) // skip payload
	_, _ = r.ReadByte()
	rsp := new(pb.RspDataHighwayHead)
	if err = proto.Unmarshal(head, rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp, nil
}

/*
const maxIdleConn = 5

type idle struct {
	conn  net.Conn
	delay int64
	next  *idle
}

// getConn ...
func (s *Session) getConn() net.Conn {
	s.idleMu.Lock()
	defer s.idleMu.Unlock()

	conn := s.idle.conn
	s.idle = s.idle.next
	return conn
}
*/
