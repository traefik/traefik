package integration

import (
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// ErrorPagesSuite test suites (using libcompose)
type ErrorPagesSuite struct{ BaseSuite }

func (ep *ErrorPagesSuite) SetUpSuite(c *check.C) {
	ep.createComposeProject(c, "error_pages")
	ep.composeProject.Start(c)
}

func (ep *ErrorPagesSuite) TestSimpleConfiguration(c *check.C) {

	errorPageHost := ep.composeProject.Container(c, "nginx2").NetworkSettings.IPAddress
	backendHost := ep.composeProject.Container(c, "nginx1").NetworkSettings.IPAddress

	file := ep.adaptFile(c, "fixtures/error_pages/simple.toml", struct {
		Server1 string
		Server2 string
	}{backendHost, errorPageHost})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:80", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.local"

	err = try.Request(frontendReq, 2*time.Second, try.BodyContains("nginx"))
	c.Assert(err, checker.IsNil)
}

func (ep *ErrorPagesSuite) TestErrorPage(c *check.C) {

	errorPageHost := ep.composeProject.Container(c, "nginx2").NetworkSettings.IPAddress
	backendHost := ep.composeProject.Container(c, "nginx1").NetworkSettings.IPAddress

	//error.toml contains a mis-configuration of the backend host
	file := ep.adaptFile(c, "fixtures/error_pages/error.toml", struct {
		Server1 string
		Server2 string
	}{backendHost, errorPageHost})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	frontendReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:80", nil)
	c.Assert(err, checker.IsNil)
	frontendReq.Host = "test.local"

	err = try.Request(frontendReq, 2*time.Second, try.BodyContains("An error occurred."))
	c.Assert(err, checker.IsNil)
}
