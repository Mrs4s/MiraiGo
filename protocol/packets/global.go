package packets

import (
	"strconv"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/protocol/crypto"
	"github.com/pkg/errors"
)

var ErrUnknownFlag = errors.New("unknown flag")
var ErrInvalidPayload = errors.New("invalid payload")
var ErrDecryptFailed = errors.New("decrypt failed")
var ErrSessionExpired = errors.New("session expired")
var ErrPacketDropped = errors.New("packet dropped")

type ISendingPacket interface {
	CommandId() uint16
	Writer() *binary.Writer
}

type IncomingPacket struct {
	SequenceId  uint16
	Flag2       byte
	CommandName string
	SessionId   []byte
	Payload     []byte
}

type IEncryptMethod interface {
	DoEncrypt([]byte, []byte) []byte
	Id() byte
}

func BuildOicqRequestPacket(uin int64, commandId uint16, encrypt IEncryptMethod, key []byte, bodyFunc func(writer *binary.Writer)) []byte {
	b := binary.NewWriter()
	bodyFunc(b)

	body := encrypt.DoEncrypt(b.Bytes(), key)
	p := binary.NewWriter()
	p.WriteByte(0x02)
	p.WriteUInt16(27 + 2 + uint16(len(body)))
	p.WriteUInt16(8001)
	p.WriteUInt16(commandId)
	p.WriteUInt16(1)
	p.WriteUInt32(uint32(uin))
	p.WriteByte(3)
	p.WriteByte(encrypt.Id())
	p.WriteByte(0)
	p.WriteUInt32(2)
	p.WriteUInt32(0)
	p.WriteUInt32(0)
	p.Write(body)
	p.WriteByte(0x03)
	return p.Bytes()
}

func BuildSsoPacket(seq uint16, appId uint32, commandName, imei string, extData, outPacketSessionId, body, ksid []byte) []byte {
	p := binary.NewWriter()
	p.WriteIntLvPacket(4, func(writer *binary.Writer) {
		writer.WriteUInt32(uint32(seq))
		writer.WriteUInt32(appId)
		writer.WriteUInt32(appId)
		writer.Write([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00})
		if len(extData) == 0 || len(extData) == 4 {
			writer.WriteUInt32(0x04)
		} else {
			writer.WriteUInt32(uint32(len(extData) + 4))
			writer.Write(extData)
		}
		writer.WriteString(commandName)
		writer.WriteIntLvPacket(4, func(w *binary.Writer) {
			w.Write(outPacketSessionId)
		})
		writer.WriteString(imei)
		writer.WriteUInt32(0x04)
		{
			writer.WriteUInt16(uint16(len(ksid)) + 2)
			writer.Write(ksid)
		}
		writer.WriteUInt32(0x04)
	})

	p.WriteIntLvPacket(4, func(writer *binary.Writer) {
		writer.Write(body)
	})
	return p.Bytes()
}

func ParseIncomingPacket(payload, d2key []byte) (*IncomingPacket, error) {
	if len(payload) < 6 {
		return nil, errors.WithStack(ErrInvalidPayload)
	}
	reader := binary.NewReader(payload)
	flag1 := reader.ReadInt32()
	flag2 := reader.ReadByte()
	if reader.ReadByte() != 0 { // flag3
		return nil, errors.WithStack(ErrUnknownFlag)
	}
	reader.ReadString() // uin string
	decrypted := func() (data []byte) {
		switch flag2 {
		case 0:
			return reader.ReadAvailable()
		case 1:
			d2 := binary.NewTeaCipher(d2key)
			return d2.Decrypt(reader.ReadAvailable())
		case 2:
			z16 := binary.NewTeaCipher(make([]byte, 16))
			return z16.Decrypt(reader.ReadAvailable())
		}
		return nil
	}()
	if len(decrypted) == 0 {
		return nil, errors.WithStack(ErrDecryptFailed)
	}
	if flag1 != 0x0A && flag1 != 0x0B {
		return nil, errors.WithStack(ErrDecryptFailed)
	}
	return parseSsoFrame(decrypted, flag2)
}

func parseSsoFrame(payload []byte, flag2 byte) (*IncomingPacket, error) {
	reader := binary.NewReader(payload)
	if reader.ReadInt32()-4 > int32(reader.Len()) {
		return nil, errors.WithStack(ErrPacketDropped)
	}
	seqId := reader.ReadInt32()
	retCode := reader.ReadInt32()
	if retCode != 0 {
		if retCode == -10008 {
			return nil, errors.WithStack(ErrSessionExpired)
		}
		return nil, errors.New("return code unsuccessful: " + strconv.FormatInt(int64(retCode), 10))
	}
	reader.ReadBytes(int(reader.ReadInt32()) - 4) // extra data
	commandName := reader.ReadString()
	sessionId := reader.ReadBytes(int(reader.ReadInt32()) - 4)
	if commandName == "Heartbeat.Alive" {
		return &IncomingPacket{
			SequenceId:  uint16(seqId),
			Flag2:       flag2,
			CommandName: commandName,
			SessionId:   sessionId,
			Payload:     []byte{},
		}, nil
	}
	compressedFlag := reader.ReadInt32()
	packet := func() []byte {
		if compressedFlag == 0 {
			pktSize := uint64(reader.ReadInt32()) & 0xffffffff
			if pktSize == uint64(reader.Len()) || pktSize == uint64(reader.Len()+4) {
				return reader.ReadAvailable()
			} else {
				return reader.ReadAvailable() // some logic
			}
		}
		if compressedFlag == 1 {
			reader.ReadBytes(4)
			return binary.ZlibUncompress(reader.ReadAvailable()) // ?
		}
		if compressedFlag == 8 {
			return reader.ReadAvailable()
		}
		return nil
	}()
	return &IncomingPacket{
		SequenceId:  uint16(seqId),
		Flag2:       flag2,
		CommandName: commandName,
		SessionId:   sessionId,
		Payload:     packet,
	}, nil
}

func (pkt *IncomingPacket) DecryptPayload(random, sessionKey []byte) ([]byte, error) {
	reader := binary.NewReader(pkt.Payload)
	if reader.ReadByte() != 2 {
		return nil, ErrUnknownFlag
	}
	reader.ReadBytes(2)
	reader.ReadBytes(2)
	reader.ReadUInt16()
	reader.ReadUInt16()
	reader.ReadInt32()
	encryptType := reader.ReadUInt16()
	reader.ReadByte()
	if encryptType == 0 {
		data := func() (decrypted []byte) {
			d := reader.ReadBytes(reader.Len() - 1)
			defer func() {
				if pan := recover(); pan != nil {
					tea := binary.NewTeaCipher(random)
					decrypted = tea.Decrypt(d)
				}
			}()
			tea := binary.NewTeaCipher(crypto.ECDH.InitialShareKey)
			decrypted = tea.Decrypt(d)
			return
		}()
		return data, nil
	}
	if encryptType == 3 {
		return func() []byte {
			d := reader.ReadBytes(reader.Len() - 1)
			return binary.NewTeaCipher(sessionKey).Decrypt(d)
		}(), nil
	}
	if encryptType == 4 {
		panic("todo")
	}
	return nil, errors.WithStack(ErrUnknownFlag)
}
