package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/containous/traefik/types"
)

var requestHostKey struct{}

// RequestHost is the struct for the middleware that adds the CanonicalDomain of the request Host into a context for later use.
type RequestHost struct{}

func (rh *RequestHost) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if next != nil {
		host := types.CanonicalDomain(parseHost(r.Host))
		next.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), requestHostKey, host)))
	}
}

func parseHost(addr string) string {
	if !strings.Contains(addr, ":") {
		return addr
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// GetCanonizedHost plucks the canonized host key from the request of a context that was put through the middleware
func GetCanonizedHost(ctx context.Context) string {
	if val, ok := ctx.Value(requestHostKey).(string); ok {
		return val
	}
	return ""
}
