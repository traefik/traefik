package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"

	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
)

func (s *SimpleSuite) TestNoOrInexistentConfigShouldFail(c *check.C) {
	cmd := exec.Command(traefikBinary)
	output, err := cmd.CombinedOutput()

	c.Assert(err, checker.NotNil)
	c.Assert(string(output), checker.Contains, "Error reading file open traefik.toml: no such file or directory")

	nonExistentFile := "non/existent/file.toml"
	cmd = exec.Command(traefikBinary, nonExistentFile)
	output, err = cmd.CombinedOutput()

	c.Assert(err, checker.NotNil)
	c.Assert(string(output), checker.Contains, fmt.Sprintf("Error reading file open %s: no such file or directory", nonExistentFile))
}

func (s *SimpleSuite) TestInvalidConfigShouldFail(c *check.C) {
	cmd := exec.Command(traefikBinary, "fixtures/invalid_configuration.toml")
	output, err := cmd.CombinedOutput()

	c.Assert(err, checker.NotNil)
	c.Assert(string(output), checker.Contains, "Error reading file Near line 1")
}

func (s *SimpleSuite) TestSimpleDefaultConfig(c *check.C) {
	cmd := exec.Command(traefikBinary, "fixtures/simple_default.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)

	time.Sleep(100 * time.Millisecond)
	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1/")

	// Expected a 404 as we did not comfigure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)

	killErr := cmd.Process.Kill()
	c.Assert(killErr, checker.IsNil)
}
