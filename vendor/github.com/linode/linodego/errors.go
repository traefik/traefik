package linodego

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-resty/resty"
)

const (
	// ErrorFromString is the Code identifying Errors created by string types
	ErrorFromString = 1
	// ErrorFromError is the Code identifying Errors created by error types
	ErrorFromError = 2
	// ErrorFromStringer is the Code identifying Errors created by fmt.Stringer types
	ErrorFromStringer = 3
)

// Error wraps the LinodeGo error with the relevant http.Response
type Error struct {
	Response *http.Response
	Code     int
	Message  string
}

// APIErrorReason is an individual invalid request message returned by the Linode API
type APIErrorReason struct {
	Reason string `json:"reason"`
	Field  string `json:"field"`
}

func (r APIErrorReason) Error() string {
	if len(r.Field) == 0 {
		return r.Reason
	}
	return fmt.Sprintf("[%s] %s", r.Field, r.Reason)
}

// APIError is the error-set returned by the Linode API when presented with an invalid request
type APIError struct {
	Errors []APIErrorReason `json:"errors"`
}

func coupleAPIErrors(r *resty.Response, err error) (*resty.Response, error) {
	if err != nil {
		return nil, NewError(err)
	}

	if r.Error() != nil {
		apiError, ok := r.Error().(*APIError)
		if !ok || (ok && len(apiError.Errors) == 0) {
			return r, nil
		}
		return nil, NewError(r)
	}

	return r, nil
}

func (e APIError) Error() string {
	var x []string
	for _, msg := range e.Errors {
		x = append(x, msg.Error())
	}
	return strings.Join(x, "; ")
}

func (g Error) Error() string {
	return fmt.Sprintf("[%03d] %s", g.Code, g.Message)
}

// NewError creates a linodego.Error with a Code identifying the source err type,
// - ErrorFromString   (1) from a string
// - ErrorFromError    (2) for an error
// - ErrorFromStringer (3) for a Stringer
// - HTTP Status Codes (100-600) for a resty.Response object
func NewError(err interface{}) *Error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *Error:
		return e
	case *resty.Response:
		apiError, ok := e.Error().(*APIError)

		if !ok {
			log.Fatalln("Unexpected Resty Error Response")
		}

		return &Error{
			Code:     e.RawResponse.StatusCode,
			Message:  apiError.Error(),
			Response: e.RawResponse,
		}
	case error:
		return &Error{Code: ErrorFromError, Message: e.Error()}
	case string:
		return &Error{Code: ErrorFromString, Message: e}
	case fmt.Stringer:
		return &Error{Code: ErrorFromStringer, Message: e.String()}
	default:
		log.Fatalln("Unsupported type to linodego.NewError")
		panic(err)
	}
}
