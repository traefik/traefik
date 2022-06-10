package tcp

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Proxy forwards a TCP request to a TCP service.
type Proxy struct {
	address          string
	tcpAddr          *net.TCPAddr
	terminationDelay time.Duration
	proxyProtocol    *dynamic.ProxyProtocol
}

// NewProxy creates a new Proxy.
func NewProxy(address string, terminationDelay time.Duration, proxyProtocol *dynamic.ProxyProtocol) (*Proxy, error) {
	if proxyProtocol != nil && (proxyProtocol.Version < 1 || proxyProtocol.Version > 2) {
		return nil, fmt.Errorf("unknown proxyProtocol version: %d", proxyProtocol.Version)
	}

	// Creates the tcpAddr only for IP based addresses,
	// because there is no need to resolve the name on every new connection,
	// and building it should happen once.
	var tcpAddr *net.TCPAddr
	if host, _, err := net.SplitHostPort(address); err == nil && net.ParseIP(host) != nil {
		tcpAddr, err = net.ResolveTCPAddr("tcp", address)
		if err != nil {
			return nil, err
		}
	}

	return &Proxy{
		address:          address,
		tcpAddr:          tcpAddr,
		terminationDelay: terminationDelay,
		proxyProtocol:    proxyProtocol,
	}, nil
}

// ServeTCP forwards the connection to a service.
func (p *Proxy) ServeTCP(conn WriteCloser) {
	log.WithoutContext().Debugf("Handling connection from %s to %s", conn.RemoteAddr(), p.address)

	// needed because of e.g. server.trackedConnection
	defer conn.Close()

	connBackend, err := p.dialBackend()
	if err != nil {
		log.WithoutContext().Errorf("Error while connecting to backend: %v", err)
		return
	}

	// maybe not needed, but just in case
	defer connBackend.Close()
	errChan := make(chan error)

	if p.proxyProtocol != nil && p.proxyProtocol.Version > 0 && p.proxyProtocol.Version < 3 {
		header := proxyproto.HeaderProxyFromAddrs(byte(p.proxyProtocol.Version), conn.RemoteAddr(), conn.LocalAddr())
		if _, err := header.WriteTo(connBackend); err != nil {
			log.WithoutContext().Errorf("Error while writing proxy protocol headers to backend connection: %v", err)
			return
		}
	}

	go p.connCopy(conn, connBackend, errChan)
	go p.connCopy(connBackend, conn, errChan)

	err = <-errChan
	if err != nil {
		log.WithoutContext().Errorf("Error during connection: %v", err)
	}

	<-errChan
}

func (p Proxy) dialBackend() (*net.TCPConn, error) {
	// Dial using directly the TCPAddr for IP based addresses.
	if p.tcpAddr != nil {
		return net.DialTCP("tcp", nil, p.tcpAddr)
	}

	log.WithoutContext().Debugf("Dial with lookup to address %s", p.address)

	// Dial with DNS lookup for host based addresses.
	conn, err := net.Dial("tcp", p.address)
	if err != nil {
		return nil, err
	}

	return conn.(*net.TCPConn), nil
}

func (p Proxy) connCopy(dst, src WriteCloser, errCh chan error) {
	_, err := io.Copy(dst, src)
	errCh <- err

	errClose := dst.CloseWrite()
	if errClose != nil {
		log.WithoutContext().Debugf("Error while terminating connection: %v", errClose)
		return
	}

	if p.terminationDelay >= 0 {
		err := dst.SetReadDeadline(time.Now().Add(p.terminationDelay))
		if err != nil {
			log.WithoutContext().Debugf("Error while setting deadline: %v", err)
		}
	}
}
