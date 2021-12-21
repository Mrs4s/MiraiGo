package packets

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
)

var (
	ErrUnknownFlag    = errors.New("unknown flag")
	ErrInvalidPayload = errors.New("invalid payload")
	ErrDecryptFailed  = errors.New("decrypt failed")
	ErrSessionExpired = errors.New("session expired")
	ErrPacketDropped  = errors.New("packet dropped")
)

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

func BuildCode2DRequestPacket(seq uint32, j uint64, cmd uint16, bodyFunc func(writer *binary.Writer)) []byte {
	return binary.NewWriterF(func(w *binary.Writer) {
		body := binary.NewWriterF(bodyFunc)
		w.WriteByte(2)
		w.WriteUInt16(uint16(43 + len(body) + 1))
		w.WriteUInt16(cmd)
		w.Write(make([]byte, 21))
		w.WriteByte(3)
		w.WriteUInt16(0)
		w.WriteUInt16(50) // version
		w.WriteUInt32(seq)
		w.WriteUInt64(j)
		w.Write(body)
		w.WriteByte(3)
	})
}

func BuildSsoPacket(seq uint16, appID, subAppID uint32, commandName, imei string, extData, outPacketSessionId, body, ksid []byte) []byte {
	return binary.NewWriterF(func(p *binary.Writer) {
		p.WriteIntLvPacket(4, func(writer *binary.Writer) {
			writer.WriteUInt32(uint32(seq))
			writer.WriteUInt32(appID)
			writer.WriteUInt32(subAppID)
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
	})
}

func ParseIncomingPacket(payload, d2key []byte) (*IncomingPacket, error) {
	if len(payload) < 6 {
		return nil, errors.WithStack(ErrInvalidPayload)
	}
	reader := binary.NewReader(payload)
	flag1 := reader.ReadInt32()
	flag2 := reader.ReadByte()
	if reader.ReadByte() != 0 { // flag3
		// return nil, errors.WithStack(ErrUnknownFlag)
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
	headLen := reader.ReadInt32()
	if headLen-4 > int32(reader.Len()) {
		return nil, errors.WithStack(ErrPacketDropped)
	}
	head := binary.NewReader(reader.ReadBytes(int(headLen) - 4))
	seqID := head.ReadInt32()
	retCode := head.ReadInt32()
	if retCode != 0 {
		if retCode == -10008 {
			return nil, errors.WithStack(ErrSessionExpired)
		}
		return nil, errors.New("return code unsuccessful: " + strconv.FormatInt(int64(retCode), 10))
	}
	head.ReadBytes(int(head.ReadInt32()) - 4) // extra data
	commandName := head.ReadString()
	sessionID := head.ReadBytes(int(head.ReadInt32()) - 4)
	if commandName == "Heartbeat.Alive" {
		return &IncomingPacket{
			SequenceId:  uint16(seqID),
			Flag2:       flag2,
			CommandName: commandName,
			SessionId:   sessionID,
			Payload:     []byte{},
		}, nil
	}
	compressedFlag := head.ReadInt32()
	reader.ReadInt32()
	packet := func() []byte {
		if compressedFlag == 0 {
			return reader.ReadAvailable()
		}
		if compressedFlag == 1 {
			return binary.ZlibUncompress(reader.ReadAvailable())
		}
		if compressedFlag == 8 {
			return reader.ReadAvailable()
		}
		return nil
	}()
	return &IncomingPacket{
		SequenceId:  uint16(seqID),
		Flag2:       flag2,
		CommandName: commandName,
		SessionId:   sessionID,
		Payload:     packet,
	}, nil
}

func (pkt *IncomingPacket) DecryptPayload(ecdhShareKey, random, sessionKey []byte) ([]byte, error) {
	reader := binary.NewReader(pkt.Payload)
	if flag := reader.ReadByte(); flag != 2 {
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
			tea := binary.NewTeaCipher(ecdhShareKey)
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
