package client

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/Mrs4s/MiraiGo/binary"
	"github.com/Mrs4s/MiraiGo/utils"
)

type Statistics struct {
	PacketReceived  atomic.Uint64 `json:"packet_received"`
	PacketSent      atomic.Uint64 `json:"packet_sent"`
	PacketLost      atomic.Uint64 `json:"packet_lost"`
	MessageReceived atomic.Uint64 `json:"message_received"`
	MessageSent     atomic.Uint64 `json:"message_sent"`
	LastMessageTime atomic.Int64  `json:"last_message_time"`
	DisconnectTimes atomic.Uint32 `json:"disconnect_times"`
	LostTimes       atomic.Uint32 `json:"lost_times"`
}

func (c *QQClient) GetStatistics() *Statistics {
	return &c.stat
}

// MarshalJSON encodes the wrapped statistics into JSON.
func (m *Statistics) MarshalJSON() ([]byte, error) {
	value := reflect.ValueOf(m).Elem()
	types := value.Type()

	byt := binary.SelectWriter()
	defer binary.PutWriter(byt)

	byt.WriteByte('{')
	for i:=0; i<value.NumField(); i++ {
		if i != 0 { byt.WriteByte(',') }
		tag := types.Field(i).Tag.Get("json")
		v   := value.Field(i)
		v    = v.Field(v.NumField()-1)
		switch v.Kind() {
			case reflect.Uint32, reflect.Uint64: byt.Write(utils.S2B(fmt.Sprintf(`"%s":%d`, tag, v.Uint())))
			case reflect.Int64: byt.Write(utils.S2B(fmt.Sprintf(`"%s":%d`, tag, v.Int())))
			// Statistics里暂时没有 atomic.Bool
			// case reflect.Bool: byt.Write(utils.S2B(fmt.Sprintf(`"%s":%v`, tag, v.Bool())))
		}
	}
	byt.WriteByte('}')

	return byt.Bytes(), nil
}
