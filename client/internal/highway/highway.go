package highway

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

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

func (a Addr) empty() bool {
	return a.IP == 0 || a.Port == 0
}

type Session struct {
	Uin        string
	AppID      int32
	SigSession []byte
	SessionKey []byte

	seq int32

	addrMu  sync.Mutex
	idx     int
	SsoAddr []Addr

	idleMu    sync.Mutex
	idleCount int
	idle      *idle
}

const highwayMaxResponseSize int32 = 1024 * 100 // 100k

func (s *Session) AddrLength() int {
	s.addrMu.Lock()
	defer s.addrMu.Unlock()
	return len(s.SsoAddr)
}

func (s *Session) AppendAddr(ip, port uint32) {
	s.addrMu.Lock()
	defer s.addrMu.Unlock()
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

func (s *Session) ping(pc *persistConn) error {
	start := time.Now()
	err := s.sendHeartbreak(pc.conn)
	if err != nil {
		return errors.Wrap(err, "echo error")
	}
	if _, err = readResponse(binary.NewNetworkReader(pc.conn)); err != nil {
		return errors.Wrap(err, "echo error")
	}
	// update delay
	pc.ping = time.Since(start).Milliseconds()
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

type persistConn struct {
	conn net.Conn
	addr Addr
	ping int64 // echo ping
}

const maxIdleConn = 7

type idle struct {
	pc   persistConn
	next *idle
}

// getIdleConn ...
func (s *Session) getIdleConn() persistConn {
	s.idleMu.Lock()
	defer s.idleMu.Unlock()

	// no idle
	if s.idle == nil {
		return persistConn{}
	}

	// switch the fastest idle conn
	conn := s.idle.pc
	s.idle = s.idle.next
	s.idleCount--
	if s.idleCount < 0 {
		panic("idle count underflow")
	}

	return conn
}

func (s *Session) putIdleConn(pc persistConn) {
	s.idleMu.Lock()
	defer s.idleMu.Unlock()

	// check persistConn
	if pc.conn == nil || pc.addr.empty() {
		panic("put bad idle conn")
	}

	cur := &idle{pc: pc}
	s.idleCount++
	if s.idle == nil { // quick path
		s.idle = cur
		return
	}

	// insert between pre and succ
	var pre, succ *idle
	succ = s.idle
	for succ != nil && succ.pc.ping < pc.ping { // keep idle list sorted by delay incremental
		pre = succ
		succ = succ.next
	}
	if pre != nil {
		pre.next = cur
	}
	cur.next = succ

	// remove the slowest idle conn if idle count greater than maxIdleConn
	if s.idleCount > maxIdleConn {
		for cur.next != nil {
			pre = cur
			cur = cur.next
		}
		pre.next = nil
		s.idleCount--
	}
}

func (s *Session) connect(addr Addr) (persistConn, error) {
	conn, err := net.DialTimeout("tcp", addr.String(), time.Second*3)
	if err != nil {
		return persistConn{}, err
	}
	_ = conn.(*net.TCPConn).SetKeepAlive(true)

	// close conn
	runtime.SetFinalizer(conn, func(conn net.Conn) {
		_ = conn.Close()
	})

	pc := persistConn{conn: conn, addr: addr}
	if err = s.ping(&pc); err != nil {
		return persistConn{}, err
	}
	return pc, nil
}

func (s *Session) nextAddr() Addr {
	s.addrMu.Lock()
	defer s.addrMu.Unlock()
	addr := s.SsoAddr[s.idx]
	s.idx = (s.idx + 1) % len(s.SsoAddr)
	return addr
}

func (s *Session) selectConn() (pc persistConn, err error) {
	for { // select from idle pc
		pc = s.getIdleConn()
		if pc.conn == nil {
			// no idle connection
			break
		}

		err = s.ping(&pc) // ping
		if err == nil {
			return
		}
	}

	try := 0
	for {
		addr := s.nextAddr()
		pc, err = s.connect(addr)
		if err == nil {
			break
		}
		try++
		if try > 5 {
			break
		}
	}
	return
}
