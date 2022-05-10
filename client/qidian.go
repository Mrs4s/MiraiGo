// 企点协议相关特殊逻辑

package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x3f6"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x6ff"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

func init() {
	decoders["qidianservice.69"] = decodeLoginExtraResponse
	decoders["HttpConn.0x6ff_501"] = decodeConnKeyResponse
}

// getQiDianAddressDetailList 外部联系人列表
func (c *QQClient) getQiDianAddressDetailList() ([]*FriendInfo, error) {
	req := &cmd0x6ff.C519ReqBody{
		SubCmd: proto.Uint32(33),
		CrmCommonHead: &cmd0x6ff.C519CRMMsgHead{
			KfUin:     proto.Uint64(uint64(c.QiDian.MasterUin)),
			VerNo:     proto.Uint32(uint32(utils.ConvertSubVersionToInt(c.version.SortVersionName))),
			CrmSubCmd: proto.Uint32(33),
			LaborUin:  proto.Uint64(uint64(c.Uin)),
		},
		GetAddressDetailListReqBody: &cmd0x6ff.GetAddressDetailListReqBody{
			Timestamp2: proto.Uint64(0),
		},
	}
	rspData, err := c.bigDataRequest(0x519, req)
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	rsp := &cmd0x6ff.C519RspBody{}
	if err = proto.Unmarshal(rspData, rsp); err != nil {
		return nil, errors.Wrap(err, "unmarshal error")
	}
	if rsp.GetAddressDetailListRspBody == nil {
		return nil, errors.New("rsp body is nil")
	}
	ret := []*FriendInfo{}
	for _, detail := range rsp.GetAddressDetailListRspBody.AddressDetail {
		if len(detail.Qq) == 0 {
			c.warning("address detail %v QQ is 0", string(detail.Name))
			continue
		}
		ret = append(ret, &FriendInfo{
			Uin:      int64(detail.Qq[0].GetAccount()),
			Nickname: string(detail.Name),
		})
	}
	return ret, nil
}

func (c *QQClient) buildLoginExtraPacket() (uint16, []byte) {
	req := &cmd0x3f6.C3F6ReqBody{
		SubCmd: proto.Uint32(69),
		CrmCommonHead: &cmd0x3f6.C3F6CRMMsgHead{
			CrmSubCmd:  proto.Uint32(69),
			VerNo:      proto.Uint32(uint32(utils.ConvertSubVersionToInt(c.version.SortVersionName))),
			Clienttype: proto.Uint32(2),
		},
		SubcmdLoginProcessCompleteReqBody: &cmd0x3f6.QDUserLoginProcessCompleteReqBody{
			Kfext:        proto.Uint64(uint64(c.Uin)),
			Pubno:        &c.version.AppId,
			Buildno:      proto.Uint32(uint32(utils.ConvertSubVersionToInt(c.version.SortVersionName))),
			TerminalType: proto.Uint32(2),
			Status:       proto.Uint32(10),
			LoginTime:    proto.Uint32(5),
			HardwareInfo: proto.String(string(c.deviceInfo.Model)),
			SoftwareInfo: proto.String(string(c.deviceInfo.Version.Release)),
			Guid:         c.deviceInfo.Guid,
			AppName:      &c.version.ApkId,
			SubAppId:     &c.version.AppId,
		},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("qidianservice.69", payload)
}

func (c *QQClient) buildConnKeyRequestPacket() (uint16, []byte) {
	req := &cmd0x6ff.C501ReqBody{
		ReqBody: &cmd0x6ff.SubCmd0X501ReqBody{
			Uin:          proto.Uint64(uint64(c.Uin)),
			IdcId:        proto.Uint32(0),
			Appid:        proto.Uint32(16),
			LoginSigType: proto.Uint32(1),
			RequestFlag:  proto.Uint32(3),
			ServiceTypes: []uint32{1},
		},
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("HttpConn.0x6ff_501", payload)
}

func (c *QQClient) bigDataRequest(subCmd uint32, req proto.Message) ([]byte, error) {
	if c.QiDian.bigDataReqSession == nil {
		return nil, errors.New("please call conn key request method before")
	}
	data, _ := proto.Marshal(req)
	head, _ := proto.Marshal(&msg.IMHead{
		HeadType: proto.Uint32(4),
		HttpconnHead: &msg.HttpConnHead{
			Uin:          proto.Uint64(uint64(c.Uin)),
			Command:      proto.Uint32(1791),
			SubCommand:   &subCmd,
			Seq:          proto.Uint32(uint32(c.nextHighwayApplySeq())),
			Version:      proto.Uint32(500), // todo: short version convert
			Flag:         proto.Uint32(1),
			CompressType: proto.Uint32(0),
			ErrorCode:    proto.Uint32(0),
		},
		LoginSig: &msg.LoginSig{
			Type: proto.Uint32(22),
			Sig:  c.QiDian.bigDataReqSession.SigSession,
		},
	})
	tea := binary.NewTeaCipher(c.QiDian.bigDataReqSession.SessionKey)
	body := tea.Encrypt(data)
	url := fmt.Sprintf("http://%v/cgi-bin/httpconn", c.QiDian.bigDataReqAddrs[0])
	httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(binary.NewWriterF(func(w *binary.Writer) {
		w.WriteByte(40)
		w.WriteUInt32(uint32(len(head)))
		w.WriteUInt32(uint32(len(body)))
		w.Write(head)
		w.Write(body)
		w.WriteByte(41)
	})))
	rsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	defer func() { _ = rsp.Body.Close() }()
	rspBody, _ := io.ReadAll(rsp.Body)
	if len(rspBody) == 0 {
		return nil, errors.Wrap(err, "request error")
	}
	r := binary.NewReader(rspBody)
	r.ReadByte()
	l1 := int(r.ReadInt32())
	l2 := int(r.ReadInt32())
	r.ReadBytes(l1)
	payload := r.ReadBytes(l2)
	return tea.Decrypt(payload), nil
}

func decodeLoginExtraResponse(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := cmd0x3f6.C3F6RspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.SubcmdLoginProcessCompleteRspBody == nil {
		return nil, errors.New("login process resp is nil")
	}
	c.QiDian = &QiDianAccountInfo{
		MasterUin:  int64(rsp.SubcmdLoginProcessCompleteRspBody.GetCorpuin()),
		ExtName:    rsp.SubcmdLoginProcessCompleteRspBody.GetExtuinName(),
		CreateTime: int64(rsp.SubcmdLoginProcessCompleteRspBody.GetOpenAccountTime()),
	}
	return nil, nil
}

func decodeConnKeyResponse(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := cmd0x6ff.C501RspBody{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if c.QiDian == nil {
		return nil, errors.New("please call login extra before")
	}
	c.QiDian.bigDataReqSession = &bigDataSessionInfo{
		SigSession: rsp.RspBody.SigSession,
		SessionKey: rsp.RspBody.SessionKey,
	}
	for _, srv := range rsp.RspBody.Addrs {
		if srv.GetServiceType() == 1 {
			for _, addr := range srv.Addrs {
				c.QiDian.bigDataReqAddrs = append(c.QiDian.bigDataReqAddrs, fmt.Sprintf("%v:%v", binary.UInt32ToIPV4Address(addr.GetIp()), addr.GetPort()))
			}
		}
	}
	return nil, nil
}
