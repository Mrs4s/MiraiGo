package auth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

//go:generate stringer -type=ProtocolType -linecomment
type ProtocolType int

const (
	Unset        ProtocolType = iota
	AndroidPhone              // Android Phone
	AndroidWatch              // Android Watch
	MacOS                     // MacOS
	QiDian                    // 企点
	IPad                      // iPad
	AndroidPad                // Android Pad
)

var (
	AppVersions = map[ProtocolType]*AppVersion{
		AndroidPhone: {
			ApkId:           "com.tencent.mobileqq",
			AppId:           537153294,
			SubAppId:        537153294,
			AppKey:          "0S200MNJT807V3GE",
			SortVersionName: "8.9.35.10440",
			BuildTime:       1676531414,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.2535",
			SSOVersion:      19,
			MiscBitmap:      150470524,
			SubSigmap:       0x10400,
			MainSigMap: WLOGIN_A5 | WLOGIN_RESERVED | WLOGIN_STWEB | WLOGIN_A2 | WLOGIN_ST |
				WLOGIN_LSKEY | WLOGIN_SKEY | WLOGIN_SIG64 | 1<<16 | WLOGIN_VKEY | WLOGIN_D2 |
				WLOGIN_SID | WLOGIN_PSKEY | WLOGIN_AQSIG | WLOGIN_LHSIG | WLOGIN_PAYTOKEN, // 16724722
			Protocol: AndroidPhone,
		},
		AndroidPad: {
			ApkId:           "com.tencent.mobileqq",
			AppId:           537152242,
			SubAppId:        537152242,
			AppKey:          "0S200MNJT807V3GE",
			SortVersionName: "8.9.35.10440",
			BuildTime:       1676531414,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.2535",
			SSOVersion:      19,
			MiscBitmap:      150470524,
			SubSigmap:       0x10400,
			MainSigMap: WLOGIN_A5 | WLOGIN_RESERVED | WLOGIN_STWEB | WLOGIN_A2 | WLOGIN_ST |
				WLOGIN_LSKEY | WLOGIN_SKEY | WLOGIN_SIG64 | 1<<16 | WLOGIN_VKEY | WLOGIN_D2 |
				WLOGIN_SID | WLOGIN_PSKEY | WLOGIN_AQSIG | WLOGIN_LHSIG | WLOGIN_PAYTOKEN, // 16724722
			Protocol: AndroidPad,
		},
		AndroidWatch: {
			ApkId:           "com.tencent.qqlite",
			AppId:           537065138,
			SubAppId:        537065138,
			SortVersionName: "2.0.8",
			BuildTime:       1559564731,
			ApkSign:         []byte{0xA6, 0xB7, 0x45, 0xBF, 0x24, 0xA2, 0xC2, 0x77, 0x52, 0x77, 0x16, 0xF6, 0xF3, 0x6E, 0xB6, 0x8D},
			SdkVersion:      "6.0.0.2365",
			SSOVersion:      5,
			MiscBitmap:      16252796,
			SubSigmap:       0x10400,
			MainSigMap: WLOGIN_A5 | WLOGIN_RESERVED | WLOGIN_STWEB | WLOGIN_A2 | WLOGIN_ST |
				WLOGIN_LSKEY | WLOGIN_SKEY | WLOGIN_SIG64 | 1<<16 | WLOGIN_VKEY | WLOGIN_D2 |
				WLOGIN_SID | WLOGIN_PSKEY | WLOGIN_AQSIG | WLOGIN_LHSIG | WLOGIN_PAYTOKEN, // 16724722
			Protocol: AndroidWatch,
		},
		IPad: {
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537151363,
			SubAppId:        537151363,
			SortVersionName: "8.9.33.614",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      19,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      WLOGIN_STWEB | WLOGIN_A2 | WLOGIN_ST | WLOGIN_SKEY | WLOGIN_VKEY | WLOGIN_D2 | WLOGIN_SID | WLOGIN_PSKEY, // 1970400
			Protocol:        IPad,
		},
		MacOS: {
			ApkId:           "com.tencent.minihd.qq",
			AppId:           537128930,
			SubAppId:        537128930,
			SortVersionName: "5.8.9",
			BuildTime:       1595836208,
			ApkSign:         []byte{170, 57, 120, 244, 31, 217, 111, 249, 145, 74, 102, 158, 24, 100, 116, 199},
			SdkVersion:      "6.0.0.2433",
			SSOVersion:      12,
			MiscBitmap:      150470524,
			SubSigmap:       66560,
			MainSigMap:      WLOGIN_STWEB | WLOGIN_A2 | WLOGIN_ST | WLOGIN_SKEY | WLOGIN_VKEY | WLOGIN_D2 | WLOGIN_SID | WLOGIN_PSKEY, // 1970400
			Protocol:        MacOS,
		},
		QiDian: {
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
			MainSigMap:      WLOGIN_STWEB | WLOGIN_A2 | WLOGIN_ST | WLOGIN_SKEY | WLOGIN_D2 | WLOGIN_PSKEY | WLOGIN_DA2, // 34869472
			Protocol:        QiDian,
		},
	}
)

// see oicq/wlogin_sdk/request/WtloginHelper.java  class SigType
const (
	_ = 1 << iota
	WLOGIN_A5
	_
	_
	WLOGIN_RESERVED
	WLOGIN_STWEB
	WLOGIN_A2
	WLOGIN_ST
	_
	WLOGIN_LSKEY
	_
	_
	WLOGIN_SKEY
	WLOGIN_SIG64
	WLOGIN_OPENKEY
	WLOGIN_TOKEN
	_
	WLOGIN_VKEY
	WLOGIN_D2
	WLOGIN_SID
	WLOGIN_PSKEY
	WLOGIN_AQSIG
	WLOGIN_LHSIG
	WLOGIN_PAYTOKEN
	WLOGIN_PF
	WLOGIN_DA2
	WLOGIN_QRPUSH
	WLOGIN_PT4Token
)

type AppVersion struct {
	ApkSign         []byte
	ApkId           string
	SortVersionName string
	SdkVersion      string
	AppId           uint32
	SubAppId        uint32
	AppKey          string
	BuildTime       uint32
	SSOVersion      uint32
	MiscBitmap      uint32
	SubSigmap       uint32
	MainSigMap      uint32
	Protocol        ProtocolType
}

func (v *AppVersion) String() string {
	return fmt.Sprintf("%s %s", v.Protocol.String(), v.SortVersionName)
}

func (v *AppVersion) UpdateFromJson(d []byte) error {
	var f appVersionFile
	if err := json.Unmarshal(d, &f); err != nil {
		return errors.Wrap(err, "failed to unmarshal json message")
	}
	v.ApkId = f.ApkId
	v.AppId = f.AppId
	v.SubAppId = f.SubAppId
	v.AppKey = f.AppKey
	v.SortVersionName = f.SortVersionName
	v.BuildTime = f.BuildTime
	v.ApkSign, _ = hex.DecodeString(f.ApkSign)
	v.SdkVersion = f.SdkVersion
	v.SSOVersion = f.SSOVersion
	v.MiscBitmap = f.MiscBitmap
	v.SubSigmap = f.SubSigmap
	v.MainSigMap = f.MainSigMap
	v.Protocol = f.ProtocolType
	return nil
}

func (i ProtocolType) Version() *AppVersion {
	return AppVersions[i]
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
	T547     []byte
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

type appVersionFile struct {
	ApkId           string       `json:"apk_id"`
	AppId           uint32       `json:"app_id"`
	SubAppId        uint32       `json:"sub_app_id"`
	AppKey          string       `json:"app_key"`
	SortVersionName string       `json:"sort_version_name"`
	BuildTime       uint32       `json:"build_time"`
	ApkSign         string       `json:"apk_sign"` // hex encoded
	SdkVersion      string       `json:"sdk_version"`
	SSOVersion      uint32       `json:"sso_version"`
	MiscBitmap      uint32       `json:"misc_bitmap"`
	MainSigMap      uint32       `json:"main_sig_map"`
	SubSigmap       uint32       `json:"sub_sig_map"`
	ProtocolType    ProtocolType `json:"protocol_type"`
}
