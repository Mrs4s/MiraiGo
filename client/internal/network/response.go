package network

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
)

type Response struct {
	Type        RequestType
	EncryptType EncryptType
	SequenceID  int32
	Uin         int64
	CommandName string
	Body        []byte

	Message string

	// Request is the original request that obtained this response.
	// Request *Request
}

var (
	ErrSessionExpired    = errors.New("session expired")
	ErrPacketDropped     = errors.New("packet dropped")
	ErrInvalidPacketType = errors.New("invalid packet type")
)

func (t *Transport) ReadResponse(head []byte) (*Response, error) {
	resp := new(Response)
	r := binary.NewReader(head)
	resp.Type = RequestType(r.ReadInt32())
	if resp.Type != RequestTypeLogin && resp.Type != RequestTypeSimple {
		return resp, ErrInvalidPacketType
	}
	resp.EncryptType = EncryptType(r.ReadByte())
	_ = r.ReadByte() // 0x00?

	resp.Uin, _ = strconv.ParseInt(r.ReadString(), 10, 64)
	body := r.ReadAvailable()
	switch resp.EncryptType {
	case EncryptTypeNoEncrypt:
		// nothing to do
	case EncryptTypeD2Key:
		body = binary.NewTeaCipher(t.Sig.D2Key).Decrypt(body)
	case EncryptTypeEmptyKey:
		emptyKey := make([]byte, 16)
		body = binary.NewTeaCipher(emptyKey).Decrypt(body)
	}
	err := t.readSSOFrame(resp, body)
	return resp, err
}

func (t *Transport) readSSOFrame(resp *Response, payload []byte) error {
	reader := binary.NewReader(payload)
	headLen := reader.ReadInt32()
	if headLen < 4 || headLen-4 > int32(reader.Len()) {
		return errors.WithStack(ErrPacketDropped)
	}

	head := binary.NewReader(reader.ReadBytes(int(headLen) - 4))
	resp.SequenceID = head.ReadInt32()
	switch retCode := head.ReadInt32(); retCode {
	case 0:
		// ok
	case -10008:
		return errors.WithStack(ErrSessionExpired)
	default:
		return errors.Errorf("return code unsuccessful: %d", retCode)
	}
	resp.Message = head.ReadString()
	resp.CommandName = head.ReadString()
	if resp.CommandName == "Heartbeat.Alive" {
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
	resp.Body = body
	return nil
}
