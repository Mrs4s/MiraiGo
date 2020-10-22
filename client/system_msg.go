package client

import (
	"github.com/Mrs4s/MiraiGo/client/pb/structmsg"
	"google.golang.org/protobuf/proto"
	"log"
)

type (
	GroupSystemMessages struct {
		InvitedRequests []*GroupInvitedRequest
		JoinRequests    []*UserJoinGroupRequest
	}

	GroupInvitedRequest struct {
		RequestId   int64
		InvitorUin  int64
		InvitorNick string
		GroupCode   int64
		GroupName   string

		client *QQClient
	}

	UserJoinGroupRequest struct {
		RequestId     int64
		Message       string
		RequesterUin  int64
		RequesterNick string
		GroupCode     int64
		GroupName     string

		client *QQClient
	}
)

// ProfileService.Pb.ReqSystemMsgNew.Group
func decodeSystemMsgGroupPacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := structmsg.RspSystemMsgNew{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, err
	}
	if len(rsp.Groupmsgs) == 0 {
		return nil, nil
	}
	ret := &GroupSystemMessages{}
	for _, st := range rsp.Groupmsgs {
		if st.Msg == nil {
			continue
		}
		if st.Msg.SubType == 1 {
			// 处理被邀请入群 或 处理成员入群申请
			switch st.Msg.GroupMsgType {
			case 1: // 成员申请
				ret.JoinRequests = append(ret.JoinRequests, &UserJoinGroupRequest{
					RequestId:     st.MsgSeq,
					Message:       st.Msg.MsgAdditional,
					RequesterUin:  st.ReqUin,
					RequesterNick: st.Msg.ReqUinNick,
					GroupCode:     st.Msg.GroupCode,
					GroupName:     st.Msg.GroupName,
					client:        c,
				})
			case 2: // 被邀请
				ret.InvitedRequests = append(ret.InvitedRequests, &GroupInvitedRequest{
					RequestId:   st.MsgSeq,
					InvitorUin:  st.Msg.ActionUin,
					InvitorNick: st.Msg.ActionUinNick,
					GroupCode:   st.Msg.GroupCode,
					GroupName:   st.Msg.GroupName,
					client:      c,
				})
			default:
				log.Println("unknown group msg:", st)
			}
		} else if st.Msg.SubType == 2 {
			// 被邀请入群, 自动同意, 不需处理
		} else if st.Msg.SubType == 3 {
			// 已被其他管理员处理
		} else if st.Msg.SubType == 5 {
			// 成员退群消息
		} else {
			log.Println("unknown group msg:", st)
		}
	}
	return ret, nil
}
