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
	// conn *TCPListener
}

func (t *Transport) packBody(req *Request, w *binary.Writer) {
	w.WriteIntLvPacket(4, func(writer *binary.Writer) {
		if req.Type == RequestTypeLogin {
			writer.WriteUInt32(uint32(req.SequenceID))
			writer.WriteUInt32(t.Version.AppId)
			writer.WriteUInt32(t.Version.SubAppId)
			writer.Write([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00})
			tgt := t.Sig.TGT
			if len(tgt) == 0 || len(tgt) == 4 {
				writer.WriteUInt32(0x04)
			} else {
				writer.WriteUInt32(uint32(len(tgt) + 4))
				writer.Write(tgt)
			}
		}
		writer.WriteString(req.CommandName)
		writer.WriteIntLvPacket(4, func(w *binary.Writer) {
			w.Write(t.Sig.OutPacketSessionID)
		})
		// writer.WriteUInt32(uint32(len(t.Sig.OutPacketSessionID) + 4))
		// w.Write(t.Sig.OutPacketSessionID)
		if req.Type == RequestTypeLogin {
			writer.WriteString(t.Device.IMEI)
			writer.WriteUInt32(0x04)
			{
				writer.WriteUInt16(uint16(len(t.Sig.Ksid)) + 2)
				writer.Write(t.Sig.Ksid)
			}
		}
		writer.WriteUInt32(0x04)
	})

	w.WriteIntLvPacket(4, func(w *binary.Writer) {
		w.Write(req.Body)
	})
	// w.WriteUInt32(uint32(len(req.Body) + 4))
	// w.Write(req.Body)
}

// PackPacket packs a packet.
func (t *Transport) PackPacket(req *Request) []byte {
	// todo(wdvxdr): combine pack packet, send packet and return the response
	if len(t.Sig.D2) == 0 {
		req.EncryptType = EncryptTypeEmptyKey
	}

	return binary.NewWriterF(func(w *binary.Writer) {
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
		w.Write(binary.NewWriterF(func(w *binary.Writer) {
			// encrypt body
			switch req.EncryptType {
			case EncryptTypeD2Key:
				wt, cl := binary.OpenWriterF(func(w *binary.Writer) { t.packBody(req, w) })
				w.EncryptAndWrite(t.Sig.D2Key, wt)
				cl()
			case EncryptTypeEmptyKey:
				wt, cl := binary.OpenWriterF(func(w *binary.Writer) { t.packBody(req, w) })
				w.EncryptAndWrite(emptyKey, wt)
				cl()
			default:
				t.packBody(req, w)
			}
		}))
		w.WriteUInt32At(pos, uint32(w.Len()))
	})
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
		req.Body = binary.NewTeaCipher(t.Sig.D2Key).Decrypt(body)
	case EncryptTypeEmptyKey:
		req.Body = binary.NewTeaCipher(emptyKey).Decrypt(body)
	}
	return req
}
