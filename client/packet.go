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

type decoderFunc = func(*QQClient, *network.Response) (interface{}, error)

func bindDecoder(c *QQClient, decoder decoderFunc) func(*network.Response) (interface{}, error) {
	return func(response *network.Response) (interface{}, error) {
		return decoder(c, response)
	}
}

//go:noinline
func (c *QQClient) uniRequest(command string, body []byte, decoder decoderFunc) *network.Request {
	seq := c.nextSeq()
	var decode func(*network.Response) (interface{}, error)
	if decoder != nil {
		decode = bindDecoder(c, decoder)
	}
	return &network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
		Decode:      decode,
	}
}

//go:noinline
func (c *QQClient) uniCall(command string, body []byte) (*network.Response, error) {
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

//go:noinline
func (c *QQClient) uniPacketWithSeq(seq uint16, command string, body []byte, decoder decoderFunc) *network.Request {
	var decode func(*network.Response) (interface{}, error)
	if decoder != nil {
		decode = bindDecoder(c, decoder)
	}
	req := network.Request{
		Type:        network.RequestTypeSimple,
		EncryptType: network.EncryptTypeD2Key,
		Uin:         c.Uin,
		SequenceID:  int32(seq),
		CommandName: command,
		Body:        body,
		Decode:      decode,
	}
	return &req
}
