package client

import (
	"bytes"
	"crypto/md5"
	binary2 "encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
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

func (c *QQClient) highwayUploadByBDH(stream io.ReadSeeker, cmdId int32, ticket, ext []byte) ([]byte, error) {
	// TODO: encrypted upload support.
	if len(c.srvSsoAddrs) == 0 {
		return nil, errors.New("srv addrs not found. maybe miss some packet?")
	}
	h := md5.New()
	length, _ := io.Copy(h, stream)
	chunkSize := 8192 * 16
	fh := h.Sum(nil)
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
			return nil, errors.New("upload failed")
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
