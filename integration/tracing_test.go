package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-check/check"
	"github.com/tidwall/gjson"
	"github.com/traefik/traefik/v3/integration/try"
	checker "github.com/vdemeester/shakers"
)

type TracingSuite struct {
	BaseSuite
	whoamiIP        string
	whoamiPort      int
	tempoIP         string
	otelCollectorIP string
}

type TracingTemplate struct {
	WhoamiIP               string
	WhoamiPort             int
	IP                     string
	TraceContextHeaderName string
	IsHTTP                 bool
}

func (s *TracingSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "tracing")
	s.composeUp(c)

	s.whoamiIP = s.getComposeServiceIP(c, "whoami")
	s.whoamiPort = 80
}

func (s *TracingSuite) SetUpTest(c *check.C) {
	s.composeUp(c, "tempo", "otel-collector", "whoami")
	s.tempoIP = s.getComposeServiceIP(c, "tempo")

	// Wait for tempo to turn ready.
	err := try.GetRequest("http://"+s.tempoIP+":3200/ready", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	s.otelCollectorIP = s.getComposeServiceIP(c, "otel-collector")

	// Wait for otel collector to turn ready.
	err = try.GetRequest("http://"+s.otelCollectorIP+":13133/", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TracingSuite) TearDownTest(c *check.C) {
	s.composeStop(c, "tempo")
}

func (s *TracingSuite) TestOpentelemetryBasic_HTTP(c *check.C) {
	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
		IsHTTP:     true,
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

	err = try.GetRequest("http://127.0.0.1:8000/basic", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "entry_point",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/basic",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router0@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.2.name": "service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.3.name":                                                           "reverse-proxy",
			"batches.0.scopeSpans.0.spans.3.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
	}

	checkTraceContent(c, s.tempoIP, contains)
}

func (s *TracingSuite) TestOpentelemetryBasic_gRPC(c *check.C) {
	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
		IsHTTP:     false,
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

	err = try.GetRequest("http://127.0.0.1:8000/basic", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "entry_point",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/basic",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router0@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.2.name": "service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.3.name":                                                           "reverse-proxy",
			"batches.0.scopeSpans.0.spans.3.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
	}

	checkTraceContent(c, s.tempoIP, contains)
}

func (s *TracingSuite) TestOpentelemetryRateLimit(c *check.C) {
	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
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

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "entry_point",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/ratelimit",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router1@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",

			"batches.0.scopeSpans.0.spans.2.name": "Retry",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.3.name": "RateLimiter",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "ratelimit-1@file",

			"batches.0.scopeSpans.0.spans.4.name": "service",
			"batches.0.scopeSpans.0.spans.4.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",

			"batches.0.scopeSpans.0.spans.5.name":                                                           "reverse-proxy",
			"batches.0.scopeSpans.0.spans.5.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "entry_point",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/ratelimit",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "429",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router1@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",

			"batches.0.scopeSpans.0.spans.2.name": "Retry",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.3.name": "RateLimiter",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "ratelimit-1@file",

			"batches.0.scopeSpans.0.spans.4.name": "Retry",
			"batches.0.scopeSpans.0.spans.4.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.resend_count\").value.intValue":          "1",

			"batches.0.scopeSpans.0.spans.5.name": "RateLimiter",
			"batches.0.scopeSpans.0.spans.5.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "ratelimit-1@file",

			"batches.0.scopeSpans.0.spans.6.name": "Retry",
			"batches.0.scopeSpans.0.spans.6.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"http.resend_count\").value.intValue":          "2",

			"batches.0.scopeSpans.0.spans.7.name": "RateLimiter",
			"batches.0.scopeSpans.0.spans.7.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "ratelimit-1@file",
		},
	}

	checkTraceContent(c, s.tempoIP, contains)
}

func (s *TracingSuite) TestOpentelemetryRetry(c *check.C) {
	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: 81,
		IP:         s.otelCollectorIP,
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

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "entry_point",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/retry",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.0.status.code":                                                    "STATUS_CODE_ERROR",

			"batches.0.scopeSpans.0.spans.1.name": "router",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router2@file",

			"batches.0.scopeSpans.0.spans.2.name": "Retry",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.3.name": "service",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.4.name":                                                           "reverse-proxy",
			"batches.0.scopeSpans.0.spans.4.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",

			"batches.0.scopeSpans.0.spans.5.name": "Retry",
			"batches.0.scopeSpans.0.spans.5.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"http.resend_count\").value.intValue":          "1",

			"batches.0.scopeSpans.0.spans.6.name": "service",
			"batches.0.scopeSpans.0.spans.6.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.7.name":                                                           "reverse-proxy",
			"batches.0.scopeSpans.0.spans.7.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",

			"batches.0.scopeSpans.0.spans.8.name": "Retry",
			"batches.0.scopeSpans.0.spans.8.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"http.resend_count\").value.intValue":          "2",

			"batches.0.scopeSpans.0.spans.9.name": "service",
			"batches.0.scopeSpans.0.spans.9.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.9.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.10.name":                                                           "reverse-proxy",
			"batches.0.scopeSpans.0.spans.10.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.10.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.10.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.10.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
	}

	checkTraceContent(c, s.tempoIP, contains)
}

func (s *TracingSuite) TestOpentelemetryAuth(c *check.C) {
	file := s.adaptFile(c, "fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
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

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "entry_point",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/auth",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "401",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router3@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service3@file",

			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "basic-auth@file",
		},
	}

	checkTraceContent(c, s.tempoIP, contains)
}

func checkTraceContent(c *check.C, tempoIP string, expectedJSON []map[string]string) {
	baseURL, err := url.Parse("http://" + tempoIP + ":3200/api/search")
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

	var contents []string
	for _, t := range out.Traces {
		baseURL, err := url.Parse("http://" + tempoIP + ":3200/api/traces/" + t.TraceID)
		c.Assert(err, checker.IsNil)

		req := &http.Request{
			Method: http.MethodGet,
			URL:    baseURL,
		}

		resp, err := try.Response(req, 5*time.Second)
		c.Assert(err, checker.IsNil)

		content, err := io.ReadAll(resp.Body)
		c.Assert(err, checker.IsNil)

		contents = append(contents, string(content))
	}

	for _, expected := range expectedJSON {
		containsAll(c, expected, contents)
	}
}

func containsAll(c *check.C, expectedJSON map[string]string, contents []string) {
	for k, v := range expectedJSON {
		found := false
		for _, content := range contents {
			if gjson.Get(content, k).String() == v {
				found = true
				break
			}
		}

		if !found {
			c.Log("[" + strings.Join(contents, ",") + "]")
			c.Errorf("missing element: \nKey: %q\nValue: %q ", k, v)
		}
	}
}

// TraceResponse contains a list of traces.
type TraceResponse struct {
	Traces []Trace `json:"traces"`
}

// Trace represents a simplified grafana tempo trace.
type Trace struct {
	TraceID string `json:"traceID"`
}
