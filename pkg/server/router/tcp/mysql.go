package tcp

import (
	"bufio"
	"bytes"
	"github.com/rs/zerolog/log"
	tcpmuxer "github.com/traefik/traefik/v3/pkg/muxer/tcp"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"io"
)

// MySQLClientSSL is the capability flag for SSL support in MySQL protocol.
const MySQLClientSSL = uint32(0x0800)

// mysqlHandshakePacketSSL is a pre-built MySQL handshake packet that forces SSL negotiation.
// This packet advertises SSL capabilities to ensure the client responds with an SSL request.
var mysqlHandshakePacketSSL = []byte{
	0x4a, 0x00, 0x00, 0x00, // Length=74, sequence=0
	0x0a,                               // Protocol version 10
	'8', '.', '0', '.', '4', '3', 0x00, // Server version "8.0.43"
	0x10, 0x00, 0x00, 0x00, // Thread ID
	'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 0x00, // Salt1 (8 bytes + null terminator)
	0xff, 0xff, // Capabilities lower (includes SSL support)
	0xff,       // Character set (utf8)
	0x02, 0x00, // Server status
	0xff, 0xdf, // Capabilities upper (includes SSL support)
	0x15,                                                       // Authentication plugin data length
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Reserved (10 bytes)
	'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 0x00, // Salt2 (12 bytes + null terminator)
	'c', 'a', 'c', 'h', 'i', 'n', 'g', '_',
	's', 'h', 'a', '2', '_', 'p', 'a', 's', 's',
	'w', 'o', 'r', 'd', 0x00, // Authentication plugin name "caching_sha2_password"
}

// serveMySQL handles MySQL protocol connections with TLS passthrough.
// It performs the following steps:
// 1. Sends a fake handshake to force SSL negotiation
// 2. Reads the client's SSL request response
// 3. Extracts SNI from the client's TLS ClientHello
// 4. Routes the connection based on SNI
// 5. Creates a mysqlConn wrapper to handle handshake replay to the backend
func (r *Router) serveMySQL(conn tcp.WriteCloser) {
	br := bufio.NewReader(conn)

	// Step 1: Send fake handshake to force SSL negotiation
	if err := r.sendMySQLHandshake(conn); err != nil {
		log.Error().Err(err).Msg("Failed to send MySQL handshake packet")
		_ = conn.Close()
		return
	}

	// Step 2: Read client response (SSL request)
	clientResp, err := readPacket(br)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read MySQL client response")
		_ = conn.Close()
		return
	}

	// Step 3: Validate SSL request
	if !hasSSLRequest(clientResp) {
		log.Error().Msg("MySQL protocol requires SSL for SNI extraction - closing connection")
		_ = conn.Close()
		return
	}

	// Step 4: Extract SNI from TLS ClientHello
	hello, err := clientHelloInfo(br)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read TLS ClientHello")
		_ = conn.Close()
		return
	}

	if !hello.isTLS {
		log.Error().Msg("Expected TLS connection but got non-TLS")
		_ = conn.Close()
		return
	}

	if hello.serverName == "" {
		log.Error().Msg("MySQL protocol requires SNI for routing - closing connection")
		_ = conn.Close()
		return
	}

	// Step 5: Route connection based on SNI
	handler, err := r.routeMySQLConnection(hello, conn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to route MySQL connection")
		_ = conn.Close()
		return
	}

	if handler == nil {
		log.Debug().Str("sni", hello.serverName).Msg("No route found for MySQL connection")
		_ = conn.Close()
		return
	}

	// Step 6: Create MySQL connection wrapper and proxy
	mysqlConn := &mysqlConn{
		WriteCloser:      r.GetConn(conn, hello.peeked),
		clientSSLRequest: clientResp,
	}
	handler.ServeTCP(mysqlConn)
}

// sendMySQLHandshake sends a fake MySQL handshake packet to force SSL negotiation.
func (r *Router) sendMySQLHandshake(conn tcp.WriteCloser) error {
	_, err := conn.Write(mysqlHandshakePacketSSL)
	return err
}

// routeMySQLConnection routes the MySQL connection based on SNI information.
func (r *Router) routeMySQLConnection(hello *clientHello, conn tcp.WriteCloser) (tcp.Handler, error) {
	connData, err := tcpmuxer.NewConnData(hello.serverName, conn, hello.protos)
	if err != nil {
		return nil, err
	}

	handler, _ := r.muxerTCPTLS.Match(connData)
	return handler, nil
}

// readPacket reads a complete MySQL packet from the buffered reader.
// MySQL's packets have a 4-byte header followed by the payload data.
// The first 3 bytes contain the payload length, the 4th byte is the sequence number.
func readPacket(br *bufio.Reader) ([]byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(br, header); err != nil {
		return nil, err
	}

	// Extract payload length from first 3 bytes (little-endian)
	length := int(header[0]) | int(header[1])<<8 | int(header[2])<<16

	// Read the payload data
	data := make([]byte, length)
	if _, err := io.ReadFull(br, data); err != nil {
		return nil, err
	}

	// Return complete packet (header + payload)
	return append(header, data...), nil
}

// hasSSLRequest checks if the MySQL packet contains SSL capability flags.
// It examines the capability flags in the client's SSL request packet.
func hasSSLRequest(packet []byte) bool {
	if len(packet) < 8 {
		return false
	}

	// Skip the 4-byte header to get to the payload
	data := packet[4:]
	if len(data) < 4 {
		return false
	}

	// Extract capability flags (little-endian)
	caps := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24

	// Check if SSL capability flag is set
	return (caps & MySQLClientSSL) != 0
}

// mysqlConn is a wrapper for TCP connections that handles MySQL protocol handshake replay.
// It intercepts the first read/write operations to replay the SSL request to the backend,
// then operates as a transparent proxy for subsequent operations.
type mysqlConn struct {
	tcp.WriteCloser

	handshakeCompleted bool
	proxyMode          bool

	clientSSLRequest []byte
}

// Read implements the io.Reader interface for mysqlConn.
// On the first call, it returns the cached client SSL request.
// On subsequent calls, it reads directly from the underlying connection.
func (c *mysqlConn) Read(p []byte) (int, error) {
	if c.handshakeCompleted {
		return c.WriteCloser.Read(p)
	}

	c.handshakeCompleted = true

	// Return the cached SSL request on first read
	copy(p, c.clientSSLRequest)
	return len(c.clientSSLRequest), nil
}

// Write implements the io.Writer interface for mysqlConn.
// On the first call, it processes the backend's handshake response.
// On subsequent calls, it writes directly to the underlying connection.
func (c *mysqlConn) Write(p []byte) (int, error) {
	if c.proxyMode {
		return c.WriteCloser.Write(p)
	}

	c.proxyMode = true

	// Process the backend's handshake response
	br := bufio.NewReader(bytes.NewReader(p))
	resp, err := readPacket(br)
	if err != nil {
		return 0, err
	}

	// Copy the processed response back to the buffer
	copy(p, resp)
	return len(resp), nil
}
