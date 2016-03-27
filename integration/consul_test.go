package main

import (
	"net/http"
	"os/exec"
	"time"

	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
)

// Consul test suites (using libcompose)
type ConsulSuite struct{ BaseSuite }

func (s *ConsulSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "consul")
}

func (s *ConsulSuite) TestSimpleConfiguration(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/consul/simple.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)
	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}
