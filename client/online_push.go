package client

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/binary/jce"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb"
	"github.com/Mrs4s/MiraiGo/client/pb/msgtype0x210"
	"github.com/Mrs4s/MiraiGo/client/pb/notify"
	"github.com/Mrs4s/MiraiGo/internal/proto"
)

var msg0x210Decoders = map[int64]func(*QQClient, []byte) error{
	0x8A: msgType0x210Sub8ADecoder, 0x8B: msgType0x210Sub8ADecoder, 0xB3: msgType0x210SubB3Decoder,
	0xD4: msgType0x210SubD4Decoder, 0x27: msgType0x210Sub27Decoder, 0x122: msgType0x210Sub122Decoder,
	0x123: msgType0x210Sub122Decoder, 0x44: msgType0x210Sub44Decoder,
}

// OnlinePush.ReqPush
func decodeOnlinePushReqPacket(c *QQClient, info *network.IncomingPacketInfo, payload []byte) (any, error) {
	request := &jce.RequestPacket{}
	request.ReadFrom(jce.NewJceReader(payload))
	data := &jce.RequestDataVersion2{}
	data.ReadFrom(jce.NewJceReader(request.SBuffer))
	jr := jce.NewJceReader(data.Map["req"]["OnlinePushPack.SvcReqPushMsg"][1:])
	uin := jr.ReadInt64(0)
	msgInfos := jr.ReadPushMessageInfos(2)
	_ = c.sendPacket(c.buildDeleteOnlinePushPacket(uin, 0, nil, info.SequenceId, msgInfos))
	for _, m := range msgInfos {
		k := fmt.Sprintf("%v%v%v", m.MsgSeq, m.MsgTime, m.MsgUid)
		if _, ok := c.onlinePushCache.Get(k); ok {
			continue
		}
		c.onlinePushCache.Add(k, unit{}, time.Second*30)
		// 0x2dc
		if m.MsgType == 732 {
			r := binary.NewReader(m.VMsg)
			groupCode := int64(uint32(r.ReadInt32()))
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

				if target != 0 {
					member := c.FindGroup(groupCode).FindMember(target)
					if t > 0 {
						member.ShutUpTimestamp = time.Now().Add(time.Second * time.Duration(t)).Unix()
					} else {
						member.ShutUpTimestamp = 0
					}
				}

				c.GroupMuteEvent.dispatch(c, &GroupMuteEvent{
					GroupCode:   groupCode,
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
						c.GroupMessageRecalledEvent.dispatch(c, &GroupMessageRecalledEvent{
							GroupCode:   groupCode,
							OperatorUin: b.OptMsgRecall.Uin,
							AuthorUin:   rm.AuthorUin,
							MessageId:   rm.Seq,
							Time:        rm.Time,
						})
					}
				}
				if b.OptGeneralGrayTip != nil {
					c.grayTipProcessor(groupCode, b.OptGeneralGrayTip)
				}
				if b.OptMsgRedTips != nil {
					if b.OptMsgRedTips.LuckyFlag == 1 { // 运气王提示
						c.GroupNotifyEvent.dispatch(c, &GroupRedBagLuckyKingNotifyEvent{
							GroupCode: groupCode,
							Sender:    int64(b.OptMsgRedTips.SenderUin),
							LuckyKing: int64(b.OptMsgRedTips.LuckyUin),
						})
					}
				}
				if b.QqGroupDigestMsg != nil {
					digest := b.QqGroupDigestMsg
					c.GroupDigestEvent.dispatch(c, &GroupDigestEvent{
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
				if b.OptMsgGrayTips != nil {
					c.msgGrayTipProcessor(groupCode, b.OptMsgGrayTips)
				}
			}
		}
		// 0x210
		if m.MsgType == 528 {
			vr := jce.NewJceReader(m.VMsg)
			subType := vr.ReadInt64(0)
			if decoder, ok := msg0x210Decoders[subType]; ok {
				protobuf := vr.ReadBytes(10)
				if err := decoder(c, protobuf); err != nil {
					return nil, errors.Wrap(err, "decode online push 0x210 error")
				}
			} else {
				c.debug("unknown online push 0x210 sub type 0x%v", strconv.FormatInt(subType, 16))
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
			c.FriendMessageRecalledEvent.dispatch(c, &FriendMessageRecalledEvent{
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
	c.NewFriendEvent.dispatch(c, &NewFriendEvent{Friend: frd})
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
		c.GroupLeaveEvent.dispatch(c, &GroupLeaveEvent{Group: g})
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
						g.Name = string(info.Value)
						c.GroupNameUpdatedEvent.dispatch(c, &GroupNameUpdatedEvent{
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
	var sender, receiver int64
	for _, templ := range t.MsgTemplParam {
		if templ.Name == "uin_str1" {
			sender, _ = strconv.ParseInt(templ.Value, 10, 64)
		} else if templ.Name == "uin_str2" {
			receiver, _ = strconv.ParseInt(templ.Value, 10, 64)
		}
	}
	if sender == 0 {
		return nil
	}
	if receiver == 0 {
		receiver = c.Uin
	}
	c.FriendNotifyEvent.dispatch(c, &FriendPokeNotifyEvent{
		Sender:   sender,
		Receiver: receiver,
	})
	return nil
}

func msgType0x210Sub44Decoder(c *QQClient, protobuf []byte) error {
	s44 := pb.Sub44{}
	if err := proto.Unmarshal(protobuf, &s44); err != nil {
		return errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if s44.GroupSyncMsg == nil {
		return nil
	}
	groupJoinLock.Lock()
	defer groupJoinLock.Unlock()
	if s44.GroupSyncMsg.GrpCode == 0 { // member sync
		return errors.New("invalid group code")
	}
	c.debug("syncing members.")
	if group := c.FindGroup(s44.GroupSyncMsg.GrpCode); group != nil {
		group.lock.Lock()
		defer group.lock.Unlock()

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
					c.GroupMemberJoinEvent.dispatch(c, &MemberJoinGroupEvent{
						Group:  group,
						Member: m,
					})
				}
			}
		}
	}
	return nil
}
