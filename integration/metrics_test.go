package integration

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

// MetricsSuite tests suite.
type MetricsSuite struct{ BaseSuite }

func TestMetricsSuite(t *testing.T) {
	suite.Run(t, new(MetricsSuite))
}

func (s *MetricsSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *MetricsSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *MetricsSuite) TestMetricsPrometheusDefaultEntryPoint() {
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

func (s *MetricsSuite) TestMetricsPrometheusTwoRoutersOneService() {
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
func (s *MetricsSuite) TestMetricsWithBufferingMiddleware() {
	s.createComposeProject("base")

	s.composeUp()
	defer s.composeDown()

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("MORE THAN TEN BYTES IN RESPONSE"))
	}))

	server.Start()
	defer server.Close()

	file := s.adaptFile("fixtures/metrics/buffering_middleware.toml", struct{ IP string }{IP: strings.TrimPrefix(server.URL, "http://")})

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

func (s *MetricsSuite) TestMetricsPrometheusTCPRouters() {
	s.createComposeProject("tcp")

	s.composeUp()
	defer s.composeDown()

	file := s.adaptFile("fixtures/metrics/tcp_routers.toml", struct {
		WhoamiA      string
		WhoamiNoCert string
		WhoamiNoTLS  string
	}{
		WhoamiA:      s.getComposeServiceIP("whoami-a") + ":8080",
		WhoamiNoCert: s.getComposeServiceIP("whoami-no-cert") + ":8080",
		WhoamiNoTLS:  s.getComposeServiceIP("whoami-no-tls") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	require.NoError(s.T(), err)

	// The TLS version and cipher are pinned to get deterministic labels.
	tlsConfig := func(serverName string) *tls.Config {
		return &tls.Config{
			ServerName:         serverName,
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS12,
			CipherSuites:       []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		}
	}

	noTLSConn, err := net.Dial("tcp", "127.0.0.1:8093")
	require.NoError(s.T(), err)

	passthroughConn, err := tls.Dial("tcp", "127.0.0.1:8093", tlsConfig("whoami-a.test"))
	require.NoError(s.T(), err)

	terminatedConn, err := tls.Dial("tcp", "127.0.0.1:8093", tlsConfig("whoami-c.test"))
	require.NoError(s.T(), err)

	conns := []net.Conn{noTLSConn, passthroughConn, terminatedConn}

	// A round trip on each connection makes sure that it reached its router, and that it is kept open.
	for _, conn := range conns {
		_, err = conn.Write([]byte("WHO"))
		require.NoError(s.T(), err)

		_, err = conn.Read(make([]byte, 2048))
		require.NoError(s.T(), err)
	}

	openConnections := []string{
		`traefik_router_tcp_open_connections{network_type="ipv4",server_port="8093",traefik_router_name="to-whoami-no-tls@file",traefik_service_name="whoami-no-tls@file",traefik_tls_mode="none"} `,
		`traefik_router_tcp_open_connections{network_type="ipv4",server_port="8093",traefik_router_name="to-whoami-a@file",traefik_service_name="whoami-a@file",traefik_tls_mode="passthrough"} `,
		`traefik_router_tcp_open_connections{network_type="ipv4",server_port="8093",traefik_router_name="to-whoami-no-cert@file",traefik_service_name="whoami-no-cert@file",traefik_tls_mode="terminated"} `,
		`traefik_router_tcp_open_tls_connections{network_type="ipv4",server_port="8093",tls_cipher="TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",tls_protocol_version="1.2",traefik_router_name="to-whoami-no-cert@file",traefik_service_name="whoami-no-cert@file"} `,
	}

	withValue := func(value string) []string {
		var metrics []string
		for _, metric := range openConnections {
			metrics = append(metrics, metric+value+"\n")
		}

		return metrics
	}

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains(withValue("1")...))
	require.NoError(s.T(), err)

	// Only the connection whose TLS is terminated by Traefik is counted as a TLS connection.
	err = try.GetRequest("http://127.0.0.1:8080/metrics", 5*time.Second, try.BodyNotContains(`traefik_router_tcp_open_tls_connections{network_type="ipv4",server_port="8093",tls_cipher="TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",tls_protocol_version="1.2",traefik_router_name="to-whoami-a@file"`))
	require.NoError(s.T(), err)

	for _, conn := range conns {
		require.NoError(s.T(), conn.Close())
	}

	err = try.GetRequest("http://127.0.0.1:8080/metrics", 5*time.Second, try.BodyContains(withValue("0")...))
	require.NoError(s.T(), err)
}
