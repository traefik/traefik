package metrics

import (
	"io"
	"net/http"
)

type BodyWrapper struct {
	io.ReadCloser
	read uint64
}

func NewBodyWrapper(body io.ReadCloser) *BodyWrapper {
	return &BodyWrapper{
		body,
		0,
	}
}

func (b *BodyWrapper) Read(p []byte) (int, error) {
	r, e := b.ReadCloser.Read(p)
	b.read += uint64(r)
	return r, e
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	sent uint64
}

func NewResponseWritrWrapper(rw http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		rw,
		0,
	}
}

func (r *ResponseWriterWrapper) Write(d []byte) (int, error) {
	r.sent += uint64(len(d))
	return r.ResponseWriter.Write(d)
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

func responseHeaderSize(h http.Header, proto string, status string) int {
	// some headers are not sent from the client
	// like X-Forwarded-Server, should they be counted or not?
	size := 1
	size += len(proto) + 1
	size += len(status) + 1
	size += headerSize(h) + 1

	return size
}

func headerSize(h http.Header) int {
	// some headers are not sent from the client
	// like X-Forwarded-Server, should they be counted or not?
	size := 1
	for k, v := range h {
		for _, e := range v {
			size += len(k) + len(e) + 3
		}
	}
	return size
}
