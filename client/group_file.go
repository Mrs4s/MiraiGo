package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/golang/protobuf/proto"
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

func (c *QQClient) GetGroupFileSystem(groupCode int64) (fs *GroupFileSystem, err error) {
	defer func() {
		if pan := recover(); pan != nil {
			c.Error("get group fs error: %v", pan)
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
	url += "?fname=" + hex.EncodeToString([]byte(fileId))
	return url
}

func (fs *GroupFileSystem) Root() ([]*GroupFile, []*GroupFolder, error) {
	return fs.GetFilesByFolder("/")
}

func (fs *GroupFileSystem) GetFilesByFolder(folderId string) ([]*GroupFile, []*GroupFolder, error) {
	var startIndex uint32 = 0
	var files []*GroupFile
	var folders []*GroupFolder
	for {
		i, err := fs.client.sendAndWait(fs.client.buildGroupFileListRequestPacket(fs.GroupCode, folderId, startIndex))
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

func (fs *GroupFileSystem) GetDownloadUrl(file *GroupFile) string {
	return fs.client.GetGroupFileUrl(file.GroupCode, file.FileId, file.BusId)
}

// DeleteFile 删除群文件，需要管理权限.
// 返回错误, 空为删除成功
func (fs *GroupFileSystem) DeleteFile(parentFolderId, fileId string, busId int32) string {
	i, err := fs.client.sendAndWait(fs.client.buildGroupFileDeleteReqPacket(fs.GroupCode, parentFolderId, fileId, busId))
	if err != nil {
		return err.Error()
	}
	return i.(string)
}

// OidbSvc.0x6d8_1
func (c *QQClient) buildGroupFileListRequestPacket(groupCode int64, folderId string, startIndex uint32) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D6D8ReqBody{FileListInfoReq: &oidb.GetFileListReqBody{
		GroupCode:    proto.Uint64(uint64(groupCode)),
		AppId:        proto.Uint32(3),
		FolderId:     &folderId,
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
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x6d8_1", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildGroupFileCountRequestPacket(groupCode int64) (uint16, []byte) {
	seq := c.nextSeq()
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
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x6d8_1", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildGroupFileSpaceRequestPacket(groupCode int64) (uint16, []byte) {
	seq := c.nextSeq()
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
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x6d8_1", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0x6d6_2
func (c *QQClient) buildGroupFileDownloadReqPacket(groupCode int64, fileId string, busId int32) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D6D6ReqBody{
		DownloadFileReq: &oidb.DownloadFileReqBody{
			GroupCode: groupCode,
			AppId:     3,
			BusId:     busId,
			FileId:    fileId,
		},
	}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:     1750,
		ServiceType: 2,
		Bodybuffer:  b,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x6d6_2", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildGroupFileDeleteReqPacket(groupCode int64, parentFolderId, fileId string, busId int32) (uint16, []byte) {
	seq := c.nextSeq()
	body := &oidb.D6D6ReqBody{DeleteFileReq: &oidb.DeleteFileReqBody{
		GroupCode:      groupCode,
		AppId:          3,
		BusId:          busId,
		ParentFolderId: parentFolderId,
		FileId:         fileId,
	}}
	b, _ := proto.Marshal(body)
	req := &oidb.OIDBSSOPkg{
		Command:       1750,
		ServiceType:   3,
		Bodybuffer:    b,
		ClientVersion: "android 8.4.8",
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x6d6_3", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func decodeOIDB6d81Response(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D8RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// OidbSvc.0x6d6_2
func decodeOIDB6d62Response(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D6RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, err
	}
	if rsp.DownloadFileRsp.DownloadUrl == nil {
		return nil, errors.New(rsp.DownloadFileRsp.ClientWording)
	}
	ip := rsp.DownloadFileRsp.DownloadIp
	url := hex.EncodeToString(rsp.DownloadFileRsp.DownloadUrl)
	return fmt.Sprintf("http://%s/ftn_handler/%s/", ip, url), nil
}

func decodeOIDB6d63Response(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D6D6RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, err
	}
	if rsp.DeleteFileRsp == nil {
		return "", nil
	}
	return rsp.DeleteFileRsp.ClientWording, nil
}
