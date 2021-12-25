package network

// Call is a client-side RPC call.
// refer to `net/rpc`
type Call struct {
	Request  *Request
	Response *Response
	Err      error
	Done     chan *Call
}

type Params map[string]interface{}

func (p Params) Bool(k string) bool {
	if p == nil {
		return false
	}
	i, ok := p[k]
	if !ok {
		return false
	}
	return i.(bool)
}

func (p Params) Int32(k string) int32 {
	if p == nil {
		return 0
	}
	i, ok := p[k]
	if !ok {
		return 0
	}
	return i.(int32)
}
