package utils

import (
	"strings"

	"github.com/Mrs4s/MiraiGo/message"
)

// SplitLongMessage 将过长的消息分割为若干个适合发送的消息
func SplitLongMessage(sendingMessage *message.SendingMessage) []*message.SendingMessage {
	// 合并连续文本消息
	sendingMessage = mergeContinuousTextMessages(sendingMessage)

	// 分割过长元素
	sendingMessage = splitElements(sendingMessage)

	// 将元素分为多组，确保各组不超过单条消息的上限
	splitMessages := splitMessages(sendingMessage)

	return splitMessages
}

// mergeContinuousTextMessages 预先将所有连续的文本消息合并为到一起，方便后续统一切割
func mergeContinuousTextMessages(sendingMessage *message.SendingMessage) *message.SendingMessage {
	// 检查下是否有连续的文本消息，若没有，则可以直接返回
	lastIsText := false
	hasContinuousText := false
	for _, msg := range sendingMessage.Elements {
		if _, ok := msg.(*message.TextElement); ok {
			if lastIsText {
				// 有连续的文本消息，需要进行处理
				hasContinuousText = true
				break
			}

			// 遇到文本元素先存放起来，方便将连续的文本元素合并
			lastIsText = true
			continue
		} else {
			lastIsText = false
		}
	}
	if !hasContinuousText {
		return sendingMessage
	}

	// 存在连续的文本消息，需要进行合并处理
	mergeContinuousTextMessages := message.NewSendingMessage()

	textBuffer := strings.Builder{}
	lastIsText = false
	for _, msg := range sendingMessage.Elements {
		if msgVal, ok := msg.(*message.TextElement); ok {
			// 遇到文本元素先存放起来，方便将连续的文本元素合并
			textBuffer.WriteString(msgVal.Content)
			lastIsText = true
			continue
		}

		// 如果之前的是文本元素（可能是多个合并起来的），则在这里将其实际放入消息中
		if lastIsText {
			mergeContinuousTextMessages.Append(message.NewText(textBuffer.String()))
			textBuffer.Reset()
		}
		lastIsText = false

		// 非文本元素则直接处理
		mergeContinuousTextMessages.Append(msg)
	}
	// 处理最后几个元素是文本的情况
	if textBuffer.Len() != 0 {
		mergeContinuousTextMessages.Append(message.NewText(textBuffer.String()))
		textBuffer.Reset()
	}

	return mergeContinuousTextMessages
}

// splitElements 将原有消息的各个元素先尝试处理，如过长的文本消息按需分割为多个元素
func splitElements(sendingMessage *message.SendingMessage) *message.SendingMessage {
	// 检查下是否存在需要文本消息，若不存在，则直接返回
	needSplit := false
	for _, msg := range sendingMessage.Elements {
		if msgVal, ok := msg.(*message.TextElement); ok {
			if textNeedSplit(msgVal.Content) {
				needSplit = true
				break
			}
		}
	}
	if !needSplit {
		return sendingMessage
	}

	// 开始尝试切割
	messageParts := message.NewSendingMessage()

	for _, msg := range sendingMessage.Elements {
		switch msgVal := msg.(type) {
		case *message.TextElement:
			messageParts.Elements = append(messageParts.Elements, splitPlainMessage(msgVal.Content)...)
		default:
			messageParts.Append(msg)
		}
	}

	return messageParts
}

// splitMessages 根据大小分为多个消息进行发送
func splitMessages(sendingMessage *message.SendingMessage) []*message.SendingMessage {
	var splitMessages []*message.SendingMessage

	messagePart := message.NewSendingMessage()
	msgSize := 0
	for _, part := range sendingMessage.Elements {
		estimateSize := message.EstimateLength([]message.IMessageElement{part})
		// 若当前分消息加上新的元素后大小会超限，且已经有元素（确保不会无限循环），则开始切分为新的一个元素
		if msgSize+estimateSize > message.MaxMessageSize && len(messagePart.Elements) > 0 {
			splitMessages = append(splitMessages, messagePart)

			messagePart = message.NewSendingMessage()
			msgSize = 0
		}

		// 加上新的元素
		messagePart.Append(part)
		msgSize += estimateSize
	}
	// 将最后一个分片加上
	if len(messagePart.Elements) != 0 {
		splitMessages = append(splitMessages, messagePart)
	}

	return splitMessages
}

func splitPlainMessage(content string) []message.IMessageElement {
	if !textNeedSplit(content) {
		return []message.IMessageElement{message.NewText(content)}
	}

	var splittedMessage []message.IMessageElement

	var part string
	remainingText := content
	for len(remainingText) != 0 {
		partSize := 0
		for _, runeValue := range remainingText {
			runeSize := len(string(runeValue))
			if partSize+runeSize > message.MaxMessageSize {
				break
			}
			partSize += runeSize
		}

		part, remainingText = remainingText[:partSize], remainingText[partSize:]
		splittedMessage = append(splittedMessage, message.NewText(part))
	}

	return splittedMessage
}

func textNeedSplit(content string) bool {
	return len(content) > message.MaxMessageSize
}