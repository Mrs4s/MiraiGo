package network

import (
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/auth"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/wrapper"
	"github.com/pkg/errors"
)

// Transport is a network transport.
type Transport struct {
	Sig     *auth.SigInfo
	Version *auth.AppVersion
	Device  *auth.Device

	// connection
	// conn *TCPClient
}

var WhiteListCommands = `
ConnAuthSvr.fast_qq_login
ConnAuthSvr.sdk_auth_api
ConnAuthSvr.sdk_auth_api_emp
FeedCloudSvr.trpc.feedcloud.commwriter.ComWriter.DoBarrage
FeedCloudSvr.trpc.feedcloud.commwriter.ComWriter.DoComment
FeedCloudSvr.trpc.feedcloud.commwriter.ComWriter.DoFollow
FeedCloudSvr.trpc.feedcloud.commwriter.ComWriter.DoLike
FeedCloudSvr.trpc.feedcloud.commwriter.ComWriter.DoPush
FeedCloudSvr.trpc.feedcloud.commwriter.ComWriter.DoReply
FeedCloudSvr.trpc.feedcloud.commwriter.ComWriter.PublishFeed
FeedCloudSvr.trpc.videocircle.circleprofile.CircleProfile.SetProfile
friendlist.addFriend
friendlist.AddFriendReq
friendlist.ModifyGroupInfoReq
MessageSvc.PbSendMsg
MsgProxy.SendMsg
OidbSvc.0x4ff_9
OidbSvc.0x4ff_9_IMCore
OidbSvc.0x56c_6
OidbSvc.0x6d9_4
OidbSvc.0x758
OidbSvc.0x758_0
OidbSvc.0x758_1
OidbSvc.0x88d_0
OidbSvc.0x89a_0
OidbSvc.0x89b_1
OidbSvc.0x8a1_0
OidbSvc.0x8a1_7
OidbSvc.0x8ba
OidbSvc.0x9fa
OidbSvc.oidb_0x758
OidbSvcTrpcTcp.0x101e_1
OidbSvcTrpcTcp.0x101e_2
OidbSvcTrpcTcp.0x1100_1
OidbSvcTrpcTcp.0x1105_1
OidbSvcTrpcTcp.0x1107_1
OidbSvcTrpcTcp.0x55f_0
OidbSvcTrpcTcp.0x6d9_4
OidbSvcTrpcTcp.0xf55_1
OidbSvcTrpcTcp.0xf57_1
OidbSvcTrpcTcp.0xf57_106
OidbSvcTrpcTcp.0xf57_9
OidbSvcTrpcTcp.0xf65_1
OidbSvcTrpcTcp.0xf65_10 
OidbSvcTrpcTcp.0xf67_1
OidbSvcTrpcTcp.0xf67_5
OidbSvcTrpcTcp.0xf6e_1
OidbSvcTrpcTcp.0xf88_1
OidbSvcTrpcTcp.0xf89_1
OidbSvcTrpcTcp.0xfa5_1
ProfileService.getGroupInfoReq
ProfileService.GroupMngReq
QChannelSvr.trpc.qchannel.commwriter.ComWriter.DoComment
QChannelSvr.trpc.qchannel.commwriter.ComWriter.DoReply
QChannelSvr.trpc.qchannel.commwriter.ComWriter.PublishFeed
qidianservice.135
qidianservice.207
qidianservice.269
qidianservice.290
SQQzoneSvc.addComment
SQQzoneSvc.addReply
SQQzoneSvc.forward
SQQzoneSvc.like
SQQzoneSvc.publishmood
SQQzoneSvc.shuoshuo
trpc.group_pro.msgproxy.sendmsg
trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginUnusualDevice
trpc.o3.ecdh_access.EcdhAccess.SsoEstablishShareKey
trpc.o3.ecdh_access.EcdhAccess.SsoSecureA2Access
trpc.o3.ecdh_access.EcdhAccess.SsoSecureA2Establish
trpc.o3.ecdh_access.EcdhAccess.SsoSecureAccess
trpc.o3.report.Report.SsoReport
trpc.passwd.manager.PasswdManager.SetPasswd
trpc.passwd.manager.PasswdManager.VerifyPasswd
trpc.qlive.relationchain_svr.RelationchainSvr.Follow
trpc.qlive.word_svr.WordSvr.NewPublicChat
trpc.qqhb.qqhb_proxy.Handler.sso_handle
trpc.springfestival.redpacket.LuckyBag.SsoSubmitGrade
wtlogin.device_lock
wtlogin.exchange_emp
wtlogin.login
wtlogin.name2uin
wtlogin.qrlogin
wtlogin.register
wtlogin.trans_emp
wtlogin_device.login
wtlogin_device.tran_sim_emp
`

func (t *Transport) packBody(req *Request, w *binary.Writer) {
	pos := w.FillUInt32()
	if req.Type == RequestTypeLogin {
		w.WriteUInt32(uint32(req.SequenceID))
		w.WriteUInt32(t.Version.AppId)
		w.WriteUInt32(t.Version.SubAppId)
		w.Write([]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00})
		tgt := t.Sig.TGT
		if len(tgt) == 0 || len(tgt) == 4 {
			w.WriteUInt32(0x04)
		} else {
			w.WriteUInt32(uint32(len(tgt) + 4))
			w.Write(tgt)
		}
	}
	w.WriteString(req.CommandName)
	w.WriteUInt32(uint32(len(t.Sig.OutPacketSessionID) + 4))
	w.Write(t.Sig.OutPacketSessionID)
	if req.Type == RequestTypeLogin {
		w.WriteString((*t.Device).IMEI)
		w.WriteUInt32(0x04)

		w.WriteUInt16(uint16(len(t.Sig.Ksid)) + 2)
		w.Write(t.Sig.Ksid)
	}
	if strings.Contains(WhiteListCommands, req.CommandName) {
		secSign := t.PackSecSign(req)
		w.WriteUInt32(uint32(len(secSign) + 4))
		w.Write(secSign)
	}

	w.WriteUInt32(0x04 + uint32(len(t.Device.QImei16)))
	w.Write([]byte(t.Device.QImei16))

	w.WriteUInt32At(pos, uint32(w.Len()-pos))

	w.WriteUInt32(uint32(len(req.Body) + 4))
	w.Write(req.Body)
}

func (t *Transport) PackSecSign(req *Request) []byte {
	if wrapper.FekitGetSign == nil {
		return []byte{}
	}
	sign, extra, token, err := wrapper.FekitGetSign(uint64(req.SequenceID), strconv.FormatInt(req.Uin, 10), req.CommandName, t.Version.QUA, req.Body)
	if err != nil {
		return []byte{}
	}
	m := &pb.SSOReserveField{
		Flag:        0,
		Qimei:       t.Device.QImei16,
		NewconnFlag: 0,
		Uid:         strconv.FormatInt(req.Uin, 10),
		Imsi:        0,
		NetworkType: 1,
		IpStackType: 1,
		MessageType: 0,
		SecInfo: &pb.SsoSecureInfo{
			SecSig:         sign,
			SecDeviceToken: token,
			SecExtra:       extra,
		},
		SsoIpOrigin: 0,
	}
	data, err := proto.Marshal(m)
	if err != nil {
		panic(errors.Wrap(err, "failed to unmarshal protobuf SSOReserveField"))
	}
	return data
}

// PackPacket packs a packet.
func (t *Transport) PackPacket(req *Request) []byte {
	// todo(wdvxdr): combine pack packet, send packet and return the response
	if len(t.Sig.D2) == 0 {
		req.EncryptType = EncryptTypeEmptyKey
	}

	w := binary.SelectWriter()
	defer binary.PutWriter(w)

	pos := w.FillUInt32()
	// vvv w.Write(head) vvv
	w.WriteUInt32(uint32(req.Type))
	w.WriteByte(byte(req.EncryptType))
	switch req.Type {
	case RequestTypeLogin:
		switch req.EncryptType {
		case EncryptTypeD2Key:
			w.WriteUInt32(uint32(len(t.Sig.D2) + 4))
			w.Write(t.Sig.D2)
		default:
			w.WriteUInt32(4)
		}
	case RequestTypeSimple:
		w.WriteUInt32(uint32(req.SequenceID))
	}
	w.WriteByte(0x00)
	w.WriteString(strconv.FormatInt(req.Uin, 10))
	// ^^^ w.Write(head) ^^^

	w2 := binary.SelectWriter()
	t.packBody(req, w2)
	body := w2.Bytes()
	// encrypt body
	switch req.EncryptType {
	case EncryptTypeD2Key:
		body = binary.NewTeaCipher(t.Sig.D2Key).Encrypt(body)
	case EncryptTypeEmptyKey:
		emptyKey := make([]byte, 16)
		body = binary.NewTeaCipher(emptyKey).Encrypt(body)
	}
	w.Write(body)
	binary.PutWriter(w2)

	w.WriteUInt32At(pos, uint32(w.Len()))
	return append([]byte(nil), w.Bytes()...)
}
