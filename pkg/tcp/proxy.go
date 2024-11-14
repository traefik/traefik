package tcp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// Proxy forwards a TCP request to a TCP service.
type Proxy struct {
	address       string
	proxyProtocol *dynamic.ProxyProtocol
	dialer        Dialer
}

// NewProxy creates a new Proxy.
func NewProxy(address string, proxyProtocol *dynamic.ProxyProtocol, dialer Dialer) (*Proxy, error) {
	if proxyProtocol != nil && (proxyProtocol.Version < 1 || proxyProtocol.Version > 2) {
		return nil, fmt.Errorf("unknown proxyProtocol version: %d", proxyProtocol.Version)
	}

	return &Proxy{
		address:       address,
		proxyProtocol: proxyProtocol,
		dialer:        dialer,
	}, nil
}

// ServeTCP forwards the connection to a service.
func (p *Proxy) ServeTCP(conn WriteCloser) {
	log.Debug().
		Str("address", p.address).
		Str("remoteAddr", conn.RemoteAddr().String()).
		Msg("Handling TCP connection")

	// needed because of e.g. server.trackedConnection
	defer conn.Close()

	connBackend, err := p.dialBackend()
	if err != nil {
		log.Error().Err(err).Msg("Error while dialing backend")
		return
	}

	// maybe not needed, but just in case
	defer connBackend.Close()
	errChan := make(chan error)

	if p.proxyProtocol != nil && p.proxyProtocol.Version > 0 && p.proxyProtocol.Version < 3 {
		header := proxyproto.HeaderProxyFromAddrs(byte(p.proxyProtocol.Version), conn.RemoteAddr(), conn.LocalAddr())
		if _, err := header.WriteTo(connBackend); err != nil {
			log.Error().Err(err).Msg("Error while writing TCP proxy protocol headers to backend connection")
			return
		}
	}

	go p.connCopy(conn, connBackend, errChan)
	go p.connCopy(connBackend, conn, errChan)

	err = <-errChan
	if err != nil {
		// Treat connection reset error during a read operation with a lower log level.
		// This allows to not report an RST packet sent by the peer as an error,
		// as it is an abrupt but possible end for the TCP session
		if isReadConnResetError(err) {
			log.Debug().Err(err).Msg("Error while handling TCP connection")
		} else {
			log.Error().Err(err).Msg("Error while handling TCP connection")
		}
	}

	<-errChan
}

func (p *Proxy) dialBackend() (WriteCloser, error) {
	conn, err := p.dialer.Dial("tcp", p.address)
	if err != nil {
		return nil, err
	}

	return conn.(WriteCloser), nil
}

func (p *Proxy) connCopy(dst, src WriteCloser, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err

	// Ends the connection with the dst connection peer.
	// It corresponds to sending a FIN packet to gracefully end the TCP session.
	errClose := dst.CloseWrite()
	if errClose != nil {
		// Calling CloseWrite() on a connection which have a socket which is "not connected" is expected to fail.
		// It happens notably when the dst connection has ended receiving an RST packet from the peer (within the other connCopy call).
		// In that case, logging the error is superfluous,
		// as in the first place we should not have needed to call CloseWrite.
		if !isSocketNotConnectedError(errClose) {
			log.Debug().Err(errClose).Msg("Error while terminating TCP connection")
		}

		return
	}

	if p.dialer.TerminationDelay() >= 0 {
		err := dst.SetReadDeadline(time.Now().Add(p.dialer.TerminationDelay()))
		if err != nil {
			log.Debug().Err(err).Msg("Error while setting TCP connection deadline")
		}
	}
}

// isSocketNotConnectedError reports whether err is a socket not connected error.
func isSocketNotConnectedError(err error) bool {
	var oerr *net.OpError
	return errors.As(err, &oerr) && errors.Is(err, syscall.ENOTCONN)
}
