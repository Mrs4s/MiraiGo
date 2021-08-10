package utils

import (
	"encoding/binary"
	"io"
	"net"
	"sync"

	"github.com/pkg/errors"
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
	err := t.rLockDo(func() error {
		_, e := t.conn.Write(buf)
		return e
	})
	if err == nil {
		return nil
	}

	t.unexpectedClose(err)
	return ErrConnectionClosed
}

func (t *TCPListener) ReadBytes(len int) ([]byte, error) {
	buf := make([]byte, len)
	err := t.rLockDo(func() error {
		_, e := io.ReadFull(t.conn, buf)
		return e
	})
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
	if !t.connIsNil() {
		t.close()
		t.invokePlannedDisconnect()
	}
}

func (t *TCPListener) unexpectedClose(err error) {
	if !t.connIsNil() {
		t.close()
		t.invokeUnexpectedDisconnect(err)
	}
}

func (t *TCPListener) close() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}
}

func (t *TCPListener) invokePlannedDisconnect() {
	if t.plannedDisconnect != nil {
		go t.plannedDisconnect(t)
	}
}

func (t *TCPListener) invokeUnexpectedDisconnect(err error) {
	if t.unexpectedDisconnect != nil {
		go t.unexpectedDisconnect(t, err)
	}
}

func (t *TCPListener) rLockDo(fn func() error) error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.conn == nil {
		return ErrConnectionClosed
	}
	return fn()
}

func (t *TCPListener) connIsNil() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.conn == nil
}
