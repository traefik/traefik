package udp

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/mastercactapus/proxyprotocol"
	"github.com/rs/zerolog/log"
)

var (
	// errNoProxyProtocol is returned when the connection does not contain the PROXY protocol header.
	errNoProxyProtocol = errors.New("not a PROXY protocol connection")
)


// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	if c.rAddr != nil {
		return c.rAddr
	}
	return c.Conn.RemoteAddr()
}

// HandleProxyProtocol reads the PROXY protocol header from the connection.
// If a valid header is found, it updates the connection's remote address.
func HandleProxyProtocol(conn net.Conn, timeout time.Duration) (*Conn, error) {
	if timeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	header, err := proxyprotocol.Read(conn)
	if timeout > 0 {
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			log.Warn().Err(err).Msg("Failed to reset read deadline")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error reading PROXY protocol header: %w", err)
	}

	if header.Addr == nil {
		return nil, errNoProxyProtocol
	}

	return &Conn{Conn: conn, rAddr: header.Addr}, nil
}