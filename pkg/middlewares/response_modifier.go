package middlewares

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ResponseModifier is a ResponseWriter to modify the response headers before sending them.
type ResponseModifier struct {
	req *http.Request
	rw  http.ResponseWriter

	headersSent bool // whether headers have already been sent
	code        int  // status code, must default to 200

	modifier    func(*http.Response) error // can be nil
	modified    bool                       // whether modifier has already been called for the current request
	modifierErr error                      // returned by modifier call
}

// NewResponseModifier returns a new ResponseModifier instance.
// The given modifier can be nil.
func NewResponseModifier(w http.ResponseWriter, r *http.Request, modifier func(*http.Response) error) http.ResponseWriter {
	return &ResponseModifier{
		req:      r,
		rw:       w,
		modifier: modifier,
		code:     http.StatusOK,
	}
}

// WriteHeader is, in the specific case of 1xx status codes, a direct call to the wrapped ResponseWriter, without marking headers as sent,
// allowing so further calls.
func (r *ResponseModifier) WriteHeader(code int) {
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
		log.Error().Err(err).Msg("Error when applying response modifier")
		r.rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.modified = true
	r.rw.WriteHeader(code)
}

func (r *ResponseModifier) Header() http.Header {
	return r.rw.Header()
}

func (r *ResponseModifier) Write(b []byte) (int, error) {
	r.WriteHeader(r.code)
	if r.modifierErr != nil {
		return 0, r.modifierErr
	}

	return r.rw.Write(b)
}

// Hijack hijacks the connection.
func (r *ResponseModifier) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.rw.(http.Hijacker); ok {
		return h.Hijack()
	}

	return nil, nil, fmt.Errorf("not a hijacker: %T", r.rw)
}

// Flush sends any buffered data to the client.
func (r *ResponseModifier) Flush() {
	if flusher, ok := r.rw.(http.Flusher); ok {
		flusher.Flush()
	}
}
