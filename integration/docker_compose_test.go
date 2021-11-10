package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	"github.com/traefik/traefik/v2/pkg/api"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
	checker "github.com/vdemeester/shakers"
)

const (
	composeProject = "minimal"
)

// Docker tests suite.
type DockerComposeSuite struct {
	BaseSuite
}

func (s *DockerComposeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, composeProject)
	err := s.dockerService.Up(context.Background(), s.composeProject, composeapi.UpOptions{})
	c.Assert(err, checker.IsNil)
}

func (s *DockerComposeSuite) TestComposeScale(c *check.C) {
	serviceCount := 2
	composeService := "whoami1"

	s.composeProject.Services[0].Scale = serviceCount

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
	defer s.killCmd(cmd)

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

	// check that we have only three routers (the one from this test + 2 unrelated internal ones)
	c.Assert(rtconf.Routers, checker.HasLen, 3)

	// check that we have only one service (not counting the internal ones) with n servers
	services := rtconf.Services
	c.Assert(services, checker.HasLen, 4)
	for name, service := range services {
		if strings.HasSuffix(name, "@internal") {
			continue
		}
		c.Assert(name, checker.Equals, composeService+"-"+composeProject+"@docker")
		c.Assert(service.LoadBalancer.Servers, checker.HasLen, serviceCount)
		// We could break here, but we don't just to keep us honest.
	}
}
