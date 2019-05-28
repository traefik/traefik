package types

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Error groups together information that defines an error. Should always be used to
type Error struct {
	Message string   `json:"message"`
	Code    int      `json:"code"`
	Details []string `json:"details,omitempty"`
	Err     error    `json:"-"`
}

// Error() gives a string representing the error; also, forces the Error type to comply with the error interface
func (e *Error) Error() string {
	msg := fmt.Sprintf("ERROR (%v): %s; \n Inner error: %s", e.Code, e.Message, e.Err)
	logrus.Debug(msg)
	return msg
}

// BadRequestError create an Error instance with http.StatusBadRequest code
func BadRequestError(message string, err error, details ...string) *Error {
	return &Error{Message: message, Err: err, Code: http.StatusBadRequest, Details: details}
}

// BadRequestError create an Error instance with http.StatusNotFound code
func NotFoundError(message string, err error, details ...string) *Error {
	return &Error{Message: message, Err: err, Code: http.StatusNotFound, Details: details}
}

// BadRequestError create an Error instance with http.StatusInternalServerError code
func InternalServerError(message string, err error, details ...string) *Error {
	return &Error{Message: message, Err: err, Code: http.StatusInternalServerError, Details: details}
}

// PanicIfError is just a wrapper to a panic call that propagates error when it's not nil
func PanicIfError(e error) {
	if e != nil {
		logrus.Errorf(e.Error())
		panic(e)
	}
}

// Panic wraps a panic call propagating the given error parameter
func Panic(e Error) {
	logrus.Errorf(e.Error())
	panic(e)
}
