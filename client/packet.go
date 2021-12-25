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
func (c *QQClient) uniRequest(command string, body []byte) *network.Request {
	seq := c.nextSeq()
	return &network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
	}
}

func (c *QQClient) commandCall(command string, body []byte) (*network.Response, error) {
	seq := c.nextSeq()
	req := network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
	}
	return c.call(&req)
}

func (c *QQClient) commandCallAndDecode(command string, body []byte, decode func(*QQClient, *network.Response) (interface{}, error)) (interface{}, error) {
	seq := c.nextSeq()
	req := network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
	}
	resp, err := c.call(&req)
	if err != nil {
		return nil, err
	}
	return decode(c, resp)
}

//go:noinline
func (c *QQClient) uniPacketWithSeq(seq uint16, command string, body []byte) *network.Request {
	req := network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
	}
	return &req
}
