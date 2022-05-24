package integration

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type TimeoutSuite struct{ BaseSuite }

func (s *TimeoutSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "timeout")
	s.composeUp(c)
}

func (s *TimeoutSuite) TestForwardingTimeouts(c *check.C) {
	timeoutEndpointIP := s.getComposeServiceIP(c, "timeoutEndpoint")
	file := s.adaptFile(c, "fixtures/timeout/forwarding_timeouts.toml", struct{ TimeoutEndpoint string }{timeoutEndpointIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Path(`/dialTimeout`)"))
	c.Assert(err, checker.IsNil)

	// This simulates a DialTimeout when connecting to the backend server.
	response, err := http.Get("http://127.0.0.1:8000/dialTimeout")
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusGatewayTimeout)

	// Check that timeout service is available
	statusURL := fmt.Sprintf("http://%s/statusTest?status=200",
		net.JoinHostPort(timeoutEndpointIP, "9000"))
	c.Assert(try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)), checker.IsNil)

	// This simulates a ResponseHeaderTimeout.
	response, err = http.Get("http://127.0.0.1:8000/responseHeaderTimeout?sleep=1000")
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusGatewayTimeout)
}
