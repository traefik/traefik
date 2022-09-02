package tcp

import (
	"bufio"
	"bytes"

	"github.com/traefik/traefik/v2/pkg/log"
	tcpmuxer "github.com/traefik/traefik/v2/pkg/muxer/tcp"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

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
