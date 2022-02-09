package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type TracingSuite struct {
	BaseSuite
	whoamiIP   string
	whoamiPort int
	tracerIP   string
}

type TracingTemplate struct {
	WhoamiIP               string
	WhoamiPort             int
	IP                     string
	TraceContextHeaderName string
}

func (s *TracingSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "tracing")
	s.composeUp(c)

	s.whoamiIP = s.getComposeServiceIP(c, "whoami")
	s.whoamiPort = 80
}

func (s *TracingSuite) startZipkin(c *check.C) {
	s.composeUp(c, "zipkin")
	s.tracerIP = s.getComposeServiceIP(c, "zipkin")

	// Wait for Zipkin to turn ready.
	err := try.GetRequest("http://"+s.tracerIP+":9411/api/v2/services", 20*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestZipkinRateLimit(c *check.C) {
	s.startZipkin(c)
	// defer s.composeStop(c, "zipkin")

	file := s.adaptFile(c, "fixtures/tracing/simple-zipkin.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.tracerIP,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

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

	err = try.GetRequest("http://"+s.tracerIP+":9411/api/v2/spans?serviceName=tracing", 20*time.Second, try.BodyContains("forward service1/router1@file", "ratelimit-1@file"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestZipkinRetry(c *check.C) {
	s.startZipkin(c)
	defer s.composeStop(c, "zipkin")

	file := s.adaptFile(c, "fixtures/tracing/simple-zipkin.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: 81,
		IP:         s.tracerIP,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/retry", 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.tracerIP+":9411/api/v2/spans?serviceName=tracing", 20*time.Second, try.BodyContains("forward service2/router2@file", "retry@file"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestZipkinAuth(c *check.C) {
	s.startZipkin(c)
	defer s.composeStop(c, "zipkin")

	file := s.adaptFile(c, "fixtures/tracing/simple-zipkin.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.tracerIP,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.tracerIP+":9411/api/v2/spans?serviceName=tracing", 20*time.Second, try.BodyContains("entrypoint web", "basic-auth@file"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) startJaeger(c *check.C) {
	s.composeUp(c, "jaeger", "whoami")
	s.tracerIP = s.getComposeServiceIP(c, "jaeger")

	// Wait for Jaeger to turn ready.
	err := try.GetRequest("http://"+s.tracerIP+":16686/api/services", 20*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestJaegerRateLimit(c *check.C) {
	s.startJaeger(c)
	defer s.composeStop(c, "jaeger")

	file := s.adaptFile(c, "fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             s.whoamiPort,
		IP:                     s.tracerIP,
		TraceContextHeaderName: "uber-trace-id",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

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

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.tracerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("forward service1/router1@file", "ratelimit-1@file"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestJaegerRetry(c *check.C) {
	s.startJaeger(c)
	defer s.composeStop(c, "jaeger")

	file := s.adaptFile(c, "fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             81,
		IP:                     s.tracerIP,
		TraceContextHeaderName: "uber-trace-id",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/retry", 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.tracerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("forward service2/router2@file", "retry@file"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestJaegerAuth(c *check.C) {
	s.startJaeger(c)
	defer s.composeStop(c, "jaeger")

	file := s.adaptFile(c, "fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             s.whoamiPort,
		IP:                     s.tracerIP,
		TraceContextHeaderName: "uber-trace-id",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.tracerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("EntryPoint web", "basic-auth@file"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestJaegerCustomHeader(c *check.C) {
	s.startJaeger(c)
	defer s.composeStop(c, "jaeger")

	file := s.adaptFile(c, "fixtures/tracing/simple-jaeger.toml", TracingTemplate{
		WhoamiIP:               s.whoamiIP,
		WhoamiPort:             s.whoamiPort,
		IP:                     s.tracerIP,
		TraceContextHeaderName: "powpow",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.tracerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("EntryPoint web", "basic-auth@file"))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestJaegerAuthCollector(c *check.C) {
	s.startJaeger(c)
	defer s.composeStop(c, "jaeger")

	file := s.adaptFile(c, "fixtures/tracing/simple-jaeger-collector.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.tracerIP,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://"+s.tracerIP+":16686/api/traces?service=tracing", 20*time.Second, try.BodyContains("EntryPoint web", "basic-auth@file"))
	c.Assert(err, checker.IsNil)
}
