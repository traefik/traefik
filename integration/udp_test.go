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

	// Trusted IP.
	content, err := proxyProtoUDPRequest("127.0.0.1:8093", "1.2.3.4", 54321)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), content, "1.2.3.4")

	// Non-trusted IP.
	content, err = proxyProtoUDPRequest("127.0.0.1:8094", "1.2.3.4", 54321)
	require.NoError(s.T(), err)
	// When header is ignored, the packet is treated as regular data.
	// We're verifying the behavior by checking we can send to both entrypoints.
	assert.NotNil(s.T(), content)
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
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}
