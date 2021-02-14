package client

import (
	"bytes"
	"crypto/md5"
	binary2 "encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (c *QQClient) highwayUpload(ip uint32, port int, updKey, data []byte, cmdId int32) error {
	return c.highwayUploadStream(ip, port, updKey, bytes.NewReader(data), cmdId)
}

func (c *QQClient) highwayUploadStream(ip uint32, port int, updKey []byte, stream io.ReadSeeker, cmdId int32) error {
	addr := net.TCPAddr{
		IP:   make([]byte, 4),
		Port: port,
	}
	binary2.LittleEndian.PutUint32(addr.IP, ip)
	h := md5.New()
	length, _ := io.Copy(h, stream)
	fh := h.Sum(nil)
	chunkSize := 8192 * 8
	_, _ = stream.Seek(0, io.SeekStart)
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		return errors.Wrap(err, "connect error")
	}
	defer conn.Close()
	offset := 0
	reader := binary.NewNetworkReader(conn)
	for {
		chunk := make([]byte, chunkSize)
		rl, err := io.ReadFull(stream, chunk)
		if err == io.EOF {
			break
		}
		if err == io.ErrUnexpectedEOF {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
			MsgBasehead: &pb.DataHighwayHead{
				Version:   1,
				Uin:       strconv.FormatInt(c.Uin, 10),
				Command:   "PicUp.DataUp",
				Seq:       c.nextGroupDataTransSeq(),
				Appid:     int32(c.version.AppId),
				Dataflag:  4096,
				CommandId: cmdId,
				LocaleId:  2052,
			},
			MsgSeghead: &pb.SegHead{
				Filesize:      length,
				Dataoffset:    int64(offset),
				Datalength:    int32(rl),
				Serviceticket: updKey,
				Md5:           ch[:],
				FileMd5:       fh[:],
			},
			ReqExtendinfo: EmptyBytes,
		})
		offset += rl
		_, err = conn.Write(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteByte(40)
			w.WriteUInt32(uint32(len(head)))
			w.WriteUInt32(uint32(len(chunk)))
			w.Write(head)
			w.Write(chunk)
			w.WriteByte(41)
		}))
		if err != nil {
			return errors.Wrap(err, "write conn error")
		}
		rspHead, _, err := highwayReadResponse(reader)
		if err != nil {
			return errors.Wrap(err, "highway upload error")
		}
		if rspHead.ErrorCode != 0 {
			return errors.New("upload failed")
		}
	}
	return nil
}

func (c *QQClient) highwayUploadByBDH(stream io.ReadSeeker, cmdId int32, ticket, ext []byte, encrypt bool) ([]byte, error) {
	if len(c.srvSsoAddrs) == 0 {
		return nil, errors.New("srv addrs not found. maybe miss some packet?")
	}
	if encrypt {
		if c.highwaySession == nil || len(c.highwaySession.SessionKey) == 0 {
			return nil, errors.New("session key not found. maybe miss some packet?")
		}
		ext = binary.NewTeaCipher(c.highwaySession.SessionKey).Encrypt(ext)
	}
	h := md5.New()
	length, _ := io.Copy(h, stream)
	fh := h.Sum(nil)
	chunkSize := 8192 * 16
	_, _ = stream.Seek(0, io.SeekStart)
	conn, err := net.DialTimeout("tcp", c.srvSsoAddrs[0], time.Second*20)
	if err != nil {
		return nil, errors.Wrap(err, "connect error")
	}
	defer conn.Close()
	offset := 0
	reader := binary.NewNetworkReader(conn)
	if err = c.highwaySendHeartbreak(conn); err != nil {
		return nil, errors.Wrap(err, "echo error")
	}
	if _, _, err = highwayReadResponse(reader); err != nil {
		return nil, errors.Wrap(err, "echo error")
	}
	var rspExt []byte
	for {
		chunk := make([]byte, chunkSize)
		rl, err := io.ReadFull(stream, chunk)
		if err == io.EOF {
			break
		}
		if err == io.ErrUnexpectedEOF {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
			MsgBasehead: &pb.DataHighwayHead{
				Version:   1,
				Uin:       strconv.FormatInt(c.Uin, 10),
				Command:   "PicUp.DataUp",
				Seq:       c.nextGroupDataTransSeq(),
				Appid:     int32(c.version.AppId),
				Dataflag:  4096,
				CommandId: cmdId,
				LocaleId:  2052,
			},
			MsgSeghead: &pb.SegHead{
				Filesize:      length,
				Dataoffset:    int64(offset),
				Datalength:    int32(rl),
				Serviceticket: ticket,
				Md5:           ch[:],
				FileMd5:       fh[:],
			},
			ReqExtendinfo: ext,
		})
		offset += rl
		_, err = conn.Write(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteByte(40)
			w.WriteUInt32(uint32(len(head)))
			w.WriteUInt32(uint32(len(chunk)))
			w.Write(head)
			w.Write(chunk)
			w.WriteByte(41)
		}))
		if err != nil {
			return nil, errors.Wrap(err, "write conn error")
		}
		rspHead, _, err := highwayReadResponse(reader)
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
			ticket = rspHead.MsgSeghead.Serviceticket
		}
	}
	return rspExt, nil
}

func (c *QQClient) highwayUploadFileMultiThreadingByBDH(path string, cmdId int32, threadCount int, ticket, ext []byte, encrypt bool) ([]byte, error) {
	if len(c.srvSsoAddrs) == 0 {
		return nil, errors.New("srv addrs not found. maybe miss some packet?")
	}
	if encrypt {
		if c.highwaySession == nil || len(c.highwaySession.SessionKey) == 0 {
			return nil, errors.New("session key not found. maybe miss some packet?")
		}
		ext = binary.NewTeaCipher(c.highwaySession.SessionKey).Encrypt(ext)
	}
	stat, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "get stat error")
	}
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, errors.Wrap(err, "open file error")
	}
	defer file.Close()
	if stat.Size() < 1024*1024*3 {
		return c.highwayUploadByBDH(file, cmdId, ticket, ext, false)
	}
	type BlockMetaData struct {
		Id          int
		BeginOffset int64
		EndOffset   int64
	}
	h := md5.New()
	_, _ = io.Copy(h, file)
	fh := h.Sum(nil)
	const blockSize int64 = 1024 * 512
	var (
		blocks        []*BlockMetaData
		rspExt        []byte
		BlockId       = ^uint32(0) // -1
		uploadedCount uint32
		lastErr       error
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
		conn, err := net.DialTimeout("tcp", c.srvSsoAddrs[0], time.Second*20)
		if err != nil {
			return errors.Wrap(err, "connect error")
		}
		defer conn.Close()
		chunk, _ := os.OpenFile(path, os.O_RDONLY, 0666)
		defer chunk.Close()
		reader := binary.NewNetworkReader(conn)
		if err = c.highwaySendHeartbreak(conn); err != nil {
			return errors.Wrap(err, "echo error")
		}
		if _, _, err = highwayReadResponse(reader); err != nil {
			return errors.Wrap(err, "echo error")
		}
		for {
			nextId := atomic.AddUint32(&BlockId, 1)
			if nextId >= uint32(len(blocks)) {
				break
			}
			block := blocks[nextId]
			if block.Id == len(blocks)-1 {
				cond.L.Lock()
				for atomic.LoadUint32(&uploadedCount) != uint32(len(blocks)-1) && lastErr == nil {
					cond.Wait()
				}
				cond.L.Unlock()
				if lastErr != nil {
					break
				}
			}
			buffer := make([]byte, blockSize)
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
					Uin:       strconv.FormatInt(c.Uin, 10),
					Command:   "PicUp.DataUp",
					Seq:       c.nextGroupDataTransSeq(),
					Appid:     int32(c.version.AppId),
					Dataflag:  4096,
					CommandId: cmdId,
					LocaleId:  2052,
				},
				MsgSeghead: &pb.SegHead{
					Filesize:      stat.Size(),
					Dataoffset:    block.BeginOffset,
					Datalength:    int32(ri),
					Serviceticket: ticket,
					Md5:           ch[:],
					FileMd5:       fh[:],
				},
				ReqExtendinfo: ext,
			})
			_, err = conn.Write(binary.NewWriterF(func(w *binary.Writer) {
				w.WriteByte(40)
				w.WriteUInt32(uint32(len(head)))
				w.WriteUInt32(uint32(len(buffer)))
				w.Write(head)
				w.Write(buffer)
				w.WriteByte(41)
			}))
			if err != nil {
				return errors.Wrap(err, "write conn error")
			}
			rspHead, _, err := highwayReadResponse(reader)
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
	wg := sync.WaitGroup{}
	wg.Add(threadCount)
	for i := 0; i < threadCount; i++ {
		go func() {
			defer wg.Done()
			defer cond.Signal()
			if err := doUpload(); err != nil {
				lastErr = err
			}
		}()
	}
	wg.Wait()
	return rspExt, lastErr
}

func (c *QQClient) highwaySendHeartbreak(conn net.Conn) error {
	head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
		MsgBasehead: &pb.DataHighwayHead{
			Version:   1,
			Uin:       strconv.FormatInt(c.Uin, 10),
			Command:   "PicUp.Echo",
			Seq:       c.nextGroupDataTransSeq(),
			Appid:     int32(c.version.AppId),
			Dataflag:  4096,
			CommandId: 0,
			LocaleId:  2052,
		},
	})
	_, err := conn.Write(binary.NewWriterF(func(w *binary.Writer) {
		w.WriteByte(40)
		w.WriteUInt32(uint32(len(head)))
		w.WriteUInt32(0)
		w.Write(head)
		w.WriteByte(41)
	}))
	return err
}

func highwayReadResponse(r *binary.NetworkReader) (*pb.RspDataHighwayHead, []byte, error) {
	_, err := r.ReadByte()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read byte")
	}
	hl, _ := r.ReadInt32()
	a2, _ := r.ReadInt32()
	head, _ := r.ReadBytes(int(hl))
	payload, _ := r.ReadBytes(int(a2))
	_, _ = r.ReadByte()
	rsp := new(pb.RspDataHighwayHead)
	if err = proto.Unmarshal(head, rsp); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp, payload, nil
}

func (c *QQClient) excitingUploadStream(stream io.ReadSeeker, cmdId int32, ticket, ext []byte) ([]byte, error) {
	fileMd5, fileLength := utils.ComputeMd5AndLength(stream)
	_, _ = stream.Seek(0, io.SeekStart)
	url := fmt.Sprintf("http://%v/cgi-bin/httpconn?htcmd=0x6FF0087&uin=%v", c.srvSsoAddrs[0], c.Uin)
	var (
		rspExt    []byte
		offset    int64 = 0
		chunkSize       = 524288
	)
	for {
		chunk := make([]byte, chunkSize)
		rl, err := io.ReadFull(stream, chunk)
		if err == io.EOF {
			break
		}
		if err == io.ErrUnexpectedEOF {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(&pb.ReqDataHighwayHead{
			MsgBasehead: &pb.DataHighwayHead{
				Version:   1,
				Uin:       strconv.FormatInt(c.Uin, 10),
				Command:   "PicUp.DataUp",
				Seq:       c.nextGroupDataTransSeq(),
				Appid:     int32(c.version.AppId),
				Dataflag:  0,
				CommandId: cmdId,
				LocaleId:  0,
			},
			MsgSeghead: &pb.SegHead{
				Filesize:      fileLength,
				Dataoffset:    offset,
				Datalength:    int32(rl),
				Serviceticket: ticket,
				Md5:           ch[:],
				FileMd5:       fileMd5,
			},
			ReqExtendinfo: ext,
		})
		offset += int64(rl)
		req, _ := http.NewRequest("POST", url, bytes.NewReader(binary.NewWriterF(func(w *binary.Writer) {
			w.WriteByte(40)
			w.WriteUInt32(uint32(len(head)))
			w.WriteUInt32(uint32(len(chunk)))
			w.Write(head)
			w.Write(chunk)
			w.WriteByte(41)
		})))
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Connection", "Keep-Alive")
		req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)")
		req.Header.Set("Pragma", "no-cache")
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "request error")
		}
		body, _ := ioutil.ReadAll(rsp.Body)
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

// 只是为了写的跟上面一样长(bushi，当然也应该是最快的玩法
func (c *QQClient) uploadPtt(ip string, port int32, updKey, fileKey, data, md5 []byte) error {
	url := make([]byte, 512)[:0]
	url = append(url, "http://"...)
	url = append(url, ip...)
	url = append(url, ':')
	url = strconv.AppendInt(url, int64(port), 10)
	url = append(url, "/?ver=4679&ukey="...)
	p := len(url)
	url = url[:p+len(updKey)*2]
	hex.Encode(url[p:], updKey)
	url = append(url, "&filekey="...)
	p = len(url)
	url = url[:p+len(fileKey)*2]
	hex.Encode(url[p:], fileKey)
	url = append(url, "&filesize="...)
	url = strconv.AppendInt(url, int64(len(data)), 10)
	url = append(url, "&bmd5="...)
	p = len(url)
	url = url[:p+32]
	hex.Encode(url[p:], md5)
	url = append(url, "&mType=pttDu&voice_encodec=1"...)
	_, err := utils.HttpPostBytes(string(url), data)
	return errors.Wrap(err, "failed to upload ptt")
}

func (c *QQClient) uploadGroupHeadPortrait(groupCode int64, img []byte) error {
	url := fmt.Sprintf(
		"http://htdata3.qq.com/cgi-bin/httpconn?htcmd=0x6ff0072&ver=5520&ukey=%v&range=0&uin=%v&seq=23&groupuin=%v&filetype=3&imagetype=5&userdata=0&subcmd=1&subver=101&clip=0_0_0_0&filesize=%v",
		c.getSKey(),
		c.Uin,
		groupCode,
		len(img),
	)
	req, err := http.NewRequest("POST", url, bytes.NewReader(img))
	req.Header["User-Agent"] = []string{"Dalvik/2.1.0 (Linux; U; Android 7.1.2; PCRT00 Build/N2G48H)"}
	req.Header["Content-Type"] = []string{"multipart/form-data;boundary=****"}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to upload group head portrait")
	}
	rsp.Body.Close()
	return nil
}
