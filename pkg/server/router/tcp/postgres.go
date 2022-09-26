package tcp

import (
	"bufio"
	"bytes"
	"errors"

	"github.com/traefik/traefik/v2/pkg/log"
	tcpmuxer "github.com/traefik/traefik/v2/pkg/muxer/tcp"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

var (
	PostgresStartTLSReply = []byte{83}                         // S
	PostgresStartTLSMsg   = []byte{0, 0, 0, 8, 4, 210, 22, 47} // int32(8) + int32(80877103)
)

// isPostGRESql determines whether the buffer contains the Postgres STARTTLS message.
func isPostGRESql(br *bufio.Reader) (bool, error) {
	for i := 0; i < len(PostgresStartTLSMsg); i++ {
		peeked, err := br.Peek(i)
		if err != nil {
			log.WithoutContext().Errorf("Error while Peeking first bytes: %s", err)
			return false, err
		}

		if !bytes.Equal(peeked, PostgresStartTLSMsg[:i]) {
			return false, nil
		}
	}
	return true, nil
}

// servePostGreSQL serves a connection with a Postgres client negotiating a STARTTLS session.
// It handles TCP TLS routing, after accepting to start the STARTTLS session.
func (r *Router) servePostGreSQL(conn tcp.WriteCloser) {
	_, err := conn.Write(PostgresStartTLSReply)
	if err != nil {
		conn.Close()
		return
	}

	br := bufio.NewReader(conn)

	b := make([]byte, len(PostgresStartTLSMsg))
	_, err = br.Read(b)
	if err != nil {
		conn.Close()
		return
	}

	hello, err := clientHelloInfo(br)
	if err != nil {
		conn.Close()
		return
	}

	if !hello.isTLS {
		conn.Close()
		return
	}

	connData, err := tcpmuxer.NewConnData(hello.serverName, conn, hello.protos)
	if err != nil {
		log.WithoutContext().Errorf("Error while reading TCP connection data: %v", err)
		conn.Close()
		return
	}

	// Contains also TCP TLS passthrough routes.
	handlerTCPTLS, _ := r.muxerTCPTLS.Match(connData)
	if handlerTCPTLS == nil {
		conn.Close()
		return
	}

	// We are in TLS mode and if the handler is not TLSHandler, we are in passthrough
	proxiedConn := r.GetConn(conn, hello.peeked)
	if _, ok := handlerTCPTLS.(*tcp.TLSHandler); !ok {
		proxiedConn = &postgresConn{WriteCloser: proxiedConn}
	}

	handlerTCPTLS.ServeTCP(proxiedConn)
}

// postgresConn is a tcp.WriteCloser that will negotiate a TLS session (STARTTLS),
// before exchanging any data.
// It enforces that the STARTTLS negotiation with the peer is successful.
type postgresConn struct {
	tcp.WriteCloser

	alreadyWritten bool
	alreadyRead    bool
	waiter         chan error
}

// Read reads bytes from the underlying connection (tcp.WriteCloser).
// On first call, it actually only reads the content the PostgresStartTLSMsg message.
// This is done to behave as a Postgres TLS client that ask to initiate a TLS session.
// Not thread safe.
func (c *postgresConn) Read(p []byte) (n int, err error) {
	if c.alreadyRead {
		if err := <-c.waiter; err != nil {
			return 0, err
		}

		return c.WriteCloser.Read(p)
	}

	if c.waiter == nil {
		c.waiter = make(chan error)
	}

	defer func() {
		c.alreadyRead = true
	}()

	copy(p, PostgresStartTLSMsg)
	return len(PostgresStartTLSMsg), nil
}

// Write writes bytes to the underlying connection (tcp.WriteCloser).
// On first call, it checks that the bytes to write are matching the PostgresStartTLSReply.
// If the check is successful, it does nothing (no actual write on the connection),
// otherwise an error is raised and transmitted to a second Read call through c.waiter.
// It is done to enforce that the STARTTLS negotiation is successful.
// Not thread safe.
func (c *postgresConn) Write(p []byte) (n int, err error) {
	if c.alreadyWritten {
		return c.WriteCloser.Write(p)
	}

	if c.waiter == nil {
		return 0, errors.New("initial read never happened")
	}

	defer func() {
		c.alreadyWritten = true
	}()

	// TODO(romain): the two assertions are split to be more accurate when returning the number of written bytes, but is it worth it?
	if len(p) != 1 {
		c.waiter <- errors.New("invalid response from PostGreSQL server")
		return len(p), nil
	}

	if p[0] != PostgresStartTLSReply[0] {
		c.waiter <- errors.New("invalid response from PostGreSQL server")
		return 1, nil
	}

	close(c.waiter)

	return 1, nil
}
