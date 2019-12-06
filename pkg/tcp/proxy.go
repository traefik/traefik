package tcp

import (
	"bytes"
	"io"
	"net"
	"strings"
	"time"

	"github.com/containous/traefik/v2/pkg/log"
)

// Proxy forwards a TCP request to a TCP service
type Proxy struct {
	target           *net.TCPAddr
	terminationDelay time.Duration
	proxyProtocol    bool
}

// NewProxy creates a new Proxy
func NewProxy(address string, terminationDelay time.Duration, proxyProtocol bool) (*Proxy, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Proxy{target: tcpAddr, terminationDelay: terminationDelay, proxyProtocol: proxyProtocol}, nil
}

// ServeTCP forwards the connection to a service
func (p *Proxy) ServeTCP(conn WriteCloser) {
	log.Debugf("Handling connection from %s", conn.RemoteAddr())

	// needed because of e.g. server.trackedConnection
	defer conn.Close()

	connBackend, err := net.DialTCP("tcp", nil, p.target)
	if err != nil {
		log.Errorf("Error while connection to backend: %v", err)
		return
	}

	// maybe not needed, but just in case
	defer connBackend.Close()

	if p.proxyProtocol {
		writeProxyHeaderV1(connBackend, conn.RemoteAddr().String(), p.target.String())
	}

	errChan := make(chan error)
	go p.connCopy(conn, connBackend, errChan)
	go p.connCopy(connBackend, conn, errChan)

	err = <-errChan
	if err != nil {
		log.WithoutContext().Errorf("Error during connection: %v", err)
	}

	<-errChan
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

func writeProxyHeaderV1(tcpConn *net.TCPConn, srcAddr, dstAddr string) {
	headerBuf := new(bytes.Buffer)
	headerBuf.WriteString("PROXY")
	headerBuf.WriteString(" ")

	if strings.HasPrefix(srcAddr, "[") {
		headerBuf.WriteString("TCP6")
	} else {
		headerBuf.WriteString("TCP4")
	}
	headerBuf.WriteString(" ")

	srcIP, srcPort, _ := net.SplitHostPort(srcAddr)
	dstIP, dstPort, _ := net.SplitHostPort(dstAddr)

	headerBuf.WriteString(srcIP)
	headerBuf.WriteString(" ")
	headerBuf.WriteString(dstIP)
	headerBuf.WriteString(" ")

	headerBuf.WriteString(srcPort)
	headerBuf.WriteString(" ")
	headerBuf.WriteString(dstPort)
	headerBuf.WriteString("\r\n")

	tcpConn.Write(headerBuf.Bytes())
}
