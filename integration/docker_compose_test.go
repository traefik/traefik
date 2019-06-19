package integration

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/pkg/api"
	"github.com/containous/traefik/pkg/testhelpers"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

const (
	composeProject = "minimal"
)

// Docker test suites
type DockerComposeSuite struct {
	BaseSuite
}

func (s *DockerComposeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, composeProject)
	s.composeProject.Start(c)
}

func (s *DockerComposeSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *DockerComposeSuite) TestComposeScale(c *check.C) {
	var serviceCount = 2
	var composeService = "whoami1"

	s.composeProject.Scale(c, composeService, serviceCount)

	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}
	file := s.adaptFile(c, "fixtures/docker/minimal.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req := testhelpers.MustNewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	req.Host = "my.super.host"

	_, err = try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	resp, err := http.Get("http://127.0.0.1:8080/api/rawdata")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()

	var rtconf api.RunTimeRepresentation
	err = json.NewDecoder(resp.Body).Decode(&rtconf)
	c.Assert(err, checker.IsNil)

	// check that we have only one router
	c.Assert(rtconf.Routers, checker.HasLen, 1)

	// check that we have only one service with n servers
	services := rtconf.Services
	c.Assert(services, checker.HasLen, 1)
	for k, v := range services {
		c.Assert(k, checker.Equals, "docker@"+composeService+"_integrationtest"+composeProject)
		c.Assert(v.LoadBalancer.Servers, checker.HasLen, serviceCount)
		// We could break here, but we don't just to keep us honest.
	}
}
