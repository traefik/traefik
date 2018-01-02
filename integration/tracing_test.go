package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

type TracingSuite struct {
	BaseSuite
	WhoAmiIP       string
	WhoAmiPort     int
	ZipkinIP       string
	TracingBackend string
}

type TracingTemplate struct {
	WhoAmiIP       string
	WhoAmiPort     int
	ZipkinIP       string
	TracingBackend string
}

func (s *TracingSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "tracing")
	s.composeProject.Start(c, "whoami")

	s.WhoAmiIP = s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress
	s.WhoAmiPort = 80
}

func (s *TracingSuite) startZipkin(c *check.C) {
	s.composeProject.Start(c, "zipkin")
	s.ZipkinIP = s.composeProject.Container(c, "zipkin").NetworkSettings.IPAddress
	s.TracingBackend = "zipkin"

	// Wait for Zipkin to turn ready.
	err := try.GetRequest("http://"+s.ZipkinIP+":9411/api/v2/services", 20*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestZipkinRateLimit(c *check.C) {
	s.startZipkin(c)
	file := s.adaptFile(c, "fixtures/tracing/simple.toml", TracingTemplate{
		WhoAmiIP:       s.WhoAmiIP,
		WhoAmiPort:     s.WhoAmiPort,
		ZipkinIP:       s.ZipkinIP,
		TracingBackend: s.TracingBackend,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	c.Assert(err, checker.IsNil)

	// sleep for 4 seconds to be certain the configured time period has elapsed
	// then test another request and verify a 200 status code
	time.Sleep(4 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// continue requests at 3 second intervals to test the other rate limit time period
	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.ZipkinIP+":9411/api/v2/spans?serviceName=tracing", 10*time.Second, try.BodyContains("forward frontend1/backend1", "rate limit"))
	c.Assert(err, checker.IsNil)

}

func (s *TracingSuite) TestZipkinRetry(c *check.C) {
	s.startZipkin(c)
	file := s.adaptFile(c, "fixtures/tracing/simple.toml", TracingTemplate{
		WhoAmiIP:       s.WhoAmiIP,
		WhoAmiPort:     81,
		ZipkinIP:       s.ZipkinIP,
		TracingBackend: s.TracingBackend,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8000/retry", 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.ZipkinIP+":9411/api/v2/spans?serviceName=tracing", 10*time.Second, try.BodyContains("forward frontend2/backend2", "retry"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestZipkinAuth(c *check.C) {
	s.startZipkin(c)
	file := s.adaptFile(c, "fixtures/tracing/simple.toml", TracingTemplate{
		WhoAmiIP:       s.WhoAmiIP,
		WhoAmiPort:     s.WhoAmiPort,
		ZipkinIP:       s.ZipkinIP,
		TracingBackend: s.TracingBackend,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.ZipkinIP+":9411/api/v2/spans?serviceName=tracing", 10*time.Second, try.BodyContains("entrypoint http", "auth basic"))
	c.Assert(err, checker.IsNil)
}
