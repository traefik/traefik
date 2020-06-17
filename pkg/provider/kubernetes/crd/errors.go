package crd

import (
	"errors"
	"fmt"
)

var (
	// ErrResourceNotFound ...
	ErrResourceNotFound = fmt.Errorf("resource not found")
)

// Error is an error wrapper with a Log property which indicates
// whether the error should be logged or not.
type Error struct {
	err error
	// Log tells wether or not this error should be logged.
	Log bool
}

// Error implements error interface.
func (e Error) Error() string {
	return e.err.Error()
}

// Unwrap allows Error to be unwrapped.
func (e Error) Unwrap() error {
	return errors.Unwrap(e.err)
}
