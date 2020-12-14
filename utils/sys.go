package utils

import (
	"reflect"
	"unsafe"
)

func IsChanClosed(ch interface{}) bool {
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		panic("object is not a channel.")
	}
	ptr := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&ch)) + unsafe.Sizeof(uint(0))))
	ptr += unsafe.Sizeof(uint(0)) * 2
	ptr += unsafe.Sizeof(uintptr(0))
	ptr += unsafe.Sizeof(uint16(0))
	return *(*uint32)(unsafe.Pointer(ptr)) > 0
}
