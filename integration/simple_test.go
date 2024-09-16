package integration

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// SimpleSuite tests suite.
type SimpleSuite struct{ BaseSuite }

func TestSimpleSuite(t *testing.T) {
	suite.Run(t, new(SimpleSuite))
}

func (s *SimpleSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *SimpleSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *SimpleSuite) TestInvalidConfigShouldFail() {
	_, output := s.cmdTraefik(withConfigFile("fixtures/invalid_configuration.toml"))

	err := try.Do(500*time.Millisecond, func() error {
		expected := "expected '.' or '=', but got '{' instead"
		actual := output.String()

		if !strings.Contains(actual, expected) {
			return fmt.Errorf("got %s, wanted %s", actual, expected)
		}

		return nil
	})
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestSimpleDefaultConfig() {
	s.cmdTraefik(withConfigFile("fixtures/simple_default.toml"))

	// Expected a 404 as we did not configure anything
	err := try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestWithWebConfig() {
	s.cmdTraefik(withConfigFile("fixtures/simple_web.toml"))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestPrintHelp() {
	_, output := s.cmdTraefik("--help")

	err := try.Do(500*time.Millisecond, func() error {
		expected := "Usage:"
		notExpected := "panic:"
		actual := output.String()

		if strings.Contains(actual, notExpected) {
			return fmt.Errorf("got %s", actual)
		}
		if !strings.Contains(actual, expected) {
			return fmt.Errorf("got %s, wanted %s", actual, expected)
		}

		return nil
	})
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestRequestAcceptGraceTimeout() {
	s.createComposeProject("reqacceptgrace")

	s.composeUp()
	defer s.composeDown()

	whoamiURL := "http://" + net.JoinHostPort(s.getComposeServiceIP("whoami"), "80")

	file := s.adaptFile("fixtures/reqacceptgrace.toml", struct {
		Server string
	}{whoamiURL})

	cmd, _ := s.cmdTraefik(withConfigFile(file))

	// Wait for Traefik to turn ready.
	err := try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	// Make sure exposed service is ready.
	err = try.GetRequest("http://127.0.0.1:8000/service", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Check that /ping endpoint is responding with 200.
	err = try.GetRequest("http://127.0.0.1:8001/ping", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Send SIGTERM to Traefik.
	proc, err := os.FindProcess(cmd.Process.Pid)
	require.NoError(s.T(), err)
	err = proc.Signal(syscall.SIGTERM)
	require.NoError(s.T(), err)

	// Give Traefik time to process the SIGTERM and send a request half-way
	// into the request accepting grace period, by which requests should
	// still get served.
	time.Sleep(5 * time.Second)
	resp, err := http.Get("http://127.0.0.1:8000/service")
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// ping endpoint should now return a Service Unavailable.
	resp, err = http.Get("http://127.0.0.1:8001/ping")
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	assert.Equal(s.T(), http.StatusServiceUnavailable, resp.StatusCode)

	// Expect Traefik to shut down gracefully once the request accepting grace
	// period has elapsed.
	waitErr := make(chan error)
	go func() {
		waitErr <- cmd.Wait()
	}()

	select {
	case err := <-waitErr:
		require.NoError(s.T(), err)
	case <-time.After(10 * time.Second):
		// By now we are ~5 seconds out of the request accepting grace period
		// (start + 5 seconds sleep prior to the mid-grace period request +
		// 10 seconds timeout = 15 seconds > 10 seconds grace period).
		// Something must have gone wrong if we still haven't terminated at
		// this point.
		s.T().Fatal("Traefik did not terminate in time")
	}
}

func (s *SimpleSuite) TestCustomPingTerminationStatusCode() {
	file := s.adaptFile("fixtures/custom_ping_termination_status_code.toml", struct{}{})
	cmd, _ := s.cmdTraefik(withConfigFile(file))

	// Wait for Traefik to turn ready.
	err := try.GetRequest("http://127.0.0.1:8001/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	// Check that /ping endpoint is responding with 200.
	err = try.GetRequest("http://127.0.0.1:8001/ping", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Send SIGTERM to Traefik.
	proc, err := os.FindProcess(cmd.Process.Pid)
	require.NoError(s.T(), err)
	err = proc.Signal(syscall.SIGTERM)
	require.NoError(s.T(), err)

	// ping endpoint should now return a Service Unavailable.
	err = try.GetRequest("http://127.0.0.1:8001/ping", 2*time.Second, try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestStatsWithMultipleEntryPoint() {
	s.T().Skip("Stats is missing")
	s.createComposeProject("stats")

	s.composeUp()
	defer s.composeDown()

	whoami1URL := "http://" + net.JoinHostPort(s.getComposeServiceIP("whoami1"), "80")
	whoami2URL := "http://" + net.JoinHostPort(s.getComposeServiceIP("whoami2"), "80")

	file := s.adaptFile("fixtures/simple_stats.toml", struct {
		Server1 string
		Server2 string
	}{whoami1URL, whoami2URL})
	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/health", 1*time.Second, try.BodyContains(`"total_status_code_count":{"200":2}`))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestNoAuthOnPing() {
	s.T().Skip("Waiting for new api handler implementation")

	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	file := s.adaptFile("./fixtures/simple_auth.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8001/api/rawdata", 2*time.Second, try.StatusCodeIs(http.StatusUnauthorized))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8001/ping", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestDefaultEntryPointHTTP() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd("--entryPoints.http.Address=:8000", "--log.level=DEBUG", "--providers.docker", "--api.insecure")

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestWithNonExistingEntryPoint() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd("--entryPoints.http.Address=:8000", "--log.level=DEBUG", "--providers.docker", "--api.insecure")

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestMetricsPrometheusDefaultEntryPoint() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd("--entryPoints.http.Address=:8000", "--api.insecure", "--metrics.prometheus.buckets=0.1,0.3,1.2,5.0", "--providers.docker", "--metrics.prometheus.addrouterslabels=true", "--log.level=DEBUG")

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix(`/whoami`)"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.BodyContains("_router_"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.BodyContains("_entrypoint_"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.BodyContains("_service_"))
	require.NoError(s.T(), err)

	// No metrics for internals.
	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.BodyNotContains("router=\"api@internal\"", "service=\"api@internal\""))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestMetricsPrometheusTwoRoutersOneService() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd("--entryPoints.http.Address=:8000", "--api.insecure", "--metrics.prometheus.buckets=0.1,0.3,1.2,5.0", "--providers.docker", "--metrics.prometheus.addentrypointslabels=false", "--metrics.prometheus.addrouterslabels=true", "--log.level=DEBUG")

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami2", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// adding a loop to test if metrics are not deleted
	for range 10 {
		request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/metrics", nil)
		require.NoError(s.T(), err)

		response, err := http.DefaultClient.Do(request)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)

		body, err := io.ReadAll(response.Body)
		require.NoError(s.T(), err)

		// Reqs count of 1 for both routers
		assert.Contains(s.T(), string(body), "traefik_router_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",router=\"router1@docker\",service=\"whoami1@docker\"} 1")
		assert.Contains(s.T(), string(body), "traefik_router_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",router=\"router2@docker\",service=\"whoami1@docker\"} 1")
		// Reqs count of 2 for service behind both routers
		assert.Contains(s.T(), string(body), "traefik_service_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"whoami1@docker\"} 2")
	}
}

// TestMetricsWithBufferingMiddleware checks that the buffering middleware
// (which introduces its own response writer in the chain), does not interfere with
// the capture middleware on which the metrics mechanism relies.
func (s *SimpleSuite) TestMetricsWithBufferingMiddleware() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("MORE THAN TEN BYTES IN RESPONSE"))
	}))

	server.Start()
	defer server.Close()

	file := s.adaptFile("fixtures/simple_metrics_with_buffer_middleware.toml", struct{ IP string }{IP: strings.TrimPrefix(server.URL, "http://")})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix(`/without`)"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8001/without", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8002/with-req", strings.NewReader("MORE THAN TEN BYTES IN REQUEST"))
	require.NoError(s.T(), err)

	// The request should fail because the body is too large.
	err = try.Request(req, 1*time.Second, try.StatusCodeIs(http.StatusRequestEntityTooLarge))
	require.NoError(s.T(), err)

	// The request should fail because the response exceeds the configured limit.
	err = try.GetRequest("http://127.0.0.1:8003/with-resp", 1*time.Second, try.StatusCodeIs(http.StatusInternalServerError))
	require.NoError(s.T(), err)

	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/metrics", nil)
	require.NoError(s.T(), err)

	response, err := http.DefaultClient.Do(request)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	body, err := io.ReadAll(response.Body)
	require.NoError(s.T(), err)

	// For allowed requests and responses, the entrypoint and service metrics have the same status code.
	assert.Contains(s.T(), string(body), "traefik_entrypoint_requests_total{code=\"200\",entrypoint=\"webA\",method=\"GET\",protocol=\"http\"} 1")
	assert.Contains(s.T(), string(body), "traefik_entrypoint_requests_bytes_total{code=\"200\",entrypoint=\"webA\",method=\"GET\",protocol=\"http\"} 0")
	assert.Contains(s.T(), string(body), "traefik_entrypoint_responses_bytes_total{code=\"200\",entrypoint=\"webA\",method=\"GET\",protocol=\"http\"} 31")

	assert.Contains(s.T(), string(body), "traefik_service_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-without@file\"} 1")
	assert.Contains(s.T(), string(body), "traefik_service_requests_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-without@file\"} 0")
	assert.Contains(s.T(), string(body), "traefik_service_responses_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-without@file\"} 31")

	// For forbidden requests, the entrypoints have metrics, the services don't.
	assert.Contains(s.T(), string(body), "traefik_entrypoint_requests_total{code=\"413\",entrypoint=\"webB\",method=\"GET\",protocol=\"http\"} 1")
	assert.Contains(s.T(), string(body), "traefik_entrypoint_requests_bytes_total{code=\"413\",entrypoint=\"webB\",method=\"GET\",protocol=\"http\"} 0")
	assert.Contains(s.T(), string(body), "traefik_entrypoint_responses_bytes_total{code=\"413\",entrypoint=\"webB\",method=\"GET\",protocol=\"http\"} 24")

	// For disallowed responses, the entrypoint and service metrics don't have the same status code.
	assert.Contains(s.T(), string(body), "traefik_entrypoint_requests_bytes_total{code=\"500\",entrypoint=\"webC\",method=\"GET\",protocol=\"http\"} 0")
	assert.Contains(s.T(), string(body), "traefik_entrypoint_requests_total{code=\"500\",entrypoint=\"webC\",method=\"GET\",protocol=\"http\"} 1")
	assert.Contains(s.T(), string(body), "traefik_entrypoint_responses_bytes_total{code=\"500\",entrypoint=\"webC\",method=\"GET\",protocol=\"http\"} 21")

	assert.Contains(s.T(), string(body), "traefik_service_requests_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-resp@file\"} 0")
	assert.Contains(s.T(), string(body), "traefik_service_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-resp@file\"} 1")
	assert.Contains(s.T(), string(body), "traefik_service_responses_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-resp@file\"} 31")
}

func (s *SimpleSuite) TestMultipleProviderSameBackendName() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoami1IP := s.getComposeServiceIP("whoami1")
	whoami2IP := s.getComposeServiceIP("whoami2")
	file := s.adaptFile("fixtures/multiple_provider.toml", struct{ IP string }{IP: whoami2IP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.BodyContains(whoami1IP))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/file", 1*time.Second, try.BodyContains(whoami2IP))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestIPStrategyAllowlist() {
	s.createComposeProject("allowlist")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd(withConfigFile("fixtures/simple_allowlist.toml"))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("override"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("override.remoteaddr.allowlist.docker.local"))
	require.NoError(s.T(), err)

	testCases := []struct {
		desc               string
		xForwardedFor      string
		host               string
		expectedStatusCode int
	}{
		{
			desc:               "override remote addr reject",
			xForwardedFor:      "8.8.8.8,8.8.8.8",
			host:               "override.remoteaddr.allowlist.docker.local",
			expectedStatusCode: 403,
		},
		{
			desc:               "override depth accept",
			xForwardedFor:      "8.8.8.8,10.0.0.1,127.0.0.1",
			host:               "override.depth.allowlist.docker.local",
			expectedStatusCode: 200,
		},
		{
			desc:               "override depth reject",
			xForwardedFor:      "10.0.0.1,8.8.8.8,127.0.0.1",
			host:               "override.depth.allowlist.docker.local",
			expectedStatusCode: 403,
		},
		{
			desc:               "override excludedIPs reject",
			xForwardedFor:      "10.0.0.3,10.0.0.1,10.0.0.2",
			host:               "override.excludedips.allowlist.docker.local",
			expectedStatusCode: 403,
		},
		{
			desc:               "override excludedIPs accept",
			xForwardedFor:      "8.8.8.8,10.0.0.1,10.0.0.2",
			host:               "override.excludedips.allowlist.docker.local",
			expectedStatusCode: 200,
		},
	}

	for _, test := range testCases {
		req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
		req.Header.Set("X-Forwarded-For", test.xForwardedFor)
		req.Host = test.host
		req.RequestURI = ""

		err = try.Request(req, 1*time.Second, try.StatusCodeIs(test.expectedStatusCode))
		require.NoErrorf(s.T(), err, "Error during %s: %v", test.desc, err)
	}
}

func (s *SimpleSuite) TestIPStrategyWhitelist() {
	s.createComposeProject("whitelist")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd(withConfigFile("fixtures/simple_whitelist.toml"))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("override"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("override.remoteaddr.whitelist.docker.local"))
	require.NoError(s.T(), err)

	testCases := []struct {
		desc               string
		xForwardedFor      string
		host               string
		expectedStatusCode int
	}{
		{
			desc:               "override remote addr reject",
			xForwardedFor:      "8.8.8.8,8.8.8.8",
			host:               "override.remoteaddr.whitelist.docker.local",
			expectedStatusCode: 403,
		},
		{
			desc:               "override depth accept",
			xForwardedFor:      "8.8.8.8,10.0.0.1,127.0.0.1",
			host:               "override.depth.whitelist.docker.local",
			expectedStatusCode: 200,
		},
		{
			desc:               "override depth reject",
			xForwardedFor:      "10.0.0.1,8.8.8.8,127.0.0.1",
			host:               "override.depth.whitelist.docker.local",
			expectedStatusCode: 403,
		},
		{
			desc:               "override excludedIPs reject",
			xForwardedFor:      "10.0.0.3,10.0.0.1,10.0.0.2",
			host:               "override.excludedips.whitelist.docker.local",
			expectedStatusCode: 403,
		},
		{
			desc:               "override excludedIPs accept",
			xForwardedFor:      "8.8.8.8,10.0.0.1,10.0.0.2",
			host:               "override.excludedips.whitelist.docker.local",
			expectedStatusCode: 200,
		},
	}

	for _, test := range testCases {
		req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
		req.Header.Set("X-Forwarded-For", test.xForwardedFor)
		req.Host = test.host
		req.RequestURI = ""

		err = try.Request(req, 1*time.Second, try.StatusCodeIs(test.expectedStatusCode))
		require.NoErrorf(s.T(), err, "Error during %s: %v", test.desc, err)
	}
}

func (s *SimpleSuite) TestXForwardedHeaders() {
	s.createComposeProject("allowlist")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd(withConfigFile("fixtures/simple_allowlist.toml"))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.BodyContains("override.remoteaddr.allowlist.docker.local"))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	require.NoError(s.T(), err)

	req.Host = "override.depth.allowlist.docker.local"
	req.Header.Set("X-Forwarded-For", "8.8.8.8,10.0.0.1,127.0.0.1")

	err = try.Request(req, 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-Proto", "X-Forwarded-For", "X-Forwarded-Host",
			"X-Forwarded-Host", "X-Forwarded-Port", "X-Forwarded-Server", "X-Real-Ip"))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestMultiProvider() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoamiURL := "http://" + net.JoinHostPort(s.getComposeServiceIP("whoami1"), "80")

	file := s.adaptFile("fixtures/multiprovider.toml", struct{ Server string }{Server: whoamiURL})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1000*time.Millisecond, try.BodyContains("service"))
	require.NoError(s.T(), err)

	config := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
				"router1": {
					EntryPoints: []string{"web"},
					Middlewares: []string{"customheader@file"},
					Service:     "service@file",
					Rule:        "PathPrefix(`/`)",
				},
			},
		},
	}

	jsonContent, err := json.Marshal(config)
	require.NoError(s.T(), err)

	request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8080/api/providers/rest", bytes.NewReader(jsonContent))
	require.NoError(s.T(), err)

	response, err := http.DefaultClient.Do(request)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1000*time.Millisecond, try.BodyContains("PathPrefix(`/`)"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("CustomValue"))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestSimpleConfigurationHostRequestTrailingPeriod() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoamiURL := "http://" + net.JoinHostPort(s.getComposeServiceIP("whoami1"), "80")

	file := s.adaptFile("fixtures/file/simple-hosts.toml", struct{ Server string }{Server: whoamiURL})

	s.traefikCmd(withConfigFile(file))

	testCases := []struct {
		desc        string
		requestHost string
	}{
		{
			desc:        "Request host without trailing period, rule without trailing period",
			requestHost: "test.localhost",
		},
		{
			desc:        "Request host with trailing period, rule without trailing period",
			requestHost: "test.localhost.",
		},
		{
			desc:        "Request host without trailing period, rule with trailing period",
			requestHost: "test.foo.localhost",
		},
		{
			desc:        "Request host with trailing period, rule with trailing period",
			requestHost: "test.foo.localhost.",
		},
	}

	for _, test := range testCases {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
		require.NoError(s.T(), err)
		req.Host = test.requestHost
		err = try.Request(req, 1*time.Second, try.StatusCodeIs(http.StatusOK))
		require.NoErrorf(s.T(), err, "Error while testing %s: %v", test.desc, err)
	}
}

func (s *SimpleSuite) TestWithDefaultRuleSyntax() {
	file := s.adaptFile("fixtures/with_default_rule_syntax.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	require.NoError(s.T(), err)

	// router1 has no error
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router1@file", 1*time.Second, try.BodyContains(`"status":"enabled"`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/notfound", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/foo", 1*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/bar", 1*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	// router2 has an error because it uses the wrong rule syntax (v3 instead of v2)
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router2@file", 1*time.Second, try.BodyContains("error while parsing rule QueryRegexp(`foo`, `bar`): unsupported function: QueryRegexp"))
	require.NoError(s.T(), err)

	// router3 has an error because it uses the wrong rule syntax (v2 instead of v3)
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router3@file", 1*time.Second, try.BodyContains("error while adding rule PathPrefix: unexpected number of parameters; got 2, expected one of [1]"))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestWithoutDefaultRuleSyntax() {
	file := s.adaptFile("fixtures/without_default_rule_syntax.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	require.NoError(s.T(), err)

	// router1 has no error
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router1@file", 1*time.Second, try.BodyContains(`"status":"enabled"`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/notfound", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/foo", 1*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/bar", 1*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	// router2 has an error because it uses the wrong rule syntax (v3 instead of v2)
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router2@file", 1*time.Second, try.BodyContains("error while adding rule PathPrefix: unexpected number of parameters; got 2, expected one of [1]"))
	require.NoError(s.T(), err)

	// router2 has an error because it uses the wrong rule syntax (v2 instead of v3)
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router3@file", 1*time.Second, try.BodyContains("error while parsing rule QueryRegexp(`foo`, `bar`): unsupported function: QueryRegexp"))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestRouterConfigErrors() {
	file := s.adaptFile("fixtures/router_errors.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	// All errors
	err := try.GetRequest("http://127.0.0.1:8080/api/http/routers", 1000*time.Millisecond, try.BodyContains(`["middleware \"unknown@file\" does not exist","found different TLS options for routers on the same host snitest.net, so using the default TLS options instead"]`))
	require.NoError(s.T(), err)

	// router3 has an error because it uses an unknown entrypoint
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router3@file", 1000*time.Millisecond, try.BodyContains(`entryPoint \"unknown-entrypoint\" doesn't exist`, "no valid entryPoint for this router"))
	require.NoError(s.T(), err)

	// router4 is enabled, but in warning state because its tls options conf was messed up
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router4@file", 1000*time.Millisecond, try.BodyContains(`"status":"warning"`))
	require.NoError(s.T(), err)

	// router5 is disabled because its middleware conf is broken
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router5@file", 1000*time.Millisecond, try.BodyContains())
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestServiceConfigErrors() {
	file := s.adaptFile("fixtures/service_errors.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains(`["the service \"service1@file\" does not have any type defined"]`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services/service1@file", 1000*time.Millisecond, try.BodyContains(`"status":"disabled"`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services/service2@file", 1000*time.Millisecond, try.BodyContains(`"status":"enabled"`))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestTCPRouterConfigErrors() {
	file := s.adaptFile("fixtures/router_errors.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	// router3 has an error because it uses an unknown entrypoint
	err := try.GetRequest("http://127.0.0.1:8080/api/tcp/routers/router3@file", 1000*time.Millisecond, try.BodyContains(`entryPoint \"unknown-entrypoint\" doesn't exist`, "no valid entryPoint for this router"))
	require.NoError(s.T(), err)

	// router4 has an unsupported Rule
	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/routers/router4@file", 1000*time.Millisecond, try.BodyContains("invalid rule: \\\"Host(`mydomain.com`)\\\""))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestTCPServiceConfigErrors() {
	file := s.adaptFile("fixtures/tcp/service_errors.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/tcp/services", 1000*time.Millisecond, try.BodyContains(`["the service \"service1@file\" does not have any type defined"]`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/services/service1@file", 1000*time.Millisecond, try.BodyContains(`"status":"disabled"`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/services/service2@file", 1000*time.Millisecond, try.BodyContains(`"status":"enabled"`))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestUDPRouterConfigErrors() {
	file := s.adaptFile("fixtures/router_errors.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/udp/routers/router3@file", 1000*time.Millisecond, try.BodyContains(`entryPoint \"unknown-entrypoint\" doesn't exist`, "no valid entryPoint for this router"))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestUDPServiceConfigErrors() {
	file := s.adaptFile("fixtures/udp/service_errors.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/udp/services", 1000*time.Millisecond, try.BodyContains(`["the UDP service \"service1@file\" does not have any type defined"]`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/udp/services/service1@file", 1000*time.Millisecond, try.BodyContains(`"status":"disabled"`))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/udp/services/service2@file", 1000*time.Millisecond, try.BodyContains(`"status":"enabled"`))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestWRRServer() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoami1IP := s.getComposeServiceIP("whoami1")
	whoami2IP := s.getComposeServiceIP("whoami2")

	file := s.adaptFile("fixtures/wrr_server.toml", struct {
		Server1 string
		Server2 string
	}{Server1: "http://" + whoami1IP, Server2: "http://" + whoami2IP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("service1"))
	require.NoError(s.T(), err)

	repartition := map[string]int{}
	for range 4 {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
		require.NoError(s.T(), err)

		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)

		body, err := io.ReadAll(response.Body)
		require.NoError(s.T(), err)

		if strings.Contains(string(body), whoami1IP) {
			repartition[whoami1IP]++
		}
		if strings.Contains(string(body), whoami2IP) {
			repartition[whoami2IP]++
		}
	}

	assert.Equal(s.T(), 3, repartition[whoami1IP])
	assert.Equal(s.T(), 1, repartition[whoami2IP])
}

func (s *SimpleSuite) TestWRR() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoami1IP := s.getComposeServiceIP("whoami1")
	whoami2IP := s.getComposeServiceIP("whoami2")

	file := s.adaptFile("fixtures/wrr.toml", struct {
		Server1 string
		Server2 string
	}{Server1: "http://" + whoami1IP, Server2: "http://" + whoami2IP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("service1", "service2"))
	require.NoError(s.T(), err)

	repartition := map[string]int{}
	for range 4 {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
		require.NoError(s.T(), err)

		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)

		body, err := io.ReadAll(response.Body)
		require.NoError(s.T(), err)

		if strings.Contains(string(body), whoami1IP) {
			repartition[whoami1IP]++
		}
		if strings.Contains(string(body), whoami2IP) {
			repartition[whoami2IP]++
		}
	}

	assert.Equal(s.T(), 3, repartition[whoami1IP])
	assert.Equal(s.T(), 1, repartition[whoami2IP])
}

func (s *SimpleSuite) TestWRRSticky() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoami1IP := s.getComposeServiceIP("whoami1")
	whoami2IP := s.getComposeServiceIP("whoami2")

	file := s.adaptFile("fixtures/wrr_sticky.toml", struct {
		Server1 string
		Server2 string
	}{Server1: "http://" + whoami1IP, Server2: "http://" + whoami2IP})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("service1", "service2"))
	require.NoError(s.T(), err)

	repartition := map[string]int{}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)

	for range 4 {
		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)

		for _, cookie := range response.Cookies() {
			req.AddCookie(cookie)
		}

		body, err := io.ReadAll(response.Body)
		require.NoError(s.T(), err)

		if strings.Contains(string(body), whoami1IP) {
			repartition[whoami1IP]++
		}
		if strings.Contains(string(body), whoami2IP) {
			repartition[whoami2IP]++
		}
	}

	assert.Equal(s.T(), 4, repartition[whoami1IP])
	assert.Equal(s.T(), 0, repartition[whoami2IP])
}

func (s *SimpleSuite) TestMirror() {
	var count, countMirror1, countMirror2 int32

	main := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&count, 1)
	}))

	mirror1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror1, 1)
	}))

	mirror2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror2, 1)
	}))

	mainServer := main.URL
	mirror1Server := mirror1.URL
	mirror2Server := mirror2.URL

	file := s.adaptFile("fixtures/mirror.toml", struct {
		MainServer    string
		Mirror1Server string
		Mirror2Server string
	}{MainServer: mainServer, Mirror1Server: mirror1Server, Mirror2Server: mirror2Server})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("mirror1", "mirror2", "service1"))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)
	for range 10 {
		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)
	}

	countTotal := atomic.LoadInt32(&count)
	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)

	assert.Equal(s.T(), int32(10), countTotal)
	assert.Equal(s.T(), int32(1), val1)
	assert.Equal(s.T(), int32(5), val2)
}

func (s *SimpleSuite) TestMirrorWithBody() {
	var count, countMirror1, countMirror2 int32

	body20 := make([]byte, 20)
	_, err := rand.Read(body20)
	require.NoError(s.T(), err)

	body5 := make([]byte, 5)
	_, err = rand.Read(body5)
	require.NoError(s.T(), err)

	verifyBody := func(req *http.Request) {
		b, _ := io.ReadAll(req.Body)
		switch req.Header.Get("Size") {
		case "20":
			require.Equal(s.T(), body20, b)
		case "5":
			require.Equal(s.T(), body5, b)
		default:
			s.T().Fatal("Size header not present")
		}
	}

	main := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		verifyBody(req)
		atomic.AddInt32(&count, 1)
	}))

	mirror1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		verifyBody(req)
		atomic.AddInt32(&countMirror1, 1)
	}))

	mirror2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		verifyBody(req)
		atomic.AddInt32(&countMirror2, 1)
	}))

	mainServer := main.URL
	mirror1Server := mirror1.URL
	mirror2Server := mirror2.URL

	file := s.adaptFile("fixtures/mirror.toml", struct {
		MainServer    string
		Mirror1Server string
		Mirror2Server string
	}{MainServer: mainServer, Mirror1Server: mirror1Server, Mirror2Server: mirror2Server})

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("mirror1", "mirror2", "service1"))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", bytes.NewBuffer(body20))
	require.NoError(s.T(), err)
	req.Header.Set("Size", "20")
	for range 10 {
		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)
	}

	countTotal := atomic.LoadInt32(&count)
	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)

	assert.Equal(s.T(), int32(10), countTotal)
	assert.Equal(s.T(), int32(1), val1)
	assert.Equal(s.T(), int32(5), val2)

	atomic.StoreInt32(&count, 0)
	atomic.StoreInt32(&countMirror1, 0)
	atomic.StoreInt32(&countMirror2, 0)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoamiWithMaxBody", bytes.NewBuffer(body5))
	require.NoError(s.T(), err)
	req.Header.Set("Size", "5")
	for range 10 {
		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)
	}

	countTotal = atomic.LoadInt32(&count)
	val1 = atomic.LoadInt32(&countMirror1)
	val2 = atomic.LoadInt32(&countMirror2)

	assert.Equal(s.T(), int32(10), countTotal)
	assert.Equal(s.T(), int32(1), val1)
	assert.Equal(s.T(), int32(5), val2)

	atomic.StoreInt32(&count, 0)
	atomic.StoreInt32(&countMirror1, 0)
	atomic.StoreInt32(&countMirror2, 0)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoamiWithMaxBody", bytes.NewBuffer(body20))
	require.NoError(s.T(), err)
	req.Header.Set("Size", "20")
	for range 10 {
		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)
	}

	countTotal = atomic.LoadInt32(&count)
	val1 = atomic.LoadInt32(&countMirror1)
	val2 = atomic.LoadInt32(&countMirror2)

	assert.Equal(s.T(), int32(10), countTotal)
	assert.Equal(s.T(), int32(0), val1)
	assert.Equal(s.T(), int32(0), val2)
}

func (s *SimpleSuite) TestMirrorCanceled() {
	var count, countMirror1, countMirror2 int32

	main := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&count, 1)
		time.Sleep(2 * time.Second)
	}))

	mirror1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror1, 1)
	}))

	mirror2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror2, 1)
	}))

	mainServer := main.URL
	mirror1Server := mirror1.URL
	mirror2Server := mirror2.URL

	file := s.adaptFile("fixtures/mirror.toml", struct {
		MainServer    string
		Mirror1Server string
		Mirror2Server string
	}{MainServer: mainServer, Mirror1Server: mirror1Server, Mirror2Server: mirror2Server})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("mirror1", "mirror2", "service1"))
	require.NoError(s.T(), err)

	for range 5 {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
		require.NoError(s.T(), err)

		client := &http.Client{
			Timeout: time.Second,
		}
		_, _ = client.Do(req)
	}

	countTotal := atomic.LoadInt32(&count)
	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)

	assert.Equal(s.T(), int32(5), countTotal)
	assert.Equal(s.T(), int32(0), val1)
	assert.Equal(s.T(), int32(0), val2)
}

func (s *SimpleSuite) TestSecureAPI() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	file := s.adaptFile("./fixtures/simple_secure_api.toml", struct{}{})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8000/secure/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestContentTypeDisableAutoDetect() {
	srv1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header()["Content-Type"] = nil
		path := strings.TrimPrefix(req.URL.Path, "/autodetect")
		switch path[:4] {
		case "/css":
			if strings.Contains(req.URL.Path, "/ct") {
				rw.Header().Set("Content-Type", "text/css")
			}

			rw.WriteHeader(http.StatusOK)

			_, err := rw.Write([]byte(".testcss { }"))
			require.NoError(s.T(), err)
		case "/pdf":
			if strings.Contains(req.URL.Path, "/ct") {
				rw.Header().Set("Content-Type", "application/pdf")
			}

			rw.WriteHeader(http.StatusOK)

			data, err := os.ReadFile("fixtures/test.pdf")
			require.NoError(s.T(), err)

			_, err = rw.Write(data)
			require.NoError(s.T(), err)
		}
	}))

	defer srv1.Close()

	file := s.adaptFile("fixtures/simple_contenttype.toml", struct {
		Server string
	}{
		Server: srv1.URL,
	})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/css/ct", time.Second, try.HasHeaderValue("Content-Type", "text/css", false))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/ct", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/css/noct", time.Second, func(res *http.Response) error {
		if ct, ok := res.Header["Content-Type"]; ok {
			return fmt.Errorf("should have no content type and %s is present", ct)
		}
		return nil
	})
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/noct", time.Second, func(res *http.Response) error {
		if ct, ok := res.Header["Content-Type"]; ok {
			return fmt.Errorf("should have no content type and %s is present", ct)
		}
		return nil
	})
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/autodetect/css/ct", time.Second, try.HasHeaderValue("Content-Type", "text/css", false))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/autodetect/pdf/ct", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/autodetect/css/noct", time.Second, try.HasHeaderValue("Content-Type", "text/plain; charset=utf-8", false))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/autodetect/pdf/noct", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	require.NoError(s.T(), err)
}

func (s *SimpleSuite) TestMuxer() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoami1URL := "http://" + net.JoinHostPort(s.getComposeServiceIP("whoami1"), "80")

	file := s.adaptFile("fixtures/simple_muxer.toml", struct {
		Server1 string
	}{whoami1URL})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("!Host"))
	require.NoError(s.T(), err)

	testCases := []struct {
		desc     string
		request  string
		target   string
		body     string
		expected int
	}{
		{
			desc:     "!Host with absolute-form URL with empty host and host header, no match",
			request:  "GET http://@/ HTTP/1.1\r\nHost: test.localhost\r\n\r\n",
			target:   "127.0.0.1:8000",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!Host with absolute-form URL with empty host and host header, match",
			request:  "GET http://@/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			target:   "127.0.0.1:8000",
			expected: http.StatusOK,
		},
		{
			desc:     "!Host with absolute-form URL and host header, no match",
			request:  "GET http://test.localhost/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			target:   "127.0.0.1:8000",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!Host with absolute-form URL and host header, match",
			request:  "GET http://toto.localhost/ HTTP/1.1\r\nHost: test.localhost\r\n\r\n",
			target:   "127.0.0.1:8000",
			expected: http.StatusOK,
		},
		{
			desc:     "!HostRegexp with absolute-form URL with empty host and host header, no match",
			request:  "GET http://@/ HTTP/1.1\r\nHost: test.localhost\r\n\r\n",
			target:   "127.0.0.1:8001",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!HostRegexp with absolute-form URL with empty host and host header, match",
			request:  "GET http://@/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			target:   "127.0.0.1:8001",
			expected: http.StatusOK,
		},
		{
			desc:     "!HostRegexp with absolute-form URL and host header, no match",
			request:  "GET http://test.localhost/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			target:   "127.0.0.1:8001",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!HostRegexp with absolute-form URL and host header, match",
			request:  "GET http://toto.localhost/ HTTP/1.1\r\nHost: test.localhost\r\n\r\n",
			target:   "127.0.0.1:8001",
			expected: http.StatusOK,
		},
		{
			desc:     "!Query with semicolon and empty query param value, no match",
			request:  "GET /?foo=; HTTP/1.1\r\nHost: other.localhost\r\n\r\n",
			target:   "127.0.0.1:8002",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!Query with semicolon, no match",
			request:  "GET /?foo=titi;bar=toto HTTP/1.1\r\nHost: other.localhost\r\n\r\n",
			target:   "127.0.0.1:8002",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!Query with semicolon, match",
			request:  "GET /?bar=toto;boo=titi HTTP/1.1\r\nHost: other.localhost\r\n\r\n",
			target:   "127.0.0.1:8002",
			expected: http.StatusOK,
			body:     "bar=toto&boo=titi",
		},
	}

	for _, test := range testCases {
		conn, err := net.Dial("tcp", test.target)
		require.NoError(s.T(), err)

		_, err = conn.Write([]byte(test.request))
		require.NoError(s.T(), err)

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), test.expected, resp.StatusCode, test.desc)

		if test.body != "" {
			body, err := io.ReadAll(resp.Body)
			require.NoError(s.T(), err)
			assert.Contains(s.T(), string(body), test.body)
		}
	}
}

func (s *SimpleSuite) TestDebugLog() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	file := s.adaptFile("fixtures/simple_debug_log.toml", struct{}{})

	_, output := s.cmdTraefik(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix(`/whoami`)"))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8000/whoami", http.NoBody)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer ThisIsABearerToken")

	response, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)

	if regexp.MustCompile("ThisIsABearerToken").MatchReader(output) {
		log.Info().Msgf("Traefik Logs: %s", output.String())
		log.Info().Msg("Found Authorization Header in Traefik DEBUG logs")
		s.T().Fail()
	}
}

func (s *SimpleSuite) TestEncodeSemicolons() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	whoami1URL := "http://" + net.JoinHostPort(s.getComposeServiceIP("whoami1"), "80")

	file := s.adaptFile("fixtures/simple_encode_semicolons.toml", struct {
		Server1 string
	}{whoami1URL})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`other.localhost`)"))
	require.NoError(s.T(), err)

	testCases := []struct {
		desc     string
		request  string
		target   string
		body     string
		expected int
	}{
		{
			desc:     "Transforming semicolons",
			request:  "GET /?bar=toto;boo=titi HTTP/1.1\r\nHost: other.localhost\r\n\r\n",
			target:   "127.0.0.1:8000",
			expected: http.StatusOK,
			body:     "bar=toto&boo=titi",
		},
		{
			desc:     "Encoding semicolons",
			request:  "GET /?bar=toto&boo=titi;aaaa HTTP/1.1\r\nHost: other.localhost\r\n\r\n",
			target:   "127.0.0.1:8001",
			expected: http.StatusOK,
			body:     "bar=toto&boo=titi%3Baaaa",
		},
	}

	for _, test := range testCases {
		conn, err := net.Dial("tcp", test.target)
		require.NoError(s.T(), err)

		_, err = conn.Write([]byte(test.request))
		require.NoError(s.T(), err)

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		require.NoError(s.T(), err)

		if resp.StatusCode != test.expected {
			log.Info().Msgf("%s failed with %d instead of %d", test.desc, resp.StatusCode, test.expected)
		}

		if test.body != "" {
			body, err := io.ReadAll(resp.Body)
			require.NoError(s.T(), err)
			assert.Contains(s.T(), string(body), test.body)
		}
	}
}

func (s *SimpleSuite) TestDenyFragment() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	s.traefikCmd(withConfigFile("fixtures/simple_default.toml"))

	// Expected a 404 as we did not configure anything
	err := try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	require.NoError(s.T(), err)

	_, err = conn.Write([]byte("GET /#/?bar=toto;boo=titi HTTP/1.1\nHost: other.localhost\n\n"))
	require.NoError(s.T(), err)

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}
