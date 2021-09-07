package metrics

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type BodyWrapper struct {
	body io.ReadCloser
	read uint64
}

func NewBodyWrapper(body io.ReadCloser) *BodyWrapper {
	return &BodyWrapper{
		body: body,
		read: 0,
	}
}

func (b *BodyWrapper) Close() error {
	return b.body.Close()
}

func (b *BodyWrapper) Read(p []byte) (int, error) {
	r, e := b.body.Read(p)
	b.read += uint64(r)
	// log.Debug().Str("body", string(p)).Int("len", r).Msg("read")
	return r, e
}

type ResponseWriterWrapper struct {
	rw   http.ResponseWriter
	sent uint64
}

func NewResponseWritrWrapper(rw http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		rw:   rw,
		sent: 0,
	}
}

func (r *ResponseWriterWrapper) Header() http.Header {
	return r.rw.Header()
}

func (r *ResponseWriterWrapper) Write(d []byte) (int, error) {
	x := uint64(len(d))
	r.sent += x
	// log.Debug().Str("body", string(d)).Uint64("len", x).Msg("sending")
	return r.rw.Write(d)
}

func (r *ResponseWriterWrapper) WriteHeader(statusCode int) {
	r.rw.WriteHeader(statusCode)
}

// Hijack hijacks the connection.
func (r *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.rw.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (r *ResponseWriterWrapper) Flush() {
	if f, ok := r.rw.(http.Flusher); ok {
		f.Flush()
	}
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
