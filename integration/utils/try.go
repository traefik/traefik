package utils

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

const maxWaitTime = 5 * time.Second

// TryGetResponse is like TryRequest, but returns the response for further
// processing at the call site.
// Conditions are not allowed since it would complicate signaling if the
// response body needs to be closed or not. Callers are expected to close on
// their own if the function returns a nil error.
func TryGetResponse(url string, timeout time.Duration) (*http.Response, error) {
	return tryGetResponse(url, timeout)
}

// TryRequest is like Try, but runs a request against the given URL and applies
// the condition on the response.
// Condition may be nil, in which case only the request against the URL must
// succeed.
func TryRequest(url string, timeout time.Duration, condition Condition) error {
	resp, err := tryGetResponse(url, timeout)
	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if condition != nil {
		return condition(resp)
	}

	return nil
}

// Try repeatedly executes an operation until no error condition occurs or the
// given timeout is reached, whatever comes first.
func Try(timeout time.Duration, operation func() error) error {
	if timeout <= 0 {
		panic("timeout must be larger than zero")
	}

	wait := time.Duration(math.Ceil(float64(timeout) / 10.0))
	if wait > maxWaitTime {
		wait = maxWaitTime
	}

	var err error
	if err = operation(); err == nil {
		return nil
	}

	for {
		select {
		case <-time.After(timeout):
			return fmt.Errorf("try operation failed: %s", err)
		case <-time.Tick(wait):
			if err = operation(); err == nil {
				return nil
			}
		}
	}
}

// Condition is a retry condition function.
// It receives a response, and returns an error
// if the response failed the condition.
type Condition func(*http.Response) error

// BodyContains returns a retry condition function.
// The condition returns an error if the request body does not contain the given
// string.
func BodyContains(s string) Condition {
	return func(res *http.Response) error {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %s", err)
		}

		if !strings.Contains(string(body), s) {
			return fmt.Errorf("could not find '%s' in body '%s'", s, string(body))
		}
		return nil
	}
}

// StatusCodeIs returns a retry condition function.
// The condition returns an error if the given response's status code is not the
// given HTTP status code.
func StatusCodeIs(status int) Condition {
	return func(res *http.Response) error {
		if res.StatusCode != status {
			return fmt.Errorf("got status code %d, wanted %d", res.StatusCode, status)
		}
		return nil
	}
}

func tryGetResponse(url string, timeout time.Duration) (*http.Response, error) {
	var resp *http.Response
	return resp, Try(timeout, func() error {
		var err error
		resp, err = http.Get(url)
		return err
	})
}
