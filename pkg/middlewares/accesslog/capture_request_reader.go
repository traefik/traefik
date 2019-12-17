package accesslog

import "io"

type captureRequestReader struct {
	// ReadCloser from where the request body is read
	source io.ReadCloser
	// Size of the request body incremented each time
	// captureRequestReader.Read is called
	size int64
}

func (r *captureRequestReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	r.size += int64(n)
	return n, err
}

func (r *captureRequestReader) Close() error {
	return r.source.Close()
}
