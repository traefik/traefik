package integration

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

// TCPHealthCheckSuite test suite for TCP health checks.
type TCPHealthCheckSuite struct {
	BaseSuite

	whoamitcp1IP string
	whoamitcp2IP string
}

func TestTCPHealthCheckSuite(t *testing.T) {
	suite.Run(t, new(TCPHealthCheckSuite))
}

func (s *TCPHealthCheckSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("tcp_healthcheck")
	s.composeUp()

	s.whoamitcp1IP = s.getComposeServiceIP("whoamitcp1")
	s.whoamitcp2IP = s.getComposeServiceIP("whoamitcp2")
}

func (s *TCPHealthCheckSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TCPHealthCheckSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/tcp_healthcheck/simple.toml", struct {
		Server1 string
		Server2 string
	}{s.whoamitcp1IP, s.whoamitcp2IP})

	s.traefikCmd(withConfigFile(file))

	// Wait for Traefik.
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("HostSNI(`*`)"))
	require.NoError(s.T(), err)

	// Test that we can consistently reach servers through load balancing.
	var (
		successfulConnectionsWhoamitcp1 int
		successfulConnectionsWhoamitcp2 int
	)
	for range 4 {
		out := s.whoIs("127.0.0.1:8093")
		require.NoError(s.T(), err)

		if strings.Contains(out, "whoamitcp1") {
			successfulConnectionsWhoamitcp1++
		}
		if strings.Contains(out, "whoamitcp2") {
			successfulConnectionsWhoamitcp2++
		}
	}

	assert.Equal(s.T(), 2, successfulConnectionsWhoamitcp1)
	assert.Equal(s.T(), 2, successfulConnectionsWhoamitcp2)

	// Stop one whoamitcp2 containers to simulate health check failure.
	conn, err := net.DialTimeout("tcp", s.whoamitcp2IP+":8080", time.Second)
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		_ = conn.Close()
	})

	s.composeStop("whoamitcp2")

	// Wait for the health check to detect the failure.
	time.Sleep(1 * time.Second)

	// Verify that the remaining server still responds.
	for range 3 {
		out := s.whoIs("127.0.0.1:8093")
		require.NoError(s.T(), err)
		assert.Contains(s.T(), out, "whoamitcp1")
	}
}

// connectTCP connects to the given TCP address and returns the response.
func (s *TCPHealthCheckSuite) whoIs(addr string) string {
	s.T().Helper()

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(s.T(), err)

	s.T().Cleanup(func() {
		_ = conn.Close()
	})

	_, err = conn.Write([]byte("WHO"))
	require.NoError(s.T(), err)

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	require.NoError(s.T(), err)

	return string(buffer[:n])
}
