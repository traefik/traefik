package integration

import (
	"net/http"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

type HostResolverSuite struct{ BaseSuite }

func (s *HostResolverSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "hostresolver")

	s.composeProject.Start(c)
	s.composeProject.Container(c, "server1")
}

func (s *HostResolverSuite) TestSimpleConfig(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/simple_hostresolver.toml"))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	testCase := []struct {
		desc   string
		host   string
		status int
	}{
		{
			desc:   "host request is resolved",
			host:   "www.github.com",
			status: http.StatusOK,
		},
		{
			desc:   "host request is not resolved",
			host:   "frontend.docker.local",
			status: http.StatusNotFound,
		},
	}

	for _, test := range testCase {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
		c.Assert(err, checker.IsNil)
		req.Host = test.host

		err = try.Request(req, 500*time.Millisecond, try.StatusCodeIs(test.status), try.HasBody())
		c.Assert(err, checker.IsNil)
	}
}
