package client

import (
	"errors"
	"sync"
)

var (
	ErrAlreadyRunning = errors.New("already running")
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

		memLock *sync.Mutex
	}

	GroupMemberInfo struct {
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

	GroupInvitedEvent struct {
		EventId     int64
		InvitorUin  int64
		InvitorNick string
		GroupCode   int64
		GroupName   string
	}

	UserJoinGroupRequest struct {
		RequestId     int64
		Message       string
		RequesterUin  int64
		RequesterNick string
		GroupCode     int64
		GroupName     string
	}

	NewFriendRequest struct {
		RequestId     int64
		Message       string
		RequesterUin  int64
		RequesterNick string
	}

	groupMemberListResponse struct {
		NextUin int64
		list    []*GroupMemberInfo
	}

	groupImageUploadResponse struct {
		ResultCode int32
		Message    string

		IsExists bool

		UploadKey  []byte
		UploadIp   []int32
		UploadPort []int32
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

func (m *GroupMemberInfo) DisplayName() string {
	if m.CardName == "" {
		return m.Nickname
	}
	return m.CardName
}
