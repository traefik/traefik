package http

import (
	"fmt"
)

const (
	// DomainNewRequest is an error during request generation.
	DomainNewRequest = "NewRequest"

	// DomainEncode is an error during request or response encoding.
	DomainEncode = "Encode"

	// DomainDo is an error during the execution phase of the request.
	DomainDo = "Do"

	// DomainDecode is an error during request or response decoding.
	DomainDecode = "Decode"
)

// Error is an error that occurred at some phase within the transport.
type Error struct {
	// Domain is the phase in which the error was generated.
	Domain string

	// Err is the concrete error.
	Err error
}

// Error implements the error interface.
func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Domain, e.Err)
}
