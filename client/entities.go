package client

import (
	"errors"
	"strings"
	"sync"
)

var (
	ErrAlreadyOnline = errors.New("already online")
)

type (
	LoginError int

	MemberPermission int

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
		memLock *sync.Mutex
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

	NewFriendEvent struct {
		Friend *FriendInfo
	}

	groupMemberListResponse struct {
		NextUin int64
		list    []*GroupMemberInfo
	}

	imageUploadResponse struct {
		ResultCode int32
		Message    string

		IsExists bool

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
	OtherLoginError              = 3
	UnsafeDeviceError            = 4
	UnknownLoginError            = -1

	Owner MemberPermission = iota
	Administrator
	Member
)

func (g *GroupInfo) UpdateName(newName string) {
	if g.AdministratorOrOwner() && newName != "" && strings.Count(newName, "") <= 20 {
		g.client.updateGroupName(g.Code, newName)
		g.Name = newName
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
		g.client.dispatchLeaveGroupEvent(&GroupLeaveEvent{Group: g})
	}
}

func (m *GroupMemberInfo) DisplayName() string {
	if m.CardName == "" {
		return m.Nickname
	}
	return m.CardName
}

func (m *GroupMemberInfo) EditCard(card string) {
	if m.Manageable() && strings.Count(card, "") <= 20 {
		m.Group.client.editMemberCard(m.Group.Code, m.Uin, card)
		m.CardName = card
	}
}

func (m *GroupMemberInfo) EditSpecialTitle(title string) {
	if m.Group.SelfPermission() == Owner && strings.Count(title, "") <= 6 {
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
	return m.Permission != Administrator
}

func (r *UserJoinGroupRequest) Accept() {
	r.client.SolveGroupJoinRequest(r, true)
}

func (r *UserJoinGroupRequest) Reject() {
	r.client.SolveGroupJoinRequest(r, false)
}

func (r *GroupInvitedRequest) Accept() {
	r.client.SolveGroupJoinRequest(r, true)
}

func (r *GroupInvitedRequest) Reject() {
	r.client.SolveGroupJoinRequest(r, false)
}

func (r *NewFriendRequest) Accept() {
	r.client.SolveFriendRequest(r, true)
}

func (r *NewFriendRequest) Reject() {
	r.client.SolveFriendRequest(r, false)
}
