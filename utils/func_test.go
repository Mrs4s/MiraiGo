package utils

import (
	"errors"
	"testing"
)

var errTest = errors.New("test error")

func TestCoverError(t *testing.T) {
	err := CoverError(nil)
	if err != nil {
		t.Errorf(`CoverError(nil) = %v, want nil`, err)
	}

	err = CoverError(func() {
		panic("test")
	})
	if err.Error() != "test" {
		t.Errorf(`CoverError(func() { panic("test") }) = %v, want "test"`, err)
	}

	err = CoverError(func() {
		panic(errTest)
	})
	if err != errTest {
		t.Errorf(`CoverError(func() { panic(errTest) }) = %v, want errTest`, err)
	}
}
