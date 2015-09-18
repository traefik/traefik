package backoff

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// This is an example that demonstrates how this package could be used
// to perform various advanced operations.
//
// It executes an HTTP GET request with exponential backoff,
// while errors are logged and failed responses are closed, as required by net/http package.
//
// Note we define a condition function which is used inside the operation to
// determine whether the operation succeeded or failed.
func Example() error {
	res, err := GetWithRetry(
		"http://localhost:9999",
		ErrorIfStatusCodeIsNot(http.StatusOK),
		NewExponentialBackOff())

	if err != nil {
		// Close response body of last (failed) attempt.
		// The Last attempt isn't handled by the notify-on-error function,
		// which closes the body of all the previous attempts.
		if e := res.Body.Close(); e != nil {
			log.Printf("error closing last attempt's response body: %s", e)
		}
		log.Printf("too many failed request attempts: %s", err)
		return err
	}
	defer res.Body.Close() // The response's Body must be closed.

	// Read body
	_, _ = ioutil.ReadAll(res.Body)

	// Do more stuff
	return nil
}

// GetWithRetry is a helper function that performs an HTTP GET request
// to the given URL, and retries with the given backoff using the given condition function.
//
// It also uses a notify-on-error function which logs
// and closes the response body of the failed request.
func GetWithRetry(url string, condition Condition, bck BackOff) (*http.Response, error) {
	var res *http.Response
	err := RetryNotify(
		func() error {
			var err error
			res, err = http.Get(url)
			if err != nil {
				return err
			}
			return condition(res)
		},
		bck,
		LogAndClose())

	return res, err
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
			return NewError(res)
		}
		return nil
	}
}

// Error is returned on ErrorIfX() condition functions throughout this package.
type Error struct {
	Response *http.Response
}

func NewError(res *http.Response) *Error {
	// Sanity check
	if res == nil {
		panic("response object is nil")
	}
	return &Error{Response: res}
}
func (err *Error) Error() string { return "request failed" }

// LogAndClose is a notify-on-error function.
// It logs the error and closes the response body.
func LogAndClose() Notify {
	return func(err error, wait time.Duration) {
		switch e := err.(type) {
		case *Error:
			defer e.Response.Body.Close()

			b, err := ioutil.ReadAll(e.Response.Body)
			var body string
			if err != nil {
				body = "can't read body"
			} else {
				body = string(b)
			}

			log.Printf("%s: %s", e.Response.Status, body)
		default:
			log.Println(err)
		}
	}
}
