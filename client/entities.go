package client

import "errors"

var (
	ErrAlreadyRunning = errors.New("already running")
)

type (
	LoginError    int
	LoginResponse struct {
		Success bool
		Error   LoginError

		// Captcha info
		CaptchaImage []byte
		CaptchaSign  []byte

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
		OwnerUin       uint32
		MemberCount    uint16
		MaxMemberCount uint16
		Members        []GroupMemberInfo
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
		Job                    string
	}

	GroupMuteEvent struct {
		GroupUin    int64
		OperatorUin int64
		TargetUin   int64
		Time        int32
	}

	GroupMessageRecalledEvent struct {
		GroupUin    int64
		OperatorUin int64
		AuthorUin   int64
		MessageId   int32
		Time        int32
	}

	groupMemberListResponse struct {
		NextUin int64
		list    []GroupMemberInfo
	}

	groupImageUploadResponse struct {
		ResultCode int32
		Message    string

		IsExists bool

		UploadKey  []byte
		UploadIp   []int32
		UploadPort []int32
	}
)

const (
	NeedCaptcha       LoginError = 1
	DeviceLockError              = 2
	OtherLoginError              = 3
	UnknownLoginError            = -1
)
