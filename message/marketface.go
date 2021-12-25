package message

import (
	"fmt"

	"github.com/Mrs4s/MiraiGo/client/pb/msg"
	"github.com/Mrs4s/MiraiGo/internal/proto"
	"github.com/Mrs4s/MiraiGo/utils"
)

type MarketFaceElement struct {
	Name       string
	FaceId     []byte // decoded = mediaType == 2 ? string(FaceId) : hex.EncodeToString(FaceId).toLower().trimSpace(); download url param?
	TabId      int32
	ItemType   int32
	SubType    int32  // image type, 0 -> None 1 -> Magic Face 2 -> GIF 3 -> PNG
	MediaType  int32  // 1 -> Voice Face 2 -> dynamic face
	EncryptKey []byte // tea + xor, see EMosmUtils.class::a maybe useful?
	MagicValue string
}

type DiceElement struct {
	*MarketFaceElement
	Value int32
}

func (e *MarketFaceElement) Type() ElementType {
	return Face
}

func (e *MarketFaceElement) Pack() []*msg.Elem {
	return []*msg.Elem{
		{
			MarketFace: &msg.MarketFace{
				FaceName:    utils.S2B(e.Name),
				ItemType:    proto.Uint32(uint32(e.ItemType)),
				FaceInfo:    proto.Uint32(1),
				FaceId:      e.FaceId,
				TabId:       proto.Uint32(uint32(e.TabId)),
				SubType:     proto.Uint32(uint32(e.SubType)),
				Key:         e.EncryptKey,
				MediaType:   proto.Uint32(uint32(e.MediaType)),
				ImageWidth:  proto.Uint32(200),
				ImageHeight: proto.Uint32(200),
				Mobileparam: utils.S2B(e.MagicValue),
			},
		},
		{
			Text: &msg.Text{Str: &e.Name},
		},
	}
}

func NewDice(value int32) IMessageElement {
	if value < 1 || value > 6 {
		return nil
	}
	return &MarketFaceElement{
		Name:       "[骰子]",
		FaceId:     []byte{72, 35, 211, 173, 177, 93, 240, 128, 20, 206, 93, 103, 150, 183, 110, 225},
		TabId:      11464,
		ItemType:   6,
		SubType:    3,
		MediaType:  0,
		EncryptKey: []byte{52, 48, 57, 101, 50, 97, 54, 57, 98, 49, 54, 57, 49, 56, 102, 57},
		MagicValue: fmt.Sprintf("rscType?1;value=%v", value-1),
	}
}

type FingerGuessingElement struct {
	*MarketFaceElement
	Value int32
	Name  string
}

var fingerGuessingName = map[int32]string{
	0: "石头",
	1: "剪刀",
	2: "布",
}

func NewFingerGuessing(value int32) IMessageElement {
	// value 0石头, 1剪子, 2布
	if value < 0 || value > 2 {
		return nil
	}
	return &MarketFaceElement{
		Name:       "[猜拳]",
		FaceId:     []byte{131, 200, 162, 147, 174, 101, 202, 20, 15, 52, 129, 32, 167, 116, 72, 238},
		TabId:      11415,
		ItemType:   6,
		SubType:    3,
		MediaType:  0,
		EncryptKey: []byte{55, 100, 101, 51, 57, 102, 101, 98, 99, 102, 52, 53, 101, 54, 100, 98},
		MagicValue: fmt.Sprintf("rscType?1;value=%v", value),
	}
}
