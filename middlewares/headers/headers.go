package headers

import (
	"net/http"
	"strings"

	"github.com/satori/go.uuid"
)

// Handler empty struct
type Handler struct{}

// NewHeaders creates new headers handler
func NewHeaders() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Header.Get("X-Request-ID") == "" {
		if strings.EqualFold(r.Header.Get("X-Request-ID-No-S"), "true") {
			r.Header.Set("X-Request-ID", uuid.NewV4().String())
		} else {
			r.Header.Set("X-Request-ID", "s"+uuid.NewV4().String())
		}
	}
	next.ServeHTTP(rw, r)
}
