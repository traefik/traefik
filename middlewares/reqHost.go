package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/containous/traefik/types"
)

var reqHostKey struct{}

// ReqHost is the struct for the middleware that adds the CanonicalDomain of the request Host into a context for later use.
type ReqHost struct{}

func (rh *ReqHost) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	reqHost := types.CanonicalDomain(parseHost(r.Host))
	if next != nil {
		next.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), reqHostKey, reqHost)))
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

// GetCanonHost plucks the canonized host key from the request of a context that was put through the middleware
func GetCanonHost(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(reqHostKey).(string)
	return val, ok
}
