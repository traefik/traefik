package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/go-check/check"

	d "github.com/libkermit/docker"
	docker "github.com/libkermit/docker-check"
	checker "github.com/vdemeester/shakers"
)

var (
	// Label added to started container to identify them as part of the integration test
	TestLabel = "io.traefik.test"

	// Images to have or pull before the build in order to make it work
	// FIXME handle this offline but loading them before build
	RequiredImages = map[string]string{
		"swarm": "1.0.0",
		"nginx": "1",
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

func (s *DockerSuite) startContainerWithConfig(c *check.C, image string, config d.ContainerConfig) string {
	if config.Name == "" {
		config.Name = namesgenerator.GetRandomName(10)
	}

	container := s.project.StartWithConfig(c, image, config)

	// FIXME(vdemeester) this is ugly (it's because of the / in front of the name in docker..)
	return strings.SplitAfter(container.Name, "/")[1]
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

	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)
	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1:8000/")

	c.Assert(err, checker.IsNil)
	// Expected a 404 as we did not comfigure anything
	c.Assert(resp.StatusCode, checker.Equals, 404)
}

func (s *DockerSuite) TestDefaultDockerContainers(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	name := s.startContainer(c, "swarm:1.0.0", "manage", "token://blablabla")

	// Start traefik
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	time.Sleep(1500 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
        req.Host = fmt.Sprintf("%s.docker.localhost", strings.Replace(name, "_", "-", -1))
	resp, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

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
		"traefik.frontend.rule": "Host:my.super.host",
	}
	s.startContainerWithLabels(c, "swarm:1.0.0", labels, "manage", "token://blabla")

	// Start traefik
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	time.Sleep(1500 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = fmt.Sprintf("my.super.host")
	resp, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

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
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	time.Sleep(1500 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = fmt.Sprintf("my.super.host")
	resp, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}
