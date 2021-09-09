package metrics

import (
	"fmt"
	"io"
	"net/http"

	gokitmetrics "github.com/go-kit/kit/metrics"
)

type bodyWrapper struct {
	io.ReadCloser
	counter gokitmetrics.Counter
}

func newBodyWrapper(body io.ReadCloser, counter gokitmetrics.Counter) *bodyWrapper {
	return &bodyWrapper{
		body,
		counter,
	}
}

// Read calls the wrapped ReadCloser Read and increments the
//      counter by the number of read bytes
func (b *bodyWrapper) Read(p []byte) (int, error) {
	r, e := b.ReadCloser.Read(p)
	b.add(r)
	return r, e
}

func (b *bodyWrapper) add(v int) {
	b.counter.Add(float64(v))
}

type responseWriterWrapper struct {
	http.ResponseWriter
	counter gokitmetrics.Counter
}

func newResponseWritrWrapper(rw http.ResponseWriter, counter gokitmetrics.Counter) *responseWriterWrapper {
	return &responseWriterWrapper{
		rw,
		counter,
	}
}

// Write calls the wrapped ResponseWriter write and increments the
//      counter by the number of written bytes
func (r *responseWriterWrapper) Write(d []byte) (int, error) {
	r.add(len(d))
	return r.ResponseWriter.Write(d)
}

func (r *responseWriterWrapper) add(v int) {
	r.counter.Add(float64(v))
}

func requestHeaderSize(req *http.Request) int {
	// some headers are not sent from the client
	// like X-Forwarded-Server, should they be counted or not?
	size := 1

	size += len("Host: ") + len(req.Host) + 1
	size += len(req.Proto) + 1
	size += len(req.URL.Path) + 1
	size += len(req.Method) + 1
	size += headerSize(req.Header) + 1

	return size
}

func responseHeaderSize(h http.Header, proto string, status int) int {
	size := 1
	size += len(proto) + 1
	size += len(fmt.Sprintf("%d", status)) + 1
	size += len(http.StatusText(status)) + 1
	size += headerSize(h) + 1

	return size
}

func headerSize(h http.Header) int {
	size := 0
	for k, v := range h {
		for _, e := range v {
			size += len(k) + len(e) + 3
		}
	}
	return size
}
