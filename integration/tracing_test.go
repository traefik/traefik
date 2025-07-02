package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
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

func (s *TracingSuite) TestOpenTelemetryBasic_HTTP_router_minimalVerbosity() {
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

	err = try.GetRequest("http://127.0.0.1:8000/basic-minimal", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.0.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/basic-minimal", net.JoinHostPort(s.whoamiIP, "80")),
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.port\").value.intValue":           "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.port\").value.intValue":                 "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue":   "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.1.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"url.path\").value.stringValue":               "/basic-minimal",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetryBasic_HTTP_entrypoint_minimalVerbosity() {
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

	err = try.GetRequest("http://127.0.0.1:8001/basic", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.0.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/basic", net.JoinHostPort(s.whoamiIP, "80")),
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.port\").value.intValue":           "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.port\").value.intValue":                 "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue":   "200",

			"batches.0.scopeSpans.0.spans.1.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.1.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"entry_point\").value.stringValue":            "web-minimal",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"url.path\").value.stringValue":               "/basic",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8001",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetryBasic_HTTP() {
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

			"batches.0.scopeSpans.0.spans.0.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.0.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/basic", net.JoinHostPort(s.whoamiIP, "80")),
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.port\").value.intValue":           "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.port\").value.intValue":                 "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue":   "200",

			"batches.0.scopeSpans.0.spans.1.name": "Metrics",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-service",

			"batches.0.scopeSpans.0.spans.2.name": "Service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.3.name": "Router",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router0@file",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.route\").value.stringValue":           "Path(`/basic`)",

			"batches.0.scopeSpans.0.spans.4.name": "Metrics",
			"batches.0.scopeSpans.0.spans.4.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",

			"batches.0.scopeSpans.0.spans.5.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.5.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"url.path\").value.stringValue":               "/basic",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetryBasic_gRPC() {
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

			"batches.0.scopeSpans.0.spans.0.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.0.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/basic", net.JoinHostPort(s.whoamiIP, "80")),
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.port\").value.intValue":           "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.port\").value.intValue":                 "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue":   "200",

			"batches.0.scopeSpans.0.spans.1.name": "Metrics",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-service",

			"batches.0.scopeSpans.0.spans.2.name": "Service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",

			"batches.0.scopeSpans.0.spans.3.name": "Router",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.service.name\").value.stringValue": "service0@file",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router0@file",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.route\").value.stringValue":           "Path(`/basic`)",

			"batches.0.scopeSpans.0.spans.4.name": "Metrics",
			"batches.0.scopeSpans.0.spans.4.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetryRateLimit() {
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

			"batches.0.scopeSpans.0.spans.0.name": "RateLimiter",
			"batches.0.scopeSpans.0.spans.0.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "ratelimit-1@file",

			"batches.0.scopeSpans.0.spans.1.name": "Router",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router1@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"http.route\").value.stringValue":           "Path(`/ratelimit`)",

			"batches.0.scopeSpans.0.spans.2.name": "Metrics",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",

			"batches.0.scopeSpans.0.spans.3.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.3.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.path\").value.stringValue":               "/ratelimit",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.response.status_code\").value.intValue": "429",
		},
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.0.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/ratelimit", net.JoinHostPort(s.whoamiIP, "80")),
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.port\").value.intValue":           "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.port\").value.intValue":                 "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue":   "200",

			"batches.0.scopeSpans.0.spans.1.name": "Metrics",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-service",

			"batches.0.scopeSpans.0.spans.2.name": "Service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",

			"batches.0.scopeSpans.0.spans.3.name": "RateLimiter",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "ratelimit-1@file",

			"batches.0.scopeSpans.0.spans.4.name": "Router",
			"batches.0.scopeSpans.0.spans.4.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.service.name\").value.stringValue": "service1@file",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router1@file",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.route\").value.stringValue":           "Path(`/ratelimit`)",

			"batches.0.scopeSpans.0.spans.5.name": "Metrics",
			"batches.0.scopeSpans.0.spans.5.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",

			"batches.0.scopeSpans.0.spans.6.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.6.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"url.path\").value.stringValue":               "/ratelimit",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetryRetry() {
	file := s.adaptFile("fixtures/tracing/simple-opentelemetry.toml", TracingTemplate{
		WhoamiIP:   s.whoamiIP,
		WhoamiPort: 81,
		IP:         s.otelCollectorIP,
	})
	defer os.Remove(file)

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("basic-auth"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/retry", 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.0.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/retry", net.JoinHostPort(s.whoamiIP, "81")),
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.port\").value.intValue":           "81",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.port\").value.intValue":                 "81",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue":   "502",
			"batches.0.scopeSpans.0.spans.0.status.code":                                                      "STATUS_CODE_ERROR",

			"batches.0.scopeSpans.0.spans.1.name": "Metrics",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-service",

			"batches.0.scopeSpans.0.spans.2.name": "Service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.3.name": "Retry",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.4.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.4.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/retry", net.JoinHostPort(s.whoamiIP, "81")),
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"network.peer.port\").value.intValue":           "81",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"server.port\").value.intValue":                 "81",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.response.status_code\").value.intValue":   "502",
			"batches.0.scopeSpans.0.spans.4.status.code":                                                      "STATUS_CODE_ERROR",

			"batches.0.scopeSpans.0.spans.5.name": "Metrics",
			"batches.0.scopeSpans.0.spans.5.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-service",

			"batches.0.scopeSpans.0.spans.6.name": "Service",
			"batches.0.scopeSpans.0.spans.6.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.7.name": "Retry",
			"batches.0.scopeSpans.0.spans.7.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.7.attributes.#(key=\"http.request.resend_count\").value.intValue":  "1",

			"batches.0.scopeSpans.0.spans.8.name":                                                             "ReverseProxy",
			"batches.0.scopeSpans.0.spans.8.kind":                                                             "SPAN_KIND_CLIENT",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"http.request.method\").value.stringValue":      "GET",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"network.protocol.version\").value.stringValue": "1.1",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"url.full\").value.stringValue":                 fmt.Sprintf("http://%s/retry", net.JoinHostPort(s.whoamiIP, "81")),
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"user_agent.original\").value.stringValue":      "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"network.peer.address\").value.stringValue":     s.whoamiIP,
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"network.peer.port\").value.intValue":           "81",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"server.address\").value.stringValue":           s.whoamiIP,
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"server.port\").value.intValue":                 "81",
			"batches.0.scopeSpans.0.spans.8.attributes.#(key=\"http.response.status_code\").value.intValue":   "502",
			"batches.0.scopeSpans.0.spans.8.status.code":                                                      "STATUS_CODE_ERROR",

			"batches.0.scopeSpans.0.spans.9.name": "Metrics",
			"batches.0.scopeSpans.0.spans.9.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.9.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-service",

			"batches.0.scopeSpans.0.spans.10.name":                                                         "Service",
			"batches.0.scopeSpans.0.spans.10.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.10.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",

			"batches.0.scopeSpans.0.spans.11.name": "Retry",
			"batches.0.scopeSpans.0.spans.11.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.11.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",
			"batches.0.scopeSpans.0.spans.11.attributes.#(key=\"http.request.resend_count\").value.intValue":  "2",

			"batches.0.scopeSpans.0.spans.12.name":                                                         "Router",
			"batches.0.scopeSpans.0.spans.12.kind":                                                         "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.12.attributes.#(key=\"traefik.service.name\").value.stringValue": "service2@file",
			"batches.0.scopeSpans.0.spans.12.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router2@file",

			"batches.0.scopeSpans.0.spans.13.name": "Metrics",
			"batches.0.scopeSpans.0.spans.13.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.13.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",

			"batches.0.scopeSpans.0.spans.14.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.14.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"url.path\").value.stringValue":               "/retry",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.14.attributes.#(key=\"http.response.status_code\").value.intValue": "502",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetryAuth() {
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

			"batches.0.scopeSpans.0.spans.0.name": "BasicAuth",
			"batches.0.scopeSpans.0.spans.0.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "basic-auth@file",
			"batches.0.scopeSpans.0.spans.0.status.message":                                                  "Authentication failed",
			"batches.0.scopeSpans.0.spans.0.status.code":                                                     "STATUS_CODE_ERROR",

			"batches.0.scopeSpans.0.spans.1.name": "Router",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.service.name\").value.stringValue": "service3@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router3@file",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"http.route\").value.stringValue":           "Path(`/auth`)",

			"batches.0.scopeSpans.0.spans.2.name": "Metrics",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",

			"batches.0.scopeSpans.0.spans.3.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.3.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.path\").value.stringValue":               "/auth",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"http.response.status_code\").value.intValue": "401",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetryAuthWithRetry() {
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

	err = try.GetRequest("http://127.0.0.1:8000/retry-auth", 500*time.Millisecond, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name": "BasicAuth",
			"batches.0.scopeSpans.0.spans.0.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "basic-auth@file",
			"batches.0.scopeSpans.0.spans.0.status.message":                                                  "Authentication failed",
			"batches.0.scopeSpans.0.spans.0.status.code":                                                     "STATUS_CODE_ERROR",

			"batches.0.scopeSpans.0.spans.1.name": "Retry",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "retry@file",

			"batches.0.scopeSpans.0.spans.2.name": "Router",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service4@file",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router4@file",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"http.route\").value.stringValue":           "Path(`/retry-auth`)",

			"batches.0.scopeSpans.0.spans.3.name": "Metrics",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",

			"batches.0.scopeSpans.0.spans.4.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.4.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"url.path\").value.stringValue":               "/retry-auth",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"url.query\").value.stringValue":              "",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.response.status_code\").value.intValue": "401",
		},
	}

	s.checkTraceContent(contains)
}

func (s *TracingSuite) TestOpenTelemetrySafeURL() {
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

	err = try.GetRequest("http://test:test@127.0.0.1:8000/auth?api_key=powpow", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	contains := []map[string]string{
		{
			"batches.0.scopeSpans.0.scope.name": "github.com/traefik/traefik",

			"batches.0.scopeSpans.0.spans.0.name":                                                           "ReverseProxy",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"url.full\").value.stringValue":               fmt.Sprintf("http://REDACTED:REDACTED@%s/auth?api_key=REDACTED", net.JoinHostPort(s.whoamiIP, "80")),
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.address\").value.stringValue":   s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"network.peer.port\").value.intValue":         "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.address\").value.stringValue":         s.whoamiIP,
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"server.port\").value.intValue":               "80",
			"batches.0.scopeSpans.0.spans.0.attributes.#(key=\"http.response.status_code\").value.intValue": "200",

			"batches.0.scopeSpans.0.spans.1.name": "Metrics",
			"batches.0.scopeSpans.0.spans.1.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.1.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-service",

			"batches.0.scopeSpans.0.spans.2.name": "Service",
			"batches.0.scopeSpans.0.spans.2.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.2.attributes.#(key=\"traefik.service.name\").value.stringValue": "service3@file",

			"batches.0.scopeSpans.0.spans.3.name": "BasicAuth",
			"batches.0.scopeSpans.0.spans.3.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.3.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "basic-auth@file",

			"batches.0.scopeSpans.0.spans.4.name": "Router",
			"batches.0.scopeSpans.0.spans.4.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.service.name\").value.stringValue": "service3@file",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"traefik.router.name\").value.stringValue":  "web-router3@file",
			"batches.0.scopeSpans.0.spans.4.attributes.#(key=\"http.route\").value.stringValue":           "Path(`/auth`)",

			"batches.0.scopeSpans.0.spans.5.name": "Metrics",
			"batches.0.scopeSpans.0.spans.5.kind": "SPAN_KIND_INTERNAL",
			"batches.0.scopeSpans.0.spans.5.attributes.#(key=\"traefik.middleware.name\").value.stringValue": "metrics-entrypoint",

			"batches.0.scopeSpans.0.spans.6.name":                                                           "EntryPoint",
			"batches.0.scopeSpans.0.spans.6.kind":                                                           "SPAN_KIND_SERVER",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"entry_point\").value.stringValue":            "web",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"http.request.method\").value.stringValue":    "GET",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"url.path\").value.stringValue":               "/auth",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"url.query\").value.stringValue":              "api_key=REDACTED",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"user_agent.original\").value.stringValue":    "Go-http-client/1.1",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"server.address\").value.stringValue":         "127.0.0.1:8000",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"network.peer.address\").value.stringValue":   "127.0.0.1",
			"batches.0.scopeSpans.0.spans.6.attributes.#(key=\"http.response.status_code\").value.intValue": "200",
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
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("basic-auth"))
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

	s.NotEmptyf(out.Traces, "expected at least one trace")

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

	s.NotEmptyf(out.Traces, "expected at least one trace")

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

	var missingElements []string
	for _, expected := range expectedJSON {
		missingElements = append(missingElements, contains(expected, contents)...)
	}

	log.Printf("Contents: [%s]\n", strings.Join(contents, ","))
	for _, element := range missingElements {
		log.Printf("Missing elements:\n%s\n", element)
	}

	require.Empty(s.T(), missingElements)
}

func contains(expectedJSON map[string]string, contents []string) []string {
	var missingElements []string

	for k, v := range expectedJSON {
		found := false
		for _, content := range contents {
			if gjson.Get(content, k).String() == v {
				found = true
				break
			}
		}
		if !found {
			missingElements = append(missingElements, "Key: "+k+", Value: "+v)
		}
	}

	return missingElements
}

// TraceResponse contains a list of traces.
type TraceResponse struct {
	Traces []Trace `json:"traces"`
}

// Trace represents a simplified grafana tempo trace.
type Trace struct {
	TraceID string `json:"traceID"`
}
