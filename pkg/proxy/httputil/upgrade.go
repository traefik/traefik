package httputil

import (
	"bufio"
	"net"
	"net/http"

	"golang.org/x/net/http/httpguts"
)

// h2cUpgradeHandler removes a client-initiated h2c upgrade before the request reaches the reverse proxy.
// Since go1.24, unencrypted HTTP/2 is served through http.Server#Protocols, which supports prior knowledge only and
// not the deprecated "Upgrade: h2c" mechanism (https://go.dev/doc/go1.24#nethttppkgnethttp).
// As Traefik no longer honors an h2c upgrade, the token has no reason to reach a backend.
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
	if httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") &&
		httpguts.HeaderValuesContainsToken([]string{req.Header.Get("Upgrade")}, "h2c") {
		// Removing the Upgrade header is enough: the reverse proxy then computes an empty upgrade type, removes
		// Connection as a hop-by-hop header, and adds neither of them back.
		delete(req.Header, "Upgrade")
	}

	// HTTP2-Settings is connection-specific (RFC 7540 section 3.2.1), so it is not forwarded either.
	delete(req.Header, "Http2-Settings")

	h.next.ServeHTTP(rw, req)
}

// tlsUpgradeHandler closes the client connection of an upgraded TLS request as soon as the backend closes its side.
// Since go1.25, once the backend side of an upgraded connection reaches EOF, httputil.ReverseProxy calls CloseWrite on the
// client connection when available, and waits for the client to close its side (https://go.dev/issue/35892).
// For a TLS connection, CloseWrite only sends a close_notify alert and leaves the TCP connection open,
// and as most clients (e.g. WebSocket clients) do not close their side in response,
// the connection stays half-open until the client writes again (https://github.com/traefik/traefik/issues/13999).
// Hiding CloseWrite restores the go1.24 behavior for TLS connections only,
// so the TCP half-close behavior of non-TLS connections is kept.
type tlsUpgradeHandler struct {
	next http.Handler
}

// newTLSUpgradeHandler wraps next with the TLS upgraded connection closing behavior.
func newTLSUpgradeHandler(next http.Handler) http.Handler {
	return &tlsUpgradeHandler{next: next}
}

func (h *tlsUpgradeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.TLS == nil || !httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		h.next.ServeHTTP(rw, req)
		return
	}

	h.next.ServeHTTP(&tlsUpgradeResponseWriter{ResponseWriter: rw}, req)
}

// tlsUpgradeResponseWriter hides the CloseWrite method of the hijacked connection.
type tlsUpgradeResponseWriter struct {
	http.ResponseWriter
}

func (w *tlsUpgradeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, brw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err != nil {
		return nil, nil, err
	}

	return noCloseWriteConn{Conn: conn}, brw, nil
}

func (w *tlsUpgradeResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// noCloseWriteConn is a net.Conn which does not expose the CloseWrite method of the wrapped connection.
type noCloseWriteConn struct {
	net.Conn
}
