package proto

type Message interface {
	Marshal() ([]byte, error)
}

func String(v string) *string { return &v }

func Bool(v bool) *bool { return &v }

func Int32(v int32) *int32 { return &v }

func Int64(v int64) *int64 { return &v }

func Uint32(v uint32) *uint32 { return &v }

func Uint64(v uint64) *uint64 { return &v }

func Float32(v float32) *float32 { return &v }

func Float64(v float64) *float64 { return &v }
