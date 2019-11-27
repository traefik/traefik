package integration

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/containous/traefik/v2/integration/try"
	"github.com/docker/docker/integration-cli/checker"
	"github.com/go-check/check"
	"github.com/hashicorp/consul/api"
)

type ConsulCatalogSuite struct {
	BaseSuite
	consulClient       *api.Client
	consulAgentClient  *api.Client
	consulAddress      string
	consulAgentAddress string
}

func (s *ConsulCatalogSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "consul_catalog")
	s.composeProject.Start(c)
	s.consulAddress = "http://" + s.composeProject.Container(c, "consul").NetworkSettings.IPAddress + ":8500"
	client, err := api.NewClient(&api.Config{
		Address: s.consulAddress,
	})
	c.Check(err, check.IsNil)
	s.consulClient = client

	// Wait for consul to elect itself leader
	err = s.waitToElectConsulLeader()
	c.Assert(err, checker.IsNil)

	s.consulAgentAddress = "http://" + s.composeProject.Container(c, "consul-agent").NetworkSettings.IPAddress + ":8500"
	clientAgent, err := api.NewClient(&api.Config{
		Address: s.consulAgentAddress,
	})
	c.Check(err, check.IsNil)
	s.consulAgentClient = clientAgent
}

func (s *ConsulCatalogSuite) waitToElectConsulLeader() error {
	return try.Do(15*time.Second, func() error {
		leader, err := s.consulClient.Status().Leader()

		if err != nil || len(leader) == 0 {
			return fmt.Errorf("leader not found. %v", err)
		}

		return nil
	})
}

func (s *ConsulCatalogSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *ConsulCatalogSuite) registerService(id, name, address, port string, tags []string, onAgent bool) error {
	iPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	client := s.consulClient
	if onAgent {
		client = s.consulAgentClient
	}

	return client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    iPort,
		Tags:    tags,
	})
}

func (s *ConsulCatalogSuite) deregisterService(id string, onAgent bool) error {
	client := s.consulClient
	if onAgent {
		client = s.consulAgentClient
	}
	return client.Agent().ServiceDeregister(id)
}

func (s *ConsulCatalogSuite) TestWithNotExposedByDefaultAndDefaultsSettings(c *check.C) {
	err := s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", []string{"traefik.enable=true"}, false)
	c.Assert(err, checker.IsNil)
	err = s.registerService("whoami2", "whoami", s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress, "80", []string{"traefik.enable=true"}, false)
	c.Assert(err, checker.IsNil)
	err = s.registerService("whoami3", "whoami", s.composeProject.Container(c, "whoami3").NetworkSettings.IPAddress, "80", []string{"traefik.enable=true"}, false)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulAddress,
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/default_not_exposed.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami2", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami3", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestByLabels(c *check.C) {
	labels := []string{
		"traefik.enable=true",
		"traefik.http.routers.router1.rule=Path(`/whoami`)",
		"traefik.http.routers.router1.service=service1",
		"traefik.http.services.service1.loadBalancer.server.url=http://" + s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}

	err := s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", labels, false)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulAddress,
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/default_not_exposed.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestSimpleConfiguration(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulAddress,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	err := s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", []string{"traefik.enable=true"}, false)
	c.Assert(err, checker.IsNil)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami.consul.localhost"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestRegisterServiceWithoutIP(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulAddress,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	err := s.registerService("whoami1", "whoami", "", "80", []string{"traefik.enable=true"}, false)
	c.Assert(err, checker.IsNil)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/http/services", nil)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("whoami@consulcatalog", "\"http://127.0.0.1:80\": \"UP\""))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestDefaultConsulService(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{

		ConsulAddress: s.consulAddress,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	err := s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", nil, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami.consul.localhost"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithTCPLabels(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulAddress,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	// Start a container with some labels
	labels := []string{
		"traefik.tcp.Routers.Super.Rule=HostSNI(`my.super.host`)",
		"traefik.tcp.Routers.Super.tls=true",
		"traefik.tcp.Services.Super.Loadbalancer.server.port=8080",
	}

	err := s.registerService("whoamitcp", "whoamitcp", s.composeProject.Container(c, "whoamitcp").NetworkSettings.IPAddress, "8080", labels, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`my.super.host`)"))
	c.Assert(err, checker.IsNil)

	who, err := guessWho("127.0.0.1:8000", "my.super.host", true)
	c.Assert(err, checker.IsNil)

	c.Assert(who, checker.Contains, "whoamitcp")

	err = s.deregisterService("whoamitcp", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithLabels(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulAddress,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	// Start a container with some labels
	labels := []string{
		"traefik.http.Routers.Super.Rule=Host(`my.super.host`)",
	}
	err := s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", labels, false)
	c.Assert(err, checker.IsNil)

	// Start another container by replacing a '.' by a '-'
	labels = []string{
		"traefik.http.Routers.SuperHost.Rule=Host(`my-super.host`)",
	}
	err = s.registerService("whoami2", "whoami", s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress, "80", labels, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my-super.host"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1"))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami2"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami2", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestSameServiceIDOnDifferentConsulAgent(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulAddress,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/default_not_exposed.toml", tempObjects)
	defer os.Remove(file)

	// Start a container with some labels
	labels := []string{
		"traefik.enable=true",
		"traefik.http.Routers.Super.service=whoami",
		"traefik.http.Routers.Super.Rule=Host(`my.super.host`)",
	}
	err := s.registerService("whoami", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", labels, false)
	c.Assert(err, checker.IsNil)

	err = s.registerService("whoami", "whoami", s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress, "80", labels, true)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2"))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/rawdata", nil)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200),
		try.BodyContainsOr(s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
			s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami2", true)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithOneMissingLabels(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulAddress,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	// Start a container with some labels
	labels := []string{
		"traefik.random.value=my.super.host",
	}
	err := s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", labels, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// FIXME Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.Request(req, 1500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}
