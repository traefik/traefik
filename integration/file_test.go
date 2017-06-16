package main

import (
	"net/http"
	"os/exec"
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
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/file/simple.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything
	try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

// #56 regression test, make sure it does not fail
func (s *FileSuite) TestSimpleConfigurationNoPanic(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/file/56-simple-panic.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected a 404 as we did not configure anything
	try.GetRequest("http://127.0.0.1:8000/", 1000*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}
