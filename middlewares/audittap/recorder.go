package audittap

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/containous/traefik/middlewares/audittap/types"
)

// Performance tweak: choose a value just big enough to hold most messages.
// It must not be bigger than MaximumEntityLength.
const initialBufferSize = 2048

type recorderResponseWriter struct {
	http.ResponseWriter
	types.ResponseInfo
}

// NewAuditResponseWriter creates a ResponseWriter that captures extra information.
func NewAuditResponseWriter(w http.ResponseWriter, maxEntityLength int) types.AuditResponseWriter {
	entity := make([]byte, 0, initialBufferSize)
	return &recorderResponseWriter{w, types.ResponseInfo{0, 0, entity, maxEntityLength}}
}

func (r *recorderResponseWriter) GetResponseInfo() types.ResponseInfo {
	return r.ResponseInfo
}

func (r *recorderResponseWriter) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.Status = code
}

func (r *recorderResponseWriter) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (r *recorderResponseWriter) Write(b []byte) (int, error) {
	if r.Status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		r.WriteHeader(http.StatusOK)
	}

	size, err := r.ResponseWriter.Write(b)
	r.Size += size

	if len(r.Entity) < r.MaxEntityLength {
		n := len(r.Entity) + len(b) - r.MaxEntityLength
		if n > 0 {
			b = b[:n]
		}
		r.Entity = append(r.Entity, b...)
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
