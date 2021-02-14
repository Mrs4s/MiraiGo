package client

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/client/pb/profilecard"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type (
	GroupInfo struct {
		Uin            int64
		Code           int64
		Name           string
		Memo           string
		OwnerUin       int64
		MemberCount    uint16
		MaxMemberCount uint16
		Members        []*GroupMemberInfo
		// 最后一条信息的SEQ,只有通过 GetGroupInfo 函数获取的 GroupInfo 才会有
		LastMsgSeq int64

		client *QQClient

		lock sync.RWMutex
	}

	GroupMemberInfo struct {
		Group                  *GroupInfo
		Uin                    int64
		Gender                 byte
		Nickname               string
		CardName               string
		Level                  uint16
		JoinTime               int64
		LastSpeakTime          int64
		SpecialTitle           string
		SpecialTitleExpireTime int64
		Permission             MemberPermission
	}

	// GroupSearchInfo 通过搜索得到的群信息
	GroupSearchInfo struct {
		Code int64  // 群号
		Name string // 群名
	}
)

func init() {
	decoders["SummaryCard.ReqSearch"] = decodeGroupSearchResponse
	decoders["OidbSvc.0x88d_0"] = decodeGroupInfoResponse
}

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
					GroupCurMsgSeq:       proto.Uint32(0),
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

// SearchGroupByKeyword 通过关键词搜索陌生群组
func (c *QQClient) SearchGroupByKeyword(keyword string) ([]GroupSearchInfo, error) {
	rsp, err := c.sendAndWait(c.buildGroupSearchPacket(keyword))
	if err != nil {
		return nil, errors.Wrap(err, "group search failed")
	}
	return rsp.([]GroupSearchInfo), nil
}

// SummaryCard.ReqSearch
func (c *QQClient) buildGroupSearchPacket(keyword string) (uint16, []byte) {
	seq := c.nextSeq()
	comm, _ := proto.Marshal(&profilecard.BusiComm{
		Ver:      proto.Int32(1),
		Seq:      proto.Int32(rand.Int31()),
		Service:  proto.Int32(80000001),
		Platform: proto.Int32(2),
		Qqver:    proto.String("8.5.0.5025"),
		Build:    proto.Int32(5025),
	})
	search, _ := proto.Marshal(&profilecard.AccountSearch{
		Start:     proto.Int32(0),
		End:       proto.Uint32(4),
		Keyword:   &keyword,
		Highlight: []string{keyword},
		UserLocation: &profilecard.Location{
			Latitude:  proto.Float64(0),
			Longitude: proto.Float64(0),
		},
		Filtertype: proto.Int32(0),
	})
	req := &jce.SummaryCardReqSearch{
		Keyword:     keyword,
		CountryCode: "+86",
		Version:     3,
		ReqServices: [][]byte{
			binary.NewWriterF(func(w *binary.Writer) {
				w.WriteByte(0x28)
				w.WriteUInt32(uint32(len(comm)))
				w.WriteUInt32(uint32(len(search)))
				w.Write(comm)
				w.Write(search)
				w.WriteByte(0x29)
			}),
		},
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
func decodeGroupSearchResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
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
		searchPb := sr.ReadBytes(int(ld2))
		searchRsp := profilecard.AccountSearch{}
		err := proto.Unmarshal(searchPb, &searchRsp)
		if err != nil {
			return nil, errors.Wrap(err, "get search result failed")
		}
		var ret []GroupSearchInfo
		for _, g := range searchRsp.GetList() {
			ret = append(ret, GroupSearchInfo{
				Code: int64(g.GetCode()),
				Name: g.GetName(),
			})
		}
		return ret, nil
	}
	return nil, nil
}

// OidbSvc.0x88d_0
func decodeGroupInfoResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D88DRspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
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
		LastMsgSeq:     int64(info.GroupInfo.GetGroupCurMsgSeq()),
		client:         c,
	}, nil
}

func (g *GroupInfo) UpdateName(newName string) {
	if g.AdministratorOrOwner() && newName != "" && strings.Count(newName, "") <= 20 {
		g.client.updateGroupName(g.Code, newName)
		g.Name = newName
	}
}

func (g *GroupInfo) UpdateMemo(newMemo string) {
	if g.AdministratorOrOwner() {
		g.client.updateGroupMemo(g.Code, newMemo)
		g.Memo = newMemo
	}
}

func (g *GroupInfo) UpdateGroupHeadPortrait(img []byte) {
	if g.AdministratorOrOwner() {
		_ = g.client.uploadGroupHeadPortrait(g.Uin, img)
	}
}

func (g *GroupInfo) MuteAll(mute bool) {
	if g.AdministratorOrOwner() {
		g.client.groupMuteAll(g.Code, mute)
	}
}

func (g *GroupInfo) MuteAnonymous(id, nick string, seconds int32) error {
	payload := fmt.Sprintf("anony_id=%v&group_code=%v&seconds=%v&anony_nick=%v&bkn=%v", url.QueryEscape(id), g.Code, seconds, nick, g.client.getCSRFToken())
	rsp, err := utils.HttpPostBytesWithCookie("https://qqweb.qq.com/c/anonymoustalk/blacklist", []byte(payload), g.client.getCookies(), "application/x-www-form-urlencoded")
	if err != nil {
		return errors.Wrap(err, "failed to request blacklist")
	}
	var muteResp struct {
		RetCode int `json:"retcode"`
		CGICode int `json:"cgicode"`
	}
	err = json.Unmarshal(rsp, &muteResp)
	if err != nil {
		return errors.Wrap(err, "failed to parse muteResp")
	}
	if muteResp.RetCode != 0 {
		return errors.Errorf("retcode %v", muteResp.RetCode)
	}
	if muteResp.CGICode != 0 {
		return errors.Errorf("retcode %v", muteResp.CGICode)
	}
	return nil
}

func (g *GroupInfo) Quit() {
	if g.SelfPermission() != Owner {
		g.client.quitGroup(g.Code)
	}
}

func (g *GroupInfo) SelfPermission() MemberPermission {
	return g.FindMember(g.client.Uin).Permission
}

func (g *GroupInfo) AdministratorOrOwner() bool {
	return g.SelfPermission() == Administrator || g.SelfPermission() == Owner
}

func (g *GroupInfo) FindMember(uin int64) *GroupMemberInfo {
	r := g.Read(func(info *GroupInfo) interface{} {
		return info.FindMemberWithoutLock(uin)
	})
	if r == nil {
		return nil
	}
	return r.(*GroupMemberInfo)
}

func (g *GroupInfo) FindMemberWithoutLock(uin int64) *GroupMemberInfo {
	for _, m := range g.Members {
		f := m
		if f.Uin == uin {
			return f
		}
	}
	return nil
}

func (g *GroupInfo) Update(f func(*GroupInfo)) {
	g.lock.Lock()
	defer g.lock.Unlock()
	f(g)
}

func (g *GroupInfo) Read(f func(*GroupInfo) interface{}) interface{} {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return f(g)
}

func (m *GroupMemberInfo) DisplayName() string {
	if m.CardName == "" {
		return m.Nickname
	}
	return m.CardName
}

func (m *GroupMemberInfo) EditCard(card string) {
	if m.CardChangable() && len(card) <= 60 {
		m.Group.client.editMemberCard(m.Group.Code, m.Uin, card)
		m.CardName = card
	}
}

func (m *GroupMemberInfo) Poke() {
	m.Group.client.SendGroupPoke(m.Group.Code, m.Uin)
}

func (m *GroupMemberInfo) SetAdmin(flag bool) {
	if m.Group.OwnerUin == m.Group.client.Uin {
		m.Group.client.setGroupAdmin(m.Group.Code, m.Uin, flag)
	}
}

func (m *GroupMemberInfo) EditSpecialTitle(title string) {
	if m.Group.SelfPermission() == Owner && len(title) <= 18 {
		m.Group.client.editMemberSpecialTitle(m.Group.Code, m.Uin, title)
		m.SpecialTitle = title
	}
}

func (m *GroupMemberInfo) Kick(msg string, block bool) error {
	if m.Uin != m.Group.client.Uin && m.Manageable() {
		m.Group.client.kickGroupMember(m.Group.Code, m.Uin, msg, block)
		return nil
	} else {
		return errors.New("not manageable")
	}
}

func (m *GroupMemberInfo) Mute(time uint32) error {
	if time >= 2592000 {
		return errors.New("time is not in range")
	}
	if m.Uin != m.Group.client.Uin && m.Manageable() {
		m.Group.client.groupMute(m.Group.Code, m.Uin, time)
		return nil
	} else {
		return errors.New("not manageable")
	}
}

func (m *GroupMemberInfo) Manageable() bool {
	if m.Uin == m.Group.client.Uin {
		return true
	}
	self := m.Group.SelfPermission()
	if self == Member || m.Permission == Owner {
		return false
	}
	return m.Permission != Administrator || self == Owner
}

func (m *GroupMemberInfo) CardChangable() bool {
	if m.Uin == m.Group.client.Uin {
		return true
	}
	self := m.Group.SelfPermission()
	if self == Member {
		return false
	}
	return m.Permission != Owner
}
