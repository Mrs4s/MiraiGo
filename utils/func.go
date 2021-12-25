package utils

import "fmt"

func CoverError(fun func()) error {
	errCh := make(chan error)
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				errCh <- err
			} else {
				errCh <- fmt.Errorf("%v", r)
			}
		}
	}()
	fun()
	errCh <- nil
	return <-errCh
}
