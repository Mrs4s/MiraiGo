package client

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/highway"
	"github.com/Mrs4s/MiraiGo/client/internal/network"
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
func (c *QQClient) SendGroupMessage(groupCode int64, m *message.SendingMessage, f ...bool) *message.GroupMessage {
	useFram := false
	if len(f) > 0 {
		useFram = f[0]
	}
	imgCount := 0
	for _, e := range m.Elements {
		switch e.Type() {
		case message.Image:
			imgCount++
		case message.Reply:
			useFram = false
		}
	}
	msgLen := message.EstimateLength(m.Elements)
	if msgLen > message.MaxMessageSize || imgCount > 50 {
		return nil
	}
	if !useFram && (msgLen > 100 || imgCount > 2) {
		lmsg, err := c.uploadGroupLongMessage(groupCode,
			message.NewForwardMessage().AddNode(&message.ForwardNode{
				SenderId:   c.Uin,
				SenderName: c.Nickname,
				Time:       int32(time.Now().Unix()),
				Message:    m.Elements,
			}))
		if err == nil {
			ret := c.sendGroupMessage(groupCode, false, &message.SendingMessage{Elements: []message.IMessageElement{lmsg}})
			ret.Elements = m.Elements
			return ret
		}
		c.error("%v", err)
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
	if !forward && serviceFlag && (imgCount > 1 || message.EstimateLength(m.Elements) > 100) {
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

func (c *QQClient) uploadGroupLongMessage(groupCode int64, m *message.ForwardMessage) (*message.ServiceElement, error) {
	ts := time.Now().UnixNano()
	seq := c.nextGroupSeq()
	data, hash := m.CalculateValidationData(seq, rand.Int31(), groupCode)
	rsp, body, err := c.multiMsgApplyUp(groupCode, data, hash, 1)
	if err != nil {
		return nil, errors.Errorf("upload long message error: %v", err)
	}
	for i, ip := range rsp.Uint32UpIp {
		addr := highway.Addr{IP: uint32(ip), Port: int(rsp.Uint32UpPort[i])}
		hash := md5.Sum(body)
		input := highway.Transaction{
			CommandID: 27,
			Ticket:    rsp.MsgSig,
			Body:      bytes.NewReader(body),
			Size:      int64(len(body)),
			Sum:       hash[:],
		}
		err := c.highwaySession.Upload(addr, input)
		if err != nil {
			c.error("highway upload long message error: %v", err)
			continue
		}
		return genLongTemplate(rsp.MsgResid, m.Brief(), ts), nil
	}
	return nil, errors.New("upload long message error: highway server list is empty or not available server.")
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
		RoutingHead: &msg.RoutingHead{Grp: &msg.Grp{GroupCode: &groupCode}},
		ContentHead: &msg.ContentHead{PkgNum: &pkgNum, PkgIndex: &pkgIndex, DivSeq: &pkgDiv},
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
		MsgRand:    &r,
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
func decodeGroupMessagePacket(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	pkt := msg.PushMessagePacket{}
	err := proto.Unmarshal(payload, &pkt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if pkt.Message.Head.GetFromUin() == c.Uin {
		c.dispatchGroupMessageReceiptEvent(&groupMessageReceiptEvent{
			Rand: pkt.Message.Body.RichText.Attr.GetRandom(),
			Seq:  pkt.Message.Head.GetMsgSeq(),
			Msg:  c.parseGroupMessage(pkt.Message),
		})
	}
	if pkt.Message.Content != nil && pkt.Message.Content.GetPkgNum() > 1 {
		seq := pkt.Message.Content.GetDivSeq()
		builder := c.messageBuilder(pkt.Message.Content.GetDivSeq())
		builder.append(pkt.Message)
		if builder.len() >= pkt.Message.Content.GetPkgNum() {
			c.msgBuilders.Delete(seq)
			if pkt.Message.Head.GetFromUin() == c.Uin {
				c.SelfGroupMessageEvent.dispatch(c, c.parseGroupMessage(builder.build()))
			} else {
				c.GroupMessageEvent.dispatch(c, c.parseGroupMessage(builder.build()))
			}
		}
		return nil, nil
	}
	if pkt.Message.Head.GetFromUin() == c.Uin {
		c.SelfGroupMessageEvent.dispatch(c, c.parseGroupMessage(pkt.Message))
	} else {
		c.GroupMessageEvent.dispatch(c, c.parseGroupMessage(pkt.Message))
	}
	return nil, nil
}

func decodeMsgSendResponse(c *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := msg.SendMessageResponse{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	switch rsp.GetResult() {
	case 0: // OK.
	case 55:
		c.error("sendPacket msg error: %v Bot has blocked target's content", rsp.GetResult())
	default:
		c.error("sendPacket msg error: %v %v", rsp.GetResult(), rsp.GetErrMsg())
	}
	return nil, nil
}

func decodeGetGroupMsgResponse(c *QQClient, info *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := msg.GetGroupMsgResp{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.GetResult() != 0 {
		c.error("get msg error: %v %v", rsp.GetResult(), rsp.GetErrmsg())
		return nil, errors.Errorf("get msg error: %v msg: %v", rsp.GetResult(), rsp.GetErrmsg())
	}
	var ret []*message.GroupMessage
	for _, m := range rsp.Msg {
		if m.Head.FromUin == nil {
			continue
		}
		if m.Content != nil && m.Content.GetPkgNum() > 1 && !info.Params.Bool("raw") {
			if m.Content.GetPkgIndex() == 0 {
				c.debug("build fragmented message from history")
				i := m.Head.GetMsgSeq() - m.Content.GetPkgNum()
				builder := &messageBuilder{}
				for {
					end := int32(math.Min(float64(i+19), float64(m.Head.GetMsgSeq()+m.Content.GetPkgNum())))
					seq, pkt := c.buildGetGroupMsgRequest(m.Head.GroupInfo.GetGroupCode(), int64(i), int64(end))
					data, err := c.sendAndWait(seq, pkt, network.RequestParams{"raw": true})
					if err != nil {
						return nil, errors.Wrap(err, "build fragmented message error")
					}
					for _, fm := range data.([]*message.GroupMessage) {
						if fm.OriginalObject.Content != nil && fm.OriginalObject.Content.GetDivSeq() == m.Content.GetDivSeq() {
							builder.append(fm.OriginalObject)
						}
					}
					if end >= m.Head.GetMsgSeq()+m.Content.GetPkgNum() {
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

func decodeAtAllRemainResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := oidb.D8A7RspBody{}
	err := unpackOIDBPackage(payload, &rsp)
	if err != nil {
		return nil, err
	}
	return &AtAllRemainInfo{
		CanAtAll:                 rsp.GetCanAtAll(),
		RemainAtAllCountForGroup: rsp.GetRemainAtAllCountForGroup(),
		RemainAtAllCountForUin:   rsp.GetRemainAtAllCountForUin(),
	}, nil
}

func (c *QQClient) parseGroupMessage(m *msg.Message) *message.GroupMessage {
	group := c.FindGroup(m.Head.GroupInfo.GetGroupCode())
	if group == nil {
		c.debug("sync group %v.", m.Head.GroupInfo.GetGroupCode())
		info, err := c.GetGroupInfo(m.Head.GroupInfo.GetGroupCode())
		if err != nil {
			c.error("error to sync group %v : %+v", m.Head.GroupInfo.GetGroupCode(), err)
			return nil
		}
		group = info
		c.GroupList = append(c.GroupList, info)
	}
	if len(group.Members) == 0 {
		mem, err := c.GetGroupMembers(group)
		if err != nil {
			c.error("error to sync group %v member : %+v", m.Head.GroupInfo.GroupCode, err)
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
		mem := group.FindMember(m.Head.GetFromUin())
		if mem == nil {
			group.Update(func(_ *GroupInfo) {
				if mem = group.FindMemberWithoutLock(m.Head.GetFromUin()); mem != nil {
					return
				}
				info, _ := c.GetMemberInfo(group.Code, m.Head.GetFromUin())
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
		Id:             m.Head.GetMsgSeq(),
		GroupCode:      group.Code,
		GroupName:      string(m.Head.GroupInfo.GroupName),
		Sender:         sender,
		Time:           m.Head.GetMsgTime(),
		Elements:       message.ParseMessageElems(m.Body.RichText.Elems),
		OriginalObject: m,
	}
	var extInfo *msg.ExtraInfo
	// pre parse
	for _, elem := range m.Body.RichText.Elems {
		// is rich long msg
		if elem.GeneralFlags != nil && elem.GeneralFlags.GetLongTextResid() != "" && len(g.Elements) == 1 {
			if f := c.GetForwardMessage(elem.GeneralFlags.GetLongTextResid()); f != nil && len(f.Nodes) == 1 {
				g = &message.GroupMessage{
					Id:             m.Head.GetMsgSeq(),
					GroupCode:      group.Code,
					GroupName:      string(m.Head.GroupInfo.GroupName),
					Sender:         sender,
					Time:           m.Head.GetMsgTime(),
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
		mem := group.FindMember(m.Head.GetFromUin())
		groupCard := m.Head.GroupInfo.GetGroupCard()
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
		g.Elements = []message.IMessageElement{
			&message.VoiceElement{
				Name: m.Body.RichText.Ptt.GetFileName(),
				Md5:  m.Body.RichText.Ptt.FileMd5,
				Size: m.Body.RichText.Ptt.GetFileSize(),
				Url:  "http://grouptalk.c2c.qq.com" + string(m.Body.RichText.Ptt.DownPara),
			},
		}
	}
	if m.Body.RichText.Attr != nil {
		g.InternalId = m.Body.RichText.Attr.GetRandom()
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
	if rsp.GetErrorCode() != 0 {
		return errors.New(rsp.GetWording())
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
	if rsp.GetErrorCode() != 0 {
		return errors.New(rsp.GetWording())
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
func decodeEssenceMsgResponse(_ *QQClient, _ *network.IncomingPacketInfo, payload []byte) (any, error) {
	rsp := &oidb.EACRspBody{}
	err := unpackOIDBPackage(payload, &rsp)
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
