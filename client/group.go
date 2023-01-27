package client

import (
	"github.com/Mrs4s/MiraiGo/message"
	"time"
)

type Group struct {
	c *QQClient
	i *GroupInfo
}

func (g *Group) Members() {

}

func (g *Group) Name() string {
	//TODO implement me
	panic("implement me")
}

func (g *Group) AvatarUrl() string {
	//TODO implement me
	panic("implement me")
}

func (g *Group) Message(m *message.SendingMessage) (*GroupMessageReceipt, bool) {
	c := g.c
	groupCode := g.i.Code

	useHighwayMessage := false
	imgCount := 0
	for _, e := range m.Elements {
		switch e.Type() {
		case message.Image:
			imgCount++
		case message.Reply:
			useHighwayMessage = true
		}
	}
	msgLen := message.EstimateLength(m.Elements)
	if msgLen > message.MaxMessageSize || imgCount > 50 {
		return nil, false
	}
	useHighwayMessage = useHighwayMessage || msgLen > 100 || imgCount > 2
	if useHighwayMessage && c.UseHighwayMessage {
		lmsg, err := c.uploadGroupLongMessage(groupCode,
			message.NewForwardMessage().AddNode(&message.ForwardNode{
				SenderId:   c.Uin,
				SenderName: c.Nickname,
				Time:       int32(time.Now().Unix()),
				Message:    m.Elements,
			}))
		if err == nil {
			ret := g.sendGroupMessage(false, &message.SendingMessage{Elements: []message.IMessageElement{lmsg}})
			ret.elems = m.Elements
			return ret, true
		}
		c.error("%v", err)
	}
	return g.sendGroupMessage(false, m), true
}

func (g *Group) Poke(target int64) {
	//TODO implement me
	panic("implement me")
}

func (g *Group) Delete() {

}
