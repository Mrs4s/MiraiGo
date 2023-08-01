package proto

import (
	"reflect"

	"github.com/RomiChan/protobuf/proto"
)

// TODO: move to a new package
const debug = false

type Message = any

func Marshal(m Message) ([]byte, error) {
	b, err := proto.Marshal(m)
	if err != nil {
		return b, err
	}
	if debug {
		t := reflect.TypeOf(m).Elem()
		n := reflect.New(t)
		err = Unmarshal(b, n.Interface())
		if err != nil {
			panic(err)
		}
		if reflect.DeepEqual(m, n) {
			panic("not equal")
		}
	}
	return b, err
}

func Unmarshal(b []byte, m Message) error {
	return proto.Unmarshal(b, m)
}

func Some[T any](val T) proto.Option[T] {
	return proto.Some(val)
}

func None[T any]() proto.Option[T] {
	return proto.None[T]()
}

// Bool stores v in a new bool value and returns a pointer to it.
func Bool(v bool) proto.Option[bool] { return proto.Some(v) }

// Int32 stores v in a new int32 value and returns a pointer to it.
func Int32(v int32) proto.Option[int32] { return proto.Some(v) }

// Int64 stores v in a new int64 value and returns a pointer to it.
func Int64(v int64) proto.Option[int64] { return proto.Some(v) }

// Float32 stores v in a new float32 value and returns a pointer to it.
func Float32(v float32) proto.Option[float32] { return proto.Some(v) }

// Float64 stores v in a new float64 value and returns a pointer to it.
func Float64(v float64) proto.Option[float64] { return proto.Some(v) }

// Uint32 stores v in a new uint32 value and returns a pointer to it.
func Uint32(v uint32) proto.Option[uint32] { return proto.Some(v) }

// Uint64 stores v in a new uint64 value and returns a pointer to it.
func Uint64(v uint64) proto.Option[uint64] { return proto.Some(v) }

// String stores v in a new string value and returns a pointer to it.
func String(v string) proto.Option[string] { return proto.Some(v) }
