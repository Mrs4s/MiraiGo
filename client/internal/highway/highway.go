package highway

import (
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

// see com/tencent/mobileqq/highway/utils/BaseConstants.java#L120-L121
const (
	_REQ_CMD_DATA        = "PicUp.DataUp"
	_REQ_CMD_HEART_BREAK = "PicUp.Echo"
)

type Session struct {
	Uin        string
	AppID      int32
	SigSession []byte
	SessionKey []byte
	SsoAddr    []Addr

	seq int32
}

const highwayMaxResponseSize int32 = 1024 * 100 // 100k

func (s *Session) AddrLength() int {
	return len(s.SsoAddr)
}

func (s *Session) AppendAddr(ip, port uint32) {
	addr := Addr{
		IP:   ip,
		Port: int(port),
	}
	s.SsoAddr = append(s.SsoAddr, addr)
}

func (s *Session) Upload(addr Addr, trans Transaction) error {
	conn, err := net.DialTimeout("tcp", addr.String(), time.Second*3)
	if err != nil {
		return errors.Wrap(err, "connect error")
	}
	defer conn.Close()

	const chunkSize = 8192 * 8
	chunk := make([]byte, chunkSize)
	offset := 0
	reader := binary.NewNetworkReader(conn)
	for {
		chunk = chunk[:chunkSize]
		rl, err := io.ReadFull(trans.Body, chunk)
		if errors.Is(err, io.EOF) {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
			MsgBasehead: s.dataHighwayHead(_REQ_CMD_DATA, 4096, trans.CommandID, 2052),
			MsgSeghead: &pb.SegHead{
				Filesize:      trans.Size,
				Dataoffset:    int64(offset),
				Datalength:    int32(rl),
				Serviceticket: trans.Ticket,
				Md5:           ch[:],
				FileMd5:       trans.Sum,
			},
			ReqExtendinfo: []byte{},
		})
		offset += rl
		frame := newFrame(head, chunk)
		_, err = frame.WriteTo(conn)
		if err != nil {
			return errors.Wrap(err, "write conn error")
		}
		rspHead, _, err := readResponse(reader)
		if err != nil {
			return errors.Wrap(err, "highway upload error")
		}
		if rspHead.ErrorCode != 0 {
			return errors.New("upload failed")
		}
	}
	return nil
}

type ExcitingInput struct {
	CommandID int32
	Body      io.ReadSeeker
	Ticket    []byte
	Ext       []byte
}

func (s *Session) UploadExciting(input ExcitingInput) ([]byte, error) {
	fileMd5, fileLength := utils.ComputeMd5AndLength(input.Body)
	_, _ = input.Body.Seek(0, io.SeekStart)
	addr := s.SsoAddr[0]
	url := fmt.Sprintf("http://%v/cgi-bin/httpconn?htcmd=0x6FF0087&Uin=%v", addr, s.Uin)
	var (
		rspExt    []byte
		offset    int64 = 0
		chunkSize       = 524288
	)
	chunk := make([]byte, chunkSize)
	w := binary.SelectWriter()
	w.Reset()
	w.Grow(600 * 1024) // 复用,600k 不要放回池中
	for {
		chunk = chunk[:chunkSize]
		rl, err := io.ReadFull(input.Body, chunk)
		if err == io.EOF {
			break
		}
		if err == io.ErrUnexpectedEOF {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
			MsgBasehead: s.dataHighwayHead(_REQ_CMD_DATA, 0, input.CommandID, 0),
			MsgSeghead: &pb.SegHead{
				Filesize:      fileLength,
				Dataoffset:    offset,
				Datalength:    int32(rl),
				Serviceticket: input.Ticket,
				Md5:           ch[:],
				FileMd5:       fileMd5,
			},
			ReqExtendinfo: input.Ext,
		})
		offset += int64(rl)
		frame := newFrame(head, chunk)
		req, _ := http.NewRequest("POST", url, &frame)
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Connection", "Keep-Alive")
		req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)")
		req.Header.Set("Pragma", "no-cache")
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "request error")
		}
		body, _ := io.ReadAll(rsp.Body)
		_ = rsp.Body.Close()
		r := binary.NewReader(body)
		r.ReadByte()
		hl := r.ReadInt32()
		a2 := r.ReadInt32()
		h := r.ReadBytes(int(hl))
		r.ReadBytes(int(a2))
		rspHead := new(pb.RspDataHighwayHead)
		if err = proto.Unmarshal(h, rspHead); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
		}
		if rspHead.ErrorCode != 0 {
			return nil, errors.Errorf("upload failed: %d", rspHead.ErrorCode)
		}
		if rspHead.RspExtendinfo != nil {
			rspExt = rspHead.RspExtendinfo
		}
	}
	return rspExt, nil
}

func (s *Session) nextSeq() int32 {
	return atomic.AddInt32(&s.seq, 2)
}

func (s *Session) dataHighwayHead(cmd string, flag, cmdID, locale int32) *pb.DataHighwayHead {
	return &pb.DataHighwayHead{
		Version:   1,
		Uin:       s.Uin,
		Command:   cmd,
		Seq:       s.nextSeq(),
		Appid:     s.AppID,
		Dataflag:  flag,
		CommandId: cmdID,
		LocaleId:  locale,
	}
}

func (s *Session) sendHeartbreak(conn net.Conn) error {
	head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
		MsgBasehead: s.dataHighwayHead(_REQ_CMD_HEART_BREAK, 4096, 0, 2052),
	})
	frame := newFrame(head, nil)
	_, err := frame.WriteTo(conn)
	return err
}

func (s *Session) sendEcho(conn net.Conn) error {
	err := s.sendHeartbreak(conn)
	if err != nil {
		return errors.Wrap(err, "echo error")
	}
	if _, _, err = readResponse(binary.NewNetworkReader(conn)); err != nil {
		return errors.Wrap(err, "echo error")
	}
	return nil
}

func readResponse(r *binary.NetworkReader) (*pb.RspDataHighwayHead, []byte, error) {
	_, err := r.ReadByte()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read byte")
	}
	hl, _ := r.ReadInt32()
	a2, _ := r.ReadInt32()
	if hl > highwayMaxResponseSize || a2 > highwayMaxResponseSize {
		return nil, nil, errors.Errorf("highway response invild. head size: %v body size: %v", hl, a2)
	}
	head, _ := r.ReadBytes(int(hl))
	payload, _ := r.ReadBytes(int(a2))
	_, _ = r.ReadByte()
	rsp := new(pb.RspDataHighwayHead)
	if err = proto.Unmarshal(head, rsp); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp, payload, nil
}
