package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

// Docker tests suite.
type DockerSuite struct {
	BaseSuite
}

func (s *DockerSuite) SetUpTest(c *check.C) {
	s.createComposeProject(c, "docker")
}

func (s *DockerSuite) TearDownTest(c *check.C) {
	s.stopServices(c)
}

func (s *DockerSuite) TestSimpleConfiguration(c *check.C) {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/docker/simple.toml", tempObjects)
	defer os.Remove(file)

	s.startServices(c)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *DockerSuite) TestDefaultDockerContainers(c *check.C) {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/docker/simple.toml", tempObjects)
	defer os.Remove(file)

	s.startServices(c, "simple")

	containers, err := s.dockerService.Ps(context.Background(), s.composeProject.Name, composeapi.PsOptions{})
	c.Assert(err, checker.IsNil)

	containers = s.filterRunning(containers)
	c.Assert(containers, checker.HasLen, 1)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = fmt.Sprintf("%s-%s.docker.localhost", containers[0].Service, containers[0].Project)

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	resp, err := try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	body, err := io.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var version map[string]interface{}

	c.Assert(json.Unmarshal(body, &version), checker.IsNil)
	c.Assert(version["Version"], checker.Equals, "swarm/1.0.0")

	err = s.dockerService.Stop(context.Background(), s.composeProject, composeapi.StopOptions{})
	c.Assert(err, checker.IsNil)
}

func (s *DockerSuite) TestDockerContainersWithTCPLabels(c *check.C) {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/docker/simple.toml", tempObjects)
	defer os.Remove(file)

	s.startServices(c, "withtcplabels")

	containers, err := s.dockerService.Ps(context.Background(), s.composeProject.Name, composeapi.PsOptions{})
	c.Assert(err, checker.IsNil)
	containers = s.filterRunning(containers)
	c.Assert(containers, checker.HasLen, 1)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`my.super.host`)"))
	c.Assert(err, checker.IsNil)

	who, err := guessWho("127.0.0.1:8000", "my.super.host", true)
	c.Assert(err, checker.IsNil)

	c.Assert(who, checker.Contains, "my.super.host")
}

func (s *DockerSuite) TestDockerContainersWithLabels(c *check.C) {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/docker/simple.toml", tempObjects)
	defer os.Remove(file)

	s.startServices(c, "withlabels1", "withlabels2")

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my-super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	_, err = try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	resp, err := try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	body, err := io.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var version map[string]interface{}

	c.Assert(json.Unmarshal(body, &version), checker.IsNil)
	c.Assert(version["Version"], checker.Equals, "swarm/1.0.0")
}

func (s *DockerSuite) TestDockerContainersWithOneMissingLabels(c *check.C) {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/docker/simple.toml", tempObjects)
	defer os.Remove(file)

	s.startServices(c, "withonelabelmissing")

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.Request(req, 1500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *DockerSuite) TestRestartDockerContainers(c *check.C) {
	tempObjects := struct {
		DockerHost  string
		DefaultRule string
	}{
		DockerHost:  s.getDockerHost(),
		DefaultRule: "Host(`{{ normalize .Name }}.docker.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/docker/simple.toml", tempObjects)
	defer os.Remove(file)

	s.startServices(c, "powpow")

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	resp, err := try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	body, err := io.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var version map[string]interface{}

	c.Assert(json.Unmarshal(body, &version), checker.IsNil)
	c.Assert(version["Version"], checker.Equals, "swarm/1.0.0")

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("powpow"))
	c.Assert(err, checker.IsNil)

	s.stopServices(c, "powpow")

	time.Sleep(5 * time.Second)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.BodyContains("powpow"))
	c.Assert(err, checker.NotNil)

	s.startServices(c, "powpow")
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("powpow"))
	c.Assert(err, checker.IsNil)
}

func (s *DockerSuite) startServices(c *check.C, services ...string) {
	var err error
	s.composeProject.Services, err = s.composeProject.GetServices(services...)
	c.Assert(err, checker.IsNil)

	err = s.dockerService.Up(context.Background(), s.composeProject, composeapi.UpOptions{Create: composeapi.CreateOptions{RemoveOrphans: true}})
	c.Assert(err, checker.IsNil)
}

func (s *DockerSuite) stopServices(c *check.C, services ...string) {
	err := s.dockerService.Stop(context.Background(), s.composeProject, composeapi.StopOptions{Services: services})
	c.Assert(err, checker.IsNil)
}

func (s *DockerSuite) filterRunning(containers []composeapi.ContainerSummary) []composeapi.ContainerSummary {
	runningContainers := make([]composeapi.ContainerSummary, 0)

	for _, c := range containers {
		if strings.EqualFold(c.State, composeapi.RUNNING) {
			runningContainers = append(runningContainers, c)
		}
	}
	return runningContainers
}
