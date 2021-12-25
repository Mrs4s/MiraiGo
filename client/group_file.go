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

func (c *QQClient) GetGroupFileSystem(groupCode int64) (fs *GroupFileSystem, err error) {
	defer func() {
		if pan := recover(); pan != nil {
			c.Error("get group fs error: %v\n%s", pan, debug.Stack())
			err = errors.New("fs error")
		}
	}()
	g := c.FindGroup(groupCode)
	if g == nil {
		return nil, errors.New("group not found")
	}
	rsp, e := c.callAndDecode(c.buildGroupFileCountRequest(groupCode), decodeOIDB6d81Response)
	if e != nil {
		return nil, e
	}
	fs = &GroupFileSystem{
		FileCount:  rsp.(*oidb.D6D8RspBody).FileCountRsp.GetAllFileCount(),
		LimitCount: rsp.(*oidb.D6D8RspBody).FileCountRsp.GetLimitCount(),
		GroupCode:  groupCode,
		client:     c,
	}
	rsp, err = c.callAndDecode(c.buildGroupFileSpaceRequest(groupCode), decodeOIDB6d81Response)
	if err != nil {
		return nil, err
	}
	fs.TotalSpace = rsp.(*oidb.D6D8RspBody).GroupSpaceRsp.GetTotalSpace()
	fs.UsedSpace = rsp.(*oidb.D6D8RspBody).GroupSpaceRsp.GetUsedSpace()
	return fs, nil
}

func (c *QQClient) GetGroupFileUrl(groupCode int64, fileId string, busId int32) string {
	i, err := c.callAndDecode(c.buildGroupFileDownloadReq(groupCode, fileId, busId), decodeOIDB6d62Response)
	if err != nil {
		return ""
	}
	url := i.(string)
	url += "?fname=" + hex.EncodeToString([]byte(fileId))
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
		req := fs.client.buildGroupFileListRequest(fs.GroupCode, folderID, startIndex)
		i, err := fs.client.callAndDecode(req, decodeOIDB6d81Response)
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
	req := fs.client.buildGroupFileUploadReq(folderId, name, fs.GroupCode, size, md5Hash, sha1Hash)
	i, err := fs.client.callAndDecode(req, decodeOIDB6d60Response)
	if err != nil {
		return errors.Wrap(err, "query upload failed")
	}
	rsp := i.(*oidb.UploadFileRspBody)
	if rsp.GetBoolFileExist() {
		req := fs.client.buildGroupFileFeedsRequest(fs.GroupCode, rsp.GetFileId(), rsp.GetBusId(), rand.Int31())
		_, err := fs.client.call(req)
		return err
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
	req = client.buildGroupFileFeedsRequest(fs.GroupCode, rsp.GetFileId(), rsp.GetBusId(), rand.Int31())
	_, err = client.call(req)
	return err
}

func (fs *GroupFileSystem) GetDownloadUrl(file *GroupFile) string {
	return fs.client.GetGroupFileUrl(file.GroupCode, file.FileId, file.BusId)
}

func (fs *GroupFileSystem) CreateFolder(parentFolder, name string) error {
	if _, err := fs.client.callAndDecode(fs.client.buildGroupFileCreateFolderRequest(fs.GroupCode, parentFolder, name), decodeOIDB6d7Response); err != nil {
		return errors.Wrap(err, "create folder error")
	}
	return nil
}

func (fs *GroupFileSystem) RenameFolder(folderId, newName string) error {
	if _, err := fs.client.callAndDecode(fs.client.buildGroupFileRenameFolderRequest(fs.GroupCode, folderId, newName), decodeOIDB6d7Response); err != nil {
		return errors.Wrap(err, "rename folder error")
	}
	return nil
}

func (fs *GroupFileSystem) DeleteFolder(folderId string) error {
	if _, err := fs.client.callAndDecode(fs.client.buildGroupFileDeleteFolderRequest(fs.GroupCode, folderId), decodeOIDB6d7Response); err != nil {
		return errors.Wrap(err, "rename folder error")
	}
	return nil
}

// DeleteFile 删除群文件，需要管理权限.
// 返回错误, 空为删除成功
func (fs *GroupFileSystem) DeleteFile(parentFolderID, fileId string, busId int32) string {
	i, err := fs.client.callAndDecode(fs.client.buildGroupFileDeleteReq(fs.GroupCode, parentFolderID, fileId, busId), decodeOIDB6d63Response)
	if err != nil {
		return err.Error()
	}
	return i.(string)
}

func (c *QQClient) buildGroupFileUploadReq(parentFolderID, fileName string, groupCode, fileSize int64, md5, sha1 []byte) *network.Request {
	b, _ := proto.Marshal(&oidb.D6D6ReqBody{UploadFileReq: &oidb.UploadFileReqBody{
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
	}})
	req := &oidb.OIDBSSOPkg{
		Command:       1750,
		ServiceType:   0,
		Bodybuffer:    b,
		ClientVersion: "android 8.4.8",
	}
	payload, _ := proto.Marshal(req)
	return c.uniRequest("OidbSvc.0x6d6_0", payload)
}

func (c *QQClient) buildGroupFileFeedsRequest(groupCode int64, fileID string, busId, msgRand int32) *network.Request {
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
	return c.uniRequest("OidbSvc.0x6d9_4", req)
}

// OidbSvc.0x6d8_1
func (c *QQClient) buildGroupFileListRequest(groupCode int64, folderID string, startIndex uint32) *network.Request {
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
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:       1752,
		ServiceType:   1,
		Bodybuffer:    b,
		ClientVersion: "android 8.4.8",
	}
	payload, _ := proto.Marshal(req)
	return c.uniRequest("OidbSvc.0x6d8_1", payload)
}

func (c *QQClient) buildGroupFileCountRequest(groupCode int64) *network.Request {
	body := &oidb.D6D8ReqBody{GroupFileCountReq: &oidb.GetFileCountReqBody{
		GroupCode: proto.Uint64(uint64(groupCode)),
		AppId:     proto.Uint32(3),
		BusId:     proto.Uint32(0),
	}}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:       1752,
		ServiceType:   2,
		Bodybuffer:    b,
		ClientVersion: "android 8.4.8",
	}
	payload, _ := proto.Marshal(req)
	return c.uniRequest("OidbSvc.0x6d8_1", payload)
}

func (c *QQClient) buildGroupFileSpaceRequest(groupCode int64) *network.Request {
	body := &oidb.D6D8ReqBody{GroupSpaceReq: &oidb.GetSpaceReqBody{
		GroupCode: proto.Uint64(uint64(groupCode)),
		AppId:     proto.Uint32(3),
	}}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:       1752,
		ServiceType:   3,
		Bodybuffer:    b,
		ClientVersion: "android 8.4.8",
	}
	payload, _ := proto.Marshal(req)
	return c.uniRequest("OidbSvc.0x6d8_1", payload)
}

func (c *QQClient) buildGroupFileCreateFolderRequest(groupCode int64, parentFolder, name string) *network.Request {
	payload := c.packOIDBPackageProto(1751, 0, &oidb.D6D7ReqBody{CreateFolderReq: &oidb.CreateFolderReqBody{
		GroupCode:      proto.Uint64(uint64(groupCode)),
		AppId:          proto.Uint32(3),
		ParentFolderId: &parentFolder,
		FolderName:     &name,
	}})
	return c.uniRequest("OidbSvc.0x6d7_0", payload)
}

func (c *QQClient) buildGroupFileRenameFolderRequest(groupCode int64, folderId, newName string) *network.Request {
	payload := c.packOIDBPackageProto(1751, 2, &oidb.D6D7ReqBody{RenameFolderReq: &oidb.RenameFolderReqBody{
		GroupCode:     proto.Uint64(uint64(groupCode)),
		AppId:         proto.Uint32(3),
		FolderId:      proto.String(folderId),
		NewFolderName: proto.String(newName),
	}})
	return c.uniRequest("OidbSvc.0x6d7_2", payload)
}

func (c *QQClient) buildGroupFileDeleteFolderRequest(groupCode int64, folderId string) *network.Request {
	payload := c.packOIDBPackageProto(1751, 1, &oidb.D6D7ReqBody{DeleteFolderReq: &oidb.DeleteFolderReqBody{
		GroupCode: proto.Uint64(uint64(groupCode)),
		AppId:     proto.Uint32(3),
		FolderId:  proto.String(folderId),
	}})
	return c.uniRequest("OidbSvc.0x6d7_1", payload)
}

// OidbSvc.0x6d6_2
func (c *QQClient) buildGroupFileDownloadReq(groupCode int64, fileId string, busId int32) *network.Request {
	body := &oidb.D6D6ReqBody{
		DownloadFileReq: &oidb.DownloadFileReqBody{
			GroupCode: &groupCode,
			AppId:     proto.Int32(3),
			BusId:     &busId,
			FileId:    &fileId,
		},
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:     1750,
		ServiceType: 2,
		Bodybuffer:  b,
	}
	payload, _ := proto.Marshal(req)
	return c.uniRequest("OidbSvc.0x6d6_2", payload)
}

func (c *QQClient) buildGroupFileDeleteReq(groupCode int64, parentFolderId, fileId string, busId int32) *network.Request {
	body := &oidb.D6D6ReqBody{DeleteFileReq: &oidb.DeleteFileReqBody{
		GroupCode:      &groupCode,
		AppId:          proto.Int32(3),
		BusId:          &busId,
		ParentFolderId: &parentFolderId,
		FileId:         &fileId,
	}}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:       1750,
		ServiceType:   3,
		Bodybuffer:    b,
		ClientVersion: "android 8.4.8",
	}
	payload, _ := proto.Marshal(req)
	return c.uniRequest("OidbSvc.0x6d6_3", payload)
}

func decodeOIDB6d81Response(_ *QQClient, resp *network.Response) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D8RspBody{}
	if err := proto.Unmarshal(resp.Body, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return &rsp, nil
}

// OidbSvc.0x6d6_2
func decodeOIDB6d62Response(_ *QQClient, resp *network.Response) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D6RspBody{}
	if err := proto.Unmarshal(resp.Body, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.DownloadFileRsp.DownloadUrl == nil {
		return nil, errors.New(rsp.DownloadFileRsp.GetClientWording())
	}
	ip := rsp.DownloadFileRsp.GetDownloadIp()
	url := hex.EncodeToString(rsp.DownloadFileRsp.DownloadUrl)
	return fmt.Sprintf("http://%s/ftn_handler/%s/", ip, url), nil
}

func decodeOIDB6d63Response(_ *QQClient, resp *network.Response) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D6RspBody{}
	if err := proto.Unmarshal(resp.Body, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.DeleteFileRsp == nil {
		return "", nil
	}
	return rsp.DeleteFileRsp.GetClientWording(), nil
}

func decodeOIDB6d60Response(_ *QQClient, resp *network.Response) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D6RspBody{}
	if err := proto.Unmarshal(resp.Body, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	return rsp.UploadFileRsp, nil
}

func decodeOIDB6d7Response(_ *QQClient, resp *network.Response) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D7RspBody{}
	if err := proto.Unmarshal(resp.Body, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.CreateFolderRsp != nil && rsp.CreateFolderRsp.GetRetCode() != 0 {
		return nil, errors.Errorf("create folder error: %v", rsp.CreateFolderRsp.GetRetCode())
	}
	if rsp.RenameFolderRsp != nil && rsp.RenameFolderRsp.GetRetCode() != 0 {
		return nil, errors.Errorf("rename folder error: %v", rsp.CreateFolderRsp.GetRetCode())
	}
	if rsp.DeleteFolderRsp != nil && rsp.DeleteFolderRsp.GetRetCode() != 0 {
		return nil, errors.Errorf("delete folder error: %v", rsp.CreateFolderRsp.GetRetCode())
	}
	return nil, nil
}
