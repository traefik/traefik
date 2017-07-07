package audittap

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/containous/traefik/middlewares/audittap/audittypes"
)

// Performance tweak: choose a value just big enough to hold most messages.
// It must not be bigger than MaximumEntityLength.
const initialBufferSize = 2048

type recorderResponseWriter struct {
	http.ResponseWriter
	status          int
	size            int
	entity          []byte
	maxEntityLength int
}

// NewAuditResponseWriter creates a ResponseWriter that captures extra information.
func NewAuditResponseWriter(w http.ResponseWriter, maxEntityLength int) audittypes.AuditResponseWriter {
	entity := make([]byte, 0, initialBufferSize)
	return &recorderResponseWriter{w, 0, 0, entity, maxEntityLength}
}

func (r *recorderResponseWriter) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.status = code
}

func (r *recorderResponseWriter) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (r *recorderResponseWriter) Write(b []byte) (int, error) {
	if r.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		r.WriteHeader(http.StatusOK)
	}

	size, err := r.ResponseWriter.Write(b)
	r.size += size

	if len(r.entity) < r.maxEntityLength {
		n := len(r.entity) + len(b) - r.maxEntityLength
		if n > 0 {
			b = b[:n]
		}
		r.entity = append(r.entity, b...)
	}

	return size, err
}

func (r *recorderResponseWriter) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (r *recorderResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (r *recorderResponseWriter) SummariseResponse(summary *audittypes.Summary) {
	hdr := NewHeaders(r.Header()).SimplifyCookies()
	flatHdr := hdr.Flatten()

	summary.ResponseStatus = strconv.Itoa(r.status)
	summary.ResponseHeaders = hdr.ResponseHeaders()
	summary.ResponsePayload = audittypes.DataMap{
		"type": flatHdr.GetString("content-type"),
	}
}
