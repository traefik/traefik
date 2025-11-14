package integration

import (
	"bytes"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type UDPSuite struct{ BaseSuite }

func TestUDPSuite(t *testing.T) {
	suite.Run(t, new(UDPSuite))
}

func (s *UDPSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("udp")
	s.composeUp()
}

func (s *UDPSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func guessWhoUDP(addr string) (string, error) {
	var conn net.Conn
	var err error

	udpAddr, err2 := net.ResolveUDPAddr("udp", addr)
	if err2 != nil {
		return "", err2
	}

	conn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return "", err
	}

	_, err = conn.Write([]byte("WHO"))
	if err != nil {
		return "", err
	}

	out := make([]byte, 2048)
	n, err := conn.Read(out)
	if err != nil {
		return "", err
	}
	return string(out[:n]), nil
}

func (s *UDPSuite) TestWRR() {
	file := s.adaptFile("fixtures/udp/wrr.toml", struct {
		WhoamiAIP string
		WhoamiBIP string
		WhoamiCIP string
		WhoamiDIP string
	}{
		WhoamiAIP: s.getComposeServiceIP("whoami-a"),
		WhoamiBIP: s.getComposeServiceIP("whoami-b"),
		WhoamiCIP: s.getComposeServiceIP("whoami-c"),
		WhoamiDIP: s.getComposeServiceIP("whoami-d"),
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("whoami-a"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8093/who", 5*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	stop := make(chan struct{})
	go func() {
		call := map[string]int{}
		for range 8 {
			out, err := guessWhoUDP("127.0.0.1:8093")
			require.NoError(s.T(), err)
			switch {
			case strings.Contains(out, "whoami-a"):
				call["whoami-a"]++
			case strings.Contains(out, "whoami-b"):
				call["whoami-b"]++
			case strings.Contains(out, "whoami-c"):
				call["whoami-c"]++
			default:
				call["unknown"]++
			}
		}
		assert.Equal(s.T(), map[string]int{"whoami-a": 3, "whoami-b": 2, "whoami-c": 3}, call)
		close(stop)
	}()

	select {
	case <-stop:
	case <-time.Tick(5 * time.Second):
		log.Info().Msg("Timeout")
	}
}

func (s *UDPSuite) TestProxyProtocol() {
	file := s.adaptFile("fixtures/udp/proxyprotocol.toml", struct {
		WhoamiIP string
	}{
		WhoamiIP: s.getComposeServiceIP("whoami-a"),
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("whoami"))
	require.NoError(s.T(), err)

	// Trusted IP.
	content, err := proxyProtoUDPRequest("127.0.0.1:8093", "1.2.3.4", 54321)
	require.NoError(s.T(), err)
	// Verify Traefik processes the packet and forwards it to the backend.
	// The backend responds with its own information (whoami-a).
	assert.Contains(s.T(), content, "whoami-a")

	// Non-trusted IP.
	content, err = proxyProtoUDPRequest("127.0.0.1:8094", "1.2.3.4", 54321)
	require.NoError(s.T(), err)
	// When header is ignored, the packet is treated as regular data.
	// The backend receives the full packet (header + payload) and echoes it back.
	// We're verifying the behavior by checking we get a response.
	assert.Contains(s.T(), content, "Received:")
}

// TestProxyProtocolMultiplePackets verifies session continuity when using Proxy Protocol.
// This test validates that after the first packet with a Proxy Protocol header establishes
// a session, subsequent packets from the same source can be sent without the header and
// are correctly associated with the existing session.
func (s *UDPSuite) TestProxyProtocolMultiplePackets() {
	file := s.adaptFile("fixtures/udp/proxyprotocol.toml", struct {
		WhoamiIP string
	}{
		WhoamiIP: s.getComposeServiceIP("whoami-a"),
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("whoami"))
	require.NoError(s.T(), err)

	// Send 3 packets to verify:
	// - First packet with Proxy Protocol header is routed correctly and response arrives.
	// - Subsequent packets without Proxy Protocol header are routed to same session.
	responses, err := proxyProtoUDPRequestMultiPacket("127.0.0.1:8093", "1.2.3.4", 54321, 3)
	require.NoError(s.T(), err)
	require.Len(s.T(), responses, 3)

	// All responses should be received, validating that responses are sent to the correct address.
	for i, response := range responses {
		assert.Contains(s.T(), response, "whoami-a", "packet %d should receive valid response", i+1)
	}
}

func proxyProtoUDPRequest(address, srcIP string, srcPort int) (string, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return "", err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Create a Proxy Protocol v2 header for UDP
	header := &proxyproto.Header{
		Version:           2,
		Command:           proxyproto.PROXY,
		TransportProtocol: proxyproto.UDPv4,
		SourceAddr: &net.UDPAddr{
			IP:   net.ParseIP(srcIP),
			Port: srcPort,
		},
		DestinationAddr: &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 8080,
		},
	}

	// Write the Proxy Protocol header
	var headerBuf bytes.Buffer
	_, err = header.WriteTo(&headerBuf)
	if err != nil {
		return "", err
	}

	// Send header + payload
	payload := []byte("WHO")
	packet := append(headerBuf.Bytes(), payload...)
	_, err = conn.Write(packet)
	if err != nil {
		return "", err
	}

	// Read response
	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		return "", err
	}

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

// proxyProtoUDPRequestMultiPacket sends multiple UDP packets to test session continuity.
// The first packet includes a Proxy Protocol header, subsequent packets do not.
// Returns responses for all packets.
func proxyProtoUDPRequestMultiPacket(address, srcIP string, srcPort, numPackets int) ([]string, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	responses := make([]string, 0, numPackets)
	for i := range numPackets {
		var packet []byte

		if i == 0 {
			// First packet: Include Proxy Protocol v2 header.
			header := &proxyproto.Header{
				Version:           2,
				Command:           proxyproto.PROXY,
				TransportProtocol: proxyproto.UDPv4,
				SourceAddr: &net.UDPAddr{
					IP:   net.ParseIP(srcIP),
					Port: srcPort,
				},
				DestinationAddr: &net.UDPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				},
			}

			var headerBuf bytes.Buffer
			_, err = header.WriteTo(&headerBuf)
			if err != nil {
				return nil, err
			}

			payload := []byte("WHO")
			packet = append(headerBuf.Bytes(), payload...)
		} else {
			// Subsequent packets: No Proxy Protocol header (session continuation).
			packet = []byte("WHO")
		}

		_, err = conn.Write(packet)
		if err != nil {
			return nil, err
		}

		// Read response.
		err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			return nil, err
		}

		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}

		responses = append(responses, string(buf[:n]))

		// Small delay between packets to ensure session is established.
		if i < numPackets-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return responses, nil
}
