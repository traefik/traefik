package tcp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/traefik/traefik/v2/pkg/log"
)

var (
	// PostgresStartTLSReply is the reply that is sent by the server to complete the StartTLS handshake.
	// This will force a secure connection.
	PostgresStartTLSReply = []byte{83} // S
	// PostgresStartTLSMsg is the message sent by the client to initiate a TLS handshake.
	PostgresStartTLSMsg = []byte{0, 0, 0, 8, 4, 210, 22, 47} // int32(8) + int32(80877103)

	// StartTLSServerFuncs maps a StartTLS flavor name to it's StartTLS handshake handler.
	StartTLSServerFuncs = map[string]func(WriteCloser) (WriteCloser, error){
		"postgres": HandlePostgresStartTLSHandshakeAsServer,
	}

	// StartTLSClientFuncs maps a StartTLS flavor name to it's StartTLS handshake handler.
	StartTLSClientFuncs = map[string]func(io.ReadWriter) error{
		"postgres": HandlePostgresStartTLSHandshakeAsClient,
	}
)

// HandlePostgresStartTLSHandshakeAsClient performs the postgres StartTLS
// handshake (acting as client). It sends the postgresStartTLSMsg and checks if
// the server response matches the expected value. It returns an error on
// read/write failures on the connection or if the server response doesn't
// match.
func HandlePostgresStartTLSHandshakeAsClient(conn io.ReadWriter) error {
	_, err := conn.Write(PostgresStartTLSMsg)
	if err != nil {
		return err
	}

	b := make([]byte, 1)
	_, err = io.ReadFull(conn, b)
	if err != nil {
		return err
	}

	if b[0] != PostgresStartTLSReply[0] {
		// we only support secure connections otherwise we wouldn't get any SNI header
		// and cannot forward.
		return fmt.Errorf("postgres does not accept tls. response got %v want %v", b, PostgresStartTLSReply[0])
	}
	return nil
}

// HandlePostgresStartTLSHandshakeAsServer performs a StartTLS Handshake (acting
// as server). It peeks into some bytes of conn and tries to find out if the
// client performs a StartTLS handshake. If the client request does contain a
// StartTLS handshake signature the handshake is performed. After this step the
// client will start a TLS session. If no StartTLS signature is found the bytes
// of the connection remain unmodified. In any case the caller must use the
// returned WriteCloser for further read/write operations. An error is returned
// if reading or writing to the WriteCloser fails.
func HandlePostgresStartTLSHandshakeAsServer(conn WriteCloser) (WriteCloser, error) {
	startTLSConn := newStartTLSConn(conn)

	buf, err := startTLSConn.Peek(len(PostgresStartTLSMsg))
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.WithoutContext().Errorf("Error on starttls handshake: %v", err)
		}
		return startTLSConn, err
	}

	if !bytes.Equal(buf, PostgresStartTLSMsg) {
		return startTLSConn, nil
	}

	// consume the bytes that we just peeked so far..
	_, err = io.ReadFull(startTLSConn.Br, buf)
	if err != nil {
		return startTLSConn, err
	}

	_, err = conn.Write(PostgresStartTLSReply)
	if err != nil {
		return startTLSConn, err
	}

	return startTLSConn, nil
}

type startTLSConn struct {
	Br *bufio.Reader
	WriteCloser
}

func newStartTLSConn(conn WriteCloser) startTLSConn {
	return startTLSConn{bufio.NewReader(conn), conn}
}

func (s startTLSConn) Peek(n int) ([]byte, error) {
	return s.Br.Peek(n)
}

func (s startTLSConn) Read(p []byte) (int, error) {
	return s.Br.Read(p)
}
