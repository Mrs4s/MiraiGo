package client

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime/debug"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/highway"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/exciting"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

type (
	GroupFileSystem struct {
		FileCount  uint32 `json:"file_count"`
		LimitCount uint32 `json:"limit_count"`
		UsedSpace  uint64 `json:"used_space"`
		TotalSpace uint64 `json:"total_space"`
		GroupCode  int64  `json:"group_id"`

		client *QQClient
	}

	GroupFile struct {
		GroupCode     int64  `json:"group_id"`
		FileId        string `json:"file_id"`
		FileName      string `json:"file_name"`
		BusId         int32  `json:"busid"`
		FileSize      int64  `json:"file_size"`
		UploadTime    int64  `json:"upload_time"`
		DeadTime      int64  `json:"dead_time"`
		ModifyTime    int64  `json:"modify_time"`
		DownloadTimes int64  `json:"download_times"`
		Uploader      int64  `json:"uploader"`
		UploaderName  string `json:"uploader_name"`
	}

	GroupFolder struct {
		GroupCode      int64  `json:"group_id"`
		FolderId       string `json:"folder_id"`
		FolderName     string `json:"folder_name"`
		CreateTime     int64  `json:"create_time"`
		Creator        int64  `json:"creator"`
		CreatorName    string `json:"creator_name"`
		TotalFileCount uint32 `json:"total_file_count"`
	}
)

var fsWaiter = utils.NewUploadWaiter()

func init() {
	decoders["OidbSvc.0x6d8_1"] = decodeOIDB6d81Response
	decoders["OidbSvc.0x6d6_0"] = decodeOIDB6d60Response
	decoders["OidbSvc.0x6d6_2"] = decodeOIDB6d62Response
	decoders["OidbSvc.0x6d6_3"] = decodeOIDB6d63Response
	decoders["OidbSvc.0x6d7_0"] = decodeOIDB6d7Response
	decoders["OidbSvc.0x6d7_1"] = decodeOIDB6d7Response
	decoders["OidbSvc.0x6d7_2"] = decodeOIDB6d7Response
	decoders["OidbSvc.0x6d9_4"] = ignoreDecoder
}

func (c *QQClient) GetGroupFileSystem(groupCode int64) (fs *GroupFileSystem, err error) {
	defer func() {
		if pan := recover(); pan != nil {
			c.error("get group fs error: %v\n%s", pan, debug.Stack())
			err = errors.New("fs error")
		}
	}()
	g := c.FindGroup(groupCode)
	if g == nil {
		return nil, errors.New("group not found")
	}
	rsp, e := c.sendAndWait(c.buildGroupFileCountRequestPacket(groupCode))
	if e != nil {
		return nil, e
	}
	fs = &GroupFileSystem{
		FileCount:  rsp.(*oidb.D6D8RspBody).FileCountRsp.GetAllFileCount(),
		LimitCount: rsp.(*oidb.D6D8RspBody).FileCountRsp.GetLimitCount(),
		GroupCode:  groupCode,
		client:     c,
	}
	rsp, err = c.sendAndWait(c.buildGroupFileSpaceRequestPacket(groupCode))
	if err != nil {
		return nil, err
	}
	fs.TotalSpace = rsp.(*oidb.D6D8RspBody).GroupSpaceRsp.GetTotalSpace()
	fs.UsedSpace = rsp.(*oidb.D6D8RspBody).GroupSpaceRsp.GetUsedSpace()
	return fs, nil
}

func (c *QQClient) GetGroupFileUrl(groupCode int64, fileId string, busId int32) string {
	i, err := c.sendAndWait(c.buildGroupFileDownloadReqPacket(groupCode, fileId, busId))
	if err != nil {
		return ""
	}
	url := i.(string)
	url += fmt.Sprintf("?fname=%x", fileId)
	return url
}

func (fs *GroupFileSystem) Root() ([]*GroupFile, []*GroupFolder, error) {
	return fs.GetFilesByFolder("/")
}

func (fs *GroupFileSystem) GetFilesByFolder(folderID string) ([]*GroupFile, []*GroupFolder, error) {
	var startIndex uint32 = 0
	var files []*GroupFile
	var folders []*GroupFolder
	for {
		i, err := fs.client.sendAndWait(fs.client.buildGroupFileListRequestPacket(fs.GroupCode, folderID, startIndex))
		if err != nil {
			return nil, nil, err
		}
		rsp := i.(*oidb.D6D8RspBody)
		if rsp.FileListInfoRsp == nil {
			break
		}
		for _, item := range rsp.FileListInfoRsp.ItemList {
			if item.FileInfo != nil {
				files = append(files, &GroupFile{
					GroupCode:     fs.GroupCode,
					FileId:        item.FileInfo.GetFileId(),
					FileName:      item.FileInfo.GetFileName(),
					BusId:         int32(item.FileInfo.GetBusId()),
					FileSize:      int64(item.FileInfo.GetFileSize()),
					UploadTime:    int64(item.FileInfo.GetUploadTime()),
					DeadTime:      int64(item.FileInfo.GetDeadTime()),
					ModifyTime:    int64(item.FileInfo.GetModifyTime()),
					DownloadTimes: int64(item.FileInfo.GetDownloadTimes()),
					Uploader:      int64(item.FileInfo.GetUploaderUin()),
					UploaderName:  item.FileInfo.GetUploaderName(),
				})
			}
			if item.FolderInfo != nil {
				folders = append(folders, &GroupFolder{
					GroupCode:      fs.GroupCode,
					FolderId:       item.FolderInfo.GetFolderId(),
					FolderName:     item.FolderInfo.GetFolderName(),
					CreateTime:     int64(item.FolderInfo.GetCreateTime()),
					Creator:        int64(item.FolderInfo.GetCreateUin()),
					CreatorName:    item.FolderInfo.GetCreatorName(),
					TotalFileCount: item.FolderInfo.GetTotalFileCount(),
				})
			}
		}
		if rsp.FileListInfoRsp.GetIsEnd() {
			break
		}
		startIndex = rsp.FileListInfoRsp.GetNextIndex()
	}
	return files, folders, nil
}

func (fs *GroupFileSystem) UploadFile(p, name, folderId string) error {
	// 同文件等待其他线程上传
	fsWaiter.Wait(p)
	defer fsWaiter.Done(p)

	file, err := os.OpenFile(p, os.O_RDONLY, 0o666)
	if err != nil {
		return errors.Wrap(err, "open file error")
	}
	defer func() { _ = file.Close() }()
	md5Hash, size := utils.ComputeMd5AndLength(file)
	_, _ = file.Seek(0, io.SeekStart)
	sha1H := sha1.New()
	_, _ = io.Copy(sha1H, file)
	sha1Hash := sha1H.Sum(nil)
	_, _ = file.Seek(0, io.SeekStart)
	i, err := fs.client.sendAndWait(fs.client.buildGroupFileUploadReqPacket(folderId, name, fs.GroupCode, size, md5Hash, sha1Hash))
	if err != nil {
		return errors.Wrap(err, "query upload failed")
	}
	rsp := i.(*oidb.UploadFileRspBody)
	if rsp.GetBoolFileExist() {
		_, pkt := fs.client.buildGroupFileFeedsRequest(fs.GroupCode, rsp.GetFileId(), rsp.GetBusId(), rand.Int31())
		return fs.client.sendPacket(pkt)
	}
	if len(rsp.UploadIpLanV4) == 0 {
		return errors.New("server requires unsupported ftn upload")
	}
	ext, _ := proto.Marshal(&exciting.GroupFileUploadExt{
		Unknown1: proto.Int32(100),
		Unknown2: proto.Int32(1),
		Entry: &exciting.GroupFileUploadEntry{
			BusiBuff: &exciting.ExcitingBusiInfo{
				BusId:       rsp.BusId,
				SenderUin:   &fs.client.Uin,
				ReceiverUin: &fs.GroupCode,
				GroupCode:   &fs.GroupCode,
			},
			FileEntry: &exciting.ExcitingFileEntry{
				FileSize:  &size,
				Md5:       md5Hash,
				Sha1:      sha1Hash,
				FileId:    []byte(rsp.GetFileId()),
				UploadKey: rsp.CheckKey,
			},
			ClientInfo: &exciting.ExcitingClientInfo{
				ClientType:   proto.Int32(2),
				AppId:        proto.String(fmt.Sprint(fs.client.version.AppId)),
				TerminalType: proto.Int32(2),
				ClientVer:    proto.String("9e9c09dc"),
				Unknown:      proto.Int32(4),
			},
			FileNameInfo: &exciting.ExcitingFileNameInfo{FileName: &name},
			Host: &exciting.ExcitingHostConfig{Hosts: []*exciting.ExcitingHostInfo{
				{
					Url: &exciting.ExcitingUrlInfo{
						Unknown: proto.Int32(1),
						Host:    &rsp.UploadIpLanV4[0],
					},
					Port: rsp.UploadPort,
				},
			}},
		},
		Unknown3: proto.Int32(0),
	})
	client := fs.client
	input := highway.ExcitingInput{
		CommandID: 71,
		Body:      file,
		Ticket:    fs.client.highwaySession.SigSession,
		Ext:       ext,
	}
	if _, err = fs.client.highwaySession.UploadExciting(input); err != nil {
		return errors.Wrap(err, "upload failed")
	}
	_, pkt := client.buildGroupFileFeedsRequest(fs.GroupCode, rsp.GetFileId(), rsp.GetBusId(), rand.Int31())
	return client.sendPacket(pkt)
}

func (fs *GroupFileSystem) GetDownloadUrl(file *GroupFile) string {
	return fs.client.GetGroupFileUrl(file.GroupCode, file.FileId, file.BusId)
}

func (fs *GroupFileSystem) CreateFolder(parentFolder, name string) error {
	if _, err := fs.client.sendAndWait(fs.client.buildGroupFileCreateFolderPacket(fs.GroupCode, parentFolder, name)); err != nil {
		return errors.Wrap(err, "create folder error")
	}
	return nil
}

func (fs *GroupFileSystem) RenameFolder(folderId, newName string) error {
	if _, err := fs.client.sendAndWait(fs.client.buildGroupFileRenameFolderPacket(fs.GroupCode, folderId, newName)); err != nil {
		return errors.Wrap(err, "rename folder error")
	}
	return nil
}

func (fs *GroupFileSystem) DeleteFolder(folderId string) error {
	if _, err := fs.client.sendAndWait(fs.client.buildGroupFileDeleteFolderPacket(fs.GroupCode, folderId)); err != nil {
		return errors.Wrap(err, "rename folder error")
	}
	return nil
}

// DeleteFile 删除群文件，需要管理权限.
// 返回错误, 空为删除成功
func (fs *GroupFileSystem) DeleteFile(parentFolderID, fileId string, busId int32) string {
	i, err := fs.client.sendAndWait(fs.client.buildGroupFileDeleteReqPacket(fs.GroupCode, parentFolderID, fileId, busId))
	if err != nil {
		return err.Error()
	}
	return i.(string)
}

func (c *QQClient) buildGroupFileUploadReqPacket(parentFolderID, fileName string, groupCode, fileSize int64, md5, sha1 []byte) (uint16, []byte) {
	body := &oidb.D6D6ReqBody{UploadFileReq: &oidb.UploadFileReqBody{
		GroupCode:          &groupCode,
		AppId:              proto.Int32(3),
		BusId:              proto.Int32(102),
		Entrance:           proto.Int32(5),
		ParentFolderId:     &parentFolderID,
		FileName:           &fileName,
		LocalPath:          proto.String("/storage/emulated/0/Pictures/files/s/" + fileName),
		Int64FileSize:      &fileSize,
		Sha:                sha1,
		Md5:                md5,
		SupportMultiUpload: proto.Bool(true),
	}}
	payload := c.packOIDBPackageProto(1750, 0, body)
	return c.uniPacket("OidbSvc.0x6d6_0", payload)
}

func (c *QQClient) buildGroupFileFeedsRequest(groupCode int64, fileID string, busId, msgRand int32) (uint16, []byte) {
	req := c.packOIDBPackageProto(1753, 4, &oidb.D6D9ReqBody{FeedsInfoReq: &oidb.FeedsReqBody{
		GroupCode: proto.Uint64(uint64(groupCode)),
		AppId:     proto.Uint32(3),
		FeedsInfoList: []*oidb.GroupFileFeedsInfo{{
			FileId:    &fileID,
			FeedFlag:  proto.Uint32(1),
			BusId:     proto.Uint32(uint32(busId)),
			MsgRandom: proto.Uint32(uint32(msgRand)),
		}},
	}})
	return c.uniPacket("OidbSvc.0x6d9_4", req)
}

// OidbSvc.0x6d8_1
func (c *QQClient) buildGroupFileListRequestPacket(groupCode int64, folderID string, startIndex uint32) (uint16, []byte) {
	body := &oidb.D6D8ReqBody{FileListInfoReq: &oidb.GetFileListReqBody{
		GroupCode:    proto.Uint64(uint64(groupCode)),
		AppId:        proto.Uint32(3),
		FolderId:     &folderID,
		FileCount:    proto.Uint32(20),
		AllFileCount: proto.Uint32(0),
		ReqFrom:      proto.Uint32(3),
		SortBy:       proto.Uint32(1),
		FilterCode:   proto.Uint32(0),
		Uin:          proto.Uint64(0),
		StartIndex:   &startIndex,
		Context:      EmptyBytes,
	}}
	payload := c.packOIDBPackageProto(1752, 1, body)
	return c.uniPacket("OidbSvc.0x6d8_1", payload)
}

func (c *QQClient) buildGroupFileCountRequestPacket(groupCode int64) (uint16, []byte) {
	body := &oidb.D6D8ReqBody{
		GroupFileCountReq: &oidb.GetFileCountReqBody{
			GroupCode: proto.Uint64(uint64(groupCode)),
			AppId:     proto.Uint32(3),
			BusId:     proto.Uint32(0),
		},
	}
	payload := c.packOIDBPackageProto(1752, 2, body)
	return c.uniPacket("OidbSvc.0x6d8_1", payload)
}

func (c *QQClient) buildGroupFileSpaceRequestPacket(groupCode int64) (uint16, []byte) {
	body := &oidb.D6D8ReqBody{GroupSpaceReq: &oidb.GetSpaceReqBody{
		GroupCode: proto.Uint64(uint64(groupCode)),
		AppId:     proto.Uint32(3),
	}}
	payload := c.packOIDBPackageProto(1752, 3, body)
	return c.uniPacket("OidbSvc.0x6d8_1", payload)
}

func (c *QQClient) buildGroupFileCreateFolderPacket(groupCode int64, parentFolder, name string) (uint16, []byte) {
	payload := c.packOIDBPackageProto(1751, 0, &oidb.D6D7ReqBody{CreateFolderReq: &oidb.CreateFolderReqBody{
		GroupCode:      proto.Uint64(uint64(groupCode)),
		AppId:          proto.Uint32(3),
		ParentFolderId: &parentFolder,
		FolderName:     &name,
	}})
	return c.uniPacket("OidbSvc.0x6d7_0", payload)
}

func (c *QQClient) buildGroupFileRenameFolderPacket(groupCode int64, folderId, newName string) (uint16, []byte) {
	payload := c.packOIDBPackageProto(1751, 2, &oidb.D6D7ReqBody{RenameFolderReq: &oidb.RenameFolderReqBody{
		GroupCode:     proto.Uint64(uint64(groupCode)),
		AppId:         proto.Uint32(3),
		FolderId:      proto.String(folderId),
		NewFolderName: proto.String(newName),
	}})
	return c.uniPacket("OidbSvc.0x6d7_2", payload)
}

func (c *QQClient) buildGroupFileDeleteFolderPacket(groupCode int64, folderId string) (uint16, []byte) {
	payload := c.packOIDBPackageProto(1751, 1, &oidb.D6D7ReqBody{DeleteFolderReq: &oidb.DeleteFolderReqBody{
		GroupCode: proto.Uint64(uint64(groupCode)),
		AppId:     proto.Uint32(3),
		FolderId:  proto.String(folderId),
	}})
	return c.uniPacket("OidbSvc.0x6d7_1", payload)
}

// OidbSvc.0x6d6_2
func (c *QQClient) buildGroupFileDownloadReqPacket(groupCode int64, fileId string, busId int32) (uint16, []byte) {
	body := &oidb.D6D6ReqBody{
		DownloadFileReq: &oidb.DownloadFileReqBody{
			GroupCode: &groupCode,
			AppId:     proto.Int32(3),
			BusId:     &busId,
			FileId:    &fileId,
		},
	}
	payload := c.packOIDBPackageProto(1750, 2, body)
	return c.uniPacket("OidbSvc.0x6d6_2", payload)
}

func (c *QQClient) buildGroupFileDeleteReqPacket(groupCode int64, parentFolderId, fileId string, busId int32) (uint16, []byte) {
	body := &oidb.D6D6ReqBody{DeleteFileReq: &oidb.DeleteFileReqBody{
		GroupCode:      &groupCode,
		AppId:          proto.Int32(3),
		BusId:          &busId,
		ParentFolderId: &parentFolderId,
		FileId:         &fileId,
	}}
	payload := c.packOIDBPackageProto(1750, 3, body)
	return c.uniPacket("OidbSvc.0x6d6_3", payload)
}

func decodeOIDB6d81Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D8RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// OidbSvc.0x6d6_2
func decodeOIDB6d62Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D6RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	if rsp.DownloadFileRsp.DownloadUrl == nil {
		return nil, errors.New(rsp.DownloadFileRsp.GetClientWording())
	}
	ip := rsp.DownloadFileRsp.GetDownloadIp()
	url := hex.EncodeToString(rsp.DownloadFileRsp.DownloadUrl)
	return fmt.Sprintf("http://%s/ftn_handler/%s/", ip, url), nil
}

func decodeOIDB6d63Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D6RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	return rsp.DeleteFileRsp.GetClientWording(), nil
}

func decodeOIDB6d60Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D6RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	return rsp.UploadFileRsp, nil
}

func decodeOIDB6d7Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D7RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	if retCode := rsp.CreateFolderRsp.GetRetCode(); retCode != 0 {
		return nil, errors.Errorf("create folder error: %v", retCode)
	}
	if retCode := rsp.RenameFolderRsp.GetRetCode(); retCode != 0 {
		return nil, errors.Errorf("rename folder error: %v", retCode)
	}
	if retCode := rsp.DeleteFolderRsp.GetRetCode(); retCode != 0 {
		return nil, errors.Errorf("delete folder error: %v", retCode)
	}
	return nil, nil
}
