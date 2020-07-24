package client

import (
	"crypto/md5"
	"errors"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/golang/protobuf/proto"
	"net"
	"time"
)

func (c *QQClient) highwayUploadImage(ser string, updKey, img []byte, cmdId int32) error {
	conn, err := net.DialTimeout("tcp", ser, time.Second*5)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return err
	}
	h := md5.Sum(img)
	pkt := c.buildImageUploadPacket(img, updKey, cmdId, h)
	for _, p := range pkt {
		_, err = conn.Write(p)
	}
	if err != nil {
		return err
	}
	r := binary.NewNetworkReader(conn)
	_, err = r.ReadByte()
	if err != nil {
		return err
	}
	hl, _ := r.ReadInt32()
	_, _ = r.ReadBytes(4)
	payload, _ := r.ReadBytes(int(hl))
	_ = conn.Close()
	rsp := pb.RspDataHighwayHead{}
	if err = proto.Unmarshal(payload, &rsp); err != nil {
		return err
	}
	if rsp.ErrorCode != 0 {
		return errors.New("upload failed")
	}
	return nil
}
