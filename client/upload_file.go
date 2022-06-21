package client

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/highway"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x346"
	"github.com/Mrs4s/MiraiGo/client/pb/exciting"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

func init() {
	decoders["OfflineFilleHandleSvr.pb_ftn_CMD_REQ_APPLY_UPLOAD_V3-1700"] = decodePrivateFileUploadReq
}

type LocalFile struct {
	FileName     string
	Body         io.ReadSeeker // LocalFile content body
	RemoteFolder string
	size         int64
	md5          []byte
	sha1         []byte
}

type fileUploadRsp struct {
	Existed   bool
	BusID     int32
	Uuid      []byte
	UploadKey []byte

	// upload group file need
	UploadIpLanV4 []string
	UploadPort    int32
}

func (f *LocalFile) init() {
	md5H := md5.New()
	sha1H := sha1.New()

	whence, _ := f.Body.Seek(0, io.SeekCurrent)
	f.size, _ = io.Copy(io.MultiWriter(md5H, sha1H), f.Body)
	_, _ = f.Body.Seek(whence, io.SeekStart) // restore

	// calculate md5&sha1 hash
	f.md5 = md5H.Sum(nil)
	f.sha1 = sha1H.Sum(nil)
}

var fsWaiter = utils.NewUploadWaiter()

func (c *QQClient) UploadFile(target message.Source, file *LocalFile) error {
	switch target.SourceType {
	case message.SourceGroup, message.SourcePrivate: // ok
	default:
		return errors.New("not implemented")
	}

	file.init()
	// 同文件等待其他线程上传
	fkey := string(file.sha1)
	fsWaiter.Wait(fkey)
	defer fsWaiter.Done(fkey)

	var seq uint16
	var pkt []byte
	if target.SourceType == message.SourcePrivate {
		seq, pkt = c.buildPrivateFileUploadReqPacket(target, file)
	} else {
		seq, pkt = c.buildGroupFileUploadReqPacket(target.PrimaryID, file)
	}
	i, err := c.sendAndWait(seq, pkt)
	if err != nil {
		return errors.Wrap(err, "query upload failed")
	}
	rsp := i.(*fileUploadRsp)

	if !rsp.Existed {
		ext := &exciting.FileUploadExt{
			Unknown1: proto.Int32(100),
			Unknown2: proto.Int32(2),
			Entry: &exciting.FileUploadEntry{
				BusiBuff: &exciting.ExcitingBusiInfo{
					BusId:       proto.Int32(rsp.BusID),
					SenderUin:   proto.Some(c.Uin),
					ReceiverUin: proto.Some(target.PrimaryID),
					GroupCode:   proto.Int64(0),
				},
				FileEntry: &exciting.ExcitingFileEntry{
					FileSize:  proto.Some(file.size),
					Md5:       file.md5,
					Sha1:      file.sha1,
					FileId:    rsp.Uuid,
					UploadKey: rsp.UploadKey,
				},
				ClientInfo: &exciting.ExcitingClientInfo{
					ClientType:   proto.Int32(2),
					AppId:        proto.String(fmt.Sprint(c.version.AppId)),
					TerminalType: proto.Int32(2),
					ClientVer:    proto.String("d92615c5"),
					Unknown:      proto.Int32(4),
				},
				FileNameInfo: &exciting.ExcitingFileNameInfo{
					FileName: proto.Some(file.FileName),
				},
			},
			Unknown200: proto.Int32(1),
		}
		if target.SourceType == message.SourceGroup {
			if len(rsp.UploadIpLanV4) == 0 {
				return errors.New("server requires unsupported ftn upload")
			}
			ext.Unknown3 = proto.Int32(0)
			ext.Unknown200 = proto.None[int32]()
			ext.Entry.BusiBuff.GroupCode = proto.Int64(target.PrimaryID)
			ext.Entry.Host = &exciting.ExcitingHostConfig{
				Hosts: []*exciting.ExcitingHostInfo{
					{
						Url: &exciting.ExcitingUrlInfo{
							Unknown: proto.Int32(1),
							Host:    proto.Some(rsp.UploadIpLanV4[0]),
						},
						Port: proto.Some(rsp.UploadPort),
					},
				},
			}
		}
		extPkt, _ := proto.Marshal(ext)
		input := highway.Transaction{
			CommandID: 71,
			Body:      file.Body,
			Size:      file.size,
			Sum:       file.md5,
			Ticket:    c.highwaySession.SigSession,
			Ext:       extPkt,
		}
		if target.SourceType == message.SourcePrivate {
			input.CommandID = 69
		}
		if _, err := c.highwaySession.UploadExciting(input); err != nil {
			return errors.Wrap(err, "upload failed")
		}
	}
	if target.SourceType == message.SourceGroup {
		_, pkt := c.buildGroupFileFeedsRequest(target.PrimaryID, string(rsp.Uuid), rsp.BusID, rand.Int31())
		return c.sendPacket(pkt)
	}
	// 私聊文件
	_, pkt = c.buildPrivateFileUploadSuccReq(target, rsp)
	err = c.sendPacket(pkt)
	if err != nil {
		return err
	}
	uid := target.PrimaryID
	msgSeq := c.nextFriendSeq()
	content, _ := proto.Marshal(&msg.SubMsgType0X4Body{
		NotOnlineFile: &msg.NotOnlineFile{
			FileType: proto.Int32(0),
			FileUuid: rsp.Uuid,
			FileMd5:  file.md5,
			FileName: []byte(file.FileName),
			FileSize: proto.Int64(file.size),
			Subcmd:   proto.Int32(1),
		},
	})
	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{
			Trans_0X211: &msg.Trans0X211{
				ToUin: proto.Uint64(uint64(uid)),
				CcCmd: proto.Uint32(4),
			},
		},
		ContentHead: &msg.ContentHead{
			PkgNum:   proto.Int32(1),
			PkgIndex: proto.Int32(0),
			DivSeq:   proto.Int32(0),
		},
		MsgBody: &msg.MessageBody{
			MsgContent: content,
		},
		MsgSeq:     proto.Some(msgSeq),
		MsgRand:    proto.Some(int32(rand.Uint32())),
		SyncCookie: syncCookie(time.Now().Unix()),
	}
	payload, _ := proto.Marshal(req)
	_, p := c.uniPacket("MessageSvc.PbSendMsg", payload)
	return c.sendPacket(p)
}

func (c *QQClient) buildPrivateFileUploadReqPacket(target message.Source, file *LocalFile) (uint16, []byte) {
	req := cmd0x346.C346ReqBody{
		Cmd: 1700,
		Seq: c.nextFriendSeq(),
		ApplyUploadReqV3: &cmd0x346.ApplyUploadReqV3{
			SenderUin:     c.Uin,
			RecverUin:     target.PrimaryID,
			FileSize:      file.size,
			FileName:      file.FileName,
			Bytes_10MMd5:  file.md5, // TODO: investigate this
			Sha:           file.sha1,
			LocalFilepath: "/storage/emulated/0/Android/data/com.tencent.mobileqq/Tencent/QQfile_recv/" + file.FileName,
			Md5:           file.md5,
		},
		BusinessId:               3,
		ClientType:               104,
		FlagSupportMediaplatform: 1,
	}
	pkg, _ := proto.Marshal(&req)
	return c.uniPacket("OfflineFilleHandleSvr.pb_ftn_CMD_REQ_APPLY_UPLOAD_V3-1700", pkg)
}

// OfflineFilleHandleSvr.pb_ftn_CMD_REQ_APPLY_UPLOAD_V3-1700
func decodePrivateFileUploadReq(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	var rsp cmd0x346.C346RspBody
	err := proto.Unmarshal(payload, &rsp)
	if err != nil {
		return nil, err
	}
	v3 := rsp.ApplyUploadRspV3
	r := &fileUploadRsp{
		Existed:   v3.BoolFileExist,
		BusID:     3,
		Uuid:      v3.Uuid,
		UploadKey: v3.MediaPlateformUploadKey,
	}
	return r, nil
}

func (c *QQClient) buildPrivateFileUploadSuccReq(target message.Source, rsp *fileUploadRsp) (uint16, []byte) {
	req := &cmd0x346.C346ReqBody{
		Cmd: 800,
		Seq: 7,
		UploadSuccReq: &cmd0x346.UploadSuccReq{
			SenderUin: c.Uin,
			RecverUin: target.PrimaryID,
			Uuid:      rsp.Uuid,
		},
		BusinessId: 3,
		ClientType: 104,
	}
	pkt, _ := proto.Marshal(req)
	return c.uniPacket("OfflineFilleHandleSvr.pb_ftn_CMD_REQ_UPLOAD_SUCC-800", pkt)
}
