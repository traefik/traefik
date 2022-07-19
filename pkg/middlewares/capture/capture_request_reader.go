package capture

import (
	"io"
)

type requestReader struct {
	// source ReadCloser from where the request body is read.
	source io.ReadCloser
	// size Counts the number of bytes read (when requestReader.Read is called).
	size int64
}

func newRequestReader(source io.ReadCloser) *requestReader {
	return &requestReader{source: source}
}

func (r *requestReader) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	r.size += int64(n)
	return n, err
}

func (r *requestReader) Close() error {
	return r.source.Close()
}

func (r *requestReader) Size() int64 {
	return r.size
}
