package accesslog

import "io"

type captureRequestReader struct {
	source io.ReadCloser
	count  int64
}

func (r *captureRequestReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	r.count += int64(n)
	return n, err
}

func (r *captureRequestReader) Close() error {
	return r.source.Close()
}
