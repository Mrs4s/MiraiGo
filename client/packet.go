package client

import (
	"github.com/Mrs4s/MiraiGo/client/internal/codec"
)

//go:noinline
func (c *QQClient) buildOicqRequestPacket(uin int64, command uint16, body []byte) []byte {
	req := codec.OICQ{
		Uin:           uint32(uin),
		Command:       command,
		EncryptMethod: c.ecdh,
		Key:           c.RandomKey,
		Body:          body,
	}
	return req.Encode()
}

//go:noinline
func (c *QQClient) uniPacketWithSeq(command string, body []byte) (uint16, []byte) {
	seq := c.nextSeq()
	req := codec.Uni{
		Uin:         c.Uin,
		Seq:         seq,
		CommandName: command,
		EncryptType: 1,
		SessionID:   c.OutGoingPacketSessionId,
		ExtraData:   EmptyBytes,
		Key:         c.sigInfo.d2Key,
		Body:        body,
	}
	return seq, req.Encode()
}

//go:noinline
func (c *QQClient) uniPacket(seq uint16, command string, body []byte) []byte {
	req := codec.Uni{
		Uin:         c.Uin,
		Seq:         seq,
		CommandName: command,
		EncryptType: 1,
		SessionID:   c.OutGoingPacketSessionId,
		ExtraData:   EmptyBytes,
		Key:         c.sigInfo.d2Key,
		Body:        body,
	}
	return req.Encode()
}
