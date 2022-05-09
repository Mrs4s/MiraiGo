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
	Sig       *auth.SigInfo
	Version   *auth.AppVersion
	Device    *auth.Device

	// connection
	// conn *TCPClient
}

func (t *Transport) packBody(req *Request, w *binary.Writer) {
	pos := w.FillUInt32()
	if req.Type == RequestTypeLogin {
		w.WriteUInt32(uint32(req.SequenceID))
		w.WriteUInt32(t.Version.AppId)
		w.WriteUInt32(t.Version.SubAppId)
		w.Write([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00})
		tgt := t.Sig.TGT
		if len(tgt) == 0 || len(tgt) == 4 {
			w.WriteUInt32(0x04)
		} else {
			w.WriteUInt32(uint32(len(tgt) + 4))
			w.Write(tgt)
		}
	}
	w.WriteString(req.CommandName)
	w.WriteUInt32(uint32(len(t.Sig.OutPacketSessionID) + 4))
	w.Write(t.Sig.OutPacketSessionID)
	if req.Type == RequestTypeLogin {
		w.WriteString(t.Device.IMEI)
		w.WriteUInt32(0x04)

		w.WriteUInt16(uint16(len(t.Sig.Ksid)) + 2)
		w.Write(t.Sig.Ksid)
	}
	w.WriteUInt32(0x04)
	w.WriteUInt32At(pos, uint32(w.Len()-pos))

	w.WriteUInt32(uint32(len(req.Body) + 4))
	w.Write(req.Body)
}

// PackPacket packs a packet.
func (t *Transport) PackPacket(req *Request) []byte {
	// todo(wdvxdr): combine pack packet, send packet and return the response
	if len(t.Sig.D2) == 0 {
		req.EncryptType = EncryptTypeEmptyKey
	}

	w := binary.SelectWriter()
	defer binary.PutWriter(w)

	pos := w.FillUInt32()
	// vvv w.Write(head) vvv
	w.WriteUInt32(uint32(req.Type))
	w.WriteByte(byte(req.EncryptType))
	switch req.Type {
	case RequestTypeLogin:
		switch req.EncryptType {
		case EncryptTypeD2Key:
			w.WriteUInt32(uint32(len(t.Sig.D2) + 4))
			w.Write(t.Sig.D2)
		default:
			w.WriteUInt32(4)
		}
	case RequestTypeSimple:
		w.WriteUInt32(uint32(req.SequenceID))
	}
	w.WriteByte(0x00)
	w.WriteString(strconv.FormatInt(req.Uin, 10))
	// ^^^ w.Write(head) ^^^

	w2 := binary.SelectWriter()
	t.packBody(req, w2)
	body := w2.Bytes()
	// encrypt body
	switch req.EncryptType {
	case EncryptTypeD2Key:
		body = binary.NewTeaCipher(t.Sig.D2Key).Encrypt(body)
	case EncryptTypeEmptyKey:
		emptyKey := make([]byte, 16)
		body = binary.NewTeaCipher(emptyKey).Encrypt(body)
	}
	w.Write(body)
	binary.PutWriter(w2)

	w.WriteUInt32At(pos, uint32(w.Len()))
	return append([]byte(nil), w.Bytes()...)
}
