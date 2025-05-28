package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type HostResolverSuite struct{ BaseSuite }

func TestHostResolverSuite(t *testing.T) {
	suite.Run(t, new(HostResolverSuite))
}

func (s *HostResolverSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("hostresolver")
	s.composeUp()
}

func (s *HostResolverSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *HostResolverSuite) TestSimpleConfig() {
	s.traefikCmd(withConfigFile("fixtures/simple_hostresolver.toml"))

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
		require.NoError(s.T(), err)
		req.Host = test.host

		err = try.Request(req, 5*time.Second, try.StatusCodeIs(test.status), try.HasBody())
		require.NoError(s.T(), err)
	}
}
