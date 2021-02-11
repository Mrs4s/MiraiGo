package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb/structmsg"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"google.golang.org/protobuf/proto"
)

type (
	GroupSystemMessages struct {
		InvitedRequests []*GroupInvitedRequest  `json:"invited_requests"`
		JoinRequests    []*UserJoinGroupRequest `json:"join_requests"`
	}

	GroupInvitedRequest struct {
		RequestId   int64  `json:"request_id"`
		InvitorUin  int64  `json:"invitor_uin"`
		InvitorNick string `json:"invitor_nick"`
		GroupCode   int64  `json:"group_id"`
		GroupName   string `json:"group_name"`

		Checked bool  `json:"checked"`
		Actor   int64 `json:"actor"`

		client *QQClient
	}

	UserJoinGroupRequest struct {
		RequestId     int64  `json:"request_id"`
		Message       string `json:"message"`
		RequesterUin  int64  `json:"requester_uin"`
		RequesterNick string `json:"requester_nick"`
		GroupCode     int64  `json:"group_id"`
		GroupName     string `json:"group_name"`

		Checked    bool  `json:"checked"`
		Actor      int64 `json:"actor"`
		Suspicious bool  `json:"suspicious"`

		client *QQClient
	}
)

func (c *QQClient) GetGroupSystemMessages() (*GroupSystemMessages, error) {
	i, err := c.sendAndWait(c.buildSystemMsgNewGroupPacket(false))
	if err != nil {
		return nil, err
	}
	msg := i.(*GroupSystemMessages)
	i, err = c.sendAndWait(c.buildSystemMsgNewGroupPacket(true))
	if err != nil {
		return nil, err
	}
	msg.InvitedRequests = append(msg.InvitedRequests, i.(*GroupSystemMessages).InvitedRequests...)
	msg.JoinRequests = append(msg.JoinRequests, i.(*GroupSystemMessages).JoinRequests...)
	return msg, nil
}

func (c *QQClient) exceptAndDispatchGroupSysMsg() {
	if c.groupSysMsgCache == nil {
		c.Error("warning: groupSysMsgCache is nil")
		c.groupSysMsgCache, _ = c.GetGroupSystemMessages()
		return
	}
	joinExists := func(req int64) bool {
		for _, msg := range c.groupSysMsgCache.JoinRequests {
			if req == msg.RequestId {
				return true
			}
		}
		return false
	}
	invExists := func(req int64) bool {
		for _, msg := range c.groupSysMsgCache.InvitedRequests {
			if req == msg.RequestId {
				return true
			}
		}
		return false
	}
	msgs, err := c.GetGroupSystemMessages()
	if err != nil {
		return
	}
	for _, msg := range msgs.JoinRequests {
		if !joinExists(msg.RequestId) {
			c.dispatchJoinGroupRequest(msg)
		}
	}
	for _, msg := range msgs.InvitedRequests {
		if !invExists(msg.RequestId) {
			c.dispatchGroupInvitedEvent(msg)
		}
	}
	c.groupSysMsgCache = msgs
}

// ProfileService.Pb.ReqSystemMsgNew.Group
func (c *QQClient) buildSystemMsgNewGroupPacket(suspicious bool) (uint16, []byte) {
	seq := c.nextSeq()
	req := &structmsg.ReqSystemMsgNew{
		MsgNum:    100,
		Version:   1000,
		Checktype: 3,
		Flag: &structmsg.FlagInfo{
			GrpMsgKickAdmin:                   1,
			GrpMsgHiddenGrp:                   1,
			GrpMsgWordingDown:                 1,
			GrpMsgGetOfficialAccount:          1,
			GrpMsgGetPayInGroup:               1,
			FrdMsgDiscuss2ManyChat:            1,
			GrpMsgNotAllowJoinGrpInviteNotFrd: 1,
			FrdMsgNeedWaitingMsg:              1,
			FrdMsgUint32NeedAllUnreadMsg:      1,
			GrpMsgNeedAutoAdminWording:        1,
			GrpMsgGetTransferGroupMsgFlag:     1,
			GrpMsgGetQuitPayGroupMsgFlag:      1,
			GrpMsgSupportInviteAutoJoin:       1,
			GrpMsgMaskInviteAutoJoin:          1,
			GrpMsgGetDisbandedByAdmin:         1,
			GrpMsgGetC2CInviteJoinGroup:       1,
		},
		FriendMsgTypeFlag: 1,
		ReqMsgType: func() int32 {
			if suspicious {
				return 2
			}
			return 1
		}(),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.Pb.ReqSystemMsgNew.Group", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ProfileService.Pb.ReqSystemMsgAction.Group
func (c *QQClient) buildSystemMsgGroupActionPacket(reqId, requester, group int64, msgType int32, isInvite, accept, block bool, reason string) (uint16, []byte) {
	seq := c.nextSeq()
	req := &structmsg.ReqSystemMsgAction{
		MsgType: msgType,
		MsgSeq:  reqId,
		ReqUin:  requester,
		SubType: 1,
		SrcId:   3,
		SubSrcId: func() int32 {
			if isInvite {
				return 10016
			}
			return 31
		}(),
		GroupMsgType: func() int32 {
			if isInvite {
				return 2
			}
			return 1
		}(),
		ActionInfo: &structmsg.SystemMsgActionInfo{
			Type: func() int32 {
				if accept {
					return 11
				}
				return 12
			}(),
			GroupCode: group,
			Blacklist: block,
			Msg:       reason,
			Sig:       EmptyBytes,
		},
		Language: 1000,
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.Pb.ReqSystemMsgAction.Group", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ProfileService.Pb.ReqSystemMsgAction.Friend
func (c *QQClient) buildSystemMsgFriendActionPacket(reqId, requester int64, accept bool) (uint16, []byte) {
	seq := c.nextSeq()
	req := &structmsg.ReqSystemMsgAction{
		MsgType:  1,
		MsgSeq:   reqId,
		ReqUin:   requester,
		SubType:  1,
		SrcId:    6,
		SubSrcId: 7,
		ActionInfo: &structmsg.SystemMsgActionInfo{
			Type: func() int32 {
				if accept {
					return 2
				}
				return 3
			}(),
			Blacklist:    false,
			AddFrdSNInfo: &structmsg.AddFrdSNInfo{},
		},
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "ProfileService.Pb.ReqSystemMsgAction.Friend", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// ProfileService.Pb.ReqSystemMsgNew.Group
func decodeSystemMsgGroupPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := structmsg.RspSystemMsgNew{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	ret := &GroupSystemMessages{}
	for _, st := range rsp.Groupmsgs {
		if st.Msg == nil {
			continue
		}
		switch st.Msg.SubType {
		case 1, 2: // 处理被邀请入群 或 处理成员入群申请
			switch st.Msg.GroupMsgType {
			case 1: // 成员申请
				ret.JoinRequests = append(ret.JoinRequests, &UserJoinGroupRequest{
					RequestId:     st.MsgSeq,
					Message:       st.Msg.MsgAdditional,
					RequesterUin:  st.ReqUin,
					RequesterNick: st.Msg.ReqUinNick,
					GroupCode:     st.Msg.GroupCode,
					GroupName:     st.Msg.GroupName,
					Checked:       st.Msg.SubType == 2,
					Actor:         st.Msg.ActorUin,
					Suspicious:    len(st.Msg.WarningTips) > 0,
					client:        c,
				})
			case 2: // 被邀请
				ret.InvitedRequests = append(ret.InvitedRequests, &GroupInvitedRequest{
					RequestId:   st.MsgSeq,
					InvitorUin:  st.Msg.ActionUin,
					InvitorNick: st.Msg.ActionUinNick,
					GroupCode:   st.Msg.GroupCode,
					GroupName:   st.Msg.GroupName,
					Checked:     st.Msg.SubType == 2,
					Actor:       st.Msg.ActorUin,
					client:      c,
				})
			case 22: // 群员邀请其他人
				ret.JoinRequests = append(ret.JoinRequests, &UserJoinGroupRequest{
					RequestId:     st.MsgSeq,
					Message:       st.Msg.MsgAdditional,
					RequesterUin:  st.ReqUin,
					RequesterNick: st.Msg.ReqUinNick,
					GroupCode:     st.Msg.GroupCode,
					GroupName:     st.Msg.GroupName,
					Checked:       st.Msg.SubType == 2,
					Actor:         st.Msg.ActorUin,
					Suspicious:    len(st.Msg.WarningTips) > 0,
					client:        c,
				})
			default:
				c.Error("unknown group message type: %v", st.Msg.GroupMsgType)
			}
		case 3: // ?
		case 5: // 自身状态变更(管理员/加群退群)
		default:
			c.Error("unknown group msg: %v", st.Msg.SubType)
		}
	}
	return ret, nil
}
