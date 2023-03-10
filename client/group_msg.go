package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/pb/cmd0x388"
	"github.com/Mrs4s/MiraiGo/client/pb/longmsg"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/multimsg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/utils"
)

func init() {
	decoders["OnlinePush.PbPushGroupMsg"] = decodeGroupMessagePacket
	decoders["MessageSvc.PbSendMsg"] = decodeMsgSendResponse
	decoders["MessageSvc.PbGetGroupMsg"] = decodeGetGroupMsgResponse
	decoders["OidbSvc.0x8a7_0"] = decodeAtAllRemainResponse
	decoders["OidbSvc.0xeac_1"] = decodeEssenceMsgResponse
	decoders["OidbSvc.0xeac_2"] = decodeEssenceMsgResponse
}

// SendGroupMessage 发送群消息
func (c *QQClient) SendGroupMessage(groupCode int64, m *message.SendingMessage) *message.GroupMessage {
	imgCount := 0
	for _, e := range m.Elements {
		switch e.Type() {
		case message.Image:
			imgCount++
		}
	}
	msgLen := message.EstimateLength(m.Elements)
	if msgLen > message.MaxMessageSize || imgCount > 50 {
		return nil
	}
	return c.sendGroupMessage(groupCode, false, m)
}

// SendGroupForwardMessage 发送群合并转发消息
func (c *QQClient) SendGroupForwardMessage(groupCode int64, m *message.ForwardElement) *message.GroupMessage {
	return c.sendGroupMessage(groupCode, true,
		&message.SendingMessage{Elements: []message.IMessageElement{m}},
	)
}

// GetGroupMessages 从服务器获取历史信息
func (c *QQClient) GetGroupMessages(groupCode, beginSeq, endSeq int64) ([]*message.GroupMessage, error) {
	seq, pkt := c.buildGetGroupMsgRequest(groupCode, beginSeq, endSeq)
	i, err := c.sendAndWait(seq, pkt, network.RequestParams{"raw": false})
	if err != nil {
		return nil, err
	}
	return i.([]*message.GroupMessage), nil
}

func (c *QQClient) GetAtAllRemain(groupCode int64) (*AtAllRemainInfo, error) {
	i, err := c.sendAndWait(c.buildAtAllRemainRequestPacket(groupCode))
	if err != nil {
		return nil, err
	}
	return i.(*AtAllRemainInfo), nil
}

func (c *QQClient) sendGroupMessage(groupCode int64, forward bool, m *message.SendingMessage) *message.GroupMessage {
	eid := utils.RandomString(6)
	mr := int32(rand.Uint32())
	ch := make(chan int32, 1)
	c.onGroupMessageReceipt(eid, func(c *QQClient, e *groupMessageReceiptEvent) {
		if e.Rand == mr {
			ch <- e.Seq
		}
	})
	defer c.onGroupMessageReceipt(eid)
	imgCount := 0
	serviceFlag := true
	for _, e := range m.Elements {
		switch e.Type() {
		case message.Image:
			imgCount++
		case message.Forward:
			forward = true
			fallthrough
		case message.Reply, message.Voice, message.Service:
			serviceFlag = false
		}
	}
	if !forward && serviceFlag && c.UseFragmentMessage && (imgCount > 1 || message.EstimateLength(m.Elements) > 100) {
		div := int32(rand.Uint32())
		fragmented := m.ToFragmented()
		for i, elems := range fragmented {
			_, pkt := c.buildGroupSendingPacket(groupCode, mr, int32(len(fragmented)), int32(i), div, forward, elems)
			_ = c.sendPacket(pkt)
		}
	} else {
		_, pkt := c.buildGroupSendingPacket(groupCode, mr, 1, 0, 0, forward, m.Elements)
		_ = c.sendPacket(pkt)
	}
	var mid int32
	ret := &message.GroupMessage{
		Id:         -1,
		InternalId: mr,
		GroupCode:  groupCode,
		Sender: &message.Sender{
			Uin:      c.Uin,
			Nickname: c.Nickname,
			IsFriend: true,
		},
		Time:     int32(time.Now().Unix()),
		Elements: m.Elements,
	}
	select {
	case mid = <-ch:
		ret.Id = mid
		return ret
	case <-time.After(time.Second * 5):
		if g, err := c.GetGroupInfo(groupCode); err == nil {
			if history, err := c.GetGroupMessages(groupCode, g.LastMsgSeq-10, g.LastMsgSeq+1); err == nil {
				for _, m := range history {
					if m.InternalId == mr {
						return m
					}
				}
			}
		}
		return ret
	}
}

func (c *QQClient) multiMsgApplyUp(groupCode int64, data []byte, hash []byte, buType int32) (*multimsg.MultiMsgApplyUpRsp, []byte, error) {
	i, err := c.sendAndWait(c.buildMultiApplyUpPacket(data, hash, buType, utils.ToGroupUin(groupCode)))
	if err != nil {
		return nil, nil, err
	}
	rsp := i.(*multimsg.MultiMsgApplyUpRsp)
	body, _ := proto.Marshal(&longmsg.LongReqBody{
		Subcmd:       1,
		TermType:     5,
		PlatformType: 9,
		MsgUpReq: []*longmsg.LongMsgUpReq{
			{
				MsgType:    3,
				DstUin:     utils.ToGroupUin(groupCode),
				MsgContent: data,
				StoreType:  2,
				MsgUkey:    rsp.MsgUkey,
			},
		},
	})
	return rsp, body, nil
}

// MessageSvc.PbSendMsg
func (c *QQClient) buildGroupSendingPacket(groupCode int64, r, pkgNum, pkgIndex, pkgDiv int32, forward bool, m []message.IMessageElement) (uint16, []byte) {
	var ptt *message.GroupVoiceElement
	if len(m) > 0 {
		if p, ok := m[0].(*message.GroupVoiceElement); ok {
			ptt = p
			m = []message.IMessageElement{}
		}
	}
	req := &msg.SendMessageRequest{
		RoutingHead: &msg.RoutingHead{Grp: &msg.Grp{GroupCode: proto.Some(groupCode)}},
		ContentHead: &msg.ContentHead{PkgNum: proto.Some(pkgNum), PkgIndex: proto.Some(pkgIndex), DivSeq: proto.Some(pkgDiv)},
		MsgBody: &msg.MessageBody{
			RichText: &msg.RichText{
				Elems: message.ToProtoElems(m, true),
				Ptt: func() *msg.Ptt {
					if ptt != nil {
						return ptt.Ptt
					}
					return nil
				}(),
			},
		},
		MsgSeq:     proto.Int32(c.nextGroupSeq()),
		MsgRand:    proto.Some(r),
		SyncCookie: EmptyBytes,
		MsgVia:     proto.Int32(1),
		MsgCtrl: func() *msg.MsgCtrl {
			if forward {
				return &msg.MsgCtrl{MsgFlag: proto.Int32(4)}
			}
			return nil
		}(),
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("MessageSvc.PbSendMsg", payload)
}

func (c *QQClient) buildGetGroupMsgRequest(groupCode, beginSeq, endSeq int64) (uint16, []byte) {
	req := &msg.GetGroupMsgReq{
		GroupCode:   proto.Uint64(uint64(groupCode)),
		BeginSeq:    proto.Uint64(uint64(beginSeq)),
		EndSeq:      proto.Uint64(uint64(endSeq)),
		PublicGroup: proto.Bool(false),
	}
	payload, _ := proto.Marshal(req)
	return c.uniPacket("MessageSvc.PbGetGroupMsg", payload)
}

func (c *QQClient) buildAtAllRemainRequestPacket(groupCode int64) (uint16, []byte) {
	payload := c.packOIDBPackageProto(2215, 0, &oidb.D8A7ReqBody{
		SubCmd:                    proto.Uint32(1),
		LimitIntervalTypeForUin:   proto.Uint32(2),
		LimitIntervalTypeForGroup: proto.Uint32(1),
		Uin:                       proto.Uint64(uint64(c.Uin)),
		GroupCode:                 proto.Uint64(uint64(groupCode)),
	})
	return c.uniPacket("OidbSvc.0x8a7_0", payload)
}

// OnlinePush.PbPushGroupMsg
func decodeGroupMessagePacket(c *QQClient, packet *network.Packet) (any, error) {
	pkt := msg.PushMessagePacket{}
	err := proto.Unmarshal(packet.Payload, &pkt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if pkt.Message.Head.FromUin.Unwrap() == c.Uin {
		c.dispatchGroupMessageReceiptEvent(&groupMessageReceiptEvent{
			Rand: pkt.Message.Body.RichText.Attr.Random.Unwrap(),
			Seq:  pkt.Message.Head.MsgSeq.Unwrap(),
			Msg:  c.parseGroupMessage(pkt.Message),
		})
	}
	if pkt.Message.Content != nil && pkt.Message.Content.PkgNum.Unwrap() > 1 {
		seq := pkt.Message.Content.DivSeq.Unwrap()
		builder := c.messageBuilder(pkt.Message.Content.DivSeq.Unwrap())
		builder.append(pkt.Message)
		if builder.len() >= pkt.Message.Content.PkgNum.Unwrap() {
			c.msgBuilders.Delete(seq)
			if pkt.Message.Head.FromUin.Unwrap() == c.Uin {
				c.SelfGroupMessageEvent.dispatch(c, c.parseGroupMessage(builder.build()))
			} else {
				c.GroupMessageEvent.dispatch(c, c.parseGroupMessage(builder.build()))
			}
		}
		return nil, nil
	}
	if pkt.Message.Head.FromUin.Unwrap() == c.Uin {
		c.SelfGroupMessageEvent.dispatch(c, c.parseGroupMessage(pkt.Message))
	} else {
		c.GroupMessageEvent.dispatch(c, c.parseGroupMessage(pkt.Message))
	}
	return nil, nil
}

func decodeMsgSendResponse(c *QQClient, pkt *network.Packet) (any, error) {
	rsp := msg.SendMessageResponse{}
	if err := proto.Unmarshal(pkt.Payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	switch rsp.Result.Unwrap() {
	case 0: // OK.
	case 46:
		c.error("sendPacket msg error: 需要使用安全设备验证")
	case 55:
		c.error("sendPacket msg error: %v Bot has been blocked ta.'s content", rsp.Result.Unwrap())
	default:
		c.error("sendPacket msg error: %v %v", rsp.Result.Unwrap(), rsp.ErrMsg.Unwrap())
	}
	return nil, nil
}

func decodeGetGroupMsgResponse(c *QQClient, pkt *network.Packet) (any, error) {
	rsp := msg.GetGroupMsgResp{}
	if err := proto.Unmarshal(pkt.Payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.Result.Unwrap() != 0 {
		c.error("get msg error: %v %v", rsp.Result.Unwrap(), rsp.Errmsg.Unwrap())
		return nil, errors.Errorf("get msg error: %v msg: %v", rsp.Result.Unwrap(), rsp.Errmsg.Unwrap())
	}
	var ret []*message.GroupMessage
	for _, m := range rsp.Msg {
		if m.Head.FromUin.IsNone() {
			continue
		}
		if m.Content != nil && m.Content.PkgNum.Unwrap() > 1 && !pkt.Params.Bool("raw") {
			if m.Content.PkgIndex.Unwrap() == 0 {
				c.debug("build fragmented message from history")
				i := m.Head.MsgSeq.Unwrap() - m.Content.PkgNum.Unwrap()
				builder := &messageBuilder{}
				for {
					end := int32(math.Min(float64(i+19), float64(m.Head.MsgSeq.Unwrap()+m.Content.PkgNum.Unwrap())))
					seq, pkt := c.buildGetGroupMsgRequest(m.Head.GroupInfo.GroupCode.Unwrap(), int64(i), int64(end))
					data, err := c.sendAndWait(seq, pkt, network.RequestParams{"raw": true})
					if err != nil {
						return nil, errors.Wrap(err, "build fragmented message error")
					}
					for _, fm := range data.([]*message.GroupMessage) {
						if fm.OriginalObject.Content != nil && fm.OriginalObject.Content.DivSeq.Unwrap() == m.Content.DivSeq.Unwrap() {
							builder.append(fm.OriginalObject)
						}
					}
					if end >= m.Head.MsgSeq.Unwrap()+m.Content.PkgNum.Unwrap() {
						break
					}
					i = end
				}
				if elem := c.parseGroupMessage(builder.build()); elem != nil {
					ret = append(ret, elem)
				}
			}
			continue
		}
		if elem := c.parseGroupMessage(m); elem != nil {
			ret = append(ret, elem)
		}
	}
	return ret, nil
}

func decodeAtAllRemainResponse(_ *QQClient, pkt *network.Packet) (any, error) {
	rsp := oidb.D8A7RspBody{}
	err := unpackOIDBPackage(pkt.Payload, &rsp)
	if err != nil {
		return nil, err
	}
	return &AtAllRemainInfo{
		CanAtAll:                 rsp.CanAtAll.Unwrap(),
		RemainAtAllCountForGroup: rsp.RemainAtAllCountForGroup.Unwrap(),
		RemainAtAllCountForUin:   rsp.RemainAtAllCountForUin.Unwrap(),
	}, nil
}

func (c *QQClient) parseGroupMessage(m *msg.Message) *message.GroupMessage {
	group := c.FindGroup(m.Head.GroupInfo.GroupCode.Unwrap())
	if group == nil {
		c.debug("sync group %v.", m.Head.GroupInfo.GroupCode.Unwrap())
		info, err := c.GetGroupInfo(m.Head.GroupInfo.GroupCode.Unwrap())
		if err != nil {
			c.error("failed to sync group %v : %+v", m.Head.GroupInfo.GroupCode.Unwrap(), err)
			return nil
		}
		group = info
		c.GroupList = append(c.GroupList, info)
	}
	if len(group.Members) == 0 {
		mem, err := c.GetGroupMembers(group)
		if err != nil {
			c.error("failed to sync group %v members : %+v", m.Head.GroupInfo.GroupCode, err)
			return nil
		}
		group.Members = mem
	}
	var anonInfo *msg.AnonymousGroupMessage
	for _, e := range m.Body.RichText.Elems {
		if e.AnonGroupMsg != nil {
			anonInfo = e.AnonGroupMsg
		}
	}
	var sender *message.Sender
	if anonInfo != nil {
		sender = &message.Sender{
			Uin:      80000000,
			Nickname: string(anonInfo.AnonNick),
			AnonymousInfo: &message.AnonymousInfo{
				AnonymousId:   base64.StdEncoding.EncodeToString(anonInfo.AnonId),
				AnonymousNick: string(anonInfo.AnonNick),
			},
			IsFriend: false,
		}
	} else {
		mem := group.FindMember(m.Head.FromUin.Unwrap())
		if mem == nil {
			group.Update(func(_ *GroupInfo) {
				if mem = group.FindMemberWithoutLock(m.Head.FromUin.Unwrap()); mem != nil {
					return
				}
				info, _ := c.GetMemberInfo(group.Code, m.Head.FromUin.Unwrap())
				if info == nil {
					return
				}
				mem = info
				group.Members = append(group.Members, mem)
				group.sort()
				go c.GroupMemberJoinEvent.dispatch(c, &MemberJoinGroupEvent{
					Group:  group,
					Member: info,
				})
			})
			if mem == nil {
				return nil
			}
		}
		sender = &message.Sender{
			Uin:      mem.Uin,
			Nickname: mem.Nickname,
			CardName: mem.CardName,
			IsFriend: c.FindFriend(mem.Uin) != nil,
		}
	}
	var g *message.GroupMessage
	g = &message.GroupMessage{
		Id:             m.Head.MsgSeq.Unwrap(),
		GroupCode:      group.Code,
		GroupName:      string(m.Head.GroupInfo.GroupName),
		Sender:         sender,
		Time:           m.Head.MsgTime.Unwrap(),
		Elements:       message.ParseMessageElems(m.Body.RichText.Elems),
		OriginalObject: m,
	}
	var extInfo *msg.ExtraInfo
	// pre parse
	for _, elem := range m.Body.RichText.Elems {
		// is rich long msg
		if elem.GeneralFlags != nil && elem.GeneralFlags.LongTextResid.Unwrap() != "" && len(g.Elements) == 1 {
			if f := c.GetForwardMessage(elem.GeneralFlags.LongTextResid.Unwrap()); f != nil && len(f.Nodes) == 1 {
				g = &message.GroupMessage{
					Id:             m.Head.MsgSeq.Unwrap(),
					GroupCode:      group.Code,
					GroupName:      string(m.Head.GroupInfo.GroupName),
					Sender:         sender,
					Time:           m.Head.MsgTime.Unwrap(),
					Elements:       f.Nodes[0].Message,
					OriginalObject: m,
				}
			}
		}
		if elem.ExtraInfo != nil {
			extInfo = elem.ExtraInfo
		}
	}
	if !sender.IsAnonymous() {
		mem := group.FindMember(m.Head.FromUin.Unwrap())
		groupCard := m.Head.GroupInfo.GroupCard.Unwrap()
		if extInfo != nil && len(extInfo.GroupCard) > 0 && extInfo.GroupCard[0] == 0x0A {
			buf := oidb.D8FCCommCardNameBuf{}
			if err := proto.Unmarshal(extInfo.GroupCard, &buf); err == nil && len(buf.RichCardName) > 0 {
				var gcard strings.Builder
				for _, e := range buf.RichCardName {
					gcard.Write(e.Text)
				}
				groupCard = gcard.String()
			}
		}
		if m.Head.GroupInfo != nil && groupCard != "" && mem.CardName != groupCard {
			old := mem.CardName
			if mem.Nickname == groupCard {
				mem.CardName = ""
			} else {
				mem.CardName = groupCard
			}
			if old != mem.CardName {
				c.MemberCardUpdatedEvent.dispatch(c, &MemberCardUpdatedEvent{
					Group:   group,
					OldCard: old,
					Member:  mem,
				})
			}
		}
	}
	if m.Body.RichText.Ptt != nil {
		var url string
		if len(m.Body.RichText.Ptt.DownPara) == 0 {
			req := &cmd0x388.D388ReqBody{
				NetType: proto.Uint32(3),
				Subcmd:  proto.Uint32(4),
				GetpttUrlReq: []*cmd0x388.GetPttUrlReq{
					{
						GroupCode:       proto.Uint64(uint64(m.Head.GroupInfo.GroupCode.Unwrap())),
						DstUin:          proto.Uint64(uint64(m.Head.ToUin.Unwrap())),
						Fileid:          proto.Uint64(uint64(m.Body.RichText.Ptt.FileId.Unwrap())),
						FileMd5:         m.Body.RichText.Ptt.FileMd5,
						ReqTerm:         proto.Uint32(5),
						ReqPlatformType: proto.Uint32(9),
						InnerIp:         proto.Uint32(0),
						BuType:          proto.Uint32(3),
						FileId:          proto.Uint64(0),
						FileKey:         m.Body.RichText.Ptt.FileKey,
						ReqTransferType: proto.Uint32(2),
						IsAuto:          proto.Uint32(1),
					},
				},
			}
			payload, _ := proto.Marshal(req)
			rsp_raw, _ := c.sendAndWaitDynamic(c.uniPacket("PttStore.GroupPttDown", payload))
			rsp := new(cmd0x388.D388RspBody)
			proto.Unmarshal(rsp_raw, rsp)
			resp := rsp.GetpttUrlRsp[0]
			url = "http://" + string(resp.DownDomain) + string(resp.DownPara)
		} else {
			url = "http://grouptalk.c2c.qq.com" + string(m.Body.RichText.Ptt.DownPara)
		}

		g.Elements = []message.IMessageElement{
			&message.VoiceElement{
				Name: m.Body.RichText.Ptt.FileName.Unwrap(),
				Md5:  m.Body.RichText.Ptt.FileMd5,
				Size: m.Body.RichText.Ptt.FileSize.Unwrap(),
				Url:  url,
			},
		}
	}
	if m.Body.RichText.Attr != nil {
		g.InternalId = m.Body.RichText.Attr.Random.Unwrap()
	}
	return g
}

// SetEssenceMessage 设为群精华消息
func (c *QQClient) SetEssenceMessage(groupCode int64, msgID, msgInternalId int32) error {
	r, err := c.sendAndWait(c.buildEssenceMsgOperatePacket(groupCode, uint32(msgID), uint32(msgInternalId), 1))
	if err != nil {
		return errors.Wrap(err, "set essence msg network")
	}
	rsp := r.(*oidb.EACRspBody)
	if rsp.ErrorCode.Unwrap() != 0 {
		return errors.New(rsp.Wording.Unwrap())
	}
	return nil
}

// DeleteEssenceMessage 移出群精华消息
func (c *QQClient) DeleteEssenceMessage(groupCode int64, msgID, msgInternalId int32) error {
	r, err := c.sendAndWait(c.buildEssenceMsgOperatePacket(groupCode, uint32(msgID), uint32(msgInternalId), 2))
	if err != nil {
		return errors.Wrap(err, "set essence msg networ")
	}
	rsp := r.(*oidb.EACRspBody)
	if rsp.ErrorCode.Unwrap() != 0 {
		return errors.New(rsp.Wording.Unwrap())
	}
	return nil
}

func (c *QQClient) buildEssenceMsgOperatePacket(groupCode int64, msgSeq, msgRand, opType uint32) (uint16, []byte) {
	commandName := "OidbSvc.0xeac_" + strconv.FormatInt(int64(opType), 10)
	payload := c.packOIDBPackageProto(3756, int32(opType), &oidb.EACReqBody{ // serviceType 2 取消
		GroupCode: proto.Uint64(uint64(groupCode)),
		Seq:       proto.Uint32(msgSeq),
		Random:    proto.Uint32(msgRand),
	})
	return c.uniPacket(commandName, payload)
}

// OidbSvc.0xeac_1/2
func decodeEssenceMsgResponse(_ *QQClient, pkt *network.Packet) (any, error) {
	rsp := &oidb.EACRspBody{}
	err := unpackOIDBPackage(pkt.Payload, &rsp)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

// GetGroupEssenceMsgList 获取群精华消息列表
func (c *QQClient) GetGroupEssenceMsgList(groupCode int64) ([]GroupDigest, error) {
	essenceURL := "https://qun.qq.com/essence/index?gc=" + strconv.FormatInt(groupCode, 10) + "&_wv=3&_wwv=128&_wvx=2&_wvxBclr=f5f6fa"
	rsp, err := utils.HttpGetBytes(essenceURL, c.getCookiesWithDomain("qun.qq.com"))
	if err != nil {
		return nil, errors.Wrap(err, "get essence msg network error")
	}
	rsp = rsp[bytes.Index(rsp, []byte("window.__INITIAL_STATE__={"))+25:]
	rsp = rsp[:bytes.Index(rsp, []byte("</script>"))]
	data := &struct {
		List []GroupDigest `json:"msgList"`
	}{}
	err = json.Unmarshal(rsp, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json")
	}
	return data.List, nil
}
