package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/opts"
	"github.com/fsouza/go-dockerclient"
	"github.com/hashicorp/consul/api"
	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
	"strings"
)

// Consul catalog test suites
type ConsulCatalogSuite struct {
	BaseSuite
	consulIP     string
	consulClient *api.Client
	consulKV     *api.KV
	dockerClient *docker.Client
}

func (s *ConsulCatalogSuite) GetContainer(name string) (*docker.Container, error) {
	return s.dockerClient.InspectContainer(name)
}

func (s *ConsulCatalogSuite) SetUpSuite(c *check.C) {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		// FIXME Handle windows -- see if dockerClient already handle that or not
		dockerHost = fmt.Sprintf("unix://%s", opts.DefaultUnixSocket)
	}
	// Make sure we can speak to docker
	dockerClient, err := docker.NewClient(dockerHost)
	c.Assert(err, checker.IsNil, check.Commentf("Error connecting to docker daemon"))
	s.dockerClient = dockerClient

	s.createComposeProject(c, "consul_catalog")
	err = s.composeProject.Up()
	c.Assert(err, checker.IsNil, check.Commentf("Error starting project"))

	consul, err := s.GetContainer("integration-test-consul_catalog_consul_1")
	c.Assert(err, checker.IsNil, check.Commentf("Error finding consul container"))

	s.consulIP = consul.NetworkSettings.IPAddress
	config := api.DefaultConfig()
	config.Address = s.consulIP + ":8500"
	consulClient, err := api.NewClient(config)
	if err != nil {
		c.Fatalf("Error creating consul client")
	}
	s.consulClient = consulClient
	s.consulKV = s.consulClient.KV()

	// Wait for consul to elect itself leader
	time.Sleep(2000 * time.Millisecond)
}

func (s *ConsulCatalogSuite) registerService(name string, address string, port int) error {
	catalog := s.consulClient.Catalog()
	_, err := catalog.Register(
		&api.CatalogRegistration{
			Node:    address,
			Address: address,
			Service: &api.AgentService{
				ID:      name,
				Service: name,
				Address: address,
				Port:    port,
			},
		},
		&api.WriteOptions{},
	)
	return err
}

func (s *ConsulCatalogSuite) deregisterService(name string, address string) error {
	catalog := s.consulClient.Catalog()
	_, err := catalog.Deregister(
		&api.CatalogDeregistration{
			Node:      address,
			Address:   address,
			ServiceID: name,
		},
		&api.WriteOptions{},
	)
	return err
}

func (s *ConsulCatalogSuite) TestSimpleConfiguration(c *check.C) {
	cmd := exec.Command(traefikBinary, "--consulCatalog", "--consulCatalog.endpoint="+s.consulIP+":8500", "--configFile=fixtures/consul_catalog/simple.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)
	// TODO validate : run on 80
	resp, err := http.Get("http://127.0.0.1:8000/")

	// Expected a 404 as we did not configure anything
	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 404)
}

func (s *ConsulCatalogSuite) TestSingleService(c *check.C) {
	cmd := exec.Command(traefikBinary, "--consulCatalog", "--consulCatalog.endpoint="+s.consulIP+":8500", "--consulCatalog.domain=consul.localhost", "--consulCatalog.prefix=/traefik", "--configFile=fixtures/consul_catalog/simple.toml")

	outfile, _ := os.Create("/tmp/stderr.txt")
	defer outfile.Close()
	defer os.Remove(outfile.Name())
	cmd.Stderr = outfile

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	nginx, err := s.GetContainer("integration-test-consul_catalog_nginx_1")
	c.Assert(err, checker.IsNil, check.Commentf("Error finding nginx container"))

	err = s.registerService("test", nginx.NetworkSettings.IPAddress, 80)
	s.consulKV.Put(&api.KVPair{Key: "traefik/test/circuitbreaker", Value: []byte("NetworkErrorRatio() > 4.2")}, &api.WriteOptions{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", nginx.NetworkSettings.IPAddress)
	defer s.consulKV.Delete("traefik/test/circuitbreaker", &api.WriteOptions{})


	time.Sleep(5000 * time.Millisecond)

	f, _ := os.Open(outfile.Name())
	str, _ := ioutil.ReadAll(f)
	c.Assert(strings.Contains(string(str), "Creating circuit breaker NetworkErrorRatio() > 4.2"), checker.Equals, true)
	f.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"
	resp, err := client.Do(req)

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, 200)

	_, err = ioutil.ReadAll(resp.Body)
	c.Assert(err, checker.IsNil)
}
