package client

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// com.tencent.mobileqq.app.QQMessageFacadeConfig::start
var c2cDecoders = map[int32]func(*QQClient, *msg.Message){
	33: troopAddMemberBroadcastDecoder,
	35: troopSystemMessageDecoder, 36: troopSystemMessageDecoder, 37: troopSystemMessageDecoder,
	45: troopSystemMessageDecoder, 46: troopSystemMessageDecoder, 84: troopSystemMessageDecoder,
	85: troopSystemMessageDecoder, 86: troopSystemMessageDecoder, 87: troopSystemMessageDecoder,
	140: tempSessionDecoder, 141: tempSessionDecoder,
	166: privateMessageDecoder, 208: privateMessageDecoder,
	187: systemMessageDecoder, 188: systemMessageDecoder, 189: systemMessageDecoder,
	190: systemMessageDecoder, 191: systemMessageDecoder,
	529: msgType0x211Decoder,
}

func (c *QQClient) c2cMessageSyncProcessor(rsp *msg.GetMessageResponse) {
	c.syncCookie = rsp.SyncCookie
	c.pubAccountCookie = rsp.PubAccountCookie
	c.msgCtrlBuf = rsp.MsgCtrlBuf
	if rsp.UinPairMsgs == nil {
		return
	}
	var delItems []*pb.MessageItem
	for _, pairMsg := range rsp.UinPairMsgs {
		for _, pMsg := range pairMsg.Messages {
			// delete message
			delItem := &pb.MessageItem{
				FromUin: pMsg.Head.GetFromUin(),
				ToUin:   pMsg.Head.GetToUin(),
				MsgType: pMsg.Head.GetMsgType(),
				MsgSeq:  pMsg.Head.GetMsgSeq(),
				MsgUid:  pMsg.Head.GetMsgUid(),
			}
			delItems = append(delItems, delItem)
			if pMsg.Head.GetToUin() != c.Uin {
				continue
			}
			if (int64(pairMsg.GetLastReadTime()) & 4294967295) > int64(pMsg.Head.GetMsgTime()) {
				continue
			}
			strKey := fmt.Sprintf("%d%d%d%d", pMsg.Head.GetFromUin(), pMsg.Head.GetToUin(), pMsg.Head.GetMsgSeq(), pMsg.Head.GetMsgUid())
			if _, ok := c.msgSvcCache.Get(strKey); ok {
				continue
			}
			c.msgSvcCache.Add(strKey, "", time.Minute)
			if decoder, ok := c2cDecoders[pMsg.Head.GetMsgType()]; ok {
				decoder(c, pMsg)
			}
			/*
				switch pMsg.Head.GetMsgType() {
				case 33: // 加群同步
					func() {
						groupJoinLock.Lock()
						defer groupJoinLock.Unlock()
						group := c.FindGroupByUin(pMsg.Head.GetFromUin())
						if pMsg.Head.GetAuthUin() == c.Uin {
							if group == nil && c.ReloadGroupList() == nil {
								c.dispatchJoinGroupEvent(c.FindGroupByUin(pMsg.Head.GetFromUin()))
							}
						} else {
							if group != nil && group.FindMember(pMsg.Head.GetAuthUin()) == nil {
								mem, err := c.getMemberInfo(group.Code, pMsg.Head.GetAuthUin())
								if err != nil {
									c.Debug("error to fetch new member info: %v", err)
									return
								}
								group.Update(func(info *GroupInfo) {
									info.Members = append(info.Members, mem)
								})
								c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
									Group:  group,
									Member: mem,
								})
							}
						}
					}()
				case 84, 87:
					c.exceptAndDispatchGroupSysMsg()
				case 141: // 临时会话
					if pMsg.Head.C2CTmpMsgHead == nil {
						continue
					}
					group := c.FindGroupByUin(pMsg.Head.C2CTmpMsgHead.GetGroupUin())
					if group == nil {
						continue
					}
					if pMsg.Head.GetFromUin() == c.Uin {
						continue
					}
					c.dispatchTempMessage(c.parseTempMessage(pMsg))
				case 166, 208: // 好友消息
					if pMsg.Head.GetFromUin() == c.Uin {
						for {
							frdSeq := atomic.LoadInt32(&c.friendSeq)
							if frdSeq < pMsg.Head.GetMsgSeq() {
								if atomic.CompareAndSwapInt32(&c.friendSeq, frdSeq, pMsg.Head.GetMsgSeq()) {
									break
								}
							} else {
								break
							}
						}
					}
					if pMsg.Body.RichText == nil || pMsg.Body.RichText.Elems == nil {
						continue
					}
					c.dispatchFriendMessage(c.parsePrivateMessage(pMsg))
				case 187:
					_, pkt := c.buildSystemMsgNewFriendPacket()
					_ = c.send(pkt)
				case 529:
					sub4 := msg.SubMsgType0X4Body{}
					if err := proto.Unmarshal(pMsg.Body.MsgContent, &sub4); err != nil {
						err = errors.Wrap(err, "unmarshal sub msg 0x4 error")
						c.Error("unmarshal sub msg 0x4 error: %v", err)
						continue
					}
					if sub4.NotOnlineFile != nil {
						rsp, err := c.sendAndWait(c.buildOfflineFileDownloadRequestPacket(sub4.NotOnlineFile.FileUuid)) // offline_file.go
						if err != nil {
							continue
						}
						c.dispatchOfflineFileEvent(&OfflineFileEvent{
							FileName:    string(sub4.NotOnlineFile.FileName),
							FileSize:    sub4.NotOnlineFile.GetFileSize(),
							Sender:      pMsg.Head.GetFromUin(),
							DownloadUrl: rsp.(string),
						})
					}
				}
			*/
		}
	}
	_, _ = c.sendAndWait(c.buildDeleteMessageRequestPacket(delItems))
	if rsp.GetSyncFlag() != msg.SyncFlag_STOP {
		c.Debug("continue sync with flag: %v", rsp.SyncFlag.String())
		_, _ = c.sendAndWait(c.buildGetMessageRequestPacket(rsp.GetSyncFlag(), time.Now().Unix()))
	}
}

func privateMessageDecoder(c *QQClient, pMsg *msg.Message) {
	if pMsg.Head.GetFromUin() == c.Uin {
		for {
			frdSeq := atomic.LoadInt32(&c.friendSeq)
			if frdSeq < pMsg.Head.GetMsgSeq() {
				if atomic.CompareAndSwapInt32(&c.friendSeq, frdSeq, pMsg.Head.GetMsgSeq()) {
					break
				}
			} else {
				break
			}
		}
	}
	if pMsg.Body.RichText == nil || pMsg.Body.RichText.Elems == nil {
		return
	}
	c.dispatchFriendMessage(c.parsePrivateMessage(pMsg))
}

func tempSessionDecoder(c *QQClient, pMsg *msg.Message) {
	if pMsg.Head.C2CTmpMsgHead == nil {
		return
	}
	group := c.FindGroupByUin(pMsg.Head.C2CTmpMsgHead.GetGroupUin())
	if group == nil {
		return
	}
	if pMsg.Head.GetFromUin() == c.Uin {
		return
	}
	c.dispatchTempMessage(c.parseTempMessage(pMsg))
}

func troopAddMemberBroadcastDecoder(c *QQClient, pMsg *msg.Message) {
	groupJoinLock.Lock()
	defer groupJoinLock.Unlock()
	group := c.FindGroupByUin(pMsg.Head.GetFromUin())
	if pMsg.Head.GetAuthUin() == c.Uin {
		if group == nil && c.ReloadGroupList() == nil {
			c.dispatchJoinGroupEvent(c.FindGroupByUin(pMsg.Head.GetFromUin()))
		}
	} else {
		if group != nil && group.FindMember(pMsg.Head.GetAuthUin()) == nil {
			mem, err := c.getMemberInfo(group.Code, pMsg.Head.GetAuthUin())
			if err != nil {
				c.Debug("error to fetch new member info: %v", err)
				return
			}
			group.Update(func(info *GroupInfo) {
				info.Members = append(info.Members, mem)
			})
			c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
				Group:  group,
				Member: mem,
			})
		}
	}
}

func systemMessageDecoder(c *QQClient, pMsg *msg.Message) {
	_, pkt := c.buildSystemMsgNewFriendPacket()
	_ = c.send(pkt)
}

func troopSystemMessageDecoder(c *QQClient, pMsg *msg.Message) {
	c.exceptAndDispatchGroupSysMsg()
}

func msgType0x211Decoder(c *QQClient, pMsg *msg.Message) {
	sub4 := msg.SubMsgType0X4Body{}
	if err := proto.Unmarshal(pMsg.Body.MsgContent, &sub4); err != nil {
		err = errors.Wrap(err, "unmarshal sub msg 0x4 error")
		c.Error("unmarshal sub msg 0x4 error: %v", err)
		return
	}
	if sub4.NotOnlineFile != nil {
		rsp, err := c.sendAndWait(c.buildOfflineFileDownloadRequestPacket(sub4.NotOnlineFile.FileUuid)) // offline_file.go
		if err != nil {
			return
		}
		c.dispatchOfflineFileEvent(&OfflineFileEvent{
			FileName:    string(sub4.NotOnlineFile.FileName),
			FileSize:    sub4.NotOnlineFile.GetFileSize(),
			Sender:      pMsg.Head.GetFromUin(),
			DownloadUrl: rsp.(string),
		})
	}
}
