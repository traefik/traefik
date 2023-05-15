package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/traefik/traefik/v3/integration/try"
)

type TracingSuite struct {
	BaseSuite
	whoamiIP        string
	whoamiPort      int
	tempoIP         string
	otelCollectorIP string
}

func TestTracingSuite(t *testing.T) {
	suite.Run(t, new(TracingSuite))
}

type TracingTemplate struct {
	WhoamiIP               string
	WhoamiPort             int
	IP                     string
	TraceContextHeaderName string
	IsHTTP                 bool
}

func (s *TracingSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("tracing")
	s.composeUp()

	s.whoamiIP = s.getComposeServiceIP("whoami")
	s.whoamiPort = 80

	// Wait for whoami to turn ready.
	err := try.GetRequest("http://"+s.whoamiIP+":80", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	s.otelCollectorIP = s.getComposeServiceIP("otel-collector")

	// Wait for otel collector to turn ready.
	err = try.GetRequest("http://"+s.otelCollectorIP+":13133/", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TracingSuite) SetupTest() {
	s.composeUp("tempo")

	s.tempoIP = s.getComposeServiceIP("tempo")

	// Wait for tempo to turn ready.
	err := try.GetRequest("http://"+s.tempoIP+":3200/ready", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *TracingSuite) TearDownTest() {
	s.composeStop("tempo")
}

func (s *TracingSuite) TestOpentelemetryBasic_HTTP() {
	file := s.adaptFile("fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
		IsHTTP:     true,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/basic", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/basic",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "Router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router0@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.2.name": "Service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.3.name":                                                           "ReverseProxy",
			"batches.0.scopeSpans.0.spans.3.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpentelemetryBasic_gRPC() {
	file := s.adaptFile("fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
		IsHTTP:     false,
	})
	defer os.Remove(file)

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/basic", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/basic",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "Router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router0@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.2.name": "Service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.3.name":                                                           "ReverseProxy",
			"batches.0.scopeSpans.0.spans.3.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpentelemetryRateLimit() {
	file := s.adaptFile("fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
	})
	defer os.Remove(file)

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	require.NoError(s.T(), err)

	// sleep for 4 seconds to be certain the configured time period has elapsed
	// then test another request and verify a 200 status code
	time.Sleep(4 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// continue requests at 3 second intervals to test the other rate limit time period
	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	time.Sleep(3 * time.Second)
	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusTooManyRequests))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/ratelimit",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "Router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router1@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",

			"batches.0.scopeSpans.0.spans.2.name": "Retry",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.3.name": "RateLimiter",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "ratelimit-1@file",

			"batches.0.scopeSpans.0.spans.4.name": "Service",
			"batches.0.scopeSpans.0.spans.4.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",

			"batches.0.scopeSpans.0.spans.5.name":                                                           "ReverseProxy",
			"batches.0.scopeSpans.0.spans.5.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.0.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/ratelimit",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "429",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "Router",
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

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpentelemetryRetry() {
	file := s.adaptFile("fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: 81,
		IP:         s.otelCollectorIP,
	})
	defer os.Remove(file)

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/retry", 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/retry",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.0.status.code":                                                    "STATUS_CODE_ERROR",

			"batches.0.scopeSpans.0.spans.1.name": "Router",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router2@file",

			"batches.0.scopeSpans.0.spans.2.name": "Retry",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.3.name": "Service",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.4.name":                                                           "ReverseProxy",
			"batches.0.scopeSpans.0.spans.4.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",

			"batches.0.scopeSpans.0.spans.5.name": "Retry",
			"batches.0.scopeSpans.0.spans.5.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"http.resend_count\").value.intValue":          "1",

			"batches.0.scopeSpans.0.spans.6.name": "Service",
			"batches.0.scopeSpans.0.spans.6.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.7.name":                                                           "ReverseProxy",
			"batches.0.scopeSpans.0.spans.7.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",

			"batches.0.scopeSpans.0.spans.8.name": "Retry",
			"batches.0.scopeSpans.0.spans.8.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"http.resend_count\").value.intValue":          "2",

			"batches.0.scopeSpans.0.spans.9.name": "Service",
			"batches.0.scopeSpans.0.spans.9.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.9.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.10.name":                                                           "ReverseProxy",
			"batches.0.scopeSpans.0.spans.10.kind":                                                           "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.10.attributes.#(key=\"url.scheme\").value.stringValue":             "http",
			"batches.0.scopeSpans.0.spans.10.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
			"batches.0.scopeSpans.0.spans.10.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpentelemetryAuth() {
	file := s.adaptFile("fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
	})
	defer os.Remove(file)

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.path\").value.stringValue":               "/auth",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "401",

			"batches.0.scopeSpans.0.spans.1.name":                                                         "Router",
			"batches.0.scopeSpans.0.spans.1.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "router3@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service3@file",

			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "basic-auth@file",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestNoInternals() {
	file := s.adaptFile("fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: s.whoamiPort,
		IP:         s.otelCollectorIP,
		IsHTTP:     true,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/ratelimit", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/ping", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8080/ping", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	baseURL, err := url.Parse("http://" + s.tempoIP + ":3200/api/search")
	require.NoError(s.T(), err)

	req := &http.Request{
		Method: http.MethodGet,
		URL:    baseURL,
	}
	// Wait for traces to be available.
	time.Sleep(10 * time.Second)
	resp, err := try.Response(req, 5*time.Second)
	require.NoError(s.T(), err)

	out := &TraceResponse{}
	content, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	err = json.Unmarshal(content, &out)
	require.NoError(s.T(), err)

	s.NotEmptyf(len(out.Traces), "expected at least one trace")

	for _, t := range out.Traces {
		baseURL, err := url.Parse("http://" + s.tempoIP + ":3200/api/traces/" + t.TraceID)
		require.NoError(s.T(), err)

		req := &http.Request{
			Method: http.MethodGet,
			URL:    baseURL,
		}

		resp, err := try.Response(req, 5*time.Second)
		require.NoError(s.T(), err)

		content, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		require.NotContains(s.T(), content, "@internal")
	}
}

func (s *TracingSuite) checkTraceContent(expectedJSON []map[string]string) {
	s.T().Helper()

	baseURL, err := url.Parse("http://" + s.tempoIP + ":3200/api/search")
	require.NoError(s.T(), err)

	req := &http.Request{
		Method: http.MethodGet,
		URL:    baseURL,
	}
	// Wait for traces to be available.
	time.Sleep(10 * time.Second)
	resp, err := try.Response(req, 5*time.Second)
	require.NoError(s.T(), err)

	out := &TraceResponse{}
	content, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	err = json.Unmarshal(content, &out)
	require.NoError(s.T(), err)

	s.NotEmptyf(len(out.Traces), "expected at least one trace")

	var contents []string
	for _, t := range out.Traces {
		baseURL, err := url.Parse("http://" + s.tempoIP + ":3200/api/traces/" + t.TraceID)
		require.NoError(s.T(), err)

		req := &http.Request{
			Method: http.MethodGet,
			URL:    baseURL,
		}

		resp, err := try.Response(req, 5*time.Second)
		require.NoError(s.T(), err)

		content, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		contents = append(contents, string(content))
	}

	for _, expected := range expectedJSON {
		containsAll(expected, contents)
	}
}

func containsAll(expectedJSON map[string]string, contents []string) {
	for k, v := range expectedJSON {
		found := false
		for _, content := range contents {
			if gjson.Get(content, k).String() == v {
				found = true
				break
			}
		}

		if !found {
			log.Info().Msgf("[" + strings.Join(contents, ",") + "]")
			log.Error().Msgf("missing element: \nKey: %q\nValue: %q ", k, v)
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
