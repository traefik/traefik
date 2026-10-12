package tap

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/http"
	"time"

	"github.com/traefik/traefik/v3/pkg/middlewares"
)

var _ middlewares.Stateful = &responseCapturer{}

// responseCapturer captures the response for the record.
// While buffering, nothing reaches the client until serve, so that the record is sent first.
// The response is buffered until its record is complete, then sent, and the rest of the response is streamed.
type responseCapturer struct {
	rw   http.ResponseWriter
	dest *destination
	send func() error

	// buffering reports whether the response is held back, and ends once the response is served.
	buffering bool
	// headers holds the headers while buffering, apart from the client ones,
	// so that an error replacing the response does not carry them.
	headers http.Header
	// sentHeaders is the snapshot taken when the status is written:
	// the headers set afterwards are trailers, or ignored.
	sentHeaders http.Header

	// buf holds the body while buffering, and the recorded part otherwise.
	buf bytes.Buffer

	status        int
	bodyTruncated bool

	sent    bool
	sendErr error
	// discarding reports whether the buffered response is dropped, as its record could not be sent.
	discarding bool
}

func newResponseCapturer(rw http.ResponseWriter, dest *destination, shouldBuffer bool, send func() error) *responseCapturer {
	r := &responseCapturer{
		rw:        rw,
		dest:      dest,
		send:      send,
		buffering: shouldBuffer,
	}

	// A copy of the client headers, so that the ones set by the previous middlewares are kept.
	if shouldBuffer {
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
	// An informational status is not the status of the response.
	// A buffered response drops it, as it carries the headers which should not be sent to the client.
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
	r.sentHeaders = r.Header().Clone()

	if !r.buffering {
		r.rw.WriteHeader(status)
		return
	}

	// With no body to record, the record is complete once the status is known.
	if !r.dest.recordBody {
		_ = r.complete()
	}
}

func (r *responseCapturer) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}

	if r.buffering {
		// Once the record is rejected, the body is discarded rather than refused,
		// as a write error would make the reverse proxy abort the connection, with no way to answer.
		if r.discarding {
			return len(p), nil
		}

		r.buf.Write(p)

		// The record is complete once its body is, which a byte past dest.maxRecordBodySize tells.
		if r.dest.maxRecordBodySize >= 0 && int64(r.buf.Len()) > r.dest.maxRecordBodySize {
			_ = r.complete()
		}

		return len(p), nil
	}

	r.capture(p)

	return r.rw.Write(p)
}

func (r *responseCapturer) Flush() {
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

	// A hijacked connection is out of our control, so the response is served first,
	// as it should not be buffered anymore.
	if err := r.serve(); err != nil {
		return nil, nil, err
	}

	conn, brw, err := h.Hijack()

	// The status is then written on the hijacked connection, which the reverse proxy only does for a 101 Switching Protocols.
	if err == nil && r.status == 0 {
		r.status = http.StatusSwitchingProtocols
	}

	return conn, brw, err
}

// capture keeps at most dest.maxRecordBodySize bytes of the body for the record.
func (r *responseCapturer) capture(p []byte) {
	if !r.dest.recordBody {
		return
	}

	if r.dest.maxRecordBodySize < 0 {
		r.buf.Write(p)
		return
	}

	remaining := r.dest.maxRecordBodySize - int64(r.buf.Len())
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

// complete sends the record, once, and serves the buffered response when the record is accepted.
func (r *responseCapturer) complete() error {
	if r.sent {
		return r.sendErr
	}

	r.sent = true
	r.sendErr = r.send()

	if r.sendErr != nil && r.buffering {
		r.discarding = true
		r.buf = bytes.Buffer{}

		return r.sendErr
	}

	// A client failing to read the response is left to the next writes to report.
	_ = r.serve()

	return r.sendErr
}

// serve serves the buffered response.
// It is a no-op for a streamed, or already served, response.
func (r *responseCapturer) serve() error {
	if !r.buffering {
		return nil
	}

	if r.discarding {
		return errors.New("response record rejected")
	}

	// Hijack can serve the response before ServeHTTP does.
	r.buffering = false

	// Serving the sent headers keeps the declared trailers from being sent as headers too.
	// Replacing rather than merging drops the upstream headers the next handler removed.
	if r.sentHeaders != nil {
		clear(r.rw.Header())
		maps.Copy(r.rw.Header(), r.sentHeaders)
	}

	if r.status != 0 {
		r.rw.WriteHeader(r.status)
	}

	if r.buf.Len() > 0 {
		if _, err := r.rw.Write(r.buf.Bytes()); err != nil {
			return fmt.Errorf("writing response: %w", err)
		}
	}

	// The server reads the trailers from the header map once the handler has returned.
	clear(r.rw.Header())
	maps.Copy(r.rw.Header(), r.headers)

	return nil
}

func (r *responseCapturer) record(duration time.Duration) *responseRecord {
	status := r.status
	if status == 0 {
		status = http.StatusOK
	}

	headers := r.sentHeaders
	if headers == nil {
		headers = r.Header().Clone()
	}

	rec := &responseRecord{
		Status:   status,
		Headers:  headers,
		Duration: duration,
	}

	if !r.dest.recordBody {
		return rec
	}

	// While buffering, buf holds the buffered body, so the limit is applied here.
	body := r.buf.Bytes()
	if r.dest.maxRecordBodySize >= 0 && int64(len(body)) > r.dest.maxRecordBodySize {
		rec.Body = body[:r.dest.maxRecordBodySize]
		rec.BodyTruncated = true

		return rec
	}

	rec.Body = body
	rec.BodyTruncated = r.bodyTruncated

	return rec
}
