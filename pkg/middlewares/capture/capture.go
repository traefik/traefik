// Package capture is a middleware that captures requests/responses size, status and headers.
//
// For another middleware to get those attributes of a requests, this middleware
// should be added before in the middleware chain.
//
//     	handler, _ := NewHandler()
//     	chain := alice.New().
//     	     Append(WrapHandler(handler)).
//     	     Append(myOtherMiddleware).
//     	     then(...)
//
// As this middleware stores those data in the request's context, the data can
// be retrieved at anytime after the ServerHTTP.
//
//     func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.Handler) {
//     ...
//     	crw := capture.GetResponseWriter(req.Context())
//     	fmt.Println(crw.Size)
//     }
package capture

import (
	"context"
	"net/http"

	"github.com/containous/alice"
)

type key string

const capturedData key = "capturedData"

// Handler will store each request data to its context.
type Handler struct{}

// WrapHandler Wraps capture handler into an Alice Constructor.
func WrapHandler(handler *Handler) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(rw, req, next)
		}), nil
	}
}

// NewHandler creates a new Handler.
func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	c := Capture{}
	ctx := req.Context()
	if req.Body != nil {
		requestReader := newRequestReader(req.Body)
		c.rr = requestReader
		req.Body = requestReader
	}
	responseWriter := newResponseWriter(rw)
	c.rw = responseWriter
	ctx = context.WithValue(ctx, capturedData, c)
	next.ServeHTTP(responseWriter, req.WithContext(ctx))
}

type Capture struct {
	rr *requestReader
	rw responseWriter
}

func GetResponseWriter(ctx context.Context) Capture {
	c, ok := ctx.Value(capturedData).(Capture)
	if !ok {
		// This should never happen as the capture middleware should be used
		// before any other middleware that want to extract data from the
		// context.
		return Capture{}
	}

	return c
}

func (c Capture) GetRequestReader() *requestReader {
	return c.rr
}

func (c Capture) GetResponseWriter() responseWriter {
	return c.rw
}
