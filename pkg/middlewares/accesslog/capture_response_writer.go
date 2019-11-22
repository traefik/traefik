package accesslog

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/vulcand/oxy/utils"
)

var (
	_ middlewares.Stateful = &captureResponseWriter{}
)

func newCaptureResponseWriter(rw http.ResponseWriter) *captureResponseWriter {
	return &captureResponseWriter{
		rw:     rw,
		status: http.StatusOK,
		header: make(http.Header),
	}
}

// captureResponseWriter is a wrapper of type http.ResponseWriter
// that tracks response status, header, and size.
type captureResponseWriter struct {
	rw     http.ResponseWriter
	status int
	size   int64
	// header is a cached version of rw.Header(), that is safe to be accessed even
	// after ServeHTTP has returned.
	header     http.Header
	headerSent bool

	tooLateMu sync.RWMutex
	// toolate defines whether ServeHTTP has already terminated, and hence whether
	// it is still allowed to access rw.Header.
	tooLate bool
}

// LockHeader declares that ServeHTTP has already terminated, and that
// thereafter calls to Header will return a cached version of the headers, instead
// of the (forbidden) live ones from the response.
func (crw *captureResponseWriter) LockHeader() {
	crw.tooLateMu.Lock()
	defer crw.tooLateMu.Unlock()
	crw.tooLate = true
}

func (crw *captureResponseWriter) Header() http.Header {
	crw.tooLateMu.RLock()
	defer crw.tooLateMu.RUnlock()
	if crw.tooLate {
		return crw.header
	}
	return crw.rw.Header()
}

func (crw *captureResponseWriter) Write(b []byte) (int, error) {
	size, err := crw.rw.Write(b)
	crw.size += int64(size)
	if !crw.headerSent {
		utils.CopyHeaders(crw.header, crw.rw.Header())
		crw.headerSent = true
	}
	return size, err
}

func (crw *captureResponseWriter) WriteHeader(s int) {
	crw.status = s
	if crw.headerSent {
		return
	}
	crw.headerSent = true
	crw.rw.WriteHeader(crw.status)
	utils.CopyHeaders(crw.header, crw.rw.Header())
}

func (crw *captureResponseWriter) Flush() {
	if f, ok := crw.rw.(http.Flusher); ok {
		f.Flush()
	}
}

func (crw *captureResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := crw.rw.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("not a hijacker: %T", crw.rw)
}

func (crw *captureResponseWriter) CloseNotify() <-chan bool {
	if c, ok := crw.rw.(http.CloseNotifier); ok {
		return c.CloseNotify()
	}
	return nil
}

func (crw *captureResponseWriter) Status() int {
	return crw.status
}

func (crw *captureResponseWriter) Size() int64 {
	return crw.size
}
