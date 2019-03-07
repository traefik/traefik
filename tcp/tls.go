package tcp

import (
	"crypto/tls"
	"net"
)

// TLSHandler handles TLS connections
type TLSHandler struct {
	Next   Handler
	Config *tls.Config
}

// ServeTCP terminates the TLS connection
func (t *TLSHandler) ServeTCP(conn net.Conn) {
	t.Next.ServeTCP(tls.Server(conn, t.Config))
}
