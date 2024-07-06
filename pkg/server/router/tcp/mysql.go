package tcp

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// isMysql determines whether the buffer contains the MySQL STARTTLS message.
// Reference: https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_basic_tls.html
//            https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_handshake.html
//            https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase_packets_protocol_ssl_request.html
func isMysql(br *bufio.Reader) (bool, error) {
	// There doesn't seem to be a good way to fingerprint MySQL connections
	return true, nil
}


func (r *Router) serveMysql(conn tcp.WriteCloser) {
	// MySQL is a server-first protocol, we need to send the handshake first, this is a base64 encoded string of a valid one taken from mysql:8.3.0
	// Reference: https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_connection_phase.html
	encodedStr := "SQAAAAo4LjMuMAALAAAAQEsOLEhSWBsA////AgD/3xUAAAAAAAAAAAAARXReNToYXyxSJW5uAGNhY2hpbmdfc2hhMl9wYXNzd29yZAA="
	decodedBytes, _ := base64.StdEncoding.DecodeString(encodedStr)
	log.Info().Msg("Sending handshake.")
	_, err := conn.Write(decodedBytes)
	if err != nil {
		log.Error().Err(err).Msg("Error while sending handshake")
		conn.Close()
		return
	}

	// We expect a SSLRequest as response, which is 36 bytes long, it is like a normal handshake response but cut short before the username
	// TODO: Parse the response and check if the client requested SSL instead of manually checking the username byte
	br := bufio.NewReader(conn)
	protocol_packet_length_raw := make([]byte, 3) // MySQL packet length is a 3 byte integer
	_, err = br.Read(protocol_packet_length_raw)
	if err != nil {
		conn.Close()
		return
	}
	protocol_packet_length_raw = append(protocol_packet_length_raw, 0) // We need to add a 0 byte to parse it as a little endian UInt32
	protocol_packet_length := binary.LittleEndian.Uint32(protocol_packet_length_raw)
	log.Info().Msgf("Received protocol packet length: %d", protocol_packet_length)

	// Read the rest of the packet, this could probably be optimised by just skipping them since we don't need them
	b := make([]byte, protocol_packet_length + 1) // +1 to include the packet number
	_, err = br.Read(b)
	if err != nil {
		conn.Close()
		return
	}

	// The first packet after the handshake should be a SSLRequest or a HandshakeResponse320/41
	// If it is a SSLRequest, it should be 32 bytes long (headers + empty username)
	// TODO: Is there a better way to check ?
	if protocol_packet_length != 32 {
		log.Info().Msg("MySQL client did not request SSL, we can't properly route this connection.")
		conn.Close()
		return
	}

	// The next packet sent should be a client hello packet
	hello, err := clientHelloInfo(br)
	if err != nil {
		conn.Close()
		return
	}

	if !hello.isTLS {
		log.Error().Msg("MySQL client did not send a TLS Client Hello, closing connection.")
		conn.Close()
		return
	}
	log.Info().Msg("MySQL client sent a TLS Client Hello, success !")
	log.Info().Msgf("Server Name: %s", hello.serverName)
	log.Info().Msgf("Protocols: %v", hello.protos)
}
