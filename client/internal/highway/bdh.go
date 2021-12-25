package highway

import (
	"crypto/md5"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

type BdhInput struct {
	CommandID int32
	File      string // upload multi-thread required
	Body      io.ReadSeeker
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

	sum, length := utils.ComputeMd5AndLength(input.Body)
	_, _ = input.Body.Seek(0, io.SeekStart)
	if err := input.encrypt(s.SessionKey); err != nil {
		return nil, errors.Wrap(err, "encrypt error")
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
	w := binary.SelectWriter()
	defer binary.PutWriter(w)
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
			MsgBasehead: &pb.DataHighwayHead{
				Version:   1,
				Uin:       s.Uin,
				Command:   "PicUp.DataUp",
				Seq:       s.nextSeq(),
				Appid:     s.AppID,
				Dataflag:  4096,
				CommandId: input.CommandID,
				LocaleId:  2052,
			},
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
		w.Reset()
		writeHeadBody(w, head, chunk)
		_, err = conn.Write(w.Bytes())
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
	if len(s.SsoAddr) == 0 {
		return nil, errors.New("srv addrs not found. maybe miss some packet?")
	}
	addr := s.SsoAddr[0].String()

	stat, err := os.Stat(input.File)
	if err != nil {
		return nil, errors.Wrap(err, "get stat error")
	}
	file, err := os.OpenFile(input.File, os.O_RDONLY, 0o666)
	if err != nil {
		return nil, errors.Wrap(err, "open file error")
	}
	sum, length := utils.ComputeMd5AndLength(file)
	_, _ = file.Seek(0, io.SeekStart)

	if err := input.encrypt(s.SessionKey); err != nil {
		return nil, errors.Wrap(err, "encrypt error")
	}

	// for small file and small thread count,
	// use UploadBDH instead of UploadBDHMultiThread
	if length < 1024*1024*3 || threadCount < 2 {
		input.Body = file
		return s.UploadBDH(input)
	}

	type BlockMetaData struct {
		Id          int
		BeginOffset int64
		EndOffset   int64
	}
	const blockSize int64 = 1024 * 512
	var (
		blocks        []*BlockMetaData
		rspExt        []byte
		BlockId       = ^uint32(0) // -1
		uploadedCount uint32
		cond          = sync.NewCond(&sync.Mutex{})
	)
	// Init Blocks
	{
		var temp int64 = 0
		for temp+blockSize < stat.Size() {
			blocks = append(blocks, &BlockMetaData{
				Id:          len(blocks),
				BeginOffset: temp,
				EndOffset:   temp + blockSize,
			})
			temp += blockSize
		}
		blocks = append(blocks, &BlockMetaData{
			Id:          len(blocks),
			BeginOffset: temp,
			EndOffset:   stat.Size(),
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
		chunk, _ := os.OpenFile(input.File, os.O_RDONLY, 0o666)
		defer chunk.Close()
		reader := binary.NewNetworkReader(conn)
		if err = s.sendEcho(conn); err != nil {
			return err
		}

		buffer := make([]byte, blockSize)
		w := binary.SelectWriter()
		w.Reset()
		w.Grow(600 * 1024) // 复用,600k 不要放回池中
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
			buffer = buffer[:blockSize]
			_, _ = chunk.Seek(block.BeginOffset, io.SeekStart)
			ri, err := io.ReadFull(chunk, buffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				if err == io.ErrUnexpectedEOF {
					buffer = buffer[:ri]
				} else {
					return err
				}
			}
			ch := md5.Sum(buffer)
			head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
				MsgBasehead: &pb.DataHighwayHead{
					Version:   1,
					Uin:       s.Uin,
					Command:   "PicUp.DataUp",
					Seq:       s.nextSeq(),
					Appid:     s.AppID,
					Dataflag:  4096,
					CommandId: input.CommandID,
					LocaleId:  2052,
				},
				MsgSeghead: &pb.SegHead{
					Filesize:      stat.Size(),
					Dataoffset:    block.BeginOffset,
					Datalength:    int32(ri),
					Serviceticket: input.Ticket,
					Md5:           ch[:],
					FileMd5:       sum,
				},
				ReqExtendinfo: input.Ext,
			})
			w.Reset()
			writeHeadBody(w, head, buffer)
			_, err = conn.Write(w.Bytes())
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
	err = group.Wait()
	return rspExt, err
}
