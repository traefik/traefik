package tap

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"net"
	"net/http"
	"time"

	"github.com/traefik/traefik/v3/pkg/middlewares"
)

var _ middlewares.Stateful = &responseCapturer{}

// FIXME the request and response should have a max body size option like in the buffering middleware,
//  to avoid keeping large body, when the body is larger than the configured max size,
//  an error can be returned.

// FIXME what about the trailers, recorded headers.

// responseCapturer captures the response for the record.
// While buffering, nothing reaches the client until serve is called,
// which is what allows the response record to be sent before the response itself.
type responseCapturer struct {
	rw   http.ResponseWriter
	dest *destination

	// buffering reports whether the response is held back instead of being streamed to the client.
	// It ends once the response has been served, from then on the response is streamed.
	buffering bool
	// headers holds the response headers while the response is buffered. They are kept apart from
	// the client ones, as the response can end up being replaced by an error that must not carry them.
	headers http.Header
	// recordedHeaders is the snapshot of the headers, taken when the response status is known.
	recordedHeaders http.Header

	// buf holds the whole response body while buffering,
	// and at most dest.maxBodySize bytes otherwise.
	buf bytes.Buffer

	status        int
	bodyTruncated bool
}

func newResponseCapturer(rw http.ResponseWriter, dest *destination, shouldBuffer bool) *responseCapturer {
	r := &responseCapturer{
		rw:        rw,
		dest:      dest,
		buffering: shouldBuffer,
	}

	if shouldBuffer {
		// The headers start as a copy of the client ones, so that the headers set by the middlewares
		// standing before this one in the chain are part of the record and of the served response.
		r.headers = rw.Header().Clone()
		if r.headers == nil {
			r.headers = make(http.Header)
		}
	}

	return r
}

func (r *responseCapturer) Header() http.Header {
	if r.buffering {
		return r.headers
	}

	return r.rw.Header()
}

func (r *responseCapturer) WriteHeader(status int) {
	// An informational response is interim: the final status is the one that follows, so it must
	// not be recorded as the status of the response.
	// A buffered response sends nothing before its record is accepted, and an informational
	// response carries the response headers, so it is dropped. A client waiting for a
	// 100 Continue still gets it, as the server sends it on its own when the request body is read.
	if status >= 100 && status <= 199 {
		if !r.buffering {
			r.rw.WriteHeader(status)
		}

		return
	}

	if r.status != 0 {
		return
	}

	r.status = status
	r.recordedHeaders = r.Header().Clone()

	if !r.buffering {
		r.rw.WriteHeader(status)
	}
}

func (r *responseCapturer) Write(p []byte) (int, error) {
	if r.status == 0 {
		// The next handler wrote a body without setting a status.
		r.WriteHeader(http.StatusOK)
	}

	if r.buffering {
		// The whole body is kept, as it still has to be served to the client.
		return r.buf.Write(p)
	}

	r.capture(p)

	return r.rw.Write(p)
}

func (r *responseCapturer) Flush() {
	// A buffered response cannot be flushed, as nothing has been served yet.
	if r.buffering {
		return
	}

	if f, ok := r.rw.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseCapturer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.rw.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("not a hijacker: %T", r.rw)
	}

	// A hijacked connection is out of our control,
	// so the response cannot be buffered any longer.
	if err := r.serve(); err != nil {
		return nil, nil, err
	}

	return h.Hijack()
}

// capture keeps at most dest.maxBodySize bytes of the body for the record.
func (r *responseCapturer) capture(p []byte) {
	if !r.dest.body {
		return
	}

	if r.dest.maxBodySize < 0 {
		r.buf.Write(p)
		return
	}

	remaining := r.dest.maxBodySize - int64(r.buf.Len())
	if remaining <= 0 {
		r.bodyTruncated = len(p) > 0
		return
	}

	if int64(len(p)) > remaining {
		r.buf.Write(p[:remaining])
		r.bodyTruncated = true
		return
	}

	r.buf.Write(p)
}

// serve serves the buffered response to the client.
// It is a no-op for a response that has been streamed, or already served.
func (r *responseCapturer) serve() error {
	if !r.buffering {
		return nil
	}

	// Hijack can serve the response before ServeHTTP does, and it must be served only once.
	r.buffering = false

	// The headers are replaced rather than merged, so that a header set upstream that the next
	// handler removed is not served.
	clear(r.rw.Header())
	maps.Copy(r.rw.Header(), r.headers)

	if r.status != 0 {
		r.rw.WriteHeader(r.status)
	}

	if r.buf.Len() == 0 {
		return nil
	}

	if _, err := r.rw.Write(r.buf.Bytes()); err != nil {
		return fmt.Errorf("writing response: %w", err)
	}

	return nil
}

func (r *responseCapturer) record(duration time.Duration) *responseRecord {
	status := r.status
	if status == 0 {
		// The next handler returned without writing anything.
		status = http.StatusOK
	}

	headers := r.recordedHeaders
	if headers == nil {
		headers = r.Header().Clone()
	}

	rec := &responseRecord{
		Status:   status,
		Headers:  headers,
		Duration: duration,
	}

	if !r.dest.body {
		return rec
	}

	// When the response is buffered, buf holds the whole body, so the limit is applied here.
	body := r.buf.Bytes()
	if r.dest.maxBodySize >= 0 && int64(len(body)) > r.dest.maxBodySize {
		rec.Body = body[:r.dest.maxBodySize]
		rec.BodyTruncated = true

		return rec
	}

	rec.Body = body
	rec.BodyTruncated = r.bodyTruncated

	return rec
}
