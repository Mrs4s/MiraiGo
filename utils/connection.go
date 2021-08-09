package utils

import (
	"github.com/pkg/errors"
	"encoding/binary"
	"io"
	"net"
	"sync"
)

type TCPListener struct {
	lock                 sync.RWMutex
	conn                 net.Conn
	plannedDisconnect    func(*TCPListener)
	unexpectedDisconnect func(*TCPListener, error)
}

var ErrConnectionClosed = errors.New("connection closed")

// PlannedDisconnect 预料中的断开连接
// 如调用 Close() Connect()
func (t *TCPListener) PlannedDisconnect(f func(*TCPListener)) {
	t.plannedDisconnect = f
}

// UnexpectedDisconnect 未预料钟的断开连接
func (t *TCPListener) UnexpectedDisconnect(f func(*TCPListener, error)) {
	t.unexpectedDisconnect = f
}

func (t *TCPListener) Connect(addr *net.TCPAddr) error {
	t.Close()
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return errors.Wrap(err, "dial tcp error")
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.conn = conn
	return nil
}

func (t *TCPListener) Write(buf []byte) error {
	if t.conn == nil {
		return ErrConnectionClosed
	}
	t.lock.RLock()
	_, err := t.conn.Write(buf)
	t.lock.RUnlock()
	if err == nil {
		return nil
	}

	t.unexpectedClose(err)
	return ErrConnectionClosed
}

func (t *TCPListener) ReadBytes(len int) ([]byte, error) {
	if t.conn == nil {
		return nil, ErrConnectionClosed
	}

	t.lock.RLock()
	buf := make([]byte, len)
	_, err := io.ReadFull(t.conn, buf)
	t.lock.RUnlock()
	if err == nil {
		return buf, nil
	}

	//time.Sleep(time.Millisecond * 100) // 服务器会发送offline包后立即断开连接, 此时还没解析, 可能还是得加锁
	t.unexpectedClose(err)
	return nil, ErrConnectionClosed
}

func (t *TCPListener) ReadInt32() (int32, error) {
	b, err := t.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}

func (t *TCPListener) Close() {
	if t.conn == nil {
		return
	}
	t.close()
	t.invokePlannedDisconnect()
}

func (t *TCPListener) close() {
	if t.conn != nil {
		t.lock.Lock()
		_ = t.conn.Close()
		t.conn = nil
		t.lock.Unlock()
	}
}

func (t *TCPListener) invokePlannedDisconnect() {
	if t.plannedDisconnect != nil {
		go t.plannedDisconnect(t)
	}
}

func (t *TCPListener) unexpectedClose(err error) {
	if t.conn != nil {
		t.close()
		t.invokeUnexpectedDisconnect(err)
	}
}

func (t *TCPListener) invokeUnexpectedDisconnect(err error) {
	if t.unexpectedDisconnect != nil {
		go t.unexpectedDisconnect(t, err)
	}
}
