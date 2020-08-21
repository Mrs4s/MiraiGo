package client

import (
	"crypto/md5"
	binary2 "encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/utils"
	"google.golang.org/protobuf/proto"
	"net"
	"strconv"
)

func (c *QQClient) highwayUploadImage(ip uint32, port int, updKey, img []byte, cmdId int32) error {
	addr := net.TCPAddr{
		IP:   make([]byte, 4),
		Port: port,
	}
	binary2.LittleEndian.PutUint32(addr.IP, ip)
	conn, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	h := md5.Sum(img)
	pkt := c.buildImageUploadPacket(img, updKey, cmdId, h)
	r := binary.NewNetworkReader(conn)
	for _, p := range pkt {
		_, err = conn.Write(p)
		if err != nil {
			return err
		}
		_, err = r.ReadByte()
		if err != nil {
			return err
		}
		hl, _ := r.ReadInt32()
		a2, _ := r.ReadInt32()
		payload, _ := r.ReadBytes(int(hl))
		_, _ = r.ReadBytes(int(a2))
		r.ReadByte()
		rsp := new(pb.RspDataHighwayHead)
		if err = proto.Unmarshal(payload, rsp); err != nil {
			return err
		}
		if rsp.ErrorCode != 0 {
			return errors.New("upload failed")
		}
	}

	return nil
}

// 只是为了写的跟上面一样长(bushi，当然也应该是最快的玩法
func (c *QQClient) uploadGroupPtt(ip, port int32, updKey, fileKey, data, md5 []byte, codec int64) error {
	url := make([]byte, 512)[:0]
	url = append(url, "http://"...)
	url = append(url, binary.UInt32ToIPV4Address(uint32(ip))...)
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
	return err
}
