package client

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Mrs4s/MiraiGo/client/pb/msgtype0x210"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/notify"
)

var msg0x210Decoders = map[int64]func(*QQClient, []byte) error{
	0x8A: msgType0x210Sub8ADecoder, 0x8B: msgType0x210Sub8ADecoder, 0xB3: msgType0x210SubB3Decoder,
	0xD4: msgType0x210SubD4Decoder, 0x27: msgType0x210Sub27Decoder, 0x122: msgType0x210Sub122Decoder,
	0x123: msgType0x210Sub122Decoder, 0x44: msgType0x210Sub44Decoder,
}

// OnlinePush.ReqPush
func decodeOnlinePushReqPacket(c *QQClient, info *incomingPacketInfo, payload []byte) (interface{}, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	jr := jce.NewJceReader(data.Map["req"]["OnlinePushPack.SvcReqPushMsg"][1:])
	msgInfos := []jce.PushMessageInfo{}
	uin := jr.ReadInt64(0)
	jr.ReadSlice(&msgInfos, 2)
	_ = c.send(c.buildDeleteOnlinePushPacket(uin, 0, nil, info.SequenceId, msgInfos))
	for _, m := range msgInfos {
		k := fmt.Sprintf("%v%v%v", m.MsgSeq, m.MsgTime, m.MsgUid)
		if _, ok := c.onlinePushCache.Get(k); ok {
			continue
		}
		c.onlinePushCache.Add(k, "", time.Second*30)
		// 0x2dc
		if m.MsgType == 732 {
			r := binary.NewReader(m.VMsg)
			groupID := int64(uint32(r.ReadInt32()))
			iType := r.ReadByte()
			r.ReadByte()
			switch iType {
			case 0x0c: // 群内禁言
				operator := int64(uint32(r.ReadInt32()))
				if operator == c.Uin {
					continue
				}
				r.ReadBytes(6)
				target := int64(uint32(r.ReadInt32()))
				t := r.ReadInt32()
				c.dispatchGroupMuteEvent(&GroupMuteEvent{
					GroupCode:   groupID,
					OperatorUin: operator,
					TargetUin:   target,
					Time:        t,
				})
			case 0x10, 0x11, 0x14, 0x15: // group notify msg
				r.ReadByte()
				b := notify.NotifyMsgBody{}
				_ = proto.Unmarshal(r.ReadAvailable(), &b)
				if b.OptMsgRecall != nil {
					for _, rm := range b.OptMsgRecall.RecalledMsgList {
						if rm.MsgType == 2 {
							continue
						}
						c.dispatchGroupMessageRecalledEvent(&GroupMessageRecalledEvent{
							GroupCode:   groupID,
							OperatorUin: b.OptMsgRecall.Uin,
							AuthorUin:   rm.AuthorUin,
							MessageId:   rm.Seq,
							Time:        rm.Time,
						})
					}
				}
				if b.OptGeneralGrayTip != nil {
					c.grayTipProcessor(groupID, b.OptGeneralGrayTip)
				}
				if b.OptMsgRedTips != nil {
					if b.OptMsgRedTips.LuckyFlag == 1 { // 运气王提示
						c.dispatchGroupNotifyEvent(&GroupRedBagLuckyKingNotifyEvent{
							GroupCode: groupID,
							Sender:    int64(b.OptMsgRedTips.SenderUin),
							LuckyKing: int64(b.OptMsgRedTips.LuckyUin),
						})
					}
				}
				if b.QqGroupDigestMsg != nil {
					digest := b.QqGroupDigestMsg
					c.dispatchGroupDigestEvent(&GroupDigestEvent{
						GroupCode:         int64(digest.GroupCode),
						MessageID:         int32(digest.Seq),
						InternalMessageID: int32(digest.Random),
						OperationType:     digest.OpType,
						OperateTime:       digest.OpTime,
						SenderUin:         int64(digest.Sender),
						OperatorUin:       int64(digest.DigestOper),
						SenderNick:        string(digest.SenderNick),
						OperatorNick:      string(digest.OperNick),
					})
				}
			}
		}
		// 0x210
		if m.MsgType == 528 {
			vr := jce.NewJceReader(m.VMsg)
			subType := vr.ReadInt64(0)
			protobuf := vr.ReadAny(10).([]byte)
			if decoder, ok := msg0x210Decoders[subType]; ok {
				if err := decoder(c, protobuf); err != nil {
					return nil, errors.Wrap(err, "decode online push 0x210 error")
				}
			} else {
				c.Debug("unknown online push 0x210 sub type 0x%v", strconv.FormatInt(subType, 16))
			}
		}
	}
	return nil, nil
}

func msgType0x210Sub8ADecoder(c *QQClient, protobuf []byte) error {
	s8a := pb.Sub8A{}
	if err := proto.Unmarshal(protobuf, &s8a); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	for _, m := range s8a.MsgInfo {
		if m.ToUin == c.Uin {
			c.dispatchFriendMessageRecalledEvent(&FriendMessageRecalledEvent{
				FriendUin: m.FromUin,
				MessageId: m.MsgSeq,
				Time:      m.MsgTime,
			})
		}
	}
	return nil
}

func msgType0x210SubB3Decoder(c *QQClient, protobuf []byte) error {
	b3 := pb.SubB3{}
	if err := proto.Unmarshal(protobuf, &b3); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	frd := &FriendInfo{
		Uin:      b3.MsgAddFrdNotify.Uin,
		Nickname: b3.MsgAddFrdNotify.Nick,
	}
	c.FriendList = append(c.FriendList, frd)
	c.dispatchNewFriendEvent(&NewFriendEvent{Friend: frd})
	return nil
}

func msgType0x210SubD4Decoder(c *QQClient, protobuf []byte) error {
	d4 := pb.SubD4{}
	if err := proto.Unmarshal(protobuf, &d4); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	groupLeaveLock.Lock()
	if g := c.FindGroup(d4.Uin); g != nil {
		if err := c.ReloadGroupList(); err != nil {
			groupLeaveLock.Unlock()
			return err
		}
		c.dispatchLeaveGroupEvent(&GroupLeaveEvent{Group: g})
	}
	groupLeaveLock.Unlock()
	return nil
}

func msgType0x210Sub27Decoder(c *QQClient, protobuf []byte) error {
	s27 := msgtype0x210.SubMsg0X27Body{}
	if err := proto.Unmarshal(protobuf, &s27); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	for _, m := range s27.ModInfos {
		if m.ModGroupProfile != nil {
			for _, info := range m.ModGroupProfile.GroupProfileInfos {
				if info.GetField() == 1 {
					if g := c.FindGroup(int64(m.ModGroupProfile.GetGroupCode())); g != nil {
						old := g.Name
						g.Name = string(info.GetValue())
						c.dispatchGroupNameUpdatedEvent(&GroupNameUpdatedEvent{
							Group:       g,
							OldName:     old,
							NewName:     g.Name,
							OperatorUin: int64(m.ModGroupProfile.GetCmdUin()),
						})
					}
				}
			}
		}
		if m.DelFriend != nil {
			frdUin := m.DelFriend.Uins[0]
			if frd := c.FindFriend(int64(frdUin)); frd != nil {
				if err := c.ReloadFriendList(); err != nil {
					return errors.Wrap(err, "failed to reload friend list")
				}
			}
		}
	}
	return nil
}

func msgType0x210Sub122Decoder(c *QQClient, protobuf []byte) error {
	t := &notify.GeneralGrayTipInfo{}
	_ = proto.Unmarshal(protobuf, t)
	var sender int64
	for _, templ := range t.MsgTemplParam {
		if templ.Name == "uin_str1" {
			sender, _ = strconv.ParseInt(templ.Value, 10, 64)
		}
	}
	if sender == 0 {
		return nil
	}
	c.dispatchFriendNotifyEvent(&FriendPokeNotifyEvent{
		Sender:   sender,
		Receiver: c.Uin,
	})
	return nil
}

func msgType0x210Sub44Decoder(c *QQClient, protobuf []byte) error {
	s44 := pb.Sub44{}
	if err := proto.Unmarshal(protobuf, &s44); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if s44.GroupSyncMsg != nil {
		func() {
			groupJoinLock.Lock()
			defer groupJoinLock.Unlock()
			if s44.GroupSyncMsg.GetGrpCode() != 0 { // member sync
				c.Debug("syncing members.")
				if group := c.FindGroup(s44.GroupSyncMsg.GetGrpCode()); group != nil {
					group.Update(func(_ *GroupInfo) {
						var lastJoinTime int64 = 0
						for _, m := range group.Members {
							if lastJoinTime < m.JoinTime {
								lastJoinTime = m.JoinTime
							}
						}
						if newMem, err := c.GetGroupMembers(group); err == nil {
							group.Members = newMem
							for _, m := range newMem {
								if lastJoinTime < m.JoinTime {
									go c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
										Group:  group,
										Member: m,
									})
								}
							}
						}
					})
				}
			}
		}()
	}
	return nil
}
