package tcp

import (
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

// Proxy forwards a TCP request to a TCP service.
type Proxy struct {
	address string
	dialer  Dialer
}

// NewProxy creates a new Proxy.
func NewProxy(address string, dialer Dialer) (*Proxy, error) {
	return &Proxy{
		address: address,
		dialer:  dialer,
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

	connBackend, err := p.dialBackend(conn)
	if err != nil {
		log.Error().Err(err).Msg("Error while dialing backend")
		return
	}

	// maybe not needed, but just in case
	defer connBackend.Close()
	errChan := make(chan error)

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

func (p *Proxy) dialBackend(clientConn net.Conn) (WriteCloser, error) {
	// The clientConn is passed to the dialer so that it can use information from it if needed,
	// to build a PROXY protocol header.
	conn, err := p.dialer.Dial("tcp", p.address, clientConn)
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
