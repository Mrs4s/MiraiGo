package client

import (
	"errors"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"strings"
	"sync"
)

var (
	ErrAlreadyOnline  = errors.New("already online")
	ErrMemberNotFound = errors.New("member not found")
)

type (
	LoginError int

	MemberPermission int

	ClientProtocol int

	LoginResponse struct {
		Success bool
		Error   LoginError

		// Captcha info
		CaptchaImage []byte
		CaptchaSign  []byte

		// Unsafe device
		VerifyUrl string

		// other error
		ErrorMessage string
	}

	FriendInfo struct {
		Uin      int64
		Nickname string
		Remark   string
		FaceId   int16

		//msgSeqList *utils.Cache
	}

	SummaryCardInfo struct {
		Uin       int64
		Sex       byte
		Age       uint8
		Nickname  string
		Level     int32
		City      string
		Sign      string
		Mobile    string
		LoginDays int64
	}

	FriendListResponse struct {
		TotalCount int32
		List       []*FriendInfo
	}

	GroupInfo struct {
		Uin            int64
		Code           int64
		Name           string
		Memo           string
		OwnerUin       int64
		MemberCount    uint16
		MaxMemberCount uint16
		Members        []*GroupMemberInfo

		client  *QQClient
		memLock sync.Mutex
	}

	GroupMemberInfo struct {
		Group                  *GroupInfo
		Uin                    int64
		Nickname               string
		CardName               string
		Level                  uint16
		JoinTime               int64
		LastSpeakTime          int64
		SpecialTitle           string
		SpecialTitleExpireTime int64
		Permission             MemberPermission
	}

	GroupMuteEvent struct {
		GroupCode   int64
		OperatorUin int64
		TargetUin   int64
		Time        int32
	}

	GroupMessageRecalledEvent struct {
		GroupCode   int64
		OperatorUin int64
		AuthorUin   int64
		MessageId   int32
		Time        int32
	}

	FriendMessageRecalledEvent struct {
		FriendUin int64
		MessageId int32
		Time      int64
	}

	GroupLeaveEvent struct {
		Group    *GroupInfo
		Operator *GroupMemberInfo
	}

	MemberJoinGroupEvent struct {
		Group  *GroupInfo
		Member *GroupMemberInfo
	}

	IGroupNotifyEvent interface {
		From() int64
		Content() string
	}

	MemberLeaveGroupEvent struct {
		Group    *GroupInfo
		Member   *GroupMemberInfo
		Operator *GroupMemberInfo
	}

	MemberPermissionChangedEvent struct {
		Group         *GroupInfo
		Member        *GroupMemberInfo
		OldPermission MemberPermission
		NewPermission MemberPermission
	}

	ClientDisconnectedEvent struct {
		Message string
	}

	GroupInvitedRequest struct {
		RequestId   int64
		InvitorUin  int64
		InvitorNick string
		GroupCode   int64
		GroupName   string

		client *QQClient
	}

	UserJoinGroupRequest struct {
		RequestId     int64
		Message       string
		RequesterUin  int64
		RequesterNick string
		GroupCode     int64
		GroupName     string

		client *QQClient
	}

	NewFriendRequest struct {
		RequestId     int64
		Message       string
		RequesterUin  int64
		RequesterNick string

		client *QQClient
	}

	LogEvent struct {
		Type    string
		Message string
	}

	ServerUpdatedEvent struct {
		Servers []jce.SsoServerInfo
	}

	NewFriendEvent struct {
		Friend *FriendInfo
	}

	OcrResponse struct {
		Texts    []*TextDetection `json:"texts"`
		Language string           `json:"language"`
	}

	TextDetection struct {
		Text        string        `json:"text"`
		Confidence  int32         `json:"confidence"`
		Coordinates []*Coordinate `json:"coordinates"`
	}

	Coordinate struct {
		X int32 `json:"x"`
		Y int32 `json:"y"`
	}

	groupMemberListResponse struct {
		NextUin int64
		list    []*GroupMemberInfo
	}

	imageUploadResponse struct {
		ResultCode int32
		Message    string

		IsExists bool
		FileId   int64
		Width    int32
		Height   int32

		ResourceId string
		UploadKey  []byte
		UploadIp   []int32
		UploadPort []int32
	}

	pttUploadResponse struct {
		ResultCode int32
		Message    string

		IsExists bool

		ResourceId string
		UploadKey  []byte
		UploadIp   []int32
		UploadPort []int32
		FileKey    []byte
	}

	groupMessageReceiptEvent struct {
		Rand int32
		Seq  int32
	}
)

const (
	NeedCaptcha       LoginError = 1
	OtherLoginError   LoginError = 3
	UnsafeDeviceError LoginError = 4
	UnknownLoginError LoginError = -1

	Owner MemberPermission = iota
	Administrator
	Member

	AndroidPhone ClientProtocol = 537062845
	AndroidPad   ClientProtocol = 537062409
	AndroidWatch ClientProtocol = 537061176
)

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

func (g *GroupInfo) Quit() {
	if g.SelfPermission() != Owner {
		g.client.quitGroup(g.Code)
	}
}

func (m *GroupMemberInfo) DisplayName() string {
	if m.CardName == "" {
		return m.Nickname
	}
	return m.CardName
}

func (m *GroupMemberInfo) EditCard(card string) {
	if m.Manageable() && len(card) <= 60 {
		m.Group.client.editMemberCard(m.Group.Code, m.Uin, card)
		m.CardName = card
	}
}

func (m *GroupMemberInfo) Poke() {
	m.Group.client.sendGroupPoke(m.Group.Code, m.Uin)
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

func (m *GroupMemberInfo) Kick(msg string) {
	if m.Uin != m.Group.client.Uin && m.Manageable() {
		m.Group.client.kickGroupMember(m.Group.Code, m.Uin, msg)
	}
}

func (m *GroupMemberInfo) Mute(time uint32) {
	if m.Uin != m.Group.client.Uin && m.Manageable() {
		if time < 2592000 {
			m.Group.client.groupMute(m.Group.Code, m.Uin, time)
		}
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

func (r *UserJoinGroupRequest) Accept() {
	r.client.SolveGroupJoinRequest(r, true, false, "")
}

func (r *UserJoinGroupRequest) Reject(block bool, reason string) {
	r.client.SolveGroupJoinRequest(r, false, block, reason)
}

func (r *GroupInvitedRequest) Accept() {
	r.client.SolveGroupJoinRequest(r, true, false, "")
}

func (r *GroupInvitedRequest) Reject(block bool, reason string) {
	r.client.SolveGroupJoinRequest(r, false, block, reason)
}

func (r *NewFriendRequest) Accept() {
	r.client.SolveFriendRequest(r, true)
}

func (r *NewFriendRequest) Reject() {
	r.client.SolveFriendRequest(r, false)
}
