package client

import (
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/internal/oicq"
)

//go:noinline
func (c *QQClient) buildOicqRequestPacket(uin int64, command uint16, body []byte) []byte {
	req := oicq.Message{
		Uin:              uint32(uin),
		Command:          command,
		EncryptionMethod: oicq.EM_ECDH,
		Body:             body,
	}
	return c.oicq.Marshal(&req)
}

//go:noinline
func (c *QQClient) uniPacket(command string, body []byte) (uint16, []byte) {
	seq := c.nextSeq()
	req := network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
	}
	return seq, c.transport.PackPacket(&req)
}

//go:noinline
func (c *QQClient) uniPacketWithSeq(seq uint16, command string, body []byte) []byte {
	req := network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
	}
	return c.transport.PackPacket(&req)
}
