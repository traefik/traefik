package snicheck

import (
	"net/http"

	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

// SNICheck is an HTTP handler that checks whether the TLS configuration for the server name is the same as for the host header.
type SNICheck struct {
	next       http.Handler
	tlsOptions string
	routerName string
}

// New creates a new SNICheck.
func New(routerName, tlsOptions string, next http.Handler) *SNICheck {
	return &SNICheck{next: next, tlsOptions: tlsOptions, routerName: routerName}
}

func (s SNICheck) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.TLS == nil {
		s.next.ServeHTTP(rw, req)
		return
	}

	tlsOptionsUsed := tcp.GetTLSOptionsNameInContext(req.Context())
	if s.tlsOptions != tlsOptionsUsed {
		log.WithoutContext().
			WithField("routerName", s.routerName).
			WithField("req.Host", req.Host).
			WithField("req.TLS.ServerName", req.TLS.ServerName).
			Debugf("TLS options difference: SNI:%s, Header:%s", tlsOptionsUsed, s.tlsOptions)
		http.Error(rw, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
		return
	}

	s.next.ServeHTTP(rw, req)
}
