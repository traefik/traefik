package integration

import (
	"bytes"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// HealthCheck test suites (using libcompose)
type HealthCheckSuite struct {
	BaseSuite
	whoami1IP string
	whoami2IP string
}

func (s *HealthCheckSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "healthcheck")
	s.composeProject.Start(c)

	s.whoami1IP = s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress
	s.whoami2IP = s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress
}

func (s *HealthCheckSuite) TestSimpleConfiguration(c *check.C) {

	file := s.adaptFile(c, "fixtures/healthcheck/simple.toml", struct {
		Server1 string
		Server2 string
	}{s.whoami1IP, s.whoami2IP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, try.BodyContains("Host:test.localhost"))
	c.Assert(err, checker.IsNil)

	frontendHealthReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	c.Assert(err, checker.IsNil)
	frontendHealthReq.Host = "test.localhost"

	err = try.Request(frontendHealthReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Fix all whoami health to 500
	client := &http.Client{}
	whoamiHosts := []string{s.whoami1IP, s.whoami2IP}
	for _, whoami := range whoamiHosts {
		statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+whoami+"/health", bytes.NewBuffer([]byte("500")))
		c.Assert(err, checker.IsNil)
		_, err = client.Do(statusInternalServerErrorReq)
		c.Assert(err, checker.IsNil)
	}

	// Waiting for Traefik healthcheck
	try.Sleep(2 * time.Second)

	// Verify no backend service is available due to failing health checks
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	c.Assert(err, checker.IsNil)

	// Change one whoami health to 200
	statusOKReq1, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBuffer([]byte("200")))
	c.Assert(err, checker.IsNil)
	_, err = client.Do(statusOKReq1)
	c.Assert(err, checker.IsNil)

	// Verify frontend health : after
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.localhost"

	// Check if whoami1 responds
	err = try.Request(frontendReq, 500*time.Millisecond, try.BodyContains(s.whoami1IP))
	c.Assert(err, checker.IsNil)

	// Check if the service with bad health check (whoami2) never respond.
	err = try.Request(frontendReq, 2*time.Second, try.BodyContains(s.whoami2IP))
	c.Assert(err, checker.Not(checker.IsNil))

	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, http.StatusNotFound)
}

func (s *HealthCheckSuite) TestMultipleEntrypointsWrr(c *check.C) {
	s.doTestMultipleEntrypoints(c, "fixtures/healthcheck/multiple-entrypoints-wrr.toml")
}

func (s *HealthCheckSuite) TestMultipleEntrypointsDrr(c *check.C) {
	s.doTestMultipleEntrypoints(c, "fixtures/healthcheck/multiple-entrypoints-drr.toml")
}

func (s *HealthCheckSuite) doTestMultipleEntrypoints(c *check.C, fixture string) {
	file := s.adaptFile(c, fixture, struct {
		Server1 string
		Server2 string
	}{s.whoami1IP, s.whoami2IP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Wait for traefik
	err = try.GetRequest("http://localhost:8080/api/providers", 60*time.Second, try.BodyContains("Host:test.localhost"))
	c.Assert(err, checker.IsNil)

	// Check entrypoint http1
	frontendHealthReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	c.Assert(err, checker.IsNil)
	frontendHealthReq.Host = "test.localhost"

	err = try.Request(frontendHealthReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Check entrypoint http2
	frontendHealthReq, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:9000/health", nil)
	c.Assert(err, checker.IsNil)
	frontendHealthReq.Host = "test.localhost"

	err = try.Request(frontendHealthReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Set one whoami health to 500
	client := &http.Client{}
	statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBuffer([]byte("500")))
	c.Assert(err, checker.IsNil)
	_, err = client.Do(statusInternalServerErrorReq)
	c.Assert(err, checker.IsNil)

	// Waiting for Traefik healthcheck
	try.Sleep(2 * time.Second)

	frontend1Req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	frontend1Req.Host = "test.localhost"

	frontend2Req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:9000/", nil)
	c.Assert(err, checker.IsNil)
	frontend2Req.Host = "test.localhost"

	// Check if whoami1 never responds
	err = try.Request(frontend2Req, 2*time.Second, try.BodyContains(s.whoami1IP))
	c.Assert(err, checker.Not(checker.IsNil))

	// Check if whoami1 never responds
	err = try.Request(frontend1Req, 2*time.Second, try.BodyContains(s.whoami1IP))
	c.Assert(err, checker.Not(checker.IsNil))
}

func (s *HealthCheckSuite) TestPortOverload(c *check.C) {

	// Set one whoami health to 200
	client := &http.Client{}
	statusInternalServerErrorReq, err := http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBuffer([]byte("200")))
	c.Assert(err, checker.IsNil)
	_, err = client.Do(statusInternalServerErrorReq)
	c.Assert(err, checker.IsNil)

	file := s.adaptFile(c, "fixtures/healthcheck/port_overload.toml", struct {
		Server1 string
	}{s.whoami1IP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// wait for traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.BodyContains("Host:test.localhost"))
	c.Assert(err, checker.IsNil)

	frontendHealthReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/health", nil)
	c.Assert(err, checker.IsNil)
	frontendHealthReq.Host = "test.localhost"

	// We test bad gateway because we use an invalid port for the backend
	err = try.Request(frontendHealthReq, 500*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	c.Assert(err, checker.IsNil)

	// Set one whoami health to 500
	statusInternalServerErrorReq, err = http.NewRequest(http.MethodPost, "http://"+s.whoami1IP+"/health", bytes.NewBuffer([]byte("500")))
	c.Assert(err, checker.IsNil)
	_, err = client.Do(statusInternalServerErrorReq)
	c.Assert(err, checker.IsNil)

	// Waiting for Traefik healthcheck
	try.Sleep(2 * time.Second)

	// Verify no backend service is available due to failing health checks
	err = try.Request(frontendHealthReq, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	c.Assert(err, checker.IsNil)

}
