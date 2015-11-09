package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/fsouza/go-dockerclient"
	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
)

var (
	// Label added to started container to identify them as part of the integration test
	TestLabel = "io.traefik.test"

	// Images to have or pull before the build in order to make it work
	// FIXME handle this offline but loading them before build
	RequiredImages = map[string]string{
		"emilevauge/whoami": "latest",
		"nginx":             "1",
	}
)

// Docker test suites
type DockerSuite struct {
	BaseSuite
	client *docker.Client
}

func (s *DockerSuite) startContainer(c *check.C, image string, args ...string) string {
	return s.startContainerWithConfig(c, docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
			Cmd:   args,
		},
	})
}

func (s *DockerSuite) startContainerWithLabels(c *check.C, image string, labels map[string]string, args ...string) string {
	return s.startContainerWithConfig(c, docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:  image,
			Cmd:    args,
			Labels: labels,
		},
	})
}

func (s *DockerSuite) startContainerWithConfig(c *check.C, config docker.CreateContainerOptions) string {
	if config.Name == "" {
		config.Name = namesgenerator.GetRandomName(10)
	}
	if config.Config.Labels == nil {
		config.Config.Labels = map[string]string{}
	}
	config.Config.Labels[TestLabel] = "true"

	container, err := s.client.CreateContainer(config)
	c.Assert(err, checker.IsNil, check.Commentf("Error creating a container using config %v", config))

	err = s.client.StartContainer(container.ID, &docker.HostConfig{})
	c.Assert(err, checker.IsNil, check.Commentf("Error starting container %v", container))

	return container.Name
}

func (s *DockerSuite) SetUpSuite(c *check.C) {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		// FIXME Handle windows -- see if dockerClient already handle that or not
		dockerHost = fmt.Sprintf("unix://%s", opts.DefaultUnixSocket)
	}
	// Make sure we can speak to docker
	dockerClient, err := docker.NewClient(dockerHost)
	c.Assert(err, checker.IsNil, check.Commentf("Error connecting to docker daemon"))

	s.client = dockerClient
	c.Assert(s.client.Ping(), checker.IsNil)

	// Pull required images
	for repository, tag := range RequiredImages {
		image := fmt.Sprintf("%s:%s", repository, tag)
		_, err := s.client.InspectImage(image)
		if err != nil {
			if err != docker.ErrNoSuchImage {
				c.Fatalf("Error while inspect image %s", image)
			}
			err = s.client.PullImage(docker.PullImageOptions{
				Repository: repository,
				Tag:        tag,
			}, docker.AuthConfiguration{})
			c.Assert(err, checker.IsNil, check.Commentf("Error while pulling image %s", image))
		}
	}
}

func (s *DockerSuite) cleanContainers(c *check.C) {
	// Clean the mess, a.k.a. the running containers with the right label
	containerList, err := s.client.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{
			"label": {fmt.Sprintf("%s=true", TestLabel)},
		},
	})
	c.Assert(err, checker.IsNil, check.Commentf("Error listing containers started by traefik"))

	for _, container := range containerList {
		err = s.client.KillContainer(docker.KillContainerOptions{
			ID: container.ID,
		})
		c.Assert(err, checker.IsNil, check.Commentf("Error killing container %v", container))
		if os.Getenv("CIRCLECI") == "" {
			// On circleci, we won't delete them — it errors out for now >_<
			err = s.client.RemoveContainer(docker.RemoveContainerOptions{
				ID:            container.ID,
				RemoveVolumes: true,
			})
			c.Assert(err, checker.IsNil, check.Commentf("Error removing container %v", container))
		}
	}
}

func (s *DockerSuite) TearDownTest(c *check.C) {
	s.cleanContainers(c)
}

func (s *DockerSuite) TearDownSuite(c *check.C) {
	// Call cleanContainers, just in case (?)
	// s.cleanContainers(c)
}

func (s *DockerSuite) TestSimpleConfiguration(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)

	cmd := exec.Command(traefikBinary, file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)
	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1/")

	c.Assert(err, checker.IsNil)
	// Expected a 404 as we did not comfigure anything
	c.Assert(resp.StatusCode, checker.Equals, 404)
}

func (s *DockerSuite) TestDefaultDockerContainers(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	name := s.startContainer(c, "emilevauge/whoami")

	// Start traefik
	cmd := exec.Command(traefikBinary, file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	time.Sleep(1500 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1/api", nil)
	c.Assert(err, checker.IsNil)
	req.Host = fmt.Sprintf("%s.docker.localhost", name)
	resp, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var result map[string]interface{}

	c.Assert(json.Unmarshal(body, &result), checker.IsNil)
	c.Assert(result["hostname"], checker.Not(checker.IsNil))
}

func (s *DockerSuite) TestDockerContainersWithLabels(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	// Start a container with some labels
	labels := map[string]string{
		"traefik.frontend.rule":  "Host",
		"traefik.frontend.value": "my.super.host",
	}
	s.startContainerWithLabels(c, "emilevauge/whoami", labels)

	// Start traefik
	cmd := exec.Command(traefikBinary, file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// FIXME Need to wait than 500 milliseconds more (for whoami or traefik to boot up ?)
	time.Sleep(1500 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1/api", nil)
	c.Assert(err, checker.IsNil)
	req.Host = fmt.Sprintf("my.super.host")
	resp, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)

	var result map[string]interface{}

	c.Assert(json.Unmarshal(body, &result), checker.IsNil)
	c.Assert(result["hostname"], checker.Not(checker.IsNil))
}

func (s *DockerSuite) TestDockerContainersWithOneMissingLabels(c *check.C) {
	file := s.adaptFileForHost(c, "fixtures/docker/simple.toml")
	defer os.Remove(file)
	// Start a container with some labels
	labels := map[string]string{
		"traefik.frontend.value": "my.super.host",
	}
	s.startContainerWithLabels(c, "emilevauge/whoami", labels)

	// Start traefik
	cmd := exec.Command(traefikBinary, file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// FIXME Need to wait than 500 milliseconds more (for whoami or traefik to boot up ?)
	time.Sleep(1500 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1/api", nil)
	c.Assert(err, checker.IsNil)
	req.Host = fmt.Sprintf("my.super.host")
	resp, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}
