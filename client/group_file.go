package client

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
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
		FileCount:  rsp.(*oidb.D6D8RspBody).FileCountRsp.AllFileCount.Unwrap(),
		LimitCount: rsp.(*oidb.D6D8RspBody).FileCountRsp.LimitCount.Unwrap(),
		GroupCode:  groupCode,
		client:     c,
	}
	rsp, err = c.sendAndWait(c.buildGroupFileSpaceRequestPacket(groupCode))
	if err != nil {
		return nil, err
	}
	fs.TotalSpace = rsp.(*oidb.D6D8RspBody).GroupSpaceRsp.TotalSpace.Unwrap()
	fs.UsedSpace = rsp.(*oidb.D6D8RspBody).GroupSpaceRsp.UsedSpace.Unwrap()
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
					FileId:        item.FileInfo.FileId.Unwrap(),
					FileName:      item.FileInfo.FileName.Unwrap(),
					BusId:         int32(item.FileInfo.BusId.Unwrap()),
					FileSize:      int64(item.FileInfo.FileSize.Unwrap()),
					UploadTime:    int64(item.FileInfo.UploadTime.Unwrap()),
					DeadTime:      int64(item.FileInfo.DeadTime.Unwrap()),
					ModifyTime:    int64(item.FileInfo.ModifyTime.Unwrap()),
					DownloadTimes: int64(item.FileInfo.DownloadTimes.Unwrap()),
					Uploader:      int64(item.FileInfo.UploaderUin.Unwrap()),
					UploaderName:  item.FileInfo.UploaderName.Unwrap(),
				})
			}
			if item.FolderInfo != nil {
				folders = append(folders, &GroupFolder{
					GroupCode:      fs.GroupCode,
					FolderId:       item.FolderInfo.FolderId.Unwrap(),
					FolderName:     item.FolderInfo.FolderName.Unwrap(),
					CreateTime:     int64(item.FolderInfo.CreateTime.Unwrap()),
					Creator:        int64(item.FolderInfo.CreateUin.Unwrap()),
					CreatorName:    item.FolderInfo.CreatorName.Unwrap(),
					TotalFileCount: item.FolderInfo.TotalFileCount.Unwrap(),
				})
			}
		}
		if rsp.FileListInfoRsp.IsEnd.Unwrap() {
			break
		}
		startIndex = rsp.FileListInfoRsp.NextIndex.Unwrap()
	}
	return files, folders, nil
}

func (fs *GroupFileSystem) UploadFile(p, name, folderId string) error {
	file, err := os.OpenFile(p, os.O_RDONLY, 0o666)
	if err != nil {
		return errors.Wrap(err, "open file error")
	}
	defer func() { _ = file.Close() }()
	f := &LocalFile{
		FileName:     name,
		Body:         file,
		RemoteFolder: folderId,
	}
	target := message.Source{
		SourceType: message.SourceGroup,
		PrimaryID:  fs.GroupCode,
	}
	return fs.client.UploadFile(target, f)
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

func (c *QQClient) buildGroupFileUploadReqPacket(groupCode int64, file *LocalFile) (uint16, []byte) {
	body := &oidb.D6D6ReqBody{UploadFileReq: &oidb.UploadFileReqBody{
		GroupCode:          proto.Some(groupCode),
		AppId:              proto.Int32(3),
		BusId:              proto.Int32(102),
		Entrance:           proto.Int32(5),
		ParentFolderId:     proto.Some(file.RemoteFolder),
		FileName:           proto.Some(file.FileName),
		LocalPath:          proto.String("/storage/emulated/0/Pictures/files/s/" + file.FileName),
		Int64FileSize:      proto.Some(file.size),
		Sha:                file.sha1,
		Md5:                file.md5,
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
			FileId:    proto.Some(fileID),
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
		FolderId:     proto.Some(folderID),
		FileCount:    proto.Uint32(20),
		AllFileCount: proto.Uint32(0),
		ReqFrom:      proto.Uint32(3),
		SortBy:       proto.Uint32(1),
		FilterCode:   proto.Uint32(0),
		Uin:          proto.Uint64(0),
		StartIndex:   proto.Some(startIndex),
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
		ParentFolderId: proto.Some(parentFolder),
		FolderName:     proto.Some(name),
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
			GroupCode: proto.Some(groupCode),
			AppId:     proto.Int32(3),
			BusId:     proto.Some(busId),
			FileId:    proto.Some(fileId),
		},
	}
	payload := c.packOIDBPackageProto(1750, 2, body)
	return c.uniPacket("OidbSvc.0x6d6_2", payload)
}

func (c *QQClient) buildGroupFileDeleteReqPacket(groupCode int64, parentFolderId, fileId string, busId int32) (uint16, []byte) {
	body := &oidb.D6D6ReqBody{DeleteFileReq: &oidb.DeleteFileReqBody{
		GroupCode:      proto.Some(groupCode),
		AppId:          proto.Int32(3),
		BusId:          proto.Some(busId),
		ParentFolderId: proto.Some(parentFolderId),
		FileId:         proto.Some(fileId),
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
		return nil, errors.New(rsp.DownloadFileRsp.ClientWording.Unwrap())
	}
	ip := rsp.DownloadFileRsp.DownloadIp.Unwrap()
	url := hex.EncodeToString(rsp.DownloadFileRsp.DownloadUrl)
	return fmt.Sprintf("http://%s/ftn_handler/%s/", ip, url), nil
}

func decodeOIDB6d63Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D6RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	return rsp.DeleteFileRsp.ClientWording.Unwrap(), nil
}

func decodeOIDB6d60Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D6RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	u := rsp.UploadFileRsp
	r := &fileUploadRsp{
		Existed:       u.BoolFileExist.Unwrap(),
		BusID:         u.BusId.Unwrap(),
		Uuid:          []byte(u.FileId.Unwrap()),
		UploadKey:     u.CheckKey,
		UploadIpLanV4: u.UploadIpLanV4,
		UploadPort:    u.UploadPort.Unwrap(),
	}
	return r, nil
}

func decodeOIDB6d7Response(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D6D7RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	if createRsp := rsp.CreateFolderRsp; createRsp != nil {
		if retCode := createRsp.RetCode.Unwrap(); retCode != 0 {
			return nil, errors.Errorf("create folder error: %v", retCode)
		}
	}
	if renameRsp := rsp.RenameFolderRsp; renameRsp != nil {
		if retCode := renameRsp.RetCode.Unwrap(); retCode != 0 {
			return nil, errors.Errorf("rename folder error: %v", retCode)
		}
	}
	if deleteRsp := rsp.DeleteFolderRsp; deleteRsp != nil {
		if retCode := deleteRsp.RetCode.Unwrap(); retCode != 0 {
			return nil, errors.Errorf("delete folder error: %v", retCode)
		}
	}
	return nil, nil
}
