package client

import (
	"errors"
	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"google.golang.org/protobuf/proto"
)

func (c *QQClient) GetGroupInfo(groupCode int64) (*GroupInfo, error) {
	i, err := c.sendAndWait(c.buildGroupInfoRequestPacket(groupCode))
	if err != nil {
		return nil, err
	}
	return i.(*GroupInfo), nil
}

// OidbSvc.0x88d_0
func (c *QQClient) buildGroupInfoRequestPacket(groupCode int64) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D88DReqBody{
		AppId: proto.Uint32(c.version.AppId),
		ReqGroupInfo: []*oidb.ReqGroupInfo{
			{
				GroupCode: proto.Uint64(uint64(groupCode)),
				Stgroupinfo: &oidb.D88DGroupInfo{
					GroupOwner:           proto.Uint64(0),
					GroupUin:             proto.Uint64(0),
					GroupCreateTime:      proto.Uint32(0),
					GroupFlag:            proto.Uint32(0),
					GroupMemberMaxNum:    proto.Uint32(0),
					GroupMemberNum:       proto.Uint32(0),
					GroupOption:          proto.Uint32(0),
					GroupLevel:           proto.Uint32(0),
					GroupFace:            proto.Uint32(0),
					GroupName:            EmptyBytes,
					GroupMemo:            EmptyBytes,
					GroupFingerMemo:      EmptyBytes,
					GroupLastMsgTime:     proto.Uint32(0),
					GroupQuestion:        EmptyBytes,
					GroupAnswer:          EmptyBytes,
					GroupGrade:           proto.Uint32(0),
					ActiveMemberNum:      proto.Uint32(0),
					HeadPortraitSeq:      proto.Uint32(0),
					MsgHeadPortrait:      &oidb.D88DGroupHeadPortrait{},
					StGroupExInfo:        &oidb.D88DGroupExInfoOnly{},
					GroupSecLevel:        proto.Uint32(0),
					CmduinPrivilege:      proto.Uint32(0),
					NoFingerOpenFlag:     proto.Uint32(0),
					NoCodeFingerOpenFlag: proto.Uint32(0),
				},
			},
		},
		PcClientVersion: proto.Uint32(0),
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:    2189,
		Bodybuffer: b,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x88d_0", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// SummaryCard.ReqSearch
func (c *QQClient) buildGroupSearchPacket(keyword string) (uint16, []byte) {
	seq := c.nextSeq()
	req := &jce.SummaryCardReqSearch{
		Keyword:     keyword,
		CountryCode: "+86",
		Version:     3,
		ReqServices: [][]byte{},
	}
	head := jce.NewJceWriter()
	head.WriteInt32(2, 0)
	buf := &jce.RequestDataVersion3{Map: map[string][]byte{
		"ReqHead":   packUniRequestData(head.Bytes()),
		"ReqSearch": packUniRequestData(req.ToBytes()),
	}}
	pkt := &jce.RequestPacket{
		IVersion:     3,
		SServantName: "SummaryCardServantObj",
		SFuncName:    "ReqSearch",
		SBuffer:      buf.ToBytes(),
		Context:      make(map[string]string),
		Status:       make(map[string]string),
	}
	packet := packets.BuildUniPacket(c.Uin, seq, "SummaryCard.ReqSearch", 1, c.OutGoingPacketSessionId, []byte{}, c.sigInfo.d2Key, pkt.ToBytes())
	return seq, packet
}

// SummaryCard.ReqSearch
func decodeGroupSearchResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	if len(data.Map["RespHead"]["SummaryCard.RespHead"]) > 20 {
		return nil, errors.New("not found")
	}
	rsp := data.Map["RespSearch"]["SummaryCard.RespSearch"][1:]
	r := jce.NewJceReader(rsp)
	rspService := r.ReadAny(2).([]interface{})[0].([]byte)
	sr := binary.NewReader(rspService)
	sr.ReadByte()
	ld1 := sr.ReadInt32()
	ld2 := sr.ReadInt32()
	if ld1 > 0 && ld2+9 < int32(len(rspService)) {
		sr.ReadBytes(int(ld1)) // busi comm
		//searchPb := sr.ReadBytes(int(ld2)) //TODO: search pb decode
	}
	return nil, nil
}

// OidbSvc.0x88d_0
func decodeGroupInfoResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D88DRspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, err
	}
	if len(rsp.RspGroupInfo) == 0 {
		return nil, errors.New(string(rsp.StrErrorInfo))
	}
	info := rsp.RspGroupInfo[0]
	if info.GroupInfo == nil {
		return nil, errors.New("group info not found")
	}
	return &GroupInfo{
		Uin:            int64(*info.GroupInfo.GroupUin),
		Code:           int64(*info.GroupCode),
		Name:           string(info.GroupInfo.GroupName),
		Memo:           string(info.GroupInfo.GroupMemo),
		OwnerUin:       int64(*info.GroupInfo.GroupOwner),
		MemberCount:    uint16(*info.GroupInfo.GroupMemberNum),
		MaxMemberCount: uint16(*info.GroupInfo.GroupMemberMaxNum),
		Members:        []*GroupMemberInfo{},
		client:         c,
	}, nil
}
