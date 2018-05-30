package integration

import (
	"time"
	"net/http"

	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
	"github.com/containous/traefik/integration/try"
)

type HostResolverSuite struct{ BaseSuite }

func (s *HostResolverSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "hostresolver")

	s.composeProject.Start(c)
	s.composeProject.Container(c, "server1")
}

func (s *HostResolverSuite) TestSimpleConfig(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/simple_hostresolver.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "frontend1.docker.local"

	err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}
