package tcp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rs/zerolog/log"
	tcpmuxer "github.com/traefik/traefik/v3/pkg/muxer/tcp"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

var (
	PostgresStartTLSMsg   = []byte{0, 0, 0, 8, 4, 210, 22, 47} // int32(8) + int32(80877103)
	PostgresStartTLSReply = []byte{83}                         // S
)

// isPostgres determines whether the buffer contains the Postgres STARTTLS message.
func isPostgres(br *bufio.Reader) (bool, error) {
	// Peek the first 8 bytes individually to prevent blocking on peek
	// if the underlying conn does not send enough bytes.
	// It could happen if a protocol start by sending less than 8 bytes,
	// and expect a response before proceeding.
	for i := 1; i < len(PostgresStartTLSMsg)+1; i++ {
		peeked, err := br.Peek(i)
		if err != nil {
			var opErr *net.OpError
			if !errors.Is(err, io.EOF) && (!errors.As(err, &opErr) || !opErr.Timeout()) {
				log.Debug().Err(err).Msg("Error while peeking first bytes")
			}
			return false, err
		}

		if !bytes.Equal(peeked, PostgresStartTLSMsg[:i]) {
			return false, nil
		}
	}
	return true, nil
}

// servePostgres serves a connection with a Postgres client negotiating a STARTTLS session.
// It handles TCP TLS routing, after accepting to start the STARTTLS session.
func (r *Router) servePostgres(conn tcp.WriteCloser) {
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
		log.Error().Err(err).Msg("Error while reading TCP connection data")
		conn.Close()
		return
	}

	// Contains also TCP TLS passthrough routes.
	handlerTCPTLS, _ := r.muxerTCPTLS.Match(connData)
	if handlerTCPTLS == nil {
		conn.Close()
		return
	}

	// We are in TLS mode and if the handler is not TLSHandler, we are in passthrough.
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

	starttlsMsgSent       bool // whether we have already sent the STARTTLS handshake to the backend.
	starttlsReplyReceived bool // whether we have already received the STARTTLS handshake reply from the backend.

	// errChan makes sure that an error is returned if the first operation to ever
	// happen on a postgresConn is a Write (because it should instead be a Read).
	errChanMu sync.Mutex
	errChan   chan error
}

// Read reads bytes from the underlying connection (tcp.WriteCloser).
// On first call, it actually only injects the PostgresStartTLSMsg,
// in order to behave as a Postgres TLS client that initiates a STARTTLS handshake.
// Read does not support concurrent calls.
func (c *postgresConn) Read(p []byte) (n int, err error) {
	if c.starttlsMsgSent {
		if err := <-c.errChan; err != nil {
			return 0, err
		}

		return c.WriteCloser.Read(p)
	}

	defer func() {
		c.starttlsMsgSent = true
		c.errChanMu.Lock()
		c.errChan = make(chan error)
		c.errChanMu.Unlock()
	}()

	copy(p, PostgresStartTLSMsg)
	return len(PostgresStartTLSMsg), nil
}

// Write writes bytes to the underlying connection (tcp.WriteCloser).
// On first call, it checks that the bytes to write (the ones provided by the backend)
// match the PostgresStartTLSReply, and if yes it drops them (as the STARTTLS
// handshake between the client and traefik has already taken place). Otherwise, an
// error is transmitted through c.errChan, so that the second Read call gets it and
// returns it up the stack.
// Write does not support concurrent calls.
func (c *postgresConn) Write(p []byte) (n int, err error) {
	if c.starttlsReplyReceived {
		return c.WriteCloser.Write(p)
	}

	c.errChanMu.Lock()
	if c.errChan == nil {
		c.errChanMu.Unlock()
		return 0, errors.New("initial read never happened")
	}
	c.errChanMu.Unlock()

	defer func() {
		c.starttlsReplyReceived = true
	}()

	if len(p) != 1 || p[0] != PostgresStartTLSReply[0] {
		c.errChan <- errors.New("invalid response from Postgres server")
		return len(p), nil
	}

	close(c.errChan)

	return 1, nil
}
