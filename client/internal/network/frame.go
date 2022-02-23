package network

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
)

var etx = []byte{0x29}

type Buffers struct {
	net.Buffers
}

var pool = sync.Pool{
	New: func() interface{} {
		lenHead := make([]byte, 9)
		lenHead[0] = 0x28
		return &Buffers{net.Buffers{lenHead, nil, nil, etx}}
	},
}

func (b *Buffers) WriteTo(w io.Writer) (n int64, err error) {
	defer pool.Put(b) // implement auto put to pool
	return b.Buffers.WriteTo(w)
}

// HeadBodyFrame 包格式
// 	* STX
// 	* head length
// 	* body length
// 	* head data
// 	* body data
// 	* ETX
// 节省内存, 可被go runtime优化为writev操作
func HeadBodyFrame(head []byte, body []byte) *Buffers {
	b := pool.Get().(*Buffers)
	if len(b.Buffers) == 0 {
		lenHead := make([]byte, 9)
		lenHead[0] = 0x28
		b.Buffers = net.Buffers{lenHead, nil, nil, etx}
	}
	b.Buffers[2] = body
	b.Buffers[1] = head
	_ = b.Buffers[0][8]
	binary.BigEndian.PutUint32(b.Buffers[0][1:], uint32(len(head)))
	binary.BigEndian.PutUint32(b.Buffers[0][5:], uint32(len(body)))
	return b
}
