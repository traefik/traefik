package headers

import (
	"net/http"
	"strings"

	"github.com/satori/go.uuid"
)

// Handler empty struct
type Handler struct {
	requestIDLabel             string
	useExistingRequestIDHeader bool
}

// NewHeaders creates new headers handler
func NewHeaders(useExistingRid bool, ridLabel string) *Handler {
	return &Handler{useExistingRequestIDHeader: useExistingRid, requestIDLabel: ridLabel}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if h.useExistingRequestIDHeader {
		if r.Header.Get("X-Request-ID") == "" {
			r.Header.Set("X-Request-ID", makeRequestID(r, h.requestIDLabel))
		}
	} else {
		r.Header.Set("X-Request-ID", makeRequestID(r, h.requestIDLabel))
	}
	next.ServeHTTP(rw, r)
}

func makeRequestID(r *http.Request, label string) string {
	var prefix string
	if len(label) > 0 {
		prefix = label + "-"
	}
	if !strings.EqualFold(r.Header.Get("X-Request-ID-No-S"), "true") {
		if prefix == "" {
			prefix = "s"
		} else {
			prefix = "s-" + prefix
		}
	}
	return prefix + uuid.NewV4().String()
}
