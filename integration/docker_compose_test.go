package integration

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

const (
	composeProject = "minimal"
)

// Docker test suites
type DockerComposeSuite struct {
	BaseSuite
}

func (s *DockerComposeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, composeProject)
	s.composeProject.Start(c)
}

func (s *DockerComposeSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *DockerComposeSuite) TestComposeScale(c *check.C) {
	var serviceCount = 2
	var composeService = "whoami1"

	s.composeProject.Scale(c, composeService, serviceCount)

	file := s.adaptFileForHost(c, "fixtures/docker/minimal.toml")
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req := testhelpers.MustNewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	req.Host = "my.super.host"

	_, err = try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	resp, err := http.Get("http://127.0.0.1:8080/api/providers/docker")
	c.Assert(err, checker.IsNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var provider types.Configuration
	c.Assert(json.Unmarshal(body, &provider), checker.IsNil)

	// check that we have only one backend with n servers
	c.Assert(provider.Backends, checker.HasLen, 1)

	myBackend := provider.Backends["backend-"+composeService+"-integrationtest"+composeProject]
	c.Assert(myBackend, checker.NotNil)
	c.Assert(myBackend.Servers, checker.HasLen, serviceCount)

	// check that we have only one frontend
	c.Assert(provider.Frontends, checker.HasLen, 1)
}
