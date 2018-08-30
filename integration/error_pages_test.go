package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// ErrorPagesSuite test suites (using libcompose)
type ErrorPagesSuite struct {
	BaseSuite
	ErrorPageIP string
	BackendIP   string
}

func (s *ErrorPagesSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "error_pages")
	s.composeProject.Start(c)

	s.ErrorPageIP = s.composeProject.Container(c, "nginx2").NetworkSettings.IPAddress
	s.BackendIP = s.composeProject.Container(c, "nginx1").NetworkSettings.IPAddress
}

func (s *ErrorPagesSuite) TestSimpleConfiguration(c *check.C) {

	file := s.adaptFile(c, "fixtures/error_pages/simple.toml", struct {
		Server1 string
		Server2 string
	}{s.BackendIP, s.ErrorPageIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.local"

	err = try.Request(frontendReq, 2*time.Second, try.BodyContains("nginx"))
	c.Assert(err, checker.IsNil)
}

func (s *ErrorPagesSuite) TestErrorPage(c *check.C) {

	// error.toml contains a mis-configuration of the backend host
	file := s.adaptFile(c, "fixtures/error_pages/error.toml", struct {
		Server1 string
		Server2 string
	}{s.BackendIP, s.ErrorPageIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.local"

	err = try.Request(frontendReq, 2*time.Second, try.BodyContains("An error occurred."))
	c.Assert(err, checker.IsNil)
}
