package tcp

import (
	"io"
	"net"

	"github.com/containous/traefik/pkg/log"
)

// Proxy forwards a TCP request to a TCP service
type Proxy struct {
	target *net.TCPAddr
}

// NewProxy creates a new Proxy
func NewProxy(address string) (*Proxy, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		target: tcpAddr,
	}, nil
}

// ServeTCP forwards the connection to a service
func (p *Proxy) ServeTCP(conn net.Conn) {
	log.Debugf("Handling connection from %s", conn.RemoteAddr())
	defer conn.Close()
	connBackend, err := net.DialTCP("tcp", nil, p.target)
	if err != nil {
		log.Errorf("Error while connection to backend: %v", err)
		return
	}
	defer connBackend.Close()

	errChan := make(chan error, 1)
	go connCopy(conn, connBackend, errChan)
	go connCopy(connBackend, conn, errChan)

	err = <-errChan
	if err != nil {
		log.Errorf("Error during connection: %v", err)
	}
}

func connCopy(dst, src net.Conn, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err
}
