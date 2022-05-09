package auth

//go:generate stringer -type=Protocol -linecomment
type Protocol int

const (
	Unset        Protocol = iota
	AndroidPhone          // Android Phone
	AndroidWatch          // Android Watch
	MacOS                 // MacOS
	QiDian                // 企点
	IPad                  // iPad
)

type AppVersion struct {
	ApkSign         []byte
	ApkId           string
	SortVersionName string
	SdkVersion      string
	AppId           uint32
	SubAppId        uint32
	BuildTime       uint32
	SSOVersion      uint32
	MiscBitmap      uint32
	SubSigmap       uint32
	MainSigMap      uint32
	Protocol        Protocol
}

func (i Protocol) Version() *AppVersion {
	switch i {
	case AndroidPhone: // Dumped by mirai from qq android v8.8.38
		return &AppVersion{
			ApkId:           "com.tencent.mobileqq",
			AppId:           537100432,
			SubAppId:        537100432,
			SortVersionName: "8.8.38",
			BuildTime:       1634310940,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.2487",
			SSOVersion:      16,
			MiscBitmap:      184024956,
			SubSigmap:       0x10400,
			MainSigMap:      34869472,
			Protocol:        i,
		}
	case AndroidWatch:
		return &AppVersion{
			ApkId:           "com.tencent.qqlite",
			AppId:           537064446,
			SubAppId:        537064446,
			SortVersionName: "2.0.5",
			BuildTime:       1559564731,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.236",
			SSOVersion:      5,
			MiscBitmap:      16252796,
			SubSigmap:       0x10400,
			MainSigMap:      34869472,
			Protocol:        i,
		}
	case IPad:
		return &AppVersion{
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537097188,
			SubAppId:        537097188,
			SortVersionName: "8.8.35",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      12,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      1970400,
			Protocol:        i,
		}
	case MacOS:
		return &AppVersion{
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537064315,
			SubAppId:        537064315,
			SortVersionName: "5.8.9",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      12,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      1970400,
			Protocol:        i,
		}
	case QiDian:
		return &AppVersion{
			ApkId:           "com.tencent.qidian",
			AppId:           537096038,
			SubAppId:        537036590,
			SortVersionName: "5.0.0",
			BuildTime:       1630062176,
			ApkSign:         []byte{160, 30, 236, 171, 133, 233, 227, 186, 43, 15, 106, 21, 140, 133, 92, 41},
			SdkVersion:      "6.0.0.2484",
			SSOVersion:      18,
			MiscBitmap:      184024956,
			SubSigmap:       66560,
			MainSigMap:      34869472,
			Protocol:        i,
		}
	}
	return nil
}

type SigInfo struct {
	LoginBitmap uint64
	TGT         []byte
	TGTKey      []byte

	SrmToken        []byte // study room manager | 0x16a
	T133            []byte
	EncryptedA1     []byte
	UserStKey       []byte
	UserStWebSig    []byte
	SKey            []byte
	SKeyExpiredTime int64
	D2              []byte
	D2Key           []byte
	DeviceToken     []byte

	PsKeyMap    map[string][]byte
	Pt4TokenMap map[string][]byte

	// Others
	OutPacketSessionID []byte
	Dpwd               []byte

	// tlv cache
	T104     []byte
	T174     []byte
	G        []byte
	T402     []byte
	RandSeed []byte // t403
	// rollbackSig []byte
	// t149        []byte
	// t150        []byte
	// t528        []byte
	// t530        []byte

	// sync info
	SyncCookie       []byte
	PubAccountCookie []byte
	Ksid             []byte
	// msgCtrlBuf       []byte
}
