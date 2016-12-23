package utils

import (
	"errors"
	"github.com/cenk/backoff"
	"net/http"
	"strconv"
	"time"
)

// TryRequest try operation timeout, and retry backoff
func TryRequest(url string, timeout time.Duration, condition Condition) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = timeout
	var res *http.Response
	err := backoff.Retry(func() error {
		var err error
		res, err = http.Get(url)
		if err != nil {
			return err
		}
		return condition(res)
	}, exponentialBackOff)
	return err
}

// Try try operation timeout, and retry backoff
func Try(timeout time.Duration, operation func() error) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = timeout
	err := backoff.Retry(operation, exponentialBackOff)
	return err
}

// Condition is a retry condition function.
// It receives a response, and returns an error
// if the response failed the condition.
type Condition func(*http.Response) error

// ErrorIfStatusCodeIsNot returns a retry condition function.
// The condition returns an error
// if the given response's status code is not the given HTTP status code.
func ErrorIfStatusCodeIsNot(status int) Condition {
	return func(res *http.Response) error {
		if res.StatusCode != status {
			return errors.New("Bad status. Got: " + res.Status + ", expected:" + strconv.Itoa(status))
		}
		return nil
	}
}
