package client

import (
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/pkg/errors"
)

var (
	ErrAlreadyOnline  = errors.New("already online")
	ErrMemberNotFound = errors.New("member not found")
	ErrNotExists      = errors.New("not exists")
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
		Qid       string
	}

	OtherClientInfo struct {
		AppId      int64
		DeviceName string
		DeviceKind string
	}

	FriendListResponse struct {
		TotalCount int32
		List       []*FriendInfo
	}

	OtherClientStatusChangedEvent struct {
		Client *OtherClientInfo
		Online bool
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

	MemberCardUpdatedEvent struct {
		Group   *GroupInfo
		OldCard string
		Member  *GroupMemberInfo
	}

	INotifyEvent interface {
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

	OfflineFileEvent struct {
		FileName    string
		FileSize    int64
		Sender      int64
		DownloadUrl string
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

	AtAllRemainInfo struct {
		CanAtAll                 bool   `json:"can_at_all"`
		RemainAtAllCountForGroup uint32 `json:"remain_at_all_count_for_group"`
		RemainAtAllCountForUin   uint32 `json:"remain_at_all_count_for_uin"`
	}

	groupMemberListResponse struct {
		NextUin int64
		list    []*GroupMemberInfo
	}

	imageUploadResponse struct {
		UploadKey  []byte
		UploadIp   []int32
		UploadPort []int32
		ResourceId string
		Message    string
		FileId     int64
		Width      int32
		Height     int32
		ResultCode int32
		IsExists   bool
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
		FileId     int64
	}

	groupMessageReceiptEvent struct {
		Rand int32
		Seq  int32
		Msg  *message.GroupMessage
	}

	highwaySessionInfo struct {
		SigSession []byte
		SessionKey []byte
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
	IPad         ClientProtocol = 2
	AndroidWatch ClientProtocol = 3
	MacOS        ClientProtocol = 4
)

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
