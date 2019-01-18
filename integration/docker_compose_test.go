package integration

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/api"
	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/testhelpers"
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

	file := s.adaptFileForHost(c, "fixtures/docker/minimal.toml")
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

	resp, err := http.Get("http://127.0.0.1:8080/api/providers/docker/services")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()

	var services []api.ServiceRepresentation
	err = json.NewDecoder(resp.Body).Decode(&services)
	c.Assert(err, checker.IsNil)

	// check that we have only one service with n servers
	c.Assert(services, checker.HasLen, 1)
	c.Assert(services[0].ID, checker.Equals, composeService+"_integrationtest"+composeProject)
	c.Assert(services[0].LoadBalancer.Servers, checker.HasLen, serviceCount)

	resp, err = http.Get("http://127.0.0.1:8080/api/providers/docker/routers")
	c.Assert(err, checker.IsNil)
	defer resp.Body.Close()

	var routers []api.RouterRepresentation
	err = json.NewDecoder(resp.Body).Decode(&routers)
	c.Assert(err, checker.IsNil)

	// check that we have only one router
	c.Assert(routers, checker.HasLen, 1)
}
