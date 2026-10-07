package snicheck

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// SNICheck is an HTTP handler that checks whether the TLS configuration for the server name is the same as for the host header.
type SNICheck struct {
	next           http.Handler
	routerName     string
	tlsOptionsName string
}

// New creates a new SNICheck.
func New(routerName, tlsOptionsName string, next http.Handler) *SNICheck {
	return &SNICheck{
		next:           next,
		routerName:     routerName,
		tlsOptionsName: tlsOptionsName,
	}
}

func (s SNICheck) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.TLS == nil {
		s.next.ServeHTTP(rw, req)
		return
	}

	tlsOptionsNameUsed := tcp.GetTLSOptionsName(req.Context())
	if s.tlsOptionsName != tlsOptionsNameUsed {
		log.Debug().
			Str("routerName", s.routerName).
			Str("req.Host", req.Host).
			Str("req.TLS.ServerName", req.TLS.ServerName).
			Msgf("TLS options difference: SNI:%s, Header:%s", tlsOptionsNameUsed, s.tlsOptionsName)

		// The TLS options name carried by the connection's context was resolved once,
		// at handshake time, from the dynamic configuration in effect at that moment.
		// If the configuration changes afterward (e.g. a router for this SNI is added
		// or its TLS options change), that per-connection value goes stale: it keeps
		// disagreeing with the router's freshly-resolved options on every subsequent
		// request, so every request on this connection would 421 forever. Closing the
		// connection forces the client to re-handshake on its next request, which
		// re-resolves the TLS options name against the current configuration and
		// self-heals.
		//
		// The "Connection: close" response header is enough to achieve this on
		// HTTP/1.1 (plain TCP close) and HTTP/2 (translated by the Go server into a
		// GOAWAY). It carries no meaning in HTTP/3, though: unlike HTTP/2, neither the
		// wire format nor quic-go interprets it as a signal to close the connection, so
		// the underlying QUIC connection must be closed directly instead.
		rw.Header().Set("Connection", "close")
		tcp.CloseConn(req.Context())
		http.Error(rw, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
		return
	}

	s.next.ServeHTTP(rw, req)
}
