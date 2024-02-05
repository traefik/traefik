package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v2/integration/try"
)

// Marathon test suites.
type MarathonSuite15 struct {
	BaseSuite
	marathonURL string
}

func TestMarathonSuite15(t *testing.T) {
	suite.Run(t, new(MarathonSuite))
}

func (s *MarathonSuite15) SetUpSuite() {
	s.BaseSuite.SetupSuite()
	s.createComposeProject("marathon15")
	s.composeUp()

	s.marathonURL = "http://" + s.getComposeServiceIP(containerNameMarathon) + ":8080"

	// Wait for Marathon readiness prior to creating the client so that we
	// don't run into the "all cluster members down" state right from the
	// start.
	err := try.GetRequest(s.marathonURL+"/v2/leader", 1*time.Minute, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *MarathonSuite15) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *MarathonSuite15) TestConfigurationUpdate() {
	s.T().Skip("doesn't work")

	// Start Traefik.
	file := s.adaptFile("fixtures/marathon/simple.toml", struct {
		MarathonURL string
	}{s.marathonURL})
	s.traefikCmd(withConfigFile(file))

	// Wait for Traefik to turn ready.
	err := try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	// Prepare Marathon client.
	config := marathon.NewDefaultConfig()
	config.URL = s.marathonURL
	client, err := marathon.NewClient(config)
	require.NoError(s.T(), err)

	// Create test application to be deployed.
	app := marathon.NewDockerApplication().
		Name("/whoami").
		CPU(0.1).
		Memory(32).
		EmptyNetworks().
		AddLabel("traefik.http.Routers.rt.Rule", "PathPrefix(`/service`)")
	app.Container.
		Expose(80).
		Docker.
		Container("traefik/whoami")
	*app.Networks = append(*app.Networks, *marathon.NewBridgePodNetwork())

	// Deploy the test application.
	s.deployApplication(client, app)

	// Query application via Traefik.
	err = try.GetRequest("http://127.0.0.1:8000/service", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	// Create test application with services to be deployed.
	app = marathon.NewDockerApplication().
		Name("/whoami").
		CPU(0.1).
		Memory(32).
		EmptyNetworks().
		AddLabel("traefik.http.Routers.app.Rule", "PathPrefix(`/app`)")
	app.Container.
		Expose(80).
		Docker.
		Container("traefik/whoami")
	*app.Networks = append(*app.Networks, *marathon.NewBridgePodNetwork())

	// Deploy the test application.
	s.deployApplication(client, app)

	// Query application via Traefik.
	err = try.GetRequest("http://127.0.0.1:8000/app", 30*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}
