package utils

import "fmt"

// CoverError == catch{}
func CoverError(fun func()) (err error) {
	if fun == nil {
		return nil
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = e
				} else {
					e = fmt.Errorf("%v", r)
				}
			}
		}()
		fun()
	}()
	return
}
