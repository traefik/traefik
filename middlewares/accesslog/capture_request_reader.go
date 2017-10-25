package accesslog

import (
	"io"
	"time"
)

type captureRequestReader struct {
	source        io.ReadCloser
	count         int64
	processingEnd time.Time
}

func (r *captureRequestReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	r.count += int64(n)
	if err == io.EOF {
		r.processingEnd = time.Now().UTC()
	}
	return n, err
}

func (r *captureRequestReader) Close() error {
	return r.source.Close()
}
