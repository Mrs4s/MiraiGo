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
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

type BdhInput struct {
	CommandID int32
	Body      io.ReadSeeker
	Ticket    []byte
	Ext       []byte
	Encrypt   bool
}

type BdhMultiThreadInput struct {
	CommandID int32
	Body      io.ReaderAt
	Sum       []byte
	Size      int64
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

func (bdh *BdhMultiThreadInput) encrypt(key []byte) error {
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

	sum, length := utils.ComputeMd5AndLength(input.Body)
	_, _ = input.Body.Seek(0, io.SeekStart)
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
			MsgBasehead: s.dataHighwayHead(4096, input.CommandID, 2052),
			MsgSeghead: &pb.SegHead{
				Filesize:      length,
				Dataoffset:    int64(offset),
				Datalength:    int32(rl),
				Serviceticket: input.Ticket,
				Md5:           ch[:],
				FileMd5:       sum,
			},
			ReqExtendinfo: input.Ext,
		})
		offset += rl
		frame := network.HeadBodyFrame(head, chunk)
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

func (s *Session) UploadBDHMultiThread(input BdhMultiThreadInput, threadCount int) ([]byte, error) {
	// for small file and small thread count,
	// use UploadBDH instead of UploadBDHMultiThread
	if input.Size < 1024*1024*3 || threadCount < 2 {
		return s.UploadBDH(BdhInput{
			CommandID: input.CommandID,
			Body:      io.NewSectionReader(input.Body, 0, input.Size),
			Ticket:    input.Ticket,
			Ext:       input.Ext,
			Encrypt:   input.Encrypt,
		})
	}

	if len(s.SsoAddr) == 0 {
		return nil, errors.New("srv addrs not found. maybe miss some packet?")
	}
	addr := s.SsoAddr[0].String()

	if err := input.encrypt(s.SessionKey); err != nil {
		return nil, err
	}

	type BlockMetaData struct {
		Id     int
		Offset int64
	}
	const blockSize int64 = 1024 * 512
	var (
		blocks        []BlockMetaData
		rspExt        []byte
		BlockId       = ^uint32(0) // -1
		uploadedCount uint32
		cond          = sync.NewCond(&sync.Mutex{})
	)
	// Init Blocks
	{
		var temp int64 = 0
		for temp+blockSize < input.Size {
			blocks = append(blocks, BlockMetaData{
				Id:     len(blocks),
				Offset: temp,
			})
			temp += blockSize
		}
		blocks = append(blocks, BlockMetaData{
			Id:     len(blocks),
			Offset: temp,
		})
	}
	doUpload := func() error {
		// send signal complete uploading
		defer cond.Signal()

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
			nextId := atomic.AddUint32(&BlockId, 1)
			if nextId >= uint32(len(blocks)) {
				break
			}
			block := blocks[nextId]
			if block.Id == len(blocks)-1 {
				cond.L.Lock()
				for atomic.LoadUint32(&uploadedCount) != uint32(len(blocks))-1 {
					cond.Wait()
				}
				cond.L.Unlock()
			}
			chunk = chunk[:blockSize]

			cond.L.Lock() // lock protect reading
			n, err := input.Body.ReadAt(chunk, block.Offset)
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
				MsgBasehead: s.dataHighwayHead(4096, input.CommandID, 2052),
				MsgSeghead: &pb.SegHead{
					Filesize:      input.Size,
					Dataoffset:    block.Offset,
					Datalength:    int32(n),
					Serviceticket: input.Ticket,
					Md5:           ch[:],
					FileMd5:       input.Sum,
				},
				ReqExtendinfo: input.Ext,
			})
			frame := network.HeadBodyFrame(head, chunk)
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
			atomic.AddUint32(&uploadedCount, 1)
		}
		return nil
	}

	group := errgroup.Group{}
	for i := 0; i < threadCount; i++ {
		group.Go(doUpload)
	}
	return rspExt, group.Wait()
}
