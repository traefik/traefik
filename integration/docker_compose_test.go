package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/api"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

// Docker tests suite.
type DockerComposeSuite struct {
	BaseSuite
}

func TestDockerComposeSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeSuite))
}

func (s *DockerComposeSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.createComposeProject("minimal")
	s.composeUp()
}

func (s *DockerComposeSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *DockerComposeSuite) TestComposeScale() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}
	file := s.adaptFile("fixtures/docker/minimal.toml", tempObjects)

	s.traefikCmd(withConfigFile(file))

	req := testhelpers.MustNewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	req.Host = "my.super.host"

	_, err := try.ResponseUntilStatusCode(req, 5*time.Second, http.StatusOK)
	require.NoError(s.T(), err)

	resp, err := http.Get("http://127.0.0.1:8080/api/rawdata")
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	var rtconf api.RunTimeRepresentation
	err = json.NewDecoder(resp.Body).Decode(&rtconf)
	require.NoError(s.T(), err)

	// check that we have only three routers (the one from this test + 2 unrelated internal ones)
	assert.Len(s.T(), rtconf.Routers, 3)

	// check that we have only one service (not counting the internal ones) with n servers
	services := rtconf.Services
	assert.Len(s.T(), services, 4)
	for name, service := range services {
		if strings.HasSuffix(name, "@internal") {
			continue
		}
		assert.Equal(s.T(), "service-mini@docker", name)
		assert.Len(s.T(), service.LoadBalancer.Servers, 2)
		// We could break here, but we don't just to keep us honest.
	}
}
