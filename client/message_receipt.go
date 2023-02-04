package client

import (
	"github.com/Mrs4s/MiraiGo/message"
)

// MessageSequence MessageInternal
// ambiguous types were used, messed up like a bunch of shit, so we use uint64 and int64 for safety
type MessageSequence = uint64
type MessageInternal = int64

type MessageReceipt interface {
	Sequence() MessageSequence
	Elems() []message.IMessageElement
	Recall() error
	Quote() *message.ReplyElement
}

type messageIdentifier struct {
	Sequence MessageSequence
	// Internal i guess this is runtime id
	Internal MessageInternal
}

var _ MessageReceipt = (*GroupMessageReceipt)(nil)

type GroupMessageReceipt struct {
	id     *messageIdentifier
	elems  []message.IMessageElement
	sender *message.Sender
	time   int32
	target *Group
}

func (r *GroupMessageReceipt) Sequence() MessageSequence {
	return r.id.Sequence
}

func (r *GroupMessageReceipt) Elems() []message.IMessageElement {
	return r.elems
}
func (r *GroupMessageReceipt) Time() int32 {
	//TODO　improve this
	return r.time
}
func (r *GroupMessageReceipt) Target() *Group {
	return r.target
}

func (r *GroupMessageReceipt) Recall() error {
	return r.target.recallGroupMessage(r.id)
}

func (r *GroupMessageReceipt) Quote() *message.ReplyElement {
	return &message.ReplyElement{
		ReplySeq: int32(r.id.Sequence),
		Sender:   r.sender.Uin,
		Time:     r.time,
		Elements: r.elems,
	}
}
