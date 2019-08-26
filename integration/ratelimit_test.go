package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/v2/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

type RateLimitSuite struct {
	BaseSuite
	ServerIP string
}

func (s *RateLimitSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "ratelimit")
	s.composeProject.Start(c)

	s.ServerIP = s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress
}

func (s *RateLimitSuite) TestSimpleConfiguration(c *check.C) {
	file := s.adaptFile(c, "fixtures/ratelimit/simple.toml", struct {
		Server1 string
	}{s.ServerIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("ratelimit"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	c.Assert(err, checker.IsNil)

	// sleep for 4 seconds to be certain the configured time period has elapsed
	// then test another request and verify a 200 status code
	time.Sleep(4 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// continue requests at 3 second intervals to test the other rate limit time period
	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	c.Assert(err, checker.IsNil)
}
