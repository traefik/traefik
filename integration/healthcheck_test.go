package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

// HealthCheck test suites.
type HealthCheckSuite struct {
	BaseSuite
	whoami1IP string
	whoami2IP string
	whoami3IP string
	whoami4IP string
}

func TestHealthCheckSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckSuite))
}

func (s *HealthCheckSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("healthcheck")
	s.composeUp()

	s.whoami1IP = s.getComposeServiceIP("whoami1")
	s.whoami2IP = s.getComposeServiceIP("whoami2")
	s.whoami3IP = s.getComposeServiceIP("whoami3")
	s.whoami4IP = s.getComposeServiceIP("whoami4")
}

func (s *HealthCheckSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *HealthCheckSuite) TestSimpleConfiguration() {
	file := s.adaptFile("fixtures/healthcheck/simple.toml", struct {
		Server1 string
		Server2 string
	}{s.whoami1IP, s.whoami2IP})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Host(`test.localhost`)"))
	require.NoError(s.T(), err)

	frontendHealthReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	require.NoError(s.T(), err)
	frontendHealthReq.Host = "test.localhost"

	err = try.Request(frontendHealthReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Fix all whoami health to 500
	client := &http.Client{}
	whoamiHosts := []string{s.whoami1IP, s.whoami2IP}
	for _, whoami := range whoamiHosts {
		statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+whoami+"/health", bytes.NewBufferString("500"))
		require.NoError(s.T(), err)
		_, err = client.Do(statusInternalServerErrorReq)
		require.NoError(s.T(), err)
	}

	// Verify no backend service is available due to failing health checks
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	// Change one whoami health to 200
	statusOKReq1, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBufferString("200"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusOKReq1)
	require.NoError(s.T(), err)

	// Verify frontend health : after
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	frontendReq.Host = "test.localhost"

	// Check if whoami1 responds
	err = try.Request(frontendReq, 500*time.Millisecond, try.BodyContains(s.whoami1IP))
	require.NoError(s.T(), err)

	// Check if the service with bad health check (whoami2) never respond.
	err = try.Request(frontendReq, 2*time.Second, try.BodyContains(s.whoami2IP))
	assert.Error(s.T(), err)

	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
}

func (s *HealthCheckSuite) TestSimpleConfiguration_Passive() {
	file := s.adaptFile("fixtures/healthcheck/simple_passive.toml", struct {
		Server1 string
	}{s.whoami1IP})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Host(`test.localhost`)"))
	require.NoError(s.T(), err)

	frontendHealthReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	require.NoError(s.T(), err)
	frontendHealthReq.Host = "test.localhost"

	err = try.Request(frontendHealthReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Fix all whoami health to 500
	client := &http.Client{}
	whoamiHosts := []string{s.whoami1IP, s.whoami2IP}
	for _, whoami := range whoamiHosts {
		statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+whoami+"/health", bytes.NewBufferString("500"))
		require.NoError(s.T(), err)
		_, err = client.Do(statusInternalServerErrorReq)
		require.NoError(s.T(), err)
	}

	// First call, the passive health check is not yet triggered, so we expect a 500.
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusInternalServerError))
	require.NoError(s.T(), err)

	// Verify no backend service is available due to failing health checks
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	// Change one whoami health to 200
	statusOKReq1, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBufferString("200"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusOKReq1)
	require.NoError(s.T(), err)

	// Verify frontend health : after
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *HealthCheckSuite) TestMultipleEntrypoints() {
	file := s.adaptFile("fixtures/healthcheck/multiple-entrypoints.toml", struct {
		Server1 string
		Server2 string
	}{s.whoami1IP, s.whoami2IP})

	s.traefikCmd(withConfigFile(file))

	// Wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Host(`test.localhost`)"))
	require.NoError(s.T(), err)

	// Check entrypoint http1
	frontendHealthReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	require.NoError(s.T(), err)
	frontendHealthReq.Host = "test.localhost"

	err = try.Request(frontendHealthReq, 5*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Check entrypoint http2
	frontendHealthReq, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:9000/health", nil)
	require.NoError(s.T(), err)
	frontendHealthReq.Host = "test.localhost"

	err = try.Request(frontendHealthReq, 5*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Set the both whoami health to 500
	client := &http.Client{}
	whoamiHosts := []string{s.whoami1IP, s.whoami2IP}
	for _, whoami := range whoamiHosts {
		statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+whoami+"/health", bytes.NewBufferString("500"))
		require.NoError(s.T(), err)
		_, err = client.Do(statusInternalServerErrorReq)
		require.NoError(s.T(), err)
	}

	// Verify no backend service is available due to failing health checks
	err = try.Request(frontendHealthReq, 5*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	// reactivate the whoami2
	statusInternalServerOkReq, err := http.NewRequest(http.MethodPost, "http://"+s.whoami2IP+"/health", bytes.NewBufferString("200"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusInternalServerOkReq)
	require.NoError(s.T(), err)

	frontend1Req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	frontend1Req.Host = "test.localhost"

	frontend2Req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:9000/", nil)
	require.NoError(s.T(), err)
	frontend2Req.Host = "test.localhost"

	// Check if whoami1 never responds
	err = try.Request(frontend2Req, 2*time.Second, try.BodyContains(s.whoami1IP))
	assert.Error(s.T(), err)

	// Check if whoami1 never responds
	err = try.Request(frontend1Req, 2*time.Second, try.BodyContains(s.whoami1IP))
	assert.Error(s.T(), err)
}

func (s *HealthCheckSuite) TestPortOverload() {
	// Set one whoami health to 200
	client := &http.Client{}
	statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBufferString("200"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusInternalServerErrorReq)
	require.NoError(s.T(), err)

	file := s.adaptFile("fixtures/healthcheck/port_overload.toml", struct {
		Server1 string
	}{s.whoami1IP})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("Host(`test.localhost`)"))
	require.NoError(s.T(), err)

	frontendHealthReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	require.NoError(s.T(), err)
	frontendHealthReq.Host = "test.localhost"

	// We test bad gateway because we use an invalid port for the backend
	err = try.Request(frontendHealthReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	require.NoError(s.T(), err)

	// Set one whoami health to 500
	statusInternalServerErrorReq, err = http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBufferString("500"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusInternalServerErrorReq)
	require.NoError(s.T(), err)

	// Verify no backend service is available due to failing health checks
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)
}

// Checks if all the loadbalancers created will correctly update the server status.
func (s *HealthCheckSuite) TestMultipleRoutersOnSameService() {
	file := s.adaptFile("fixtures/healthcheck/multiple-routers-one-same-service.toml", struct {
		Server1 string
	}{s.whoami1IP})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Host(`test.localhost`)"))
	require.NoError(s.T(), err)

	// Set whoami health to 200 to be sure to start with the wanted status
	client := &http.Client{}
	statusOkReq, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBufferString("200"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusOkReq)
	require.NoError(s.T(), err)

	// check healthcheck on web1 entrypoint
	healthReqWeb1, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	require.NoError(s.T(), err)
	healthReqWeb1.Host = "test.localhost"
	err = try.Request(healthReqWeb1, 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// check healthcheck on web2 entrypoint
	healthReqWeb2, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:9000/health", nil)
	require.NoError(s.T(), err)
	healthReqWeb2.Host = "test.localhost"

	err = try.Request(healthReqWeb2, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Set whoami health to 500
	statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBufferString("500"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusInternalServerErrorReq)
	require.NoError(s.T(), err)

	// Verify no backend service is available due to failing health checks
	err = try.Request(healthReqWeb1, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	err = try.Request(healthReqWeb2, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	// Change one whoami health to 200
	statusOKReq1, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBufferString("200"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusOKReq1)
	require.NoError(s.T(), err)

	// Verify health check
	err = try.Request(healthReqWeb1, 3*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.Request(healthReqWeb2, 3*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *HealthCheckSuite) TestPropagate() {
	file := s.adaptFile("fixtures/healthcheck/propagate.toml", struct {
		Server1 string
		Server2 string
		Server3 string
		Server4 string
	}{s.whoami1IP, s.whoami2IP, s.whoami3IP, s.whoami4IP})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Host(`root.localhost`)"))
	require.NoError(s.T(), err)

	rootReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	require.NoError(s.T(), err)
	rootReq.Host = "root.localhost"

	err = try.Request(rootReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Bring whoami1 and whoami3 down
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	whoamiHosts := []string{s.whoami1IP, s.whoami3IP}
	for _, whoami := range whoamiHosts {
		statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+whoami+"/health", bytes.NewBufferString("500"))
		require.NoError(s.T(), err)
		_, err = client.Do(statusInternalServerErrorReq)
		require.NoError(s.T(), err)
	}

	try.Sleep(time.Second)

	want2 := `IP: ` + s.whoami2IP
	want4 := `IP: ` + s.whoami4IP

	// Verify load-balancing on root still works, and that we're getting an alternation between wsp2, and wsp4.
	reachedServers := make(map[string]int)
	for range 4 {
		resp, err := client.Do(rootReq)
		require.NoError(s.T(), err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		if reachedServers[s.whoami4IP] > reachedServers[s.whoami2IP] {
			assert.Contains(s.T(), string(body), want2)
			reachedServers[s.whoami2IP]++
			continue
		}

		if reachedServers[s.whoami2IP] > reachedServers[s.whoami4IP] {
			assert.Contains(s.T(), string(body), want4)
			reachedServers[s.whoami4IP]++
			continue
		}

		// First iteration, so we can't tell whether it's going to be wsp2, or wsp4.
		if strings.Contains(string(body), `IP: `+s.whoami4IP) {
			reachedServers[s.whoami4IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami2IP) {
			reachedServers[s.whoami2IP]++
			continue
		}
	}

	assert.Equal(s.T(), 2, reachedServers[s.whoami2IP])
	assert.Equal(s.T(), 2, reachedServers[s.whoami4IP])

	fooReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	require.NoError(s.T(), err)
	fooReq.Host = "foo.localhost"

	// Verify load-balancing on foo still works, and that we're getting wsp2, wsp2, wsp2, wsp2, etc.
	want := `IP: ` + s.whoami2IP
	for range 4 {
		resp, err := client.Do(fooReq)
		require.NoError(s.T(), err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		assert.Contains(s.T(), string(body), want)
	}

	barReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	require.NoError(s.T(), err)
	barReq.Host = "bar.localhost"

	// Verify load-balancing on bar still works, and that we're getting wsp2, wsp2, wsp2, wsp2, etc.
	want = `IP: ` + s.whoami2IP
	for range 4 {
		resp, err := client.Do(barReq)
		require.NoError(s.T(), err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		assert.Contains(s.T(), string(body), want)
	}

	// Bring whoami2 and whoami4 down
	whoamiHosts = []string{s.whoami2IP, s.whoami4IP}
	for _, whoami := range whoamiHosts {
		statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+whoami+"/health", bytes.NewBufferString("500"))
		require.NoError(s.T(), err)
		_, err = client.Do(statusInternalServerErrorReq)
		require.NoError(s.T(), err)
	}

	try.Sleep(time.Second)

	// Verify that everything is down, and that we get 503s everywhere.
	for range 2 {
		resp, err := client.Do(rootReq)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusServiceUnavailable, resp.StatusCode)

		resp, err = client.Do(fooReq)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusServiceUnavailable, resp.StatusCode)

		resp, err = client.Do(barReq)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusServiceUnavailable, resp.StatusCode)
	}

	// Bring everything back up.
	whoamiHosts = []string{s.whoami1IP, s.whoami2IP, s.whoami3IP, s.whoami4IP}
	for _, whoami := range whoamiHosts {
		statusOKReq, err := http.NewRequest(http.MethodPost, "http://"+whoami+"/health", bytes.NewBufferString("200"))
		require.NoError(s.T(), err)
		_, err = client.Do(statusOKReq)
		require.NoError(s.T(), err)
	}

	try.Sleep(time.Second)

	// Verify everything is up on root router.
	reachedServers = make(map[string]int)
	for range 4 {
		resp, err := client.Do(rootReq)
		require.NoError(s.T(), err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		if strings.Contains(string(body), `IP: `+s.whoami1IP) {
			reachedServers[s.whoami1IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami2IP) {
			reachedServers[s.whoami2IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami3IP) {
			reachedServers[s.whoami3IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami4IP) {
			reachedServers[s.whoami4IP]++
			continue
		}
	}

	assert.Equal(s.T(), 1, reachedServers[s.whoami1IP])
	assert.Equal(s.T(), 1, reachedServers[s.whoami2IP])
	assert.Equal(s.T(), 1, reachedServers[s.whoami3IP])
	assert.Equal(s.T(), 1, reachedServers[s.whoami4IP])

	// Verify everything is up on foo router.
	reachedServers = make(map[string]int)
	for range 4 {
		resp, err := client.Do(fooReq)
		require.NoError(s.T(), err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		if strings.Contains(string(body), `IP: `+s.whoami1IP) {
			reachedServers[s.whoami1IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami2IP) {
			reachedServers[s.whoami2IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami3IP) {
			reachedServers[s.whoami3IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami4IP) {
			reachedServers[s.whoami4IP]++
			continue
		}
	}

	assert.Equal(s.T(), 2, reachedServers[s.whoami1IP])
	assert.Equal(s.T(), 1, reachedServers[s.whoami2IP])
	assert.Equal(s.T(), 1, reachedServers[s.whoami3IP])
	assert.Equal(s.T(), 0, reachedServers[s.whoami4IP])

	// Verify everything is up on bar router.
	reachedServers = make(map[string]int)
	for range 4 {
		resp, err := client.Do(barReq)
		require.NoError(s.T(), err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(s.T(), err)

		if strings.Contains(string(body), `IP: `+s.whoami1IP) {
			reachedServers[s.whoami1IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami2IP) {
			reachedServers[s.whoami2IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami3IP) {
			reachedServers[s.whoami3IP]++
			continue
		}

		if strings.Contains(string(body), `IP: `+s.whoami4IP) {
			reachedServers[s.whoami4IP]++
			continue
		}
	}

	assert.Equal(s.T(), 2, reachedServers[s.whoami1IP])
	assert.Equal(s.T(), 1, reachedServers[s.whoami2IP])
	assert.Equal(s.T(), 1, reachedServers[s.whoami3IP])
	assert.Equal(s.T(), 0, reachedServers[s.whoami4IP])
}

func (s *HealthCheckSuite) TestPropagateNoHealthCheck() {
	file := s.adaptFile("fixtures/healthcheck/propagate_no_healthcheck.toml", struct {
		Server1 string
	}{s.whoami1IP})

	s.traefikCmd(withConfigFile(file))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Host(`noop.localhost`)"), try.BodyNotContains("Host(`root.localhost`)"))
	require.NoError(s.T(), err)

	rootReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	require.NoError(s.T(), err)
	rootReq.Host = "root.localhost"

	err = try.Request(rootReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *HealthCheckSuite) TestPropagateReload() {
	// Setup a WSP service without the healthcheck enabled (wsp-service1)
	withoutHealthCheck := s.adaptFile("fixtures/healthcheck/reload_without_healthcheck.toml", struct {
		Server1 string
		Server2 string
	}{s.whoami1IP, s.whoami2IP})
	withHealthCheck := s.adaptFile("fixtures/healthcheck/reload_with_healthcheck.toml", struct {
		Server1 string
		Server2 string
	}{s.whoami1IP, s.whoami2IP})

	s.traefikCmd(withConfigFile(withoutHealthCheck))

	// wait for traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("Host(`root.localhost`)"))
	require.NoError(s.T(), err)

	// Allow one of the underlying services on it to fail all servers HC (whoami2)
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	statusOKReq, err := http.NewRequest(http.MethodPost, "http://"+s.whoami2IP+"/health", bytes.NewBufferString("500"))
	require.NoError(s.T(), err)
	_, err = client.Do(statusOKReq)
	require.NoError(s.T(), err)

	rootReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000", nil)
	require.NoError(s.T(), err)
	rootReq.Host = "root.localhost"

	// Check the failed service (whoami2) is getting requests, but answer 500
	err = try.Request(rootReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)

	// Enable the healthcheck on the root WSP (wsp-service1) and let Traefik reload the config
	fr1, err := os.OpenFile(withoutHealthCheck, os.O_APPEND|os.O_WRONLY, 0o644)
	assert.NotNil(s.T(), fr1)
	require.NoError(s.T(), err)
	err = fr1.Truncate(0)
	require.NoError(s.T(), err)

	fr2, err := os.ReadFile(withHealthCheck)
	require.NoError(s.T(), err)
	_, err = fmt.Fprint(fr1, string(fr2))
	require.NoError(s.T(), err)
	err = fr1.Close()
	require.NoError(s.T(), err)

	// Waiting for the reflected change.
	err = try.Request(rootReq, 5*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Check the failed service (whoami2) is not getting requests
	wantIPs := []string{s.whoami1IP, s.whoami1IP, s.whoami1IP, s.whoami1IP}
	for _, ip := range wantIPs {
		err = try.Request(rootReq, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("IP: "+ip))
		require.NoError(s.T(), err)
	}
}
