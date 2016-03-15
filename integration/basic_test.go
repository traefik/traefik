package main

import (
	"net/http"
	"os/exec"
	"time"

	"fmt"
	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
)

// SimpleSuite
type SimpleSuite struct{ BaseSuite }

func (s *SimpleSuite) TestNoOrInexistentConfigShouldFail(c *check.C) {
	cmd := exec.Command(traefikBinary)
	output, err := cmd.CombinedOutput()

	c.Assert(err, checker.NotNil)
	c.Assert(string(output), checker.Contains, "Error reading file: open : no such file or directory")

	nonExistentFile := "non/existent/file.toml"
	cmd = exec.Command(traefikBinary, "--configFile="+nonExistentFile)
	output, err = cmd.CombinedOutput()

	c.Assert(err, checker.NotNil)
	c.Assert(string(output), checker.Contains, fmt.Sprintf("Error reading file: open %s: no such file or directory", nonExistentFile))
}

func (s *SimpleSuite) TestInvalidConfigShouldFail(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/invalid_configuration.toml")
	output, err := cmd.CombinedOutput()

	c.Assert(err, checker.NotNil)
	c.Assert(string(output), checker.Contains, "Error reading file: While parsing config: Near line 1")
}

func (s *SimpleSuite) TestSimpleDefaultConfig(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/simple_default.toml")
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

func (s *SimpleSuite) TestWithWebConfig(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/simple_web.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:8080/api")
	// Expected a 200
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)
}
