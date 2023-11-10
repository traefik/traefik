package integration

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v3/integration/try"
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

func (s *TracingSuite) startTempo(c *check.C) {
	s.composeUp(c, "tempo", "whoami")
	s.tracerIP = s.getComposeServiceIP(c, "tempo")

	// Wait for tempo to turn ready.
	err := try.GetRequest("http://"+s.tracerIP+":3200/ready", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestOpentelemetryRateLimit(c *check.C) {
	s.startTempo(c)
	defer s.composeStop(c, "tempo")

	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
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

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	c.Assert(err, checker.IsNil)

	checkTraceContent(c, s.tracerIP, "forward service1/router1@file", "ratelimit-1@file")
}

func (s *TracingSuite) TestOpentelemetryRetry(c *check.C) {
	s.startTempo(c)
	defer s.composeStop(c, "tempo")

	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
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

	checkTraceContent(c, s.tracerIP, "forward service2/router2@file", "retry@file")
}

func (s *TracingSuite) TestOpentelemetryAuth(c *check.C) {
	s.startTempo(c)
	defer s.composeStop(c, "tempo")

	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
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

	checkTraceContent(c, s.tracerIP, "EntryPoint web", "basic-auth@file")
}

func checkTraceContent(c *check.C, tracerIP string, bodyContains ...string) {
	baseURL, err := url.Parse("http://" + tracerIP + ":3200/api/search")
	c.Assert(err, checker.IsNil)

	req := &http.Request{
		Method: http.MethodGet,
		URL:    baseURL,
	}
	resp, err := try.Response(req, 20*time.Second)
	c.Assert(err, checker.IsNil)

	out := &TraceResponse{}
	err = json.NewDecoder(resp.Body).Decode(&out)
	c.Assert(err, checker.IsNil)

	if len(out.Traces) == 0 {
		c.Fatalf("expected at least one trace, got %d", len(out.Traces))
	}
	traceID := out.Traces[0].TraceID
	err = try.GetRequest("http://"+tracerIP+":3200/api/traces/"+traceID, 20*time.Second, try.BodyContains(bodyContains...))
	c.Assert(err, checker.IsNil)
}

// TraceResponse contains a list of traces.
type TraceResponse struct {
	Traces []Trace `json:"traces"`
}

// Trace represents a simplified grafana tempo trace.
type Trace struct {
	TraceID string `json:"traceID"`
}
