package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

// Docker tests suite.
type DockerSuite struct {
	BaseSuite
}

func TestDockerSuite(t *testing.T) {
	suite.Run(t, new(DockerSuite))
}

func (s *DockerSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.createComposeProject("docker")
}

func (s *DockerSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *DockerSuite) TearDownTest() {
	s.composeStop("simple", "withtcplabels", "withlabels1", "withlabels2", "withonelabelmissing", "powpow", "nonRunning")
}

func (s *DockerSuite) TestSimpleConfiguration() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile("fixtures/docker/simple.toml", tempObjects)

	s.composeUp()

	s.traefikCmd(withConfigFile(file))

	// Expected a 404 as we did not configure anything
	err := try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *DockerSuite) TestDefaultDockerContainers() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile("fixtures/docker/simple.toml", tempObjects)

	s.composeUp("simple")

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	require.NoError(s.T(), err)
	req.Host = "simple.docker.localhost"

	resp, err := try.ResponseUntilStatusCode(req, 3*time.Second, http.StatusOK)
	require.NoError(s.T(), err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)

	var version map[string]any

	assert.NoError(s.T(), json.Unmarshal(body, &version))
	assert.Equal(s.T(), "swarm/1.0.0", version["Version"])
}

func (s *DockerSuite) TestDockerContainersWithTCPLabels() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile("fixtures/docker/simple.toml", tempObjects)

	s.composeUp("withtcplabels")

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`my.super.host`)"))
	require.NoError(s.T(), err)

	who, err := guessWho("127.0.0.1:8000", "my.super.host", true)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), who, "my.super.host")
}

func (s *DockerSuite) TestDockerContainersWithLabels() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile("fixtures/docker/simple.toml", tempObjects)

	s.composeUp("withlabels1", "withlabels2")

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	require.NoError(s.T(), err)
	req.Host = "my-super.host"

	_, err = try.ResponseUntilStatusCode(req, 3*time.Second, http.StatusOK)
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	require.NoError(s.T(), err)
	req.Host = "my.super.host"

	resp, err := try.ResponseUntilStatusCode(req, 3*time.Second, http.StatusOK)
	require.NoError(s.T(), err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)

	var version map[string]any

	assert.NoError(s.T(), json.Unmarshal(body, &version))
	assert.Equal(s.T(), "swarm/1.0.0", version["Version"])
}

func (s *DockerSuite) TestDockerContainersWithOneMissingLabels() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile("fixtures/docker/simple.toml", tempObjects)

	s.composeUp("withonelabelmissing")

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	require.NoError(s.T(), err)
	req.Host = "my.super.host"

	// Expected a 404 as we did not configure anything
	err = try.Request(req, 3*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *DockerSuite) TestRestartDockerContainers() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile("fixtures/docker/simple.toml", tempObjects)

	s.composeUp("powpow")

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	require.NoError(s.T(), err)
	req.Host = "my.super.host"

	// TODO Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	resp, err := try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	require.NoError(s.T(), err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)

	var version map[string]any

	assert.NoError(s.T(), json.Unmarshal(body, &version))
	assert.Equal(s.T(), "swarm/1.0.0", version["Version"])

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("powpow"))
	require.NoError(s.T(), err)

	s.composeStop("powpow")

	time.Sleep(5 * time.Second)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("powpow"))
	assert.Error(s.T(), err)

	s.composeUp("powpow")
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("powpow"))
	require.NoError(s.T(), err)
}

func (s *DockerSuite) TestDockerAllowNonRunning() {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile("fixtures/docker/simple.toml", tempObjects)

	s.composeUp("nonRunning")

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	// Verify the container is working when running
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "non.running.host"

	resp, err := try.ResponseUntilStatusCode(req, 3*time.Second, http.StatusOK)
	require.NoError(s.T(), err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), string(body), "Hostname:")

	// Verify the router exists in Traefik configuration
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers", 1*time.Second, try.BodyContains("NonRunning"))
	require.NoError(s.T(), err)

	// Stop the container
	s.composeStop("nonRunning")

	// Wait a bit for container stop to be detected
	time.Sleep(2 * time.Second)

	// Verify the router still exists in configuration even though container is stopped
	// This is the key test - the router should persist due to allowNonRunning=true
	err = try.GetRequest("http://127.0.0.1:8080/api/http/routers", 10*time.Second, try.BodyContains("NonRunning"))
	require.NoError(s.T(), err)

	// Verify the service still exists in configuration
	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1*time.Second, try.BodyContains("nonRunning"))
	require.NoError(s.T(), err)

	// HTTP requests should fail (502 Bad Gateway) since container is stopped but router exists
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "non.running.host"

	err = try.Request(req, 3*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable))
	require.NoError(s.T(), err)
}
