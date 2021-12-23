package message

import (
	"strings"
	"testing"
)

func Test_mergeContinuousTextMessages(t *testing.T) {
	msg := NewSendingMessage()
	msg.Append(NewText("短片段一"))
	msg.Append(NewText(strings.Repeat("长一", 800))) // 6*800
	msg.Append(NewText("短片段二"))
	msg.Append(NewText(strings.Repeat("长二", 1200))) // 6*1200
	msg.Append(NewText("短片段三"))

	// 总长度为 12036
	totalSize := EstimateLength(msg.Elements)
	expectedPart := (totalSize + MaxMessageSize - 1) / MaxMessageSize

	messages := SplitLongMessage(msg)
	// 应分为 3段
	if len(messages) != expectedPart {
		t.Errorf("should split into %v part", expectedPart)
	}
	partsSize := 0
	for idx, message := range messages {
		partSize := EstimateLength(message.Elements)
		if partSize > MaxMessageSize {
			t.Errorf("part %v size=%v is more than %v", idx, partSize, MaxMessageSize)
		}
		partsSize += partSize
	}
	if partsSize != totalSize {
		t.Errorf("parts size sum=%v is not equal to total size=%v", partsSize, totalSize)
	}
}
