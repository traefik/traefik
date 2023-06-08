package headers

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/log"
)

type responseModifier struct {
	req *http.Request
	rw  http.ResponseWriter

	headersSent bool // whether headers have already been sent
	code        int  // status code, must default to 200

	modifier    func(*http.Response) error // can be nil
	modified    bool                       // whether modifier has already been called for the current request
	modifierErr error                      // returned by modifier call
}

// modifier can be nil.
func newResponseModifier(w http.ResponseWriter, r *http.Request, modifier func(*http.Response) error) http.ResponseWriter {
	rm := &responseModifier{
		req:      r,
		rw:       w,
		modifier: modifier,
		code:     http.StatusOK,
	}

	if _, ok := w.(http.CloseNotifier); ok {
		return responseModifierWithCloseNotify{responseModifier: rm}
	}
	return rm
}

// WriteHeader is, in the specific case of 1xx status codes, a direct call to the wrapped ResponseWriter, without marking headers as sent,
// allowing so further calls.
func (r *responseModifier) WriteHeader(code int) {
	if r.headersSent {
		return
	}

	// Handling informational headers.
	if code >= 100 && code <= 199 {
		r.rw.WriteHeader(code)
		return
	}

	defer func() {
		r.code = code
		r.headersSent = true
	}()

	if r.modifier == nil || r.modified {
		r.rw.WriteHeader(code)
		return
	}

	resp := http.Response{
		Header:  r.rw.Header(),
		Request: r.req,
	}

	if err := r.modifier(&resp); err != nil {
		r.modifierErr = err
		// we are propagating when we are called in Write, but we're logging anyway,
		// because we could be called from another place which does not take care of
		// checking w.modifierErr.
		log.WithoutContext().Errorf("Error when applying response modifier: %v", err)
		r.rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.modified = true
	r.rw.WriteHeader(code)
}

func (r *responseModifier) Header() http.Header {
	return r.rw.Header()
}

func (r *responseModifier) Write(b []byte) (int, error) {
	r.WriteHeader(r.code)
	if r.modifierErr != nil {
		return 0, r.modifierErr
	}

	return r.rw.Write(b)
}

// Hijack hijacks the connection.
func (r *responseModifier) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.rw.(http.Hijacker); ok {
		return h.Hijack()
	}

	return nil, nil, fmt.Errorf("not a hijacker: %T", r.rw)
}

// Flush sends any buffered data to the client.
func (r *responseModifier) Flush() {
	if flusher, ok := r.rw.(http.Flusher); ok {
		flusher.Flush()
	}
}

type responseModifierWithCloseNotify struct {
	*responseModifier
}

// CloseNotify implements http.CloseNotifier.
func (r *responseModifierWithCloseNotify) CloseNotify() <-chan bool {
	return r.responseModifier.rw.(http.CloseNotifier).CloseNotify()
}
