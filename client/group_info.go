package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/Mrs4s/MiraiGo/internal/packets"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/client/pb/profilecard"
	"github.com/Mrs4s/MiraiGo/utils"
)

type (
	GroupInfo struct {
		Uin             int64
		Code            int64
		Name            string
		Memo            string
		OwnerUin        int64
		GroupCreateTime uint32
		GroupLevel      uint32
		MemberCount     uint16
		MaxMemberCount  uint16
		Members         []*GroupMemberInfo
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
		ShutUpTimestamp        int64
		Permission             MemberPermission
	}

	// GroupSearchInfo 通过搜索得到的群信息
	GroupSearchInfo struct {
		Code int64  // 群号
		Name string // 群名
		Memo string // 简介
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
func decodeGroupSearchResponse(_ *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
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
				Memo: g.GetBrief(),
			})
		}
		return ret, nil
	}
	return nil, nil
}

// OidbSvc.0x88d_0
func decodeGroupInfoResponse(c *QQClient, _ *incomingPacketInfo, payload []byte) (interface{}, error) {
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
		Uin:             int64(*info.GroupInfo.GroupUin),
		Code:            int64(*info.GroupCode),
		Name:            string(info.GroupInfo.GroupName),
		Memo:            string(info.GroupInfo.GroupMemo),
		GroupCreateTime: *info.GroupInfo.GroupCreateTime,
		GroupLevel:      *info.GroupInfo.GroupLevel,
		OwnerUin:        int64(*info.GroupInfo.GroupOwner),
		MemberCount:     uint16(*info.GroupInfo.GroupMemberNum),
		MaxMemberCount:  uint16(*info.GroupInfo.GroupMemberMaxNum),
		Members:         []*GroupMemberInfo{},
		LastMsgSeq:      int64(info.GroupInfo.GetGroupCurMsgSeq()),
		client:          c,
	}, nil
}

func (g *GroupInfo) UpdateName(newName string) error {
	if !g.AdministratorOrOwner() {
		return errors.New("not manageable")
	}
	if newName == "" || strings.Count(newName, "") > 20 {
		return errors.New("group name length should be between 1 and 20 unicode runes")
	}
	if err := g.client.updateGroupName(g.Code, newName); err != nil {
		return err
	}
	g.Name = newName
	return nil
}

func (g *GroupInfo) UpdateMemo(newMemo string) error {
	if !g.AdministratorOrOwner() {
		return errors.New("not manageable")
	}
	if err := g.client.updateGroupMemo(g.Code, newMemo); err != nil {
		return err
	}
	g.Memo = newMemo
	return nil
}

func (g *GroupInfo) UpdateGroupHeadPortrait(img []byte) error {
	if !g.AdministratorOrOwner() {
		return errors.New("not manageable")
	}
	return g.client.uploadGroupHeadPortrait(g.Uin, img)
}

func (g *GroupInfo) MuteAll(mute bool) error {
	if !g.AdministratorOrOwner() {
		return errors.New("not manageable")
	}
	return g.client.groupMuteAll(g.Code, mute)
}

func (g *GroupInfo) MuteAnonymous(id, nick string, seconds int32) error {
	token, err := g.client.getCSRFToken()
	if err != nil {
		return err
	}
	payload := fmt.Sprintf("anony_id=%v&group_code=%v&seconds=%v&anony_nick=%v&bkn=%v", url.QueryEscape(id), g.Code, seconds, nick, token)
	cookies, err := g.client.getCookies()
	if err != nil {
		return err
	}
	rsp, err := utils.HttpPostBytesWithCookie("https://qqweb.qq.com/c/anonymoustalk/blacklist", []byte(payload), cookies, "application/x-www-form-urlencoded")
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

func (g *GroupInfo) Quit() error {
	if g.SelfPermission() == Owner {
		return errors.Errorf("group owner cannot quit group")
	}
	return g.client.quitGroup(g.Code)
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
	i := sort.Search(len(g.Members), func(i int) bool {
		return g.Members[i].Uin >= uin
	})
	if i >= len(g.Members) || g.Members[i].Uin != uin {
		return nil
	}
	return g.Members[i]
}

// sort call this method must hold the lock
func (g *GroupInfo) sort() {
	sort.Slice(g.Members, func(i, j int) bool {
		return g.Members[i].Uin < g.Members[j].Uin
	})
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

func (m *GroupMemberInfo) EditCard(card string) error {
	if !m.CardChangable() {
		return errors.Errorf("card name not changable")
	}
	if len(card) > 60 {
		return errors.Errorf("card name length must not exceed 60 bytes")
	}
	if err := m.Group.client.editMemberCard(m.Group.Code, m.Uin, card); err != nil {
		return err
	}
	m.CardName = card
	return nil
}

func (m *GroupMemberInfo) Poke() error {
	return m.Group.client.SendGroupPoke(m.Group.Code, m.Uin)
}

func (m *GroupMemberInfo) SetAdmin(flag bool) error {
	if m.Group.OwnerUin == m.Group.client.Uin {
		return m.Group.client.setGroupAdmin(m.Group.Code, m.Uin, flag)
	}
	return errors.New("not manageable")
}

func (m *GroupMemberInfo) EditSpecialTitle(title string) error {
	if m.Group.SelfPermission() == Owner && len(title) <= 18 {
		if err := m.Group.client.editMemberSpecialTitle(m.Group.Code, m.Uin, title); err != nil {
			return err
		}
		m.SpecialTitle = title
	}
	return nil
}

func (m *GroupMemberInfo) Kick(msg string, block bool) error {
	if m.Uin != m.Group.client.Uin && m.Manageable() {
		return m.Group.client.kickGroupMember(m.Group.Code, m.Uin, msg, block)
	} else {
		return errors.New("not manageable")
	}
}

func (m *GroupMemberInfo) Mute(time uint32) error {
	if time >= 30*24*60*60 {
		return errors.New("time must be less than 30 days")
	}
	if m.Uin != m.Group.client.Uin && m.Manageable() {
		return m.Group.client.groupMute(m.Group.Code, m.Uin, time)
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
