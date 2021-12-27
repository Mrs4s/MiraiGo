package network

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/pkg/errors"
)

type TCPListener struct {
	//lock                 sync.RWMutex
	conn *net.TCPConn
	//connected            bool

	// PlannedDisconnect 预料中的断开连接
	// 如调用 Close() Connect()
	PlannedDisconnect func(*TCPListener)

	// UnexpectedDisconnect 未预料的断开连接
	UnexpectedDisconnect func(*TCPListener, error)
}

func (t *TCPListener) getConn() *net.TCPConn {
	return (*net.TCPConn)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&t.conn))))
}

func (t *TCPListener) setConn(conn *net.TCPConn) (swapped bool) {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&t.conn)), unsafe.Pointer(nil), unsafe.Pointer(conn))
}

func (t *TCPListener) closeConn() *net.TCPConn {
	return (*net.TCPConn)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&t.conn)), unsafe.Pointer(nil)))
}

func (t *TCPListener) Connected() bool {
	// 等同于 t.getConn() != nil (? copilot写的)
	return t.getConn() != nil
}

func (t *TCPListener) Connect(addr *net.TCPAddr) error {
	t.Close()
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return errors.Wrap(err, "dial tcp error")
	}
	t.setConn(conn)
	//t.lock.Lock()
	//defer t.lock.Unlock()
	//t.conn = conn
	return nil
}

// ConnectFastest 连接到最快的服务器
// TODO 禁用不可用服务器
func (t *TCPListener) ConnectFastest(addr []*net.TCPAddr) (*net.TCPAddr, error) {
	ch := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(addr))
	for _, remote := range addr {
		go func(remote *net.TCPAddr) {
			defer wg.Done()
			conn, err := net.DialTCP("tcp", nil, remote)
			if err != nil {
				return
			}
			//addrs = append(addrs, remote)
			if !t.setConn(conn) {
				_ = conn.Close()
				return
			}
			ch <- nil
		}(remote)
	}
	go func() {
		wg.Wait()
		if t.getConn() == nil {
			ch <- errors.New("All addr are unreachable")
		}
	}()
	err := <-ch
	if err != nil {
		return nil, err
	}
	conn := t.getConn()
	return conn.RemoteAddr().(*net.TCPAddr), nil
}

func (t *TCPListener) Write(buf []byte) error {
	if conn := t.getConn(); conn != nil {
		_, err := conn.Write(buf)
		if err != nil {
			t.unexpectedClose(err)
			return ErrConnectionBroken
		}
		return nil
	}

	return ErrConnectionBroken
}

func (t *TCPListener) ReadBytes(len int) ([]byte, error) {
	buf := make([]byte, len)
	if conn := t.getConn(); conn != nil {
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			// time.Sleep(time.Millisecond * 100) // 服务器会发送offline包后立即断开连接, 此时还没解析, 可能还是得加锁
			t.unexpectedClose(err)
			return nil, ErrConnectionBroken
		}
		return buf, nil
	}

	return nil, ErrConnectionBroken
}

func (t *TCPListener) ReadInt32() (int32, error) {
	b, err := t.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

func (t *TCPListener) Close() {
	t.close()
	t.invokePlannedDisconnect()
}

func (t *TCPListener) unexpectedClose(err error) {
	t.close()
	t.invokeUnexpectedDisconnect(err)
}

func (t *TCPListener) close() {
	if conn := t.closeConn(); conn != nil {
		_ = conn.Close()
	}
}

func (t *TCPListener) invokePlannedDisconnect() {
	//if t.Connected() {
	t.PlannedDisconnect(t)
	//}
	//t.lock.RLock()
	//defer t.lock.RUnlock()
	//if t.plannedDisconnect != nil && t.connected {
	//	go t.plannedDisconnect(t)
	//	t.connected = false
	//}
}

func (t *TCPListener) invokeUnexpectedDisconnect(err error) {
	//if t.Connected() {
	t.UnexpectedDisconnect(t, err)
	//}
	//t.lock.RLock()
	//defer t.lock.RUnlock()
	//if t.unexpectedDisconnect != nil && t.connected {
	//	go t.unexpectedDisconnect(t, err)
	//	t.connected = false
	//}
}

//func (t *TCPListener) getConn() net.Conn {
//	t.lock.RLock()
//	defer t.lock.RUnlock()
//	return t.conn
//}
