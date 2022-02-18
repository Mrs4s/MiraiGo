package network

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
)

type Response struct {
	SequenceID  int32
	CommandName string
	Body        []byte

	// Request is the original request that obtained this response.
	Request *Request
}

func (r *Response) Params() Params {
	if r.Request == nil {
		return nil
	}
	return r.Request.Params
}

var (
	ErrSessionExpired    = errors.New("session expired")
	ErrPacketDropped     = errors.New("packet dropped")
	ErrInvalidPacketType = errors.New("invalid packet type")
	ErrConnectionBroken  = errors.New("connection broken")
)

func (t *Transport) ReadRequest(head []byte) (*Request, error) {
	req := new(Request)
	r := binary.NewReader(head)
	req.Type = RequestType(r.ReadInt32())
	if req.Type != RequestTypeLogin && req.Type != RequestTypeSimple {
		return req, ErrInvalidPacketType
	}
	req.EncryptType = EncryptType(r.ReadByte())
	_ = r.ReadByte() // 0x00?

	req.Uin, _ = strconv.ParseInt(r.ReadString(), 10, 64)
	body := r.ReadAvailable()
	switch req.EncryptType {
	case EncryptTypeNoEncrypt:
		// nothing to do
	case EncryptTypeD2Key:
		body = binary.NewTeaCipher(t.Sig.D2Key).Decrypt(body)
	case EncryptTypeEmptyKey:
		body = binary.NewTeaCipher(emptyKey).Decrypt(body)
	}
	err := t.readSSOFrame(req, body)
	return req, err
}

func (t *Transport) readSSOFrame(req *Request, payload []byte) error {
	reader := binary.NewReader(payload)
	headLen := reader.ReadInt32()
	if headLen-4 > int32(reader.Len()) {
		return errors.WithStack(ErrPacketDropped)
	}

	head := binary.NewReader(reader.ReadBytes(int(headLen) - 4))
	req.SequenceID = head.ReadInt32()
	retCode := head.ReadInt32()
	message := head.ReadString()
	switch retCode {
	case 0:
		// ok
	case -10008:
		return errors.WithMessage(ErrSessionExpired, message)
	default:
		return errors.Errorf("return code unsuccessful: %d message: %s", retCode, message)
	}
	req.CommandName = head.ReadString()
	if req.CommandName == "Heartbeat.Alive" {
		return nil
	}
	_ = head.ReadInt32Bytes() // session id
	compressedFlag := head.ReadInt32()

	bodyLen := reader.ReadInt32() - 4
	body := reader.ReadAvailable()
	if bodyLen > 0 && bodyLen < int32(len(body)) {
		body = body[:bodyLen]
	}
	switch compressedFlag {
	case 0, 8:
	case 1:
		body = binary.ZlibUncompress(body)
	}
	req.Body = body
	return nil
}
