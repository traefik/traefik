package udp

import (
	"io"
	"net"

	"github.com/traefik/traefik/v2/pkg/log"
)

// Proxy is a reverse-proxy implementation of the Handler interface.
type Proxy struct {
	// TODO: maybe optimize by pre-resolving it at proxy creation time
	target string
}

// NewProxy creates a new Proxy.
func NewProxy(address string) (*Proxy, error) {
	return &Proxy{target: address}, nil
}

// ServeUDP implements the Handler interface.
func (p *Proxy) ServeUDP(conn *Conn) {
	log.Debugf("Handling connection from %s", conn.rAddr)

	// needed because of e.g. server.trackedConnection
	defer conn.Close()

	connBackend, err := net.Dial("udp", p.target)
	if err != nil {
		log.Errorf("Error while connecting to backend: %v", err)
		return
	}

	// maybe not needed, but just in case
	defer connBackend.Close()

	errChan := make(chan error)
	go p.connCopy(conn, connBackend, errChan)
	go p.connCopy(connBackend, conn, errChan)

	err = <-errChan
	if err != nil {
		log.WithoutContext().Errorf("Error while serving UDP: %v", err)
	}

	<-errChan
}

func (p Proxy) connCopy(dst io.WriteCloser, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err

	if err := dst.Close(); err != nil {
		log.WithoutContext().Debugf("Error while terminating connection: %v", err)
	}
}
