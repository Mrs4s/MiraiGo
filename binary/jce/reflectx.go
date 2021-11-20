package jce

import (
	"reflect"
	"unsafe"
)

type value struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
	flag uintptr
}

func pointerOf(v reflect.Value) unsafe.Pointer {
	return (*value)(unsafe.Pointer(&v)).data
}
