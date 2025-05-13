package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
	s.composeStop("simple", "withtcplabels", "withlabels1", "withlabels2", "withonelabelmissing", "powpow")
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

func (s *DockerSuite) TestWRRServer() {
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

	whoami1IP := s.getComposeServiceIP("wrr-server")
	whoami2IP := s.getComposeServiceIP("wrr-server2")

	// Expected a 404 as we did not configure anything
	err := try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/http/services", 1000*time.Millisecond, try.BodyContains("wrr-server"))
	require.NoError(s.T(), err)

	repartition := map[string]int{}
	for range 4 {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
		req.Host = "my.wrr.host"
		require.NoError(s.T(), err)

		response, err := http.DefaultClient.Do(req)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, response.StatusCode)

		body, err := io.ReadAll(response.Body)
		require.NoError(s.T(), err)

		if strings.Contains(string(body), whoami1IP) {
			repartition[whoami1IP]++
		}
		if strings.Contains(string(body), whoami2IP) {
			repartition[whoami2IP]++
		}
	}

	assert.Equal(s.T(), 3, repartition[whoami1IP])
	assert.Equal(s.T(), 1, repartition[whoami2IP])
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

	var version map[string]interface{}

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

	var version map[string]interface{}

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

	var version map[string]interface{}

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
