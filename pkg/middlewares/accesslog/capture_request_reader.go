package accesslog

import (
	"io"
	"net/http"
)

type captureRequestReader struct {
	req   *http.Request
	count int64
}

func (r *captureRequestReader) Read(p []byte) (int, error) {
	if !r.req.Close {
		n, err := r.req.Body.Read(p)
		if err != nil {
			return 0, err
		}
		r.count += int64(n)
		return n, err
	}
	return 0, io.ErrClosedPipe
}

func (r *captureRequestReader) Close() error {
	if !r.req.Close {
		return r.req.Body.Close()
	}
	return io.ErrClosedPipe
}
