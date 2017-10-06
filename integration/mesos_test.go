package integration

import (
	"net/http"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// Mesos test suites (using libcompose)
type MesosSuite struct{ BaseSuite }

func (s *MesosSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "mesos")
}

func (s *MesosSuite) TestSimpleConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/mesos/simple.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}
