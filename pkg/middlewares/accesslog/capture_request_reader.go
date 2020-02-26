package accesslog

import "io"

type captureRequestReader struct {
	// source ReadCloser from where the request body is read.
	source io.ReadCloser
	// count Counts the number of bytes read (when captureRequestReader.Read is called).
	count int64
}

func (r *captureRequestReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	r.count += int64(n)
	return n, err
}

func (r *captureRequestReader) Close() error {
	return r.source.Close()
}
