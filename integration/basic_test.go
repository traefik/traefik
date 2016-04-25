package main

import (
	"net/http"
	"os/exec"
	"time"

	"fmt"
	"github.com/go-check/check"

	"bytes"
	checker "github.com/vdemeester/shakers"
)

// SimpleSuite
type SimpleSuite struct{ BaseSuite }

func (s *SimpleSuite) TestNoOrInexistentConfigShouldFail(c *check.C) {
	cmd := exec.Command(traefikBinary)

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	cmd.Start()
	time.Sleep(500 * time.Millisecond)
	output := b.Bytes()

	c.Assert(string(output), checker.Contains, "No configuration file found")
	cmd.Process.Kill()

	nonExistentFile := "non/existent/file.toml"
	cmd = exec.Command(traefikBinary, "--configFile="+nonExistentFile)

	cmd.Stdout = &b
	cmd.Stderr = &b

	cmd.Start()
	time.Sleep(500 * time.Millisecond)
	output = b.Bytes()

	c.Assert(string(output), checker.Contains, fmt.Sprintf("Error reading configuration file: open %s: no such file or directory", nonExistentFile))
	cmd.Process.Kill()
}

func (s *SimpleSuite) TestInvalidConfigShouldFail(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/invalid_configuration.toml")

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	cmd.Start()
	time.Sleep(500 * time.Millisecond)
	defer cmd.Process.Kill()
	output := b.Bytes()

	c.Assert(string(output), checker.Contains, "While parsing config: Near line 0 (last key parsed ''): Bare keys cannot contain '{'")
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
