package headers

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/log"
)

type responseModifier struct {
	r *http.Request
	w http.ResponseWriter

	headersSent bool // whether headers have already been sent
	code        int  // status code, must default to 200

	modifier    func(*http.Response) error // can be nil
	modified    bool                       // whether modifier has already been called for the current request
	modifierErr error                      // returned by modifier call
}

// modifier can be nil.
func newResponseModifier(w http.ResponseWriter, r *http.Request, modifier func(*http.Response) error) *responseModifier {
	return &responseModifier{
		r:        r,
		w:        w,
		modifier: modifier,
		code:     http.StatusOK,
	}
}

func (w *responseModifier) WriteHeader(code int) {
	if w.headersSent {
		return
	}
	defer func() {
		w.code = code
		w.headersSent = true
	}()

	if w.modifier == nil || w.modified {
		w.w.WriteHeader(code)
		return
	}

	resp := http.Response{
		Header:  w.w.Header(),
		Request: w.r,
	}

	if err := w.modifier(&resp); err != nil {
		w.modifierErr = err
		// we are propagating when we are called in Write, but we're logging anyway,
		// because we could be called from another place which does not take care of
		// checking w.modifierErr.
		log.WithoutContext().Errorf("Error when applying response modifier: %v", err)
		w.w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.modified = true
	w.w.WriteHeader(code)
}

func (w *responseModifier) Header() http.Header {
	return w.w.Header()
}

func (w *responseModifier) Write(b []byte) (int, error) {
	w.WriteHeader(w.code)
	if w.modifierErr != nil {
		return 0, w.modifierErr
	}

	return w.w.Write(b)
}

// Hijack hijacks the connection.
func (w *responseModifier) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.w.(http.Hijacker); ok {
		return h.Hijack()
	}

	return nil, nil, fmt.Errorf("not a hijacker: %T", w.w)
}

// Flush sends any buffered data to the client.
func (w *responseModifier) Flush() {
	if flusher, ok := w.w.(http.Flusher); ok {
		flusher.Flush()
	}
}
