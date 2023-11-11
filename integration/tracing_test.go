package integration

import (
	"encoding/json"
	"io"
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

func (s *TracingSuite) startTempo(c *check.C) {
	s.composeUp(c, "tempo", "whoami")
	s.tracerIP = s.getComposeServiceIP(c, "tempo")

	// Wait for tempo to turn ready.
	err := try.GetRequest("http://"+s.tracerIP+":3200/ready", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TestZipkinRateLimit(c *check.C) {
	s.startTempo(c)
	defer s.composeStop(c, "tempo")

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

	checkTraceContent(c, s.tracerIP, "forward service1/router1@file", "ratelimit-1@file")
}

func (s *TracingSuite) TestZipkinRetry(c *check.C) {
	s.startTempo(c)
	defer s.composeStop(c, "tempo")

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

	checkTraceContent(c, s.tracerIP, "forward service2/router2@file", "retry@file")
}

func (s *TracingSuite) TestZipkinAuth(c *check.C) {
	s.startTempo(c)
	defer s.composeStop(c, "tempo")

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

	checkTraceContent(c, s.tracerIP, "entrypoint web 127.0.0.1:8000", "basic-auth@file")
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

	checkTraceContent(c, s.tracerIP, "EntryPoint web 127.0.0.1:8000", "basic-auth@file")
}

func checkTraceContent(c *check.C, tracerIP string, bodyContains ...string) {
	baseURL, err := url.Parse("http://" + tracerIP + ":3200/api/search")
	c.Assert(err, checker.IsNil)

	req := &http.Request{
		Method: http.MethodGet,
		URL:    baseURL,
	}
	// Wait for traces to be available.
	time.Sleep(10 * time.Second)
	resp, err := try.Response(req, 5*time.Second)
	c.Assert(err, checker.IsNil)

	out := &TraceResponse{}
	content, err := io.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)
	err = json.Unmarshal(content, &out)
	c.Assert(err, checker.IsNil)

	if len(out.Traces) == 0 {
		c.Fatalf("expected at least one trace, got %d (%s)", len(out.Traces), string(content))
	}

	containsMap := make(map[string]struct{}, len(bodyContains))
	for _, b := range bodyContains {
		containsMap[b] = struct{}{}
	}

	for _, t := range out.Traces {
		baseURL, err := url.Parse("http://" + tracerIP + ":3200/api/traces/" + t.TraceID)
		c.Assert(err, checker.IsNil)

		req := &http.Request{
			Method: http.MethodGet,
			URL:    baseURL,
		}

		resp, err := try.Response(req, 5*time.Second)
		c.Assert(err, checker.IsNil)

		content, err := io.ReadAll(resp.Body)
		c.Assert(err, checker.IsNil)

		out := &BatchesResponse{}
		err = json.Unmarshal(content, &out)
		c.Assert(err, checker.IsNil)

		for _, b := range out.Batches {
			for _, s := range b.ScopeSpans {
				for _, span := range s.Spans {
					delete(containsMap, span.Name)
					c.Logf("found %s", span.Name)
				}
			}
		}
	}

	if len(containsMap) > 0 {
		notFound := []string{}
		for v := range containsMap {
			notFound = append(notFound, v)
		}
		c.Errorf("expected traces to contain %v, but not found %v", bodyContains, notFound)
	}
}

// TraceResponse contains a list of traces.
type TraceResponse struct {
	Traces []Trace `json:"traces"`
}

// Trace represents a simplified grafana tempo trace.
type Trace struct {
	TraceID       string `json:"traceID"`
	RootTraceName string `json:"rootTraceName"`
	DurationMs    int    `json:"durationMs"`
}

// BatchesResponse contains a list of batches.
type BatchesResponse struct {
	Batches []Batch `json:"batches"`
}

// Batch represents a simplified grafana tempo batch.
type Batch struct {
	ScopeSpans []ScopeSpan `json:"scopeSpans"`
}

// ScopeSpan represents a simplified grafana tempo scope span.
type ScopeSpan struct {
	Spans []Span `json:"spans"`
}

// Span represents a simplified grafana tempo Span.
type Span struct {
	Name string `json:"name"`
}
