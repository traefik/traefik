package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/provider/label"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/go-check/check"
	d "github.com/libkermit/docker"
	"github.com/libkermit/docker-check"
	checker "github.com/vdemeester/shakers"
)

var (
	// Label added to started container to identify them as part of the integration test
	TestLabel = "io.traefik.test"

	// Images to have or pull before the build in order to make it work
	// FIXME handle this offline but loading them before build
	RequiredImages = map[string]string{
		"swarm":             "1.0.0",
		"emilevauge/whoami": "latest",
	}
)

// Docker test suites
type DockerSuite struct {
	BaseSuite
	project *docker.Project
}

func (s *DockerSuite) startContainer(c *check.C, image string, args ...string) string {
	return s.startContainerWithConfig(c, image, d.ContainerConfig{
		Cmd: args,
	})
}

func (s *DockerSuite) startContainerWithLabels(c *check.C, image string, labels map[string]string, args ...string) string {
	return s.startContainerWithConfig(c, image, d.ContainerConfig{
		Cmd:    args,
		Labels: labels,
	})
}

func (s *DockerSuite) startContainerWithNameAndLabels(c *check.C, name string, image string, labels map[string]string, args ...string) string {
	return s.startContainerWithConfig(c, image, d.ContainerConfig{
		Name:   name,
		Cmd:    args,
		Labels: labels,
	})
}

func (s *DockerSuite) startContainerWithConfig(c *check.C, image string, config d.ContainerConfig) string {
	if config.Name == "" {
		config.Name = namesgenerator.GetRandomName(10)
	}

	container := s.project.StartWithConfig(c, image, config)

	// FIXME(vdemeester) this is ugly (it's because of the / in front of the name in docker..)
	return strings.SplitAfter(container.Name, "/")[1]
}

func (s *DockerSuite) stopAndRemoveContainerByName(c *check.C, name string) {
	s.project.Stop(c, name)
	s.project.Remove(c, name)
}

func (s *DockerSuite) SetUpSuite(c *check.C) {
	project := docker.NewProjectFromEnv(c)
	s.project = project

	// Pull required images
	for repository, tag := range RequiredImages {
		image := fmt.Sprintf("%s:%s", repository, tag)
		s.project.Pull(c, image)
	}
}

func (s *DockerSuite) TearDownTest(c *check.C) {
	s.project.Clean(c, os.Getenv("CIRCLECI") != "")
}

func (s *DockerSuite) TestSimpleConfiguration(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *DockerSuite) TestDefaultDockerContainers(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	name := s.startContainer(c, "swarm:1.0.0", "manage", "token://blablabla")

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = fmt.Sprintf("%s.docker.localhost", strings.Replace(name, "_", "-", -1))

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	resp, err := try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var version map[string]interface{}

	c.Assert(json.Unmarshal(body, &version), checker.IsNil)
	c.Assert(version["Version"], checker.Equals, "swarm/1.0.0")
}

func (s *DockerSuite) TestDockerContainersWithLabels(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	// Start a container with some labels
	labels := map[string]string{
		label.TraefikFrontendRule: "Host:my.super.host",
	}
	s.startContainerWithLabels(c, "swarm:1.0.0", labels, "manage", "token://blabla")

	// Start another container by replacing a '.' by a '-'
	labels = map[string]string{
		label.TraefikFrontendRule: "Host:my-super.host",
	}
	s.startContainerWithLabels(c, "swarm:1.0.0", labels, "manage", "token://blablabla")
	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

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

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var version map[string]interface{}

	c.Assert(json.Unmarshal(body, &version), checker.IsNil)
	c.Assert(version["Version"], checker.Equals, "swarm/1.0.0")
}

func (s *DockerSuite) TestDockerContainersWithOneMissingLabels(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	// Start a container with some labels
	labels := map[string]string{
		"traefik.frontend.value": "my.super.host",
	}
	s.startContainerWithLabels(c, "swarm:1.0.0", labels, "manage", "token://blabla")

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	// TODO validate : run on 80
	// Expected a 404 as we did not comfigure anything
	err = try.Request(req, 1500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

// TestDockerContainersWithServiceLabels allows cheking the labels behavior
// Use service label if defined and compete information with container labels.
func (s *DockerSuite) TestDockerContainersWithServiceLabels(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	// Start a container with some labels
	labels := map[string]string{
		label.Prefix + "servicename.frontend.rule": "Host:my.super.host",
		label.TraefikFrontendRule:                  "Host:my.wrong.host",
		label.TraefikPort:                          "2375",
	}
	s.startContainerWithLabels(c, "swarm:1.0.0", labels, "manage", "token://blabla")

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	resp, err := try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var version map[string]interface{}

	c.Assert(json.Unmarshal(body, &version), checker.IsNil)
	c.Assert(version["Version"], checker.Equals, "swarm/1.0.0")
}

func (s *DockerSuite) TestRestartDockerContainers(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	// Start a container with some labels
	labels := map[string]string{
		label.Prefix + "frontend.rule": "Host:my.super.host",
		label.TraefikPort:              "2375",
	}
	s.startContainerWithNameAndLabels(c, "powpow", "swarm:1.0.0", labels, "manage", "token://blabla")

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	resp, err := try.ResponseUntilStatusCode(req, 1500*time.Millisecond, http.StatusOK)
	c.Assert(err, checker.IsNil)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var version map[string]interface{}

	c.Assert(json.Unmarshal(body, &version), checker.IsNil)
	c.Assert(version["Version"], checker.Equals, "swarm/1.0.0")

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/docker/backends", 60*time.Second, try.BodyContains("powpow"))
	c.Assert(err, checker.IsNil)

	s.stopAndRemoveContainerByName(c, "powpow")
	defer s.project.Remove(c, "powpow")

	time.Sleep(5 * time.Second)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/docker/backends", 10*time.Second, try.BodyContains("powpow"))
	c.Assert(err, checker.NotNil)

	s.startContainerWithNameAndLabels(c, "powpow", "swarm:1.0.0", labels, "manage", "token://blabla")

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/docker/backends", 60*time.Second, try.BodyContains("powpow"))
	c.Assert(err, checker.IsNil)
}
