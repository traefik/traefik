package middlewares

import (
	"github.com/satori/go.uuid"
	"net/http"
)

// RequestID is the header to be used to add unique request ID
const RequestID = "X-Request-ID"

// AddRequestID is a middleware used to add a uniq request ID header if not present
type AddRequestID struct {
	Handler http.Handler
}

func (s *AddRequestID) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	requestID := r.Header.Get(RequestID)
	if requestID == "" {
		requestID = uuid.NewV4().String()
	}
	r.Header.Set(RequestID, requestID)
	rw.Header().Set(RequestID, requestID)
	if next != nil {
		next.ServeHTTP(rw, r)
	}
}

// SetHandler sets handler
func (s *AddRequestID) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}
