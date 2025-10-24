package udp

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildProxyProtocolV2Header creates a Proxy Protocol v2 UDP header with given addresses.
func buildProxyProtocolV2Header(t *testing.T, srcIP string, srcPort int, dstIP string, dstPort int, isV6 bool) []byte {
	t.Helper()

	var header *proxyproto.Header
	if isV6 {
		header = &proxyproto.Header{
			Version:           2,
			Command:           proxyproto.PROXY,
			TransportProtocol: proxyproto.UDPv6,
			SourceAddr:        &net.UDPAddr{IP: net.ParseIP(srcIP), Port: srcPort},
			DestinationAddr:   &net.UDPAddr{IP: net.ParseIP(dstIP), Port: dstPort},
		}
	} else {
		header = &proxyproto.Header{
			Version:           2,
			Command:           proxyproto.PROXY,
			TransportProtocol: proxyproto.UDPv4,
			SourceAddr:        &net.UDPAddr{IP: net.ParseIP(srcIP), Port: srcPort},
			DestinationAddr:   &net.UDPAddr{IP: net.ParseIP(dstIP), Port: dstPort},
		}
	}

	var buf bytes.Buffer
	_, err := header.WriteTo(&buf)
	require.NoError(t, err)
	return buf.Bytes()
}

// setupTestListener creates a test UDP listener with given Proxy Protocol config.
func setupTestListener(t *testing.T, ppConfig *ProxyProtocolConfig) (*Listener, *net.UDPConn) {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	listener, err := ListenPacketConn(conn, 3*time.Second, ppConfig)
	require.NoError(t, err)

	clientConn, err := net.DialUDP("udp", nil, listener.Addr().(*net.UDPAddr))
	require.NoError(t, err)

	return listener, clientConn
}

// TestBuildProxyProtocolV2Header_IPv4 verifies the helper creates valid v2 headers.
func TestBuildProxyProtocolV2Header_IPv4(t *testing.T) {
	header := buildProxyProtocolV2Header(t, "192.0.2.1", 12345, "198.51.100.1", 5000, false)
	require.NotNil(t, header)
	// Proxy Protocol v2 header for UDP/IPv4: 16 byte preamble + 8 bytes addresses + 4 bytes ports = 28 bytes.
	assert.Equal(t, 28, len(header))
}

// TestBuildProxyProtocolV2Header_IPv6 verifies the helper creates valid IPv6 headers.
func TestBuildProxyProtocolV2Header_IPv6(t *testing.T) {
	header := buildProxyProtocolV2Header(t, "2001:db8::1", 54321, "2001:db8::2", 5000, true)
	require.NotNil(t, header)
	// Proxy Protocol v2 header for UDP/IPv6: 16 byte preamble + 32 bytes addresses + 4 bytes ports = 52 bytes.
	assert.Equal(t, 52, len(header))
}

// TestProxyProtocol_IPv4_Insecure tests IPv4 Proxy Protocol parsing in insecure mode.
func TestProxyProtocol_IPv4_Insecure(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure: true,
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	header := buildProxyProtocolV2Header(t, "192.0.2.1", 12345, "198.51.100.1", 5000, false)
	payload := []byte("test payload")
	packet := append(header, payload...)

	_, err := clientConn.Write(packet)
	require.NoError(t, err)

	conn, err := listener.Accept()
	require.NoError(t, err)

	// Verify IPv4source address from Proxy Protocol header, not actual client.
	assert.Equal(t, "192.0.2.1:12345", conn.RemoteAddr().String())

	// Verify payload received without header.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])
}

// TestProxyProtocol_IPv6_Insecure tests IPv6 Proxy Protocol parsing in insecure mode.
func TestProxyProtocol_IPv6_Insecure(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure: true,
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	header := buildProxyProtocolV2Header(t, "2001:db8::1", 54321, "2001:db8::2", 5000, true)
	payload := []byte("ipv6 test")
	packet := append(header, payload...)

	_, err := clientConn.Write(packet)
	require.NoError(t, err)

	conn, err := listener.Accept()
	require.NoError(t, err)

	// Verify IPv6 source address from Proxy Protocol header, not actual client.
	assert.Equal(t, "[2001:db8::1]:54321", conn.RemoteAddr().String())

	// Verify payload received without header.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])
}

// TestProxyProtocol_TrustedIPs_Allow tests TrustedIPs mode when source is trusted.
func TestProxyProtocol_TrustedIPs_Allow(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure:   false,
		TrustedIPs: []string{"127.0.0.1/32"},
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	// Client is 127.0.0.1, which is in TrustedIPs.
	header := buildProxyProtocolV2Header(t, "192.0.2.100", 9999, "198.51.100.1", 5000, false)
	payload := []byte("trusted source")
	packet := append(header, payload...)

	_, err := clientConn.Write(packet)
	require.NoError(t, err)

	conn, err := listener.Accept()
	require.NoError(t, err)

	// Header should be parsed because source is trusted.
	assert.Equal(t, "192.0.2.100:9999", conn.RemoteAddr().String())

	// Verify payload received without header.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])
}

// TestProxyProtocol_TrustedIPs_Deny tests TrustedIPs mode when source is untrusted.
func TestProxyProtocol_TrustedIPs_Deny(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure:   false,
		TrustedIPs: []string{"10.0.0.0/8"},
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	// Send packet with Proxy Protocol header from untrusted source (127.0.0.1).
	header := buildProxyProtocolV2Header(t, "192.0.2.200", 8888, "198.51.100.1", 5000, false)
	payload := []byte("untrusted")
	packet := append(header, payload...)

	_, err := clientConn.Write(packet)
	require.NoError(t, err)

	conn, err := listener.Accept()
	require.NoError(t, err)

	// Header is ignored, actual source used.
	remoteAddr := conn.RemoteAddr().String()
	assert.Contains(t, remoteAddr, "127.0.0.1")
	assert.NotContains(t, remoteAddr, "192.0.2.200")

	// Entire packet (header+payload) should be received since header ignored.
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, packet, buf[:n])
}

// TestProxyProtocol_NoHeader tests that packets without Proxy Protocol headers are processed normally.
func TestProxyProtocol_NoHeader(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure: true,
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	// Send packet WITHOUT Proxy Protocol header.
	payload := []byte("regular udp packet")
	_, err := clientConn.Write(payload)
	require.NoError(t, err)

	conn, err := listener.Accept()
	require.NoError(t, err)

	// Should use actual client address.
	remoteAddr := conn.RemoteAddr().String()
	assert.Contains(t, remoteAddr, "127.0.0.1")

	// Should receive full payload.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])
}

// TestProxyProtocol_MalformedHeader tests that malformed Proxy Protocol headers cause packet to be dropped.
func TestProxyProtocol_MalformedHeader(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure: true,
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	// Send malformed header: Proxy Protocol v2 signature but invalid data.
	// This should be dropped and not create a session.
	malformed := []byte{
		0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A, // Signature
		0x21,       // Version 2, PROXY command
		0x12,       // UDP over IPv4
		0xFF, 0xFF, // Invalid length
		// Missing address data
	}
	_, err := clientConn.Write(malformed)
	require.NoError(t, err)

	// Send a valid packet with Proxy Protocol header to verify malformed packet was dropped.
	// If malformed packet created a session, we'd get that session.
	// If malformed packet was dropped, we get a new session from this valid packet.
	header := buildProxyProtocolV2Header(t, "192.0.2.99", 6666, "198.51.100.1", 5000, false)
	payload := []byte("valid packet after malformed")
	validPacket := append(header, payload...)
	_, err = clientConn.Write(validPacket)
	require.NoError(t, err)

	// Accept should return a session from the valid packet, not the malformed one.
	conn, err := listener.Accept()
	require.NoError(t, err)

	// Verify we got the session from the valid packet (with Proxy Protocol header source IP).
	assert.Equal(t, "192.0.2.99:6666", conn.RemoteAddr().String())

	// Verify payload is received without header.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])
}

// TestProxyProtocol_SubsequentPackets tests that subsequent packets without headers are routed to the same session.
func TestProxyProtocol_SubsequentPackets(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure: true,
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	// First packet: with Proxy Protocol header.
	header := buildProxyProtocolV2Header(t, "192.0.2.50", 7777, "198.51.100.1", 5000, false)
	payload1 := []byte("first packet")
	packet1 := append(header, payload1...)
	_, err := clientConn.Write(packet1)
	require.NoError(t, err)

	conn, err := listener.Accept()
	require.NoError(t, err)

	// Verify session keyed by header source IP.
	assert.Equal(t, "192.0.2.50:7777", conn.RemoteAddr().String())

	// Read first packet.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload1, buf[:n])

	// Second packet: without Proxy Protocol header (from same client).
	payload2 := []byte("second packet no header")
	_, err = clientConn.Write(payload2)
	require.NoError(t, err)

	// Should go to same session.
	n, err = conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload2, buf[:n])
}

// TestProxyProtocol_SessionIsolation tests that different Proxy Protocol source IPs create separate sessions.
func TestProxyProtocol_SessionIsolation(t *testing.T) {
	ppConfig := &ProxyProtocolConfig{
		Insecure: true,
	}
	listener, clientConn := setupTestListener(t, ppConfig)
	defer listener.Close()
	defer clientConn.Close()

	// Session 1: source 192.0.2.10:1000.
	header1 := buildProxyProtocolV2Header(t, "192.0.2.10", 1000, "198.51.100.1", 5000, false)
	payload1 := []byte("session 1 data")
	packet1 := append(header1, payload1...)
	_, err := clientConn.Write(packet1)
	require.NoError(t, err)

	conn1, err := listener.Accept()
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.10:1000", conn1.RemoteAddr().String())

	// Session 2: source 192.0.2.20:2000 (different source).
	header2 := buildProxyProtocolV2Header(t, "192.0.2.20", 2000, "198.51.100.1", 5000, false)
	payload2 := []byte("session 2 data")
	packet2 := append(header2, payload2...)
	_, err = clientConn.Write(packet2)
	require.NoError(t, err)

	conn2, err := listener.Accept()
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.20:2000", conn2.RemoteAddr().String())

	// Verify data isolated to correct sessions.
	buf := make([]byte, 1024)

	n, err := conn1.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload1, buf[:n])

	n, err = conn2.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload2, buf[:n])
}
