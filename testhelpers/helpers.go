package testhelpers

import (
	"fmt"
	"io"
	"net/http"
)

// Intp returns a pointer to the given integer value.
func Intp(i int) *int {
	return &i
}

// MustNewRequest creates a new http get request or panics if it can't
func MustNewRequest(method, urlStr string, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(fmt.Sprintf("failed to create HTTP %s Request for '%s': %s", method, urlStr, err))
	}
	return request
}
