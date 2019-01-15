package integration

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

type TimeoutSuite struct{ BaseSuite }

func (s *TimeoutSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "timeout")
	s.composeProject.Start(c)
}

func (s *TimeoutSuite) TestForwardingTimeouts(c *check.C) {
	httpTimeoutEndpoint := s.composeProject.Container(c, "timeoutEndpoint").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/timeout/forwarding_timeouts.toml", struct {
		TimeoutEndpoint string
	}{httpTimeoutEndpoint})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, try.BodyContains("Path:/dialTimeout"))
	c.Assert(err, checker.IsNil)

	// This simulates a DialTimeout when connecting to the backend server.
	response, err := http.Get("http://127.0.0.1:8000/dialTimeout")
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusGatewayTimeout)

	// Check that timeout service is available
	statusURL := fmt.Sprintf("http://%s:9000/statusTest?status=200", httpTimeoutEndpoint)
	c.Assert(try.GetRequest(statusURL, 60*time.Second, try.StatusCodeIs(http.StatusOK)), checker.IsNil)

	// This simulates a ResponseHeaderTimeout.
	response, err = http.Get("http://127.0.0.1:8000/responseHeaderTimeout?sleep=1000")
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusGatewayTimeout)
}
