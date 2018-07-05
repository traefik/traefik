package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/containous/traefik/types"
)

var reqHostKey struct{}

// ReqHostMiddleware is the struct for the middleware that adds the CanonicalDomain of the requst Host into a context for later use.
type ReqHostMiddleware struct{}

// NewReqHostMiddleware is a middleware that adds the CanonicalDomain of the requst Host into a context for later use.
func NewReqHostMiddleware() *ReqHostMiddleware {
	return &ReqHostMiddleware{}
}

func (rhm *ReqHostMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	reqHost := r.Host
	if strings.IndexByte(reqHost, ':') >= 0 {
		var err error
		reqHost, _, err = net.SplitHostPort(reqHost)
		if err != nil {
			reqHost = r.Host
		}
	}
	reqHost = types.CanonicalDomain(reqHost)
	if next != nil {
		next.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), reqHostKey, reqHost)))
	}
}

// GetCannonHost plucks the cannonized host key from the requst of a context that was put through the middleware
func GetCannonHost(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(reqHostKey).(string)
	return val, ok
}
