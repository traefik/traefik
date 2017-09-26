package integration

import (
	"net/http"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// File test suites
type FileSuite struct{ BaseSuite }

func (s *FileSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "file")

	s.composeProject.Start(c)
}

func (s *FileSuite) TestSimpleConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/file/simple.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

// #56 regression test, make sure it does not fail
func (s *FileSuite) TestSimpleConfigurationNoPanic(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/file/56-simple-panic.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *FileSuite) TestDirectoryConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/file/directory.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything at /test
	err = try.GetRequest("http://127.0.0.1:8000/test", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	// Expected a 502 as there is no backend server
	err = try.GetRequest("http://127.0.0.1:8000/test2", 1000*time.Millisecond, try.StatusCodeIs(http.StatusBadGateway))
	c.Assert(err, checker.IsNil)
}
