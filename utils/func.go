package utils

import "fmt"

// CoverError == catch{}
func CoverError(fun func()) error {
	if fun == nil {
		return nil
	}
	errCh := make(chan error, 1)
	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					errCh <- err
				} else {
					errCh <- fmt.Errorf("%v", r)
				}
			} else {
				errCh <- nil
			}
		}()
		fun()
	}()
	return <-errCh
}
