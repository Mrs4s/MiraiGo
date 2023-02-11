package network

type Packet struct {
	SequenceId  uint16
	CommandName string
	Payload     []byte
	Params      RequestParams
}

type RequestParams map[string]any

func (p RequestParams) Bool(k string) bool {
	if p == nil {
		return false
	}
	i, ok := p[k]
	if !ok {
		return false
	}
	return i.(bool)
}

func (p RequestParams) Int32(k string) int32 {
	if p == nil {
		return 0
	}
	i, ok := p[k]
	if !ok {
		return 0
	}
	return i.(int32)
}
