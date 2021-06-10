package tcpmux

import (
	"bufio"

	"github.com/traefik/traefik/v2/pkg/tcp"
	"github.com/traefik/traefik/v2/pkg/types"
)

type Matcher interface {
	Match(conn tcp.WriteCloser) bool
}

// ClientIP matches a connection based on the client IP.
type ClientIP struct {
	ip string
}

// NewClientIP returns a new clientIP with the specified IP.
func NewClientIP(ip string) *ClientIP {
	return &ClientIP{ip: ip}
}

func (c ClientIP) Match(conn tcp.WriteCloser) bool {
	// Does the Remote Address match our matcher IP.
	return c.ip == conn.RemoteAddr().String()
}

// SNIHost matches the SNI Host of the connection.
type SniHost struct {
	host string
}

// NewSNIHost returns a new SNIHost with the speficied host.
func NewSNIHost(host string) *SniHost {
	return &SniHost{host: host}
}

func (s SniHost) Match(conn tcp.WriteCloser) bool {
	// Does the SNI Host of the connection match our matcher host.
	br := bufio.NewReader(conn)
	serverName, tls, _, err := tcp.ClientHelloServerName(br)
	if err != nil {
		return false
	}
	if !tls {
		return false
	}
	serverName = types.CanonicalDomain(serverName)
	if serverName == "" {
		return false
	}

	return s.host == serverName
}
