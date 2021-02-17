package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/Mrs4s/MiraiGo/client/pb/longmsg"
	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/client/pb/multimsg"
	"github.com/Mrs4s/MiraiGo/client/pb/oidb"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/MiraiGo/protocol/packets"
	"github.com/Mrs4s/MiraiGo/utils"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
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
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	if useFram {
		if m.Any(func(e message.IMessageElement) bool { return e.Type() == message.Reply }) {
			useFram = false
		}
	}
	msgLen := message.EstimateLength(m.Elements, 5000)
	if msgLen > 5000 || imgCount > 50 {
		return nil
	}
	if (msgLen > 100 || imgCount > 2) && !useFram {
		ret := c.sendGroupMessage(groupCode, false,
			&message.SendingMessage{Elements: []message.IMessageElement{
				c.uploadGroupLongMessage(groupCode,
					&message.ForwardMessage{Nodes: []*message.ForwardNode{
						{
							SenderId:   c.Uin,
							SenderName: c.Nickname,
							Time:       int32(time.Now().Unix()),
							Message:    m.Elements,
						},
					}},
				),
			}},
		)
		return &message.GroupMessage{
			Id:         ret.Id,
			InternalId: ret.InternalId,
			GroupCode:  ret.GroupCode,
			Sender:     ret.Sender,
			Time:       ret.Time,
			Elements:   m.Elements,
		}
	}
	return c.sendGroupMessage(groupCode, false, m)
}

// SendGroupForwardMessage 发送群合并转发消息
func (c *QQClient) SendGroupForwardMessage(groupCode int64, m *message.ForwardMessage) *message.GroupMessage {
	return c.sendGroupMessage(groupCode, true,
		&message.SendingMessage{Elements: []message.IMessageElement{c.UploadGroupForwardMessage(groupCode, m)}},
	)
}

// GetGroupMessages 从服务器获取历史信息
func (c *QQClient) GetGroupMessages(groupCode, beginSeq, endSeq int64) ([]*message.GroupMessage, error) {
	i, err := c.sendAndWait(c.buildGetGroupMsgRequest(groupCode, beginSeq, endSeq))
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
	ch := make(chan int32)
	c.onGroupMessageReceipt(eid, func(c *QQClient, e *groupMessageReceiptEvent) {
		if e.Rand == mr && !utils.IsChanClosed(ch) {
			ch <- e.Seq
		}
	})
	defer c.onGroupMessageReceipt(eid)
	imgCount := m.Count(func(e message.IMessageElement) bool { return e.Type() == message.Image })
	msgLen := message.EstimateLength(m.Elements, 703)
	if (msgLen > 100 || imgCount > 1) && !forward && !m.Any(func(e message.IMessageElement) bool {
		_, ok := e.(*message.GroupVoiceElement)
		_, ok2 := e.(*message.ServiceElement)
		_, ok3 := e.(*message.ReplyElement)
		if _, ok4 := e.(*message.ForwardElement); ok4 {
			forward = true
			return true
		}
		return ok || ok2 || ok3
	}) {
		div := int32(rand.Uint32())
		fragmented := m.ToFragmented()
		for i, elems := range fragmented {
			_, pkt := c.buildGroupSendingPacket(groupCode, mr, int32(len(fragmented)), int32(i), div, forward, elems)
			_ = c.send(pkt)
		}
	} else {
		_, pkt := c.buildGroupSendingPacket(groupCode, mr, 1, 0, 0, forward, m.Elements)
		_ = c.send(pkt)
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

func (c *QQClient) uploadGroupLongMessage(groupCode int64, m *message.ForwardMessage) *message.ServiceElement {
	if len(m.Nodes) >= 200 {
		return nil
	}
	ts := time.Now().UnixNano()
	seq := c.nextGroupSeq()
	data, hash := m.CalculateValidationData(seq, rand.Int31(), groupCode)
	rsp, body, err := c.multiMsgApplyUp(groupCode, data, hash, 1)
	if err != nil {
		return nil
	}
	for i, ip := range rsp.Uint32UpIp {
		err := c.highwayUpload(uint32(ip), int(rsp.Uint32UpPort[i]), rsp.MsgSig, body, 27)
		if err == nil {
			bri := func() string {
				var r string
				for _, n := range m.Nodes {
					r += message.ToReadableString(n.Message)
					if len(r) >= 27 {
						break
					}
				}
				return r
			}()
			return genLongTemplate(rsp.MsgResid, bri, ts)
		}
	}
	return nil
}

func (c *QQClient) UploadGroupForwardMessage(groupCode int64, m *message.ForwardMessage) *message.ForwardElement {
	if len(m.Nodes) >= 200 {
		return nil
	}
	ts := time.Now().UnixNano()
	seq := c.nextGroupSeq()
	data, hash, items := m.CalculateValidationDataForward(seq, rand.Int31(), groupCode)
	rsp, body, err := c.multiMsgApplyUp(groupCode, data, hash, 2)
	if err != nil {
		return nil
	}
	for i, ip := range rsp.Uint32UpIp {
		err := c.highwayUpload(uint32(ip), int(rsp.Uint32UpPort[i]), rsp.MsgSig, body, 27)
		if err == nil {
			var pv string
			for i := 0; i < int(math.Min(4, float64(len(m.Nodes)))); i++ {
				pv += fmt.Sprintf(`<title size="26" color="#777777">%s: %s</title>`, XmlEscape(m.Nodes[i].SenderName), XmlEscape(message.ToReadableString(m.Nodes[i].Message)))
			}
			return genForwardTemplate(rsp.MsgResid, pv, "群聊的聊天记录", "[聊天记录]", "聊天记录", fmt.Sprintf("查看 %d 条转发消息", len(m.Nodes)), ts, items)
		}
	}
	return nil
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
	seq := c.nextSeq()
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
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbSendMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildGetGroupMsgRequest(groupCode, beginSeq, endSeq int64) (uint16, []byte) {
	seq := c.nextSeq()
	req := &msg.GetGroupMsgReq{
		GroupCode:   proto.Uint64(uint64(groupCode)),
		BeginSeq:    proto.Uint64(uint64(beginSeq)),
		EndSeq:      proto.Uint64(uint64(endSeq)),
		PublicGroup: proto.Bool(false),
	}
	payload, _ := proto.Marshal(req)
	packet := packets.BuildUniPacket(c.Uin, seq, "MessageSvc.PbGetGroupMsg", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

func (c *QQClient) buildAtAllRemainRequestPacket(groupCode int64) (uint16, []byte) {
	seq := c.nextSeq()
	payload := c.packOIDBPackageProto(2215, 0, &oidb.D8A7ReqBody{
		SubCmd:                    proto.Uint32(1),
		LimitIntervalTypeForUin:   proto.Uint32(2),
		LimitIntervalTypeForGroup: proto.Uint32(1),
		Uin:                       proto.Uint64(uint64(c.Uin)),
		GroupCode:                 proto.Uint64(uint64(groupCode)),
	})
	packet := packets.BuildUniPacket(c.Uin, seq, "OidbSvc.0x8a7_0", 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OnlinePush.PbPushGroupMsg
func decodeGroupMessagePacket(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
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
		var builder *groupMessageBuilder
		i, ok := c.groupMsgBuilders.Load(pkt.Message.Content.GetDivSeq())
		if !ok {
			builder = &groupMessageBuilder{}
			c.groupMsgBuilders.Store(pkt.Message.Content.GetDivSeq(), builder)
		} else {
			builder = i.(*groupMessageBuilder)
		}
		builder.MessageSlices = append(builder.MessageSlices, pkt.Message)
		if int32(len(builder.MessageSlices)) >= pkt.Message.Content.GetPkgNum() {
			c.groupMsgBuilders.Delete(pkt.Message.Content.GetDivSeq())
			if pkt.Message.Head.GetFromUin() == c.Uin {
				c.dispatchGroupMessageSelf(c.parseGroupMessage(builder.build()))
			} else {
				c.dispatchGroupMessage(c.parseGroupMessage(builder.build()))
			}
		}
		return nil, nil
	}
	if pkt.Message.Head.GetFromUin() == c.Uin {
		c.dispatchGroupMessageSelf(c.parseGroupMessage(pkt.Message))
	} else {
		c.dispatchGroupMessage(c.parseGroupMessage(pkt.Message))
	}
	return nil, nil
}

func decodeMsgSendResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.SendMessageResponse{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.GetResult() != 0 {
		c.Error("send msg error: %v %v", rsp.GetResult(), rsp.GetErrMsg())
	}
	return nil, nil
}

func decodeGetGroupMsgResponse(c *QQClient, _ uint16, payload []byte) (interface{}, error) {
	rsp := msg.GetGroupMsgResp{}
	if err := proto.Unmarshal(payload, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if rsp.GetResult() != 0 {
		c.Error("get msg error: %v %v", rsp.GetResult(), rsp.GetErrmsg())
		return nil, errors.Errorf("get msg error: %v msg: %v", rsp.GetResult(), rsp.GetErrmsg())
	}
	var ret []*message.GroupMessage
	for _, m := range rsp.Msg {
		if m.Head.FromUin == nil {
			continue
		}
		if elem := c.parseGroupMessage(m); elem != nil {
			ret = append(ret, elem)
		}
	}
	return ret, nil
}

func decodeAtAllRemainResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := oidb.D8A7RspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, &rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
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
		c.Debug("sync group %v.", m.Head.GroupInfo.GetGroupCode())
		info, err := c.GetGroupInfo(m.Head.GroupInfo.GetGroupCode())
		if err != nil {
			c.Error("error to sync group %v : %+v", m.Head.GroupInfo.GetGroupCode(), err)
			return nil
		}
		group = info
		c.GroupList = append(c.GroupList, info)
	}
	if len(group.Members) == 0 {
		mem, err := c.GetGroupMembers(group)
		if err != nil {
			c.Error("error to sync group %v member : %+v", m.Head.GroupInfo.GroupCode, err)
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
				info, _ := c.getMemberInfo(group.Code, m.Head.GetFromUin())
				if info == nil {
					return
				}
				mem = info
				group.Members = append(group.Members, mem)
				go c.dispatchNewMemberEvent(&MemberJoinGroupEvent{
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
				groupCard = ""
				for _, e := range buf.RichCardName {
					groupCard += string(e.Text)
				}
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
				go c.dispatchMemberCardUpdatedEvent(&MemberCardUpdatedEvent{
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
func (c *QQClient) SetEssenceMessage(groupCode int64, msgId, msgInternalId int32) error {
	r, err := c.sendAndWait(c.buildEssenceMsgOperatePacket(groupCode, uint32(msgId), uint32(msgInternalId), 1))
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
func (c *QQClient) DeleteEssenceMessage(groupCode int64, msgId, msgInternalId int32) error {
	r, err := c.sendAndWait(c.buildEssenceMsgOperatePacket(groupCode, uint32(msgId), uint32(msgInternalId), 2))
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
	seq := c.nextSeq()
	commandName := "OidbSvc.0xeac_" + strconv.FormatInt(int64(opType), 10)
	payload := c.packOIDBPackageProto(3756, int32(opType), &oidb.EACReqBody{ // serviceType 2 取消
		GroupCode: proto.Uint64(uint64(groupCode)),
		Seq:       proto.Uint32(msgSeq),
		Random:    proto.Uint32(msgRand),
	})
	packet := packets.BuildUniPacket(c.Uin, seq, commandName, 1, c.OutGoingPacketSessionId, EmptyBytes, c.sigInfo.d2Key, payload)
	return seq, packet
}

// OidbSvc.0xeac_1/2
func decodeEssenceMsgResponse(_ *QQClient, _ uint16, payload []byte) (interface{}, error) {
	pkg := oidb.OIDBSSOPkg{}
	rsp := &oidb.EACRspBody{}
	if err := proto.Unmarshal(payload, &pkg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
	}
	if err := proto.Unmarshal(pkg.Bodybuffer, rsp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal protobuf message")
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
	var data = &struct {
		List []GroupDigest `json:"msgList"`
	}{}
	err = json.Unmarshal(rsp, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json")
	}
	return data.List, nil
}
