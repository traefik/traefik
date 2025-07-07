package integration

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

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
