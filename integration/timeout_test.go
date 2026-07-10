package integration

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type TimeoutSuite struct{ BaseSuite }

func TestTimeoutSuite(t *testing.T) {
	suite.Run(t, new(TimeoutSuite))
}

func (s *TimeoutSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("timeout")
	s.composeUp()
}

func (s *TimeoutSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TimeoutSuite) TestForwardingTimeouts() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/forwarding_timeouts.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/dialTimeout`)"))
	require.NoError(s.T(), err)

	// This simulates a DialTimeout when connecting to the backend server.
	response, err := http.Get("http://127.0.0.1:8000/dialTimeout")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	// Check that timeout service is available
	statusURL := fmt.Sprintf("http://%s/statusTest?status=200",
		net.JoinHostPort(timeoutEndpointIP, "9000"))
	assert.NoError(s.T(), try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)))

	// This simulates a ResponseHeaderTimeout.
	response, err = http.Get("http://127.0.0.1:8000/responseHeaderTimeout?sleep=1000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)
}

func (s *TimeoutSuite) TestRouterRespondingTimeouts() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/router_responding_timeouts.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/roundTrip`)"))
	require.NoError(s.T(), err)

	// Check that timeout service is available.
	statusURL := fmt.Sprintf("http://%s/statusTest?status=200",
		net.JoinHostPort(timeoutEndpointIP, "9000"))
	require.NoError(s.T(), try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)))

	client := &http.Client{}

	// A backend answering after the router roundTrip timeout yields a 504.
	response, err := client.Get("http://127.0.0.1:8000/roundTrip?sleep=2000")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	// Drain the response to allow the keep-alive connection reuse.
	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// A backend answering below the router roundTrip timeout is not affected,
	// even on the same keep-alive connection (no leaked connection deadline).
	response, err = client.Get("http://127.0.0.1:8000/roundTrip?sleep=100")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func (s *TimeoutSuite) TestRouterRespondingTimeoutsHTTP3() {
	timeoutEndpointIP := s.getComposeServiceIP("timeoutEndpoint")
	file := s.adaptFile("fixtures/timeout/router_responding_timeouts_http3.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/roundTrip`)"))
	require.NoError(s.T(), err)

	transport := &http3.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: "snitest.com"}}
	defer func() { _ = transport.Close() }()

	// Wait for the HTTP/3 entry point to be ready with a below-timeout request.
	fastReq, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/roundTrip?sleep=100", nil)
	require.NoError(s.T(), err)
	fastReq.Host = "snitest.com"
	require.NoError(s.T(), try.RequestWithTransport(fastReq, 60*time.Second, transport, try.StatusCodeIs(http.StatusOK)))

	client := &http.Client{Transport: transport}

	// A backend answering after the router roundTrip timeout yields a 504 over HTTP/3.
	slowReq, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/roundTrip?sleep=2000", nil)
	require.NoError(s.T(), err)
	slowReq.Host = "snitest.com"

	response, err := client.Do(slowReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusGatewayTimeout, response.StatusCode)

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())

	// A subsequent below-timeout request on the same HTTP/3 connection is unaffected,
	// as the deadline is per-stream and cannot leak to another stream.
	fastReq2, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/roundTrip?sleep=100", nil)
	require.NoError(s.T(), err)
	fastReq2.Host = "snitest.com"

	response, err = client.Do(fastReq2)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(s.T(), err)
	require.NoError(s.T(), response.Body.Close())
}
