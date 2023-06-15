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
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	checker "github.com/vdemeester/shakers"
)

// SimpleSuite tests suite.
type SimpleSuite struct{ BaseSuite }

func (s *SimpleSuite) TestInvalidConfigShouldFail(c *check.C) {
	cmd, output := s.cmdTraefik(withConfigFile("fixtures/invalid_configuration.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.Do(500*time.Millisecond, func() error {
		expected := "expected '.' or '=', but got '{' instead"
		actual := output.String()

		if !strings.Contains(actual, expected) {
			return fmt.Errorf("got %s, wanted %s", actual, expected)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestSimpleDefaultConfig(c *check.C) {
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/simple_default.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWithWebConfig(c *check.C) {
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/simple_web.toml"))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestPrintHelp(c *check.C) {
	cmd, output := s.cmdTraefik("--help")

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.Do(500*time.Millisecond, func() error {
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
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestRequestAcceptGraceTimeout(c *check.C) {
	s.createComposeProject(c, "reqacceptgrace")

	s.composeUp(c)
	defer s.composeDown(c)

	whoamiURL := "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "whoami"), "80")

	file := s.adaptFile(c, "fixtures/reqacceptgrace.toml", struct {
		Server string
	}{whoamiURL})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	// Make sure exposed service is ready.
	err = try.GetRequest("http://127.0.0.1:8000/service", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Check that /ping endpoint is responding with 200.
	err = try.GetRequest("http://127.0.0.1:8001/ping", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Send SIGTERM to Traefik.
	proc, err := os.FindProcess(cmd.Process.Pid)
	c.Assert(err, checker.IsNil)
	err = proc.Signal(syscall.SIGTERM)
	c.Assert(err, checker.IsNil)

	// Give Traefik time to process the SIGTERM and send a request half-way
	// into the request accepting grace period, by which requests should
	// still get served.
	time.Sleep(5 * time.Second)
	resp, err := http.Get("http://127.0.0.1:8000/service")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, checker.Equals, http.StatusOK)

	// ping endpoint should now return a Service Unavailable.
	resp, err = http.Get("http://127.0.0.1:8001/ping")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()
	c.Assert(resp.StatusCode, checker.Equals, http.StatusServiceUnavailable)

	// Expect Traefik to shut down gracefully once the request accepting grace
	// period has elapsed.
	waitErr := make(chan error)
	go func() {
		waitErr <- cmd.Wait()
	}()

	select {
	case err := <-waitErr:
		c.Assert(err, checker.IsNil)
	case <-time.After(10 * time.Second):
		// By now we are ~5 seconds out of the request accepting grace period
		// (start + 5 seconds sleep prior to the mid-grace period request +
		// 10 seconds timeout = 15 seconds > 10 seconds grace period).
		// Something must have gone wrong if we still haven't terminated at
		// this point.
		c.Fatal("Traefik did not terminate in time")
	}
}

func (s *SimpleSuite) TestCustomPingTerminationStatusCode(c *check.C) {
	file := s.adaptFile(c, "fixtures/custom_ping_termination_status_code.toml", struct{}{})
	defer os.Remove(file)
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8001/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	// Check that /ping endpoint is responding with 200.
	err = try.GetRequest("http://127.0.0.1:8001/ping", 3*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Send SIGTERM to Traefik.
	proc, err := os.FindProcess(cmd.Process.Pid)
	c.Assert(err, checker.IsNil)
	err = proc.Signal(syscall.SIGTERM)
	c.Assert(err, checker.IsNil)

	// ping endpoint should now return a Service Unavailable.
	err = try.GetRequest("http://127.0.0.1:8001/ping", 2*time.Second, try.StatusCodeIs(http.StatusNoContent))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestStatsWithMultipleEntryPoint(c *check.C) {
	c.Skip("Stats is missing")
	s.createComposeProject(c, "stats")

	s.composeUp(c)
	defer s.composeDown(c)

	whoami1URL := "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "whoami1"), "80")
	whoami2URL := "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "whoami2"), "80")

	file := s.adaptFile(c, "fixtures/simple_stats.toml", struct {
		Server1 string
		Server2 string
	}{whoami1URL, whoami2URL})
	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/health", 1*time.Second, try.BodyContains(`"total_status_code_count":{"200":2}`))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestNoAuthOnPing(c *check.C) {
	c.Skip("Waiting for new api handler implementation")

	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	file := s.adaptFile(c, "./fixtures/simple_auth.toml", struct{}{})
	defer os.Remove(file)
	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8001/api/rawdata", 2*time.Second, try.StatusCodeIs(http.StatusUnauthorized))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8001/ping", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestDefaultEntryPointHTTP(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	cmd, output := s.traefikCmd("--entryPoints.http.Address=:8000", "--log.level=DEBUG", "--providers.docker", "--api.insecure")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWithNonExistingEntryPoint(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	cmd, output := s.traefikCmd("--entryPoints.http.Address=:8000", "--log.level=DEBUG", "--providers.docker", "--api.insecure")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestMetricsPrometheusDefaultEntryPoint(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	cmd, output := s.traefikCmd("--entryPoints.http.Address=:8000", "--api.insecure", "--metrics.prometheus.buckets=0.1,0.3,1.2,5.0", "--providers.docker", "--metrics.prometheus.addrouterslabels=true", "--log.level=DEBUG")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix(`/whoami`)"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.BodyContains("_router_"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.BodyContains("_entrypoint_"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 1*time.Second, try.BodyContains("_service_"))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestMetricsPrometheusTwoRoutersOneService(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	cmd, output := s.traefikCmd("--entryPoints.http.Address=:8000", "--api.insecure", "--metrics.prometheus.buckets=0.1,0.3,1.2,5.0", "--providers.docker", "--metrics.prometheus.addentrypointslabels=false", "--metrics.prometheus.addrouterslabels=true", "--log.level=DEBUG")
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami2", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// adding a loop to test if metrics are not deleted
	for i := 0; i < 10; i++ {
		request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/metrics", nil)
		c.Assert(err, checker.IsNil)

		response, err := http.DefaultClient.Do(request)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)

		body, err := io.ReadAll(response.Body)
		c.Assert(err, checker.IsNil)

		// Reqs count of 1 for both routers
		c.Assert(string(body), checker.Contains, "traefik_router_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",router=\"router1@docker\",service=\"whoami1-traefik-integration-test-base@docker\"} 1")
		c.Assert(string(body), checker.Contains, "traefik_router_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",router=\"router2@docker\",service=\"whoami1-traefik-integration-test-base@docker\"} 1")
		// Reqs count of 2 for service behind both routers
		c.Assert(string(body), checker.Contains, "traefik_service_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"whoami1-traefik-integration-test-base@docker\"} 2")
	}
}

// TestMetricsWithBufferingMiddleware checks that the buffering middleware
// (which introduces its own response writer in the chain), does not interfere with
// the capture middleware on which the metrics mechanism relies.
func (s *SimpleSuite) TestMetricsWithBufferingMiddleware(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("MORE THAN TEN BYTES IN RESPONSE"))
	}))

	server.Start()
	defer server.Close()

	file := s.adaptFile(c, "fixtures/simple_metrics_with_buffer_middleware.toml", struct{ IP string }{IP: strings.TrimPrefix(server.URL, "http://")})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix(`/without`)"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8001/without", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8002/with-req", strings.NewReader("MORE THAN TEN BYTES IN REQUEST"))
	c.Assert(err, checker.IsNil)

	// The request should fail because the body is too large.
	err = try.Request(req, 1*time.Second, try.StatusCodeIs(http.StatusRequestEntityTooLarge))
	c.Assert(err, checker.IsNil)

	// The request should fail because the response exceeds the configured limit.
	err = try.GetRequest("http://127.0.0.1:8003/with-resp", 1*time.Second, try.StatusCodeIs(http.StatusInternalServerError))
	c.Assert(err, checker.IsNil)

	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/metrics", nil)
	c.Assert(err, checker.IsNil)

	response, err := http.DefaultClient.Do(request)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusOK)

	body, err := io.ReadAll(response.Body)
	c.Assert(err, checker.IsNil)

	// For allowed requests and responses, the entrypoint and service metrics have the same status code.
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_requests_total{code=\"200\",entrypoint=\"webA\",method=\"GET\",protocol=\"http\"} 1")
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_requests_bytes_total{code=\"200\",entrypoint=\"webA\",method=\"GET\",protocol=\"http\"} 0")
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_responses_bytes_total{code=\"200\",entrypoint=\"webA\",method=\"GET\",protocol=\"http\"} 31")

	c.Assert(string(body), checker.Contains, "traefik_service_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-without@file\"} 1")
	c.Assert(string(body), checker.Contains, "traefik_service_requests_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-without@file\"} 0")
	c.Assert(string(body), checker.Contains, "traefik_service_responses_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-without@file\"} 31")

	// For forbidden requests, the entrypoints have metrics, the services don't.
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_requests_total{code=\"413\",entrypoint=\"webB\",method=\"GET\",protocol=\"http\"} 1")
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_requests_bytes_total{code=\"413\",entrypoint=\"webB\",method=\"GET\",protocol=\"http\"} 0")
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_responses_bytes_total{code=\"413\",entrypoint=\"webB\",method=\"GET\",protocol=\"http\"} 24")

	// For disallowed responses, the entrypoint and service metrics don't have the same status code.
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_requests_bytes_total{code=\"500\",entrypoint=\"webC\",method=\"GET\",protocol=\"http\"} 0")
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_requests_total{code=\"500\",entrypoint=\"webC\",method=\"GET\",protocol=\"http\"} 1")
	c.Assert(string(body), checker.Contains, "traefik_entrypoint_responses_bytes_total{code=\"500\",entrypoint=\"webC\",method=\"GET\",protocol=\"http\"} 21")

	c.Assert(string(body), checker.Contains, "traefik_service_requests_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-resp@file\"} 0")
	c.Assert(string(body), checker.Contains, "traefik_service_requests_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-resp@file\"} 1")
	c.Assert(string(body), checker.Contains, "traefik_service_responses_bytes_total{code=\"200\",method=\"GET\",protocol=\"http\",service=\"service-resp@file\"} 31")
}

func (s *SimpleSuite) TestMultipleProviderSameBackendName(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	whoami1IP := s.getComposeServiceIP(c, "whoami1")
	whoami2IP := s.getComposeServiceIP(c, "whoami2")
	file := s.adaptFile(c, "fixtures/multiple_provider.toml", struct{ IP string }{IP: whoami2IP})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 1*time.Second, try.BodyContains(whoami1IP))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/file", 1*time.Second, try.BodyContains(whoami2IP))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestIPStrategyWhitelist(c *check.C) {
	s.createComposeProject(c, "whitelist")

	s.composeUp(c)
	defer s.composeDown(c)

	cmd, output := s.traefikCmd(withConfigFile("fixtures/simple_whitelist.toml"))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("override"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("override.remoteaddr.whitelist.docker.local"))
	c.Assert(err, checker.IsNil)

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
		if err != nil {
			c.Fatalf("Error while %s: %v", test.desc, err)
		}
	}
}

func (s *SimpleSuite) TestXForwardedHeaders(c *check.C) {
	s.createComposeProject(c, "whitelist")

	s.composeUp(c)
	defer s.composeDown(c)

	cmd, output := s.traefikCmd(withConfigFile("fixtures/simple_whitelist.toml"))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.BodyContains("override.remoteaddr.whitelist.docker.local"))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	c.Assert(err, checker.IsNil)

	req.Host = "override.depth.whitelist.docker.local"
	req.Header.Set("X-Forwarded-For", "8.8.8.8,10.0.0.1,127.0.0.1")

	err = try.Request(req, 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-Proto", "X-Forwarded-For", "X-Forwarded-Host",
			"X-Forwarded-Host", "X-Forwarded-Port", "X-Forwarded-Server", "X-Real-Ip"))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestMultiProvider(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	whoamiURL := "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "whoami1"), "80")

	file := s.adaptFile(c, "fixtures/multiprovider.toml", struct{ Server string }{Server: whoamiURL})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1000*time.Millisecond, try.BodyContains("service"))
	c.Assert(err, checker.IsNil)

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
	c.Assert(err, checker.IsNil)

	request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8080/api/providers/rest", bytes.NewReader(jsonContent))
	c.Assert(err, checker.IsNil)

	response, err := http.DefaultClient.Do(request)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusOK)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1000*time.Millisecond, try.BodyContains("PathPrefix(`/`)"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/", 1*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("CustomValue"))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestSimpleConfigurationHostRequestTrailingPeriod(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	whoamiURL := "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "whoami1"), "80")

	file := s.adaptFile(c, "fixtures/file/simple-hosts.toml", struct{ Server string }{Server: whoamiURL})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

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
		c.Assert(err, checker.IsNil)
		req.Host = test.requestHost
		err = try.Request(req, 1*time.Second, try.StatusCodeIs(http.StatusOK))
		if err != nil {
			c.Fatalf("Error while testing %s: %v", test.desc, err)
		}
	}
}

func (s *SimpleSuite) TestRouterConfigErrors(c *check.C) {
	file := s.adaptFile(c, "fixtures/router_errors.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// All errors
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers", 1000*time.Millisecond, try.BodyContains(`["middleware \"unknown@file\" does not exist","found different TLS options for routers on the same host snitest.net, so using the default TLS options instead"]`))
	c.Assert(err, checker.IsNil)

	// router3 has an error because it uses an unknown entrypoint
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router3@file", 1000*time.Millisecond, try.BodyContains(`entryPoint \"unknown-entrypoint\" doesn't exist`, "no valid entryPoint for this router"))
	c.Assert(err, checker.IsNil)

	// router4 is enabled, but in warning state because its tls options conf was messed up
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router4@file", 1000*time.Millisecond, try.BodyContains(`"status":"warning"`))
	c.Assert(err, checker.IsNil)

	// router5 is disabled because its middleware conf is broken
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers/router5@file", 1000*time.Millisecond, try.BodyContains())
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestServiceConfigErrors(c *check.C) {
	file := s.adaptFile(c, "fixtures/service_errors.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains(`["the service \"service1@file\" does not have any type defined"]`))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services/service1@file", 1000*time.Millisecond, try.BodyContains(`"status":"disabled"`))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services/service2@file", 1000*time.Millisecond, try.BodyContains(`"status":"enabled"`))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestTCPRouterConfigErrors(c *check.C) {
	file := s.adaptFile(c, "fixtures/router_errors.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// router3 has an error because it uses an unknown entrypoint
	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/routers/router3@file", 1000*time.Millisecond, try.BodyContains(`entryPoint \"unknown-entrypoint\" doesn't exist`, "no valid entryPoint for this router"))
	c.Assert(err, checker.IsNil)

	// router4 has an unsupported Rule
	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/routers/router4@file", 1000*time.Millisecond, try.BodyContains("invalid rule: \\\"Host(`mydomain.com`)\\\""))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestTCPServiceConfigErrors(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/service_errors.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/services", 1000*time.Millisecond, try.BodyContains(`["the service \"service1@file\" does not have any type defined"]`))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/services/service1@file", 1000*time.Millisecond, try.BodyContains(`"status":"disabled"`))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/tcp/services/service2@file", 1000*time.Millisecond, try.BodyContains(`"status":"enabled"`))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestUDPRouterConfigErrors(c *check.C) {
	file := s.adaptFile(c, "fixtures/router_errors.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/udp/routers/router3@file", 1000*time.Millisecond, try.BodyContains(`entryPoint \"unknown-entrypoint\" doesn't exist`, "no valid entryPoint for this router"))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestUDPServiceConfigErrors(c *check.C) {
	file := s.adaptFile(c, "fixtures/udp/service_errors.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/udp/services", 1000*time.Millisecond, try.BodyContains(`["the udp service \"service1@file\" does not have any type defined"]`))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/udp/services/service1@file", 1000*time.Millisecond, try.BodyContains(`"status":"disabled"`))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/udp/services/service2@file", 1000*time.Millisecond, try.BodyContains(`"status":"enabled"`))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestWRR(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	whoami1IP := s.getComposeServiceIP(c, "whoami1")
	whoami2IP := s.getComposeServiceIP(c, "whoami2")

	file := s.adaptFile(c, "fixtures/wrr.toml", struct {
		Server1 string
		Server2 string
	}{Server1: "http://" + whoami1IP, Server2: "http://" + whoami2IP})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("service1", "service2"))
	c.Assert(err, checker.IsNil)

	repartition := map[string]int{}
	for i := 0; i < 4; i++ {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
		c.Assert(err, checker.IsNil)

		response, err := http.DefaultClient.Do(req)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)

		body, err := io.ReadAll(response.Body)
		c.Assert(err, checker.IsNil)

		if strings.Contains(string(body), whoami1IP) {
			repartition[whoami1IP]++
		}
		if strings.Contains(string(body), whoami2IP) {
			repartition[whoami2IP]++
		}
	}

	c.Assert(repartition[whoami1IP], checker.Equals, 3)
	c.Assert(repartition[whoami2IP], checker.Equals, 1)
}

func (s *SimpleSuite) TestWRRSticky(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	whoami1IP := s.getComposeServiceIP(c, "whoami1")
	whoami2IP := s.getComposeServiceIP(c, "whoami2")

	file := s.adaptFile(c, "fixtures/wrr_sticky.toml", struct {
		Server1 string
		Server2 string
	}{Server1: "http://" + whoami1IP, Server2: "http://" + whoami2IP})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("service1", "service2"))
	c.Assert(err, checker.IsNil)

	repartition := map[string]int{}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	c.Assert(err, checker.IsNil)

	for i := 0; i < 4; i++ {
		response, err := http.DefaultClient.Do(req)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)

		for _, cookie := range response.Cookies() {
			req.AddCookie(cookie)
		}

		body, err := io.ReadAll(response.Body)
		c.Assert(err, checker.IsNil)

		if strings.Contains(string(body), whoami1IP) {
			repartition[whoami1IP]++
		}
		if strings.Contains(string(body), whoami2IP) {
			repartition[whoami2IP]++
		}
	}

	c.Assert(repartition[whoami1IP], checker.Equals, 4)
	c.Assert(repartition[whoami2IP], checker.Equals, 0)
}

func (s *SimpleSuite) TestMirror(c *check.C) {
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

	file := s.adaptFile(c, "fixtures/mirror.toml", struct {
		MainServer    string
		Mirror1Server string
		Mirror2Server string
	}{MainServer: mainServer, Mirror1Server: mirror1Server, Mirror2Server: mirror2Server})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("mirror1", "mirror2", "service1"))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	c.Assert(err, checker.IsNil)
	for i := 0; i < 10; i++ {
		response, err := http.DefaultClient.Do(req)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
	}

	countTotal := atomic.LoadInt32(&count)
	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)

	c.Assert(countTotal, checker.Equals, int32(10))
	c.Assert(val1, checker.Equals, int32(1))
	c.Assert(val2, checker.Equals, int32(5))
}

func (s *SimpleSuite) TestMirrorWithBody(c *check.C) {
	var count, countMirror1, countMirror2 int32

	body20 := make([]byte, 20)
	_, err := rand.Read(body20)
	c.Assert(err, checker.IsNil)

	body5 := make([]byte, 5)
	_, err = rand.Read(body5)
	c.Assert(err, checker.IsNil)

	verifyBody := func(req *http.Request) {
		b, _ := io.ReadAll(req.Body)
		switch req.Header.Get("Size") {
		case "20":
			if !bytes.Equal(b, body20) {
				c.Fatalf("Not Equals \n%v \n%v", body20, b)
			}
		case "5":
			if !bytes.Equal(b, body5) {
				c.Fatalf("Not Equals \n%v \n%v", body5, b)
			}
		default:
			c.Fatal("Size header not present")
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

	file := s.adaptFile(c, "fixtures/mirror.toml", struct {
		MainServer    string
		Mirror1Server string
		Mirror2Server string
	}{MainServer: mainServer, Mirror1Server: mirror1Server, Mirror2Server: mirror2Server})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("mirror1", "mirror2", "service1"))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", bytes.NewBuffer(body20))
	c.Assert(err, checker.IsNil)
	req.Header.Set("Size", "20")
	for i := 0; i < 10; i++ {
		response, err := http.DefaultClient.Do(req)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
	}

	countTotal := atomic.LoadInt32(&count)
	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)

	c.Assert(countTotal, checker.Equals, int32(10))
	c.Assert(val1, checker.Equals, int32(1))
	c.Assert(val2, checker.Equals, int32(5))

	atomic.StoreInt32(&count, 0)
	atomic.StoreInt32(&countMirror1, 0)
	atomic.StoreInt32(&countMirror2, 0)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoamiWithMaxBody", bytes.NewBuffer(body5))
	c.Assert(err, checker.IsNil)
	req.Header.Set("Size", "5")
	for i := 0; i < 10; i++ {
		response, err := http.DefaultClient.Do(req)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
	}

	countTotal = atomic.LoadInt32(&count)
	val1 = atomic.LoadInt32(&countMirror1)
	val2 = atomic.LoadInt32(&countMirror2)

	c.Assert(countTotal, checker.Equals, int32(10))
	c.Assert(val1, checker.Equals, int32(1))
	c.Assert(val2, checker.Equals, int32(5))

	atomic.StoreInt32(&count, 0)
	atomic.StoreInt32(&countMirror1, 0)
	atomic.StoreInt32(&countMirror2, 0)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoamiWithMaxBody", bytes.NewBuffer(body20))
	c.Assert(err, checker.IsNil)
	req.Header.Set("Size", "20")
	for i := 0; i < 10; i++ {
		response, err := http.DefaultClient.Do(req)
		c.Assert(err, checker.IsNil)
		c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
	}

	countTotal = atomic.LoadInt32(&count)
	val1 = atomic.LoadInt32(&countMirror1)
	val2 = atomic.LoadInt32(&countMirror2)

	c.Assert(countTotal, checker.Equals, int32(10))
	c.Assert(val1, checker.Equals, int32(0))
	c.Assert(val2, checker.Equals, int32(0))
}

func (s *SimpleSuite) TestMirrorCanceled(c *check.C) {
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

	file := s.adaptFile(c, "fixtures/mirror.toml", struct {
		MainServer    string
		Mirror1Server string
		Mirror2Server string
	}{MainServer: mainServer, Mirror1Server: mirror1Server, Mirror2Server: mirror2Server})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("mirror1", "mirror2", "service1"))
	c.Assert(err, checker.IsNil)

	for i := 0; i < 5; i++ {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
		c.Assert(err, checker.IsNil)

		client := &http.Client{
			Timeout: time.Second,
		}
		_, _ = client.Do(req)
	}

	countTotal := atomic.LoadInt32(&count)
	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)

	c.Assert(countTotal, checker.Equals, int32(5))
	c.Assert(val1, checker.Equals, int32(0))
	c.Assert(val2, checker.Equals, int32(0))
}

func (s *SimpleSuite) TestSecureAPI(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	file := s.adaptFile(c, "./fixtures/simple_secure_api.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8000/secure/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestContentTypeDisableAutoDetect(c *check.C) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header()["Content-Type"] = nil
		switch req.URL.Path[:4] {
		case "/css":
			if !strings.Contains(req.URL.Path, "noct") {
				rw.Header().Set("Content-Type", "text/css")
			}

			rw.WriteHeader(http.StatusOK)

			_, err := rw.Write([]byte(".testcss { }"))
			c.Assert(err, checker.IsNil)
		case "/pdf":
			if !strings.Contains(req.URL.Path, "noct") {
				rw.Header().Set("Content-Type", "application/pdf")
			}

			rw.WriteHeader(http.StatusOK)

			data, err := os.ReadFile("fixtures/test.pdf")
			c.Assert(err, checker.IsNil)

			_, err = rw.Write(data)
			c.Assert(err, checker.IsNil)
		}
	}))

	defer srv1.Close()

	file := s.adaptFile(c, "fixtures/simple_contenttype.toml", struct {
		Server string
	}{
		Server: srv1.URL,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file), "--log.level=DEBUG")
	defer display(c)

	err := cmd.Start()
	c.Assert(err, check.IsNil)
	defer s.killCmd(cmd)

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("127.0.0.1"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/css/ct/nomiddleware", time.Second, try.HasHeaderValue("Content-Type", "text/css", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/ct/nomiddleware", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/css/ct/middlewareauto", time.Second, try.HasHeaderValue("Content-Type", "text/css", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/ct/nomiddlewareauto", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/css/ct/middlewarenoauto", time.Second, try.HasHeaderValue("Content-Type", "text/css", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/ct/nomiddlewarenoauto", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/css/noct/nomiddleware", time.Second, try.HasHeaderValue("Content-Type", "text/plain; charset=utf-8", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/noct/nomiddleware", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/css/noct/middlewareauto", time.Second, try.HasHeaderValue("Content-Type", "text/plain; charset=utf-8", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/noct/nomiddlewareauto", time.Second, try.HasHeaderValue("Content-Type", "application/pdf", false))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/css/noct/middlewarenoauto", time.Second, func(res *http.Response) error {
		if ct, ok := res.Header["Content-Type"]; ok {
			return fmt.Errorf("should have no content type and %s is present", ct)
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/pdf/noct/middlewarenoauto", time.Second, func(res *http.Response) error {
		if ct, ok := res.Header["Content-Type"]; ok {
			return fmt.Errorf("should have no content type and %s is present", ct)
		}
		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *SimpleSuite) TestMuxer(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	whoami1URL := "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "whoami1"), "80")

	file := s.adaptFile(c, "fixtures/simple_muxer.toml", struct {
		Server1 string
	}{whoami1URL})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("!Host"))
	c.Assert(err, checker.IsNil)

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
			desc:     "!Query with semicolon, no match",
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
		c.Assert(err, checker.IsNil)

		_, err = conn.Write([]byte(test.request))
		c.Assert(err, checker.IsNil)

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		c.Assert(err, checker.IsNil)

		if resp.StatusCode != test.expected {
			c.Errorf("%s failed with %d instead of %d", test.desc, resp.StatusCode, test.expected)
		}

		if test.body != "" {
			body, err := io.ReadAll(resp.Body)
			c.Assert(err, checker.IsNil)
			c.Assert(string(body), checker.Contains, test.body)
		}
	}
}

func (s *SimpleSuite) TestDebugLog(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	file := s.adaptFile(c, "fixtures/simple_debug_log.toml", struct{}{})
	defer os.Remove(file)

	cmd, output := s.cmdTraefik(withConfigFile(file))

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("PathPrefix(`/whoami`)"))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8000/whoami", http.NoBody)
	c.Assert(err, checker.IsNil)
	req.Header.Set("Autorization", "Bearer ThisIsABearerToken")

	response, err := http.DefaultClient.Do(req)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusOK)

	if regexp.MustCompile("ThisIsABearerToken").MatchReader(output) {
		c.Logf("Traefik Logs: %s", output.String())
		c.Log("Found Authorization Header in Traefik DEBUG logs")
		c.Fail()
	}
}

func (s *SimpleSuite) TestEncodeSemicolons(c *check.C) {
	s.createComposeProject(c, "base")

	s.composeUp(c)
	defer s.composeDown(c)

	whoami1URL := "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "whoami1"), "80")

	file := s.adaptFile(c, "fixtures/simple_encode_semicolons.toml", struct {
		Server1 string
	}{whoami1URL})
	defer os.Remove(file)

	cmd, output := s.traefikCmd(withConfigFile(file))
	defer output(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`other.localhost`)"))
	c.Assert(err, checker.IsNil)

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
		c.Assert(err, checker.IsNil)

		_, err = conn.Write([]byte(test.request))
		c.Assert(err, checker.IsNil)

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		c.Assert(err, checker.IsNil)

		if resp.StatusCode != test.expected {
			c.Errorf("%s failed with %d instead of %d", test.desc, resp.StatusCode, test.expected)
		}

		if test.body != "" {
			body, err := io.ReadAll(resp.Body)
			c.Assert(err, checker.IsNil)
			c.Assert(string(body), checker.Contains, test.body)
		}
	}
}
