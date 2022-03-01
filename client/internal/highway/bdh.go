package highway

import (
	"crypto/md5"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

type BdhInput struct {
	CommandID int32
	Body      io.Reader
	Sum       []byte // md5 sum of body
	Size      int64  // body size
	Ticket    []byte
	Ext       []byte
	Encrypt   bool
}

func (bdh *BdhInput) encrypt(key []byte) error {
	if bdh.Encrypt {
		if len(key) == 0 {
			return errors.New("session key not found. maybe miss some packet?")
		}
		bdh.Ext = binary.NewTeaCipher(key).Encrypt(bdh.Ext)
	}
	return nil
}

func (s *Session) UploadBDH(input BdhInput) ([]byte, error) {
	if len(s.SsoAddr) == 0 {
		return nil, errors.New("srv addrs not found. maybe miss some packet?")
	}
	addr := s.SsoAddr[0].String()

	if err := input.encrypt(s.SessionKey); err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second*20)
	if err != nil {
		return nil, errors.Wrap(err, "connect error")
	}
	defer conn.Close()

	reader := binary.NewNetworkReader(conn)
	if err = s.sendEcho(conn); err != nil {
		return nil, err
	}

	const chunkSize = 256 * 1024
	var rspExt []byte
	offset := 0
	chunk := make([]byte, chunkSize)
	for {
		chunk = chunk[:chunkSize]
		rl, err := io.ReadFull(input.Body, chunk)
		if errors.Is(err, io.EOF) {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
			MsgBasehead: s.dataHighwayHead(_REQ_CMD_DATA, 4096, input.CommandID, 2052),
			MsgSeghead: &pb.SegHead{
				Filesize:      input.Size,
				Dataoffset:    int64(offset),
				Datalength:    int32(rl),
				Serviceticket: input.Ticket,
				Md5:           ch[:],
				FileMd5:       input.Sum,
			},
			ReqExtendinfo: input.Ext,
		})
		offset += rl
		frame := newFrame(head, chunk)
		_, err = frame.WriteTo(conn)
		if err != nil {
			return nil, errors.Wrap(err, "write conn error")
		}
		rspHead, _, err := readResponse(reader)
		if err != nil {
			return nil, errors.Wrap(err, "highway upload error")
		}
		if rspHead.ErrorCode != 0 {
			return nil, errors.Errorf("upload failed: %d", rspHead.ErrorCode)
		}
		if rspHead.RspExtendinfo != nil {
			rspExt = rspHead.RspExtendinfo
		}
		if rspHead.MsgSeghead != nil && rspHead.MsgSeghead.Serviceticket != nil {
			input.Ticket = rspHead.MsgSeghead.Serviceticket
		}
	}
	return rspExt, nil
}

func (s *Session) UploadBDHMultiThread(input BdhInput, threadCount int) ([]byte, error) {
	// for small file and small thread count,
	// use UploadBDH instead of UploadBDHMultiThread
	if input.Size < 1024*1024*3 || threadCount < 2 {
		return s.UploadBDH(input)
	}

	if len(s.SsoAddr) == 0 {
		return nil, errors.New("srv addrs not found. maybe miss some packet?")
	}
	addr := s.SsoAddr[0].String()

	if err := input.encrypt(s.SessionKey); err != nil {
		return nil, err
	}

	const blockSize int64 = 1024 * 512
	var (
		rspExt          []byte
		completedThread uint32
		cond            = sync.NewCond(&sync.Mutex{})
		offset          = int64(0)
		count           = (input.Size + blockSize - 1) / blockSize
		id              = 0
	)
	doUpload := func() error {
		// send signal complete uploading
		defer func() {
			atomic.AddUint32(&completedThread, 1)
			cond.Signal()
		}()

		conn, err := net.DialTimeout("tcp", addr, time.Second*20)
		if err != nil {
			return errors.Wrap(err, "connect error")
		}
		defer conn.Close()
		reader := binary.NewNetworkReader(conn)
		if err = s.sendEcho(conn); err != nil {
			return err
		}

		chunk := make([]byte, blockSize)
		for {
			cond.L.Lock() // lock protect reading
			off := offset
			offset += blockSize
			id++
			if int64(id) == count { // last
				for atomic.LoadUint32(&completedThread) != uint32(threadCount-1) {
					cond.Wait()
				}
			} else if int64(id) > count {
				cond.L.Unlock()
				break
			}
			chunk = chunk[:blockSize]
			n, err := io.ReadFull(input.Body, chunk)
			cond.L.Unlock()

			if err != nil {
				if err == io.EOF {
					break
				}
				if err == io.ErrUnexpectedEOF {
					chunk = chunk[:n]
				} else {
					return err
				}
			}
			ch := md5.Sum(chunk)
			head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
				MsgBasehead: s.dataHighwayHead(_REQ_CMD_DATA, 4096, input.CommandID, 2052),
				MsgSeghead: &pb.SegHead{
					Filesize:      input.Size,
					Dataoffset:    off,
					Datalength:    int32(n),
					Serviceticket: input.Ticket,
					Md5:           ch[:],
					FileMd5:       input.Sum,
				},
				ReqExtendinfo: input.Ext,
			})
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
				return errors.Errorf("upload failed: %d", rspHead.ErrorCode)
			}
			if rspHead.RspExtendinfo != nil {
				rspExt = rspHead.RspExtendinfo
			}
		}
		return nil
	}

	group := errgroup.Group{}
	for i := 0; i < threadCount; i++ {
		group.Go(doUpload)
	}
	return rspExt, group.Wait()
}
