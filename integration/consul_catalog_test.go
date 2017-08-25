package integration

import (
	"fmt"
	"net/http"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	"github.com/hashicorp/consul/api"
	checker "github.com/vdemeester/shakers"
)

// Consul catalog test suites
type ConsulCatalogSuite struct {
	BaseSuite
	consulIP     string
	consulClient *api.Client
}

func (s *ConsulCatalogSuite) SetUpSuite(c *check.C) {

	s.createComposeProject(c, "consul_catalog")
	s.composeProject.Start(c)

	consul := s.composeProject.Container(c, "consul")

	s.consulIP = consul.NetworkSettings.IPAddress
	config := api.DefaultConfig()
	config.Address = s.consulIP + ":8500"
	consulClient, err := api.NewClient(config)
	if err != nil {
		c.Fatalf("Error creating consul client. %v", err)
	}
	s.consulClient = consulClient

	// Wait for consul to elect itself leader
	err = try.Do(3*time.Second, func() error {
		leader, err := consulClient.Status().Leader()

		if err != nil || len(leader) == 0 {
			return fmt.Errorf("Leader not found. %v", err)
		}

		return nil
	})
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) registerService(name string, address string, port int, tags []string) error {
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
				Tags:    tags,
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
	cmd, _ := s.cmdTraefik(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.endpoint="+s.consulIP+":8500")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestSingleService(c *check.C) {
	cmd, _ := s.cmdTraefik(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	nginx := s.composeProject.Container(c, "nginx")

	err = s.registerService("test", nginx.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", nginx.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestExposedByDefaultFalseSingleService(c *check.C) {
	cmd, _ := s.cmdTraefik(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=false",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	nginx := s.composeProject.Container(c, "nginx")

	err = s.registerService("test", nginx.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", nginx.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestExposedByDefaultFalseSimpleServiceMultipleNode(c *check.C) {
	cmd, _ := s.cmdTraefik(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=false",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	nginx := s.composeProject.Container(c, "nginx")
	nginx2 := s.composeProject.Container(c, "nginx2")

	err = s.registerService("test", nginx.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", nginx.NetworkSettings.IPAddress)

	err = s.registerService("test", nginx2.NetworkSettings.IPAddress, 80, []string{"traefik.enable=true"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", nginx2.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestExposedByDefaultTrueSimpleServiceMultipleNode(c *check.C) {
	cmd, _ := s.cmdTraefik(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	nginx := s.composeProject.Container(c, "nginx")
	nginx2 := s.composeProject.Container(c, "nginx2")

	err = s.registerService("test", nginx.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", nginx.NetworkSettings.IPAddress)

	err = s.registerService("test", nginx2.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", nginx2.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}
