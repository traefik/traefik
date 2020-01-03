package accesslog

import (
	"errors"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

type captureRequestReader struct {
	req    *http.Request
	count  int64
	mu     sync.Mutex
	logger *logrus.Logger
}

func (r *captureRequestReader) GetCount() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.count
}

func (r *captureRequestReader) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Access to req.Body is unsafe so we add recovery management
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Errorf("error while reading req.Body: %#v", rec)
			n = 0
			switch x := rec.(type) {
			case error:
				err = x
			case string:
				err = errors.New(x)
			default:
				// Fallback err (per specs, error strings should be lowercase w/o punctuation
				err = errors.New("error while reading req.Body: unknown panic")
			}
		}
	}()

	n, err = r.req.Body.Read(p)
	r.count += int64(n)
	return n, err
}

func (r *captureRequestReader) Close() (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Access to req.Body is unsafe so we add recovery management
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Errorf("error while closing req.Body: %#v", rec)
			switch x := rec.(type) {
			case error:
				err = x
			case string:
				err = errors.New(x)
			default:
				// Fallback err (per specs, error strings should be lowercase w/o punctuation
				err = errors.New("error while closing req.Body: unknown panic")
			}
		}
	}()

	return r.req.Body.Close()
}
