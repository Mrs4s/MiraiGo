package auth

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/rand"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

type OSVersion struct {
	Incremental []byte
	Release     []byte
	CodeName    []byte
	SDK         uint32
}

type Device struct {
	Display      []byte
	Product      []byte
	Device       []byte
	Board        []byte
	Brand        []byte
	Model        []byte
	Bootloader   []byte
	FingerPrint  []byte
	BootId       []byte
	ProcVersion  []byte
	BaseBand     []byte
	SimInfo      []byte
	OSType       []byte
	MacAddress   []byte
	IpAddress    []byte
	WifiBSSID    []byte
	WifiSSID     []byte
	IMSIMd5      []byte
	IMEI         string
	AndroidId    []byte
	APN          []byte
	VendorName   []byte
	VendorOSName []byte
	Guid         []byte
	TgtgtKey     []byte
	Protocol     Protocol
	Version      *OSVersion
}

func (info *Device) ToJson() []byte {
	f := &deviceFile{
		Display:     string(info.Display),
		Product:     string(info.Product),
		Device:      string(info.Device),
		Board:       string(info.Board),
		Model:       string(info.Model),
		FingerPrint: string(info.FingerPrint),
		BootId:      string(info.BootId),
		ProcVersion: string(info.ProcVersion),
		IMEI:        info.IMEI,
		Brand:       string(info.Brand),
		Bootloader:  string(info.Bootloader),
		BaseBand:    string(info.BaseBand),
		AndroidId:   string(info.AndroidId),
		Version: &osVersionFile{
			Incremental: string(info.Version.Incremental),
			Release:     string(info.Version.Release),
			Codename:    string(info.Version.CodeName),
			Sdk:         info.Version.SDK,
		},
		SimInfo:      string(info.SimInfo),
		OsType:       string(info.OSType),
		MacAddress:   string(info.MacAddress),
		IpAddress:    []int32{int32(info.IpAddress[0]), int32(info.IpAddress[1]), int32(info.IpAddress[2]), int32(info.IpAddress[3])},
		WifiBSSID:    string(info.WifiBSSID),
		WifiSSID:     string(info.WifiSSID),
		ImsiMd5:      hex.EncodeToString(info.IMSIMd5),
		Apn:          string(info.APN),
		VendorName:   string(info.VendorName),
		VendorOSName: string(info.VendorOSName),
		Protocol:     int(info.Protocol),
	}
	d, _ := json.Marshal(f)
	return d
}

func (info *Device) ReadJson(d []byte) error {
	var f deviceFile
	if err := json.Unmarshal(d, &f); err != nil {
		return errors.Wrap(err, "failed to unmarshal json message")
	}
	setIfNotEmpty := func(trg *[]byte, str string) {
		if str != "" {
			*trg = []byte(str)
		}
	}
	setIfNotEmpty(&info.Display, f.Display)
	setIfNotEmpty(&info.Product, f.Product)
	setIfNotEmpty(&info.Device, f.Device)
	setIfNotEmpty(&info.Board, f.Board)
	setIfNotEmpty(&info.Brand, f.Brand)
	setIfNotEmpty(&info.Model, f.Model)
	setIfNotEmpty(&info.Bootloader, f.Bootloader)
	setIfNotEmpty(&info.FingerPrint, f.FingerPrint)
	setIfNotEmpty(&info.BootId, f.BootId)
	setIfNotEmpty(&info.ProcVersion, f.ProcVersion)
	setIfNotEmpty(&info.BaseBand, f.BaseBand)
	setIfNotEmpty(&info.SimInfo, f.SimInfo)
	setIfNotEmpty(&info.OSType, f.OsType)
	setIfNotEmpty(&info.MacAddress, f.MacAddress)
	if len(f.IpAddress) == 4 {
		info.IpAddress = []byte{byte(f.IpAddress[0]), byte(f.IpAddress[1]), byte(f.IpAddress[2]), byte(f.IpAddress[3])}
	}
	setIfNotEmpty(&info.WifiBSSID, f.WifiBSSID)
	setIfNotEmpty(&info.WifiSSID, f.WifiSSID)
	if len(f.ImsiMd5) != 0 {
		imsiMd5, err := hex.DecodeString(f.ImsiMd5)
		if err != nil {
			info.IMSIMd5 = imsiMd5
		}
	}
	if f.IMEI != "" {
		info.IMEI = f.IMEI
	}
	setIfNotEmpty(&info.APN, f.Apn)
	setIfNotEmpty(&info.VendorName, f.VendorName)
	setIfNotEmpty(&info.VendorOSName, f.VendorOSName)

	setIfNotEmpty(&info.AndroidId, f.AndroidId)
	if f.AndroidId == "" {
		info.AndroidId = info.Display // ?
	}

	switch f.Protocol {
	case 1, 2, 3, 4, 5:
		info.Protocol = Protocol(f.Protocol)
	default:
		info.Protocol = IPad
	}
	info.GenNewGuid()
	info.GenNewTgtgtKey()
	return nil
}

func (info *Device) GenNewGuid() {
	t := md5.Sum(append(info.AndroidId, info.MacAddress...))
	info.Guid = t[:]
}

func (info *Device) GenNewTgtgtKey() {
	r := make([]byte, 16)
	rand.Read(r)
	h := md5.New()
	h.Write(r)
	h.Write(info.Guid)
	info.TgtgtKey = h.Sum(nil)
}

func (info *Device) GenDeviceInfoData() []byte {
	m := &pb.DeviceInfo{
		Bootloader:   string(info.Bootloader),
		ProcVersion:  string(info.ProcVersion),
		Codename:     string(info.Version.CodeName),
		Incremental:  string(info.Version.Incremental),
		Fingerprint:  string(info.FingerPrint),
		BootId:       string(info.BootId),
		AndroidId:    string(info.AndroidId),
		BaseBand:     string(info.BaseBand),
		InnerVersion: string(info.Version.Incremental),
	}
	data, err := proto.Marshal(m)
	if err != nil {
		panic(errors.Wrap(err, "failed to unmarshal protobuf message"))
	}
	return data
}

type deviceFile struct {
	Display      string         `json:"display"`
	Product      string         `json:"product"`
	Device       string         `json:"device"`
	Board        string         `json:"board"`
	Model        string         `json:"model"`
	FingerPrint  string         `json:"finger_print"`
	BootId       string         `json:"boot_id"`
	ProcVersion  string         `json:"proc_version"`
	Protocol     int            `json:"protocol"` // 0: Pad 1: Phone 2: Watch
	IMEI         string         `json:"imei"`
	Brand        string         `json:"brand"`
	Bootloader   string         `json:"bootloader"`
	BaseBand     string         `json:"base_band"`
	Version      *osVersionFile `json:"version"`
	SimInfo      string         `json:"sim_info"`
	OsType       string         `json:"os_type"`
	MacAddress   string         `json:"mac_address"`
	IpAddress    []int32        `json:"ip_address"`
	WifiBSSID    string         `json:"wifi_bssid"`
	WifiSSID     string         `json:"wifi_ssid"`
	ImsiMd5      string         `json:"imsi_md5"`
	AndroidId    string         `json:"android_id"`
	Apn          string         `json:"apn"`
	VendorName   string         `json:"vendor_name"`
	VendorOSName string         `json:"vendor_os_name"`
}

type osVersionFile struct {
	Incremental string `json:"incremental"`
	Release     string `json:"release"`
	Codename    string `json:"codename"`
	Sdk         uint32 `json:"sdk"`
}
