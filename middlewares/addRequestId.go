package middlewares

import (
	"github.com/satori/go.uuid"
	"net/http"
)

const RequestId = "X-Request-ID"

// AddRequestId is a middleware used to add a uniq request ID header if not present
type AddRequestId struct {
	Handler http.Handler
}

func (s *AddRequestId) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	requestId := r.Header.Get(RequestId)
	if requestId == "" {
		requestId = uuid.NewV4().String()
	}
	r.Header.Set(RequestId, requestId)
	rw.Header().Set(RequestId, requestId)
	if next != nil {
		next.ServeHTTP(rw, r)
	}
}

// SetHandler sets handler
func (s *AddRequestId) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}
