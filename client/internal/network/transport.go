package network

import (
	goBinary "encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"strconv"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/auth"
)

// Transport is a network transport.
type Transport struct {
	// sessionMu sync.Mutex
	Sig     *auth.SigInfo
	Version *auth.AppVersion
	Device  *auth.Device

	// connection
	conn TCPListener
}

func (t *Transport) GetStatistics() *Statistics {
	return t.conn.getStatistics()
}

func (t *Transport) PlannedDisconnect(fun func(*TCPListener)) {
	t.conn.PlannedDisconnect = fun
}

func (t *Transport) UnexpectedDisconnect(fun func(*TCPListener, error)) {
	t.conn.UnexpectedDisconnect = fun
}

func (t *Transport) ConnectFastest(servers []*net.TCPAddr) (chosen *net.TCPAddr, err error) {
	return t.conn.ConnectFastest(servers)
}

func (t *Transport) Close() {
	t.conn.Close()
}

func (t *Transport) Write(data []byte) error {
	return t.conn.Write(data)
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
		body = binary.NewTeaCipher(emptyKey).Encrypt(body)
	}
	w.Write(body)
	binary.PutWriter(w2)

	w.WriteUInt32At(pos, uint32(w.Len()))
	return append([]byte(nil), w.Bytes()...)
}

type PktHandler func(pkt *Request, netErr error)
type RequestHandler func(head []byte) (*Request, error)

func (t *Transport) NetLoop(pktHandler PktHandler, respHandler RequestHandler) {
	go t.netLoop(pktHandler, respHandler)
}

// readPacket 帮助函数(Helper function)
func readPacket(conn *net.TCPConn, minSize, maxSize uint32) ([]byte, error) {
	lBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lBuf)
	if err != nil {
		return nil, err
	}
	l := goBinary.BigEndian.Uint32(lBuf)
	if l < minSize || l > maxSize {
		return nil, fmt.Errorf("parse incoming packet error: invalid packet length %v", l)
	}
	data := make([]byte, l-4)
	_, err = io.ReadFull(conn, data)
	return data, err
}

// netLoop 整个函数周期使用同一个连接，确保不会发生串线这种奇怪的事情
func (t *Transport) netLoop(pktHandler PktHandler, respHandler RequestHandler) {
	conn := t.conn.getConn()
	defer func() {
		if r := recover(); r != nil {
			pktHandler(nil, fmt.Errorf("panic: %v", r))
		}
		t.conn.Close()
	}()
	errCount := 0
	for {
		data, err := readPacket(conn, 4, 10<<20) // max 10MB
		if err != nil {
			// 在且仅在没有新连接建立时断线才被认为是意外的
			if t.conn.getConn() == conn {
				pktHandler(nil, errors.Wrap(ErrConnectionBroken, err.Error()))
			}
			return
		}
		req, err := respHandler(data)
		if err == nil {
			errCount = 0
			goto ok
		}
		errCount++
		if errCount > 2 {
			err = errors.Wrap(ErrConnectionBroken, err.Error())
		}
	ok:
		go pktHandler(req, err)
	}
}
