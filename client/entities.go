package client

import (
	"errors"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/message"
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

		// SMS needed
		SMSPhone string

		// other error
		ErrorMessage string
	}

	FriendInfo struct {
		Uin      int64  `json:"uin"`
		Nickname string `json:"nickname"`
		Remark   string `json:"remark"`
		FaceId   int16  `json:"face_id"`

		//msgSeqList *utils.Cache
	}

	SummaryCardInfo struct {
		Uin       int64  `json:"uin"`
		Sex       byte   `json:"sex"`
		Age       uint8  `json:"age"`
		Nickname  string `json:"nickname"`
		Level     int32  `json:"level"`
		City      string `json:"city"`
		Sign      string `json:"sign"`
		Mobile    string `json:"mobile"`
		LoginDays int64  `json:"login_days"`
	}

	FriendListResponse struct {
		TotalCount int32
		List       []*FriendInfo
	}

	GroupInfo struct {
		Uin            int64              `json:"uin"`
		Code           int64              `json:"code"`
		Name           string             `json:"name"`
		Memo           string             `json:"memo"`
		OwnerUin       int64              `json:"owner_uin"`
		MemberCount    uint16             `json:"member_count"`
		MaxMemberCount uint16             `json:"max_member_count"`
		Members        []*GroupMemberInfo `json:"members"`

		client *QQClient
		lock   sync.RWMutex
	}

	GroupMemberInfo struct {
		Group                  *GroupInfo       `json:"group"`
		Uin                    int64            `json:"uin"`
		Gender                 byte             `json:"gender"`
		Nickname               string           `json:"nickname"`
		CardName               string           `json:"card_name"`
		Level                  uint16           `json:"level"`
		JoinTime               int64            `json:"join_time"`
		LastSpeakTime          int64            `json:"last_speak_time"`
		SpecialTitle           string           `json:"special_title"`
		SpecialTitleExpireTime int64            `json:"special_title_expire_time"`
		Permission             MemberPermission `json:"permission"`
	}

	GroupMuteEvent struct {
		GroupCode   int64 `json:"group_code"`
		OperatorUin int64 `json:"operator_uin"`
		TargetUin   int64 `json:"target_uin"`
		Time        int32 `json:"time"`
	}

	GroupMessageRecalledEvent struct {
		GroupCode   int64 `json:"group_code"`
		OperatorUin int64 `json:"operator_uin"`
		AuthorUin   int64 `json:"author_uin"`
		MessageId   int32 `json:"message_id"`
		Time        int32 `json:"time"`
	}

	FriendMessageRecalledEvent struct {
		FriendUin int64 `json:"friend_uin"`
		MessageId int32 `json:"message_id"`
		Time      int64 `json:"time"`
	}

	GroupLeaveEvent struct {
		Group    *GroupInfo       `json:"group"`
		Operator *GroupMemberInfo `json:"operator"`
	}

	MemberJoinGroupEvent struct {
		Group  *GroupInfo       `json:"group"`
		Member *GroupMemberInfo `json:"member"`
	}

	MemberCardUpdatedEvent struct {
		Group   *GroupInfo       `json:"group"`
		OldCard string           `json:"old_card"`
		Member  *GroupMemberInfo `json:"member"`
	}

	IGroupNotifyEvent interface {
		From() int64
		Content() string
	}

	MemberLeaveGroupEvent struct {
		Group    *GroupInfo       `json:"group"`
		Member   *GroupMemberInfo `json:"member"`
		Operator *GroupMemberInfo `json:"operator"`
	}

	MemberPermissionChangedEvent struct {
		Group         *GroupInfo       `json:"group"`
		Member        *GroupMemberInfo `json:"member"`
		OldPermission MemberPermission `json:"old_permission"`
		NewPermission MemberPermission `json:"new_permission"`
	}

	ClientDisconnectedEvent struct {
		Message string `json:"message"`
	}

	GroupInvitedRequest struct {
		RequestId   int64  `json:"request_id"`
		InvitorUin  int64  `json:"invitor_uin"`
		InvitorNick string `json:"invitor_nick"`
		GroupCode   int64  `json:"group_code"`
		GroupName   string `json:"group_name"`

		client *QQClient
	}

	UserJoinGroupRequest struct {
		RequestId     int64  `json:"request_id"`
		Message       string `json:"message"`
		RequesterUin  int64  `json:"requester_uin"`
		RequesterNick string `json:"requester_nick"`
		GroupCode     int64  `json:"group_code"`
		GroupName     string `json:"group_name"`

		client *QQClient
	}

	NewFriendRequest struct {
		RequestId     int64  `json:"request_id"`
		Message       string `json:"message"`
		RequesterUin  int64  `json:"requester_uin"`
		RequesterNick string `json:"requester_nick"`

		client *QQClient
	}

	LogEvent struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	ServerUpdatedEvent struct {
		Servers []jce.SsoServerInfo
	}

	NewFriendEvent struct {
		Friend *FriendInfo `json:"friend"`
	}

	OfflineFileEvent struct {
		FileName    string `json:"file_name"`
		FileSize    int64  `json:"file_size"`
		Sender      int64  `json:"sender"`
		DownloadUrl string `json:"download_url"`
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
		UploadIp   []string
		UploadPort []int32
		FileKey    []byte
	}

	groupMessageReceiptEvent struct {
		Rand int32
		Seq  int32
		Msg  *message.GroupMessage
	}
)

const (
	NeedCaptcha            LoginError = 1
	OtherLoginError        LoginError = 3
	UnsafeDeviceError      LoginError = 4
	SMSNeededError         LoginError = 5
	TooManySMSRequestError LoginError = 6
	SMSOrVerifyNeededError LoginError = 7
	SliderNeededError      LoginError = 8
	UnknownLoginError      LoginError = -1

	Owner MemberPermission = iota
	Administrator
	Member

	AndroidPhone ClientProtocol = 1
	AndroidPad   ClientProtocol = 2
	AndroidWatch ClientProtocol = 3
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

func (g *GroupInfo) SelfPermission() MemberPermission {
	return g.FindMember(g.client.Uin).Permission
}

func (g *GroupInfo) AdministratorOrOwner() bool {
	return g.SelfPermission() == Administrator || g.SelfPermission() == Owner
}

func (g *GroupInfo) FindMember(uin int64) *GroupMemberInfo {
	r := g.Read(func(info *GroupInfo) interface{} {
		for _, m := range info.Members {
			f := m
			if f.Uin == uin {
				return f
			}
		}
		return nil
	})
	if r == nil {
		return nil
	}
	return r.(*GroupMemberInfo)
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
