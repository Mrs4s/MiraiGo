package highway

import (
	"encoding/binary"
	"net"
)

var etx = []byte{0x29}

// newFrame 包格式
//
//   - STX: 0x28(40)
//   - head length
//   - body length
//   - head data
//   - body data
//   - ETX: 0x29(41)
//
// 节省内存, 可被go runtime优化为writev操作
func newFrame(head []byte, body []byte) net.Buffers {
	buffers := make(net.Buffers, 4)
	// buffer0 format:
	// 	- STX
	// 	- head length
	// 	- body length
	buffer0 := make([]byte, 9)
	buffer0[0] = 0x28
	binary.BigEndian.PutUint32(buffer0[1:], uint32(len(head)))
	binary.BigEndian.PutUint32(buffer0[5:], uint32(len(body)))
	buffers[0] = buffer0
	buffers[1] = head
	buffers[2] = body
	buffers[3] = etx
	return buffers
}
