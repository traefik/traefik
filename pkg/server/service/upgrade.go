package service

import (
	"net/http"

	"golang.org/x/net/http/httpguts"
)

// h2cUpgradeHandler removes a client-initiated h2c upgrade before the request reaches the reverse proxy.
// Upgrade is hop-by-hop, and Traefik does not implement the h2c one, so it has no reason to request it from a backend.
//
// This is a temporary workaround for httputil.ReverseProxy, which forwards the token, see https://go.dev/issue/80416.
// It has to be removed once the go directive in go.mod requires a Go release carrying that fix.
type h2cUpgradeHandler struct {
	next http.Handler
}

// newH2CUpgradeHandler wraps next with the h2c upgrade removal behavior.
func newH2CUpgradeHandler(next http.Handler) http.Handler {
	return &h2cUpgradeHandler{next: next}
}

func (h *h2cUpgradeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	isH2C := httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") &&
		httpguts.HeaderValuesContainsToken([]string{req.Header.Get("Upgrade")}, "h2c")

	_, hasSettings := req.Header["Http2-Settings"]

	if !isH2C && !hasSettings {
		h.next.ServeHTTP(rw, req)
		return
	}

	outReq := req.Clone(req.Context())

	if isH2C {
		// Removing the Upgrade header is enough: the reverse proxy then computes an empty upgrade type, removes
		// Connection as a hop-by-hop header, and adds neither of them back.
		outReq.Header.Del("Upgrade")
	}

	// HTTP2-Settings is connection-specific (RFC 7540 section 3.2.1), so it is not forwarded either.
	outReq.Header.Del("HTTP2-Settings")

	h.next.ServeHTTP(rw, outReq)
}
