package testhelpers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// MustNewRequest creates a new http get request or panics if it can't
func MustNewRequest(method, urlStr string, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(fmt.Sprintf("failed to create HTTP %s Request for '%s': %s", method, urlStr, err))
	}
	return request
}

// MustParseURL parses a URL or panics if it can't
func MustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL '%s': %s", rawURL, err))
	}
	return u
}
