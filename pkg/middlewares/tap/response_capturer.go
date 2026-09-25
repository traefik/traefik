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

var (
	_ middlewares.Stateful = &responseCapturer{}
	_ middlewares.Stateful = &statusRecorder{}
)

// responseCapturer captures the response for the record.
// When withheld, nothing reaches the client until serve is called, which is what allows
// the response record to be sent before the response itself.
type responseCapturer struct {
	rw   http.ResponseWriter
	dest *destination

	// withheld reports whether the response is buffered instead of being streamed to the client.
	withheld bool
	// headers holds the response headers, either the client ones or, when withheld, a detached map.
	headers http.Header
	// recordedHeaders is the snapshot of the headers, taken when the response status is known.
	recordedHeaders http.Header

	// buf holds the whole response body when withheld, and at most dest.maxBodySize bytes otherwise.
	buf bytes.Buffer

	status        int
	bodyTruncated bool
	written       bool
	hijacked      bool
}

func newResponseCapturer(rw http.ResponseWriter, dest *destination, withhold bool) *responseCapturer {
	// A detached header map is needed, because the response can end up being replaced by an error.
	// It starts as a copy of the client one, so that the headers set by the middlewares standing
	// before this one in the chain are part of the record.
	headers := rw.Header()
	if withhold {
		headers = headers.Clone()
		if headers == nil {
			headers = make(http.Header)
		}
	}

	return &responseCapturer{rw: rw, dest: dest, withheld: withhold, headers: headers}
}

func (r *responseCapturer) Header() http.Header {
	return r.headers
}

func (r *responseCapturer) WriteHeader(status int) {
	// An informational response is interim: it is forwarded as it comes, even when the
	// response is withheld, because it commits nothing. The final status is the one that
	// follows, so it must not be recorded as the status of the response.
	if status >= http.StatusContinue && status < http.StatusOK {
		r.rw.WriteHeader(status)
		return
	}

	if r.status != 0 {
		return
	}

	r.status = status
	r.recordedHeaders = r.headers.Clone()

	if !r.withheld {
		r.rw.WriteHeader(status)
	}
}

func (r *responseCapturer) Write(p []byte) (int, error) {
	if r.status == 0 {
		// The next handler wrote a body without setting a status.
		r.WriteHeader(http.StatusOK)
	}

	if r.withheld && !r.hijacked {
		// The whole body is kept, as it still has to be served to the client.
		return r.buf.Write(p)
	}

	r.written = true
	r.capture(p)

	return r.rw.Write(p)
}

func (r *responseCapturer) Flush() {
	// A withheld response cannot be flushed, as nothing has been served yet.
	if r.withheld && !r.hijacked {
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

	// A hijacked connection is out of our control, so the response cannot be withheld any longer.
	if r.withheld && !r.hijacked {
		if err := r.serve(); err != nil {
			return nil, nil, err
		}

		r.hijacked = true
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

// served reports whether a part of the response already reached the client,
// which means it is too late to replace it with an error.
func (r *responseCapturer) served() bool {
	return !r.withheld || r.hijacked || r.written
}

// serve serves the withheld response to the client. It is a no-op for a response that has been streamed.
func (r *responseCapturer) serve() error {
	if !r.withheld || r.hijacked {
		return nil
	}

	// The response is served only once, as Hijack can serve it before ServeHTTP returns.
	r.withheld = false
	r.written = true

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
		headers = r.headers.Clone()
	}

	rec := &responseRecord{
		Status:   status,
		Headers:  headers,
		Duration: duration,
	}

	if !r.dest.body {
		return rec
	}

	// When the response is withheld, buf holds the whole body, so the limit is applied here.
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

// statusRecorder is the http.ResponseWriter given to the tap services, discarding their response body.
type statusRecorder struct {
	header http.Header
	status int
}

func (s *statusRecorder) Header() http.Header {
	return s.header
}

func (s *statusRecorder) WriteHeader(status int) {
	if s.status == 0 {
		s.status = status
	}
}

func (s *statusRecorder) Write(p []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}

	return len(p), nil
}

func (s *statusRecorder) Flush() {}

func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("connection on %T cannot be hijacked", s)
}
