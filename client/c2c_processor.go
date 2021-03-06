package client

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"

	"github.com/Mrs4s/MiraiGo/binary"

	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

var c2cDecoders = map[int32]func(*QQClient, *msg.Message, *incomingPacketInfo){
	33: troopAddMemberBroadcastDecoder,
	35: troopSystemMessageDecoder, 36: troopSystemMessageDecoder, 37: troopSystemMessageDecoder,
	45: troopSystemMessageDecoder, 46: troopSystemMessageDecoder, 84: troopSystemMessageDecoder,
	85: troopSystemMessageDecoder, 86: troopSystemMessageDecoder, 87: troopSystemMessageDecoder,
	140: tempSessionDecoder, 141: tempSessionDecoder,
	9: privateMessageDecoder, 10: privateMessageDecoder, 31: privateMessageDecoder,
	79: privateMessageDecoder, 97: privateMessageDecoder, 120: privateMessageDecoder,
	132: privateMessageDecoder, 133: privateMessageDecoder, 166: privateMessageDecoder,
	167: privateMessageDecoder,
	208: privatePttDecoder,
	187: systemMessageDecoder, 188: systemMessageDecoder, 189: systemMessageDecoder,
	190: systemMessageDecoder, 191: systemMessageDecoder,
	529: msgType0x211Decoder,
}

func (c *QQClient) c2cMessageSyncProcessor(rsp *msg.GetMessageResponse, info *incomingPacketInfo) {
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
			if _, ok := c.msgSvcCache.GetAndUpdate(strKey, time.Minute*5); ok {
				c.Debug("c2c msg %v already exists in cache. skip.", pMsg.Head.GetMsgUid())
				continue
			}
			c.msgSvcCache.Add(strKey, "", time.Minute*5)
			if info.Params.bool("init") {
				continue
			}
			if decoder, ok := c2cDecoders[pMsg.Head.GetMsgType()]; ok {
				decoder(c, pMsg, info)
			} else {
				c.Debug("unknown msg type on c2c processor: %v", pMsg.Head.GetMsgType())
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
		seq, pkt := c.buildGetMessageRequestPacket(rsp.GetSyncFlag(), time.Now().Unix())
		_, _ = c.sendAndWait(seq, pkt, info.Params)
	}
}

func privateMessageDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
	switch pMsg.Head.GetC2CCmd() {
	case 11, 175: // friend msg
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
	default:
		c.Debug("unknown c2c cmd on private msg decoder: %v", pMsg.Head.GetC2CCmd())
	}
}

func privatePttDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
	if pMsg.Body == nil || pMsg.Body.RichText == nil || pMsg.Body.RichText.Ptt == nil {
		return
	}
	if len(pMsg.Body.RichText.Ptt.Reserve) != 0 {
		// m := binary.NewReader(pMsg.Body.RichText.Ptt.Reserve[1:]).ReadTlvMap(1)
		// T3 -> timestamp T8 -> voiceType T9 -> voiceLength T10 -> PbReserveStruct
	}
	c.dispatchFriendMessage(c.parsePrivateMessage(pMsg))
}

func tempSessionDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
	if pMsg.Head.C2CTmpMsgHead == nil || pMsg.Body == nil {
		return
	}
	cond := pMsg.Head.GetMsgType() == 529 && pMsg.Head.GetC2CCmd() == 6 ||
		pMsg.Head.GetMsgType() == 141 && pMsg.Head.GetC2CCmd() == 11
	if cond {
		group := c.FindGroup(pMsg.Head.C2CTmpMsgHead.GetGroupCode())
		if group == nil {
			return
		}
		if pMsg.Head.GetFromUin() == c.Uin {
			return
		}
		c.dispatchTempMessage(c.parseTempMessage(pMsg))
	} else if pMsg.Head.GetMsgType() == 140 {
		// FIXME: pMsg.Head.GetMsgType() == 140 && pMsg.Head.GetC2CCmd() == ??
		// 如果有人找到了这个cmd是什么可以帮忙提个PR
		log.Infof("请在MiraiGo开启一个issue, 标题为: tempSession type 140 cmd %v 内容为空, 感谢你的帮忙!", pMsg.Head.GetC2CCmd())
	}
}

func troopAddMemberBroadcastDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
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

func systemMessageDecoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
	_, pkt := c.buildSystemMsgNewFriendPacket()
	_ = c.send(pkt)
}

func troopSystemMessageDecoder(c *QQClient, pMsg *msg.Message, info *incomingPacketInfo) {
	if !info.Params.bool("used_reg_proxy") && pMsg.Head.GetMsgType() != 85 && pMsg.Head.GetMsgType() != 36 {
		c.exceptAndDispatchGroupSysMsg()
	}
	if len(pMsg.Body.GetMsgContent()) == 0 {
		return
	}
	reader := binary.NewReader(pMsg.GetBody().GetMsgContent())
	groupCode := uint32(reader.ReadInt32())
	if info := c.FindGroup(int64(groupCode)); info != nil && pMsg.Head.GetGroupName() != "" && info.Name != pMsg.Head.GetGroupName() {
		c.Debug("group %v name updated. %v -> %v", groupCode, info.Name, pMsg.Head.GetGroupName())
		info.Name = pMsg.Head.GetGroupName()
	}
}

func msgType0x211Decoder(c *QQClient, pMsg *msg.Message, _ *incomingPacketInfo) {
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
