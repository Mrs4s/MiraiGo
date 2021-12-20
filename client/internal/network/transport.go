package network

import (
	"strconv"
	"sync"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/auth"
)

// Transport is a network transport.
type Transport struct {
	sessionMu sync.Mutex
	// todo: combine session fields to a struct
	tgt       []byte
	d2key     []byte
	sessionID []byte
	ksid      []byte

	version *auth.AppVersion
	device  *auth.Device

	// connection
	conn *TCPListener
}

func (t *Transport) packBody(req *Request, w *binary.Writer) {
	w.WriteIntLvPacket(4, func(writer *binary.Writer) {
		if req.Type == RequestTypeLogin {
			writer.WriteUInt32(uint32(req.SequenceID))
			writer.WriteUInt32(t.version.AppId)
			writer.WriteUInt32(t.version.SubAppId)
			writer.Write([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00})
			if len(t.tgt) == 0 || len(t.tgt) == 4 {
				writer.WriteUInt32(0x04)
			} else {
				writer.WriteUInt32(uint32(len(t.tgt) + 4))
				writer.Write(t.tgt)
			}
		}

		writer.WriteString(req.Method)
		writer.WriteUInt32(uint32(len(t.sessionID) + 4))
		w.Write(t.sessionID)
		if req.Type == RequestTypeLogin {
			writer.WriteString(t.device.IMEI)
			writer.WriteUInt32(0x04)
			{
				writer.WriteUInt16(uint16(len(t.ksid)) + 2)
				writer.Write(t.ksid)
			}
		}
		writer.WriteUInt32(0x04)
	})

	w.WriteUInt32(uint32(len(req.Body) + 4))
	w.Write(req.Body)
}

func (t *Transport) Send(req *Request) error {
	// todo: return response
	head := binary.NewWriterF(func(w *binary.Writer) {
		w.WriteUInt32(uint32(req.Type))
		w.WriteUInt32(uint32(req.EncryptType))
		switch req.Type {
		case RequestTypeLogin:
			switch req.EncryptType {
			case EncryptTypeD2Key:
				w.WriteUInt32(uint32(len(t.d2key) + 4))
				w.Write(t.d2key)
			default:
				w.WriteUInt32(4)
			}
		case RequestTypeSimple:
			w.WriteUInt32(uint32(req.SequenceID))
		}
		w.WriteString(strconv.FormatInt(req.Uin, 10))
	})

	w := binary.SelectWriter()
	defer binary.PutWriter(w)
	t.packBody(req, w)
	body := w.Bytes()

	// encrypt body
	switch req.EncryptType {
	case EncryptTypeD2Key:
		body = binary.NewTeaCipher(t.d2key).Encrypt(body)
	case EncryptTypeEmptyKey:
		body = binary.NewTeaCipher(emptyKey).Encrypt(body)
	}

	w2 := binary.SelectWriter()
	defer binary.PutWriter(w2)
	w2.WriteUInt32(uint32(len(head) + len(body) + 4))
	w2.Write(head)
	w2.Write(body)
	err := t.conn.Write(w2.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (t *Transport) parse(head []byte) *Request {
	req := new(Request)
	r := binary.NewReader(head)
	req.Type = RequestType(r.ReadInt32())
	encryptType := EncryptType(r.ReadInt32())
	switch req.Type {
	case RequestTypeLogin:
	// req.Key = r.ReadBytes(int(encryptType))
	case RequestTypeSimple:
		req.SequenceID = r.ReadInt32()
	}
	_ = r.ReadString() // req.Uin
	body := r.ReadAvailable()
	switch encryptType {
	case EncryptTypeNoEncrypt:
		req.Body = body
	case EncryptTypeD2Key:
		req.Body = binary.NewTeaCipher(t.d2key).Decrypt(body)
	case EncryptTypeEmptyKey:
		req.Body = binary.NewTeaCipher(emptyKey).Decrypt(body)
	}
	return req
}
