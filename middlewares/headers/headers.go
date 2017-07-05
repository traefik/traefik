package headers

import (
	"net/http"

	"github.com/satori/go.uuid"
)

// Handler empty struct
type Handler struct{}

// NewHeaders creates new headers handler
func NewHeaders() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	u1 := "s" + uuid.NewV4().String()
	r.Header.Set("X-Request-ID", u1)
	next.ServeHTTP(rw, r)
}
