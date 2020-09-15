package integration

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/hashicorp/consul/api"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
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
			return fmt.Errorf("leader not found. %w", err)
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

func (s *ConsulCatalogSuite) registerService(reg *api.AgentServiceRegistration, onAgent bool) error {
	client := s.consulClient
	if onAgent {
		client = s.consulAgentClient
	}

	return client.Agent().ServiceRegister(reg)
}

func (s *ConsulCatalogSuite) deregisterService(id string, onAgent bool) error {
	client := s.consulClient
	if onAgent {
		client = s.consulAgentClient
	}
	return client.Agent().ServiceDeregister(id)
}

func (s *ConsulCatalogSuite) TestWithNotExposedByDefaultAndDefaultsSettings(c *check.C) {
	reg1 := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}
	err := s.registerService(reg1, false)
	c.Assert(err, checker.IsNil)

	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami2",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress,
	}
	err = s.registerService(reg2, false)
	c.Assert(err, checker.IsNil)

	reg3 := &api.AgentServiceRegistration{
		ID:      "whoami3",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.composeProject.Container(c, "whoami3").NetworkSettings.IPAddress,
	}
	err = s.registerService(reg3, false)
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

	err = try.Request(req, 2*time.Second,
		try.StatusCodeIs(200),
		try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.StatusCodeIs(200),
		try.BodyContains(
			fmt.Sprintf(`"http://%s:80":"UP"`, reg1.Address),
			fmt.Sprintf(`"http://%s:80":"UP"`, reg2.Address),
			fmt.Sprintf(`"http://%s:80":"UP"`, reg3.Address),
		))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami2", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami3", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestByLabels(c *check.C) {
	containerIP := s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress

	reg := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers.router1.rule=Path(`/whoami`)",
			"traefik.http.routers.router1.service=service1",
			"traefik.http.services.service1.loadBalancer.server.url=http://" + containerIP,
		},
		Port:    80,
		Address: containerIP,
	}
	err := s.registerService(reg, false)
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

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}
	err := s.registerService(reg, false)
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

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: "",
	}
	err := s.registerService(reg, false)
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

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Port:    80,
		Address: s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}
	err := s.registerService(reg, false)
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

	// Start a container with some tags
	reg := &api.AgentServiceRegistration{
		ID:   "whoamitcp",
		Name: "whoamitcp",
		Tags: []string{
			"traefik.tcp.Routers.Super.Rule=HostSNI(`my.super.host`)",
			"traefik.tcp.Routers.Super.tls=true",
			"traefik.tcp.Services.Super.Loadbalancer.server.port=8080",
		},
		Port:    8080,
		Address: s.composeProject.Container(c, "whoamitcp").NetworkSettings.IPAddress,
	}

	err := s.registerService(reg, false)
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

	// Start a container with some tags
	reg1 := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.http.Routers.Super.Rule=Host(`my.super.host`)",
		},
		Port:    80,
		Address: s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}

	err := s.registerService(reg1, false)
	c.Assert(err, checker.IsNil)

	// Start another container by replacing a '.' by a '-'
	reg2 := &api.AgentServiceRegistration{
		ID:   "whoami2",
		Name: "whoami",
		Tags: []string{
			"traefik.http.Routers.SuperHost.Rule=Host(`my-super.host`)",
		},
		Port:    80,
		Address: s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress,
	}
	err = s.registerService(reg2, false)
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

	// Start a container with some tags
	tags := []string{
		"traefik.enable=true",
		"traefik.http.Routers.Super.service=whoami",
		"traefik.http.Routers.Super.Rule=Host(`my.super.host`)",
	}

	reg1 := &api.AgentServiceRegistration{
		ID:      "whoami",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}
	err := s.registerService(reg1, false)
	c.Assert(err, checker.IsNil)

	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress,
	}
	err = s.registerService(reg2, true)
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

	// Start a container with some tags
	reg := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.random.value=my.super.host",
		},
		Port:    80,
		Address: s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}

	err := s.registerService(reg, false)
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

func (s *ConsulCatalogSuite) TestConsulServiceWithHealthCheck(c *check.C) {
	tags := []string{
		"traefik.enable=true",
		"traefik.http.routers.router1.rule=Path(`/whoami`)",
		"traefik.http.routers.router1.service=service1",
		"traefik.http.services.service1.loadBalancer.server.url=http://" + s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}

	reg1 := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
		Check: &api.AgentServiceCheck{
			CheckID:  "some-failed-check",
			TCP:      "127.0.0.1:1234",
			Name:     "some-failed-check",
			Interval: "1s",
			Timeout:  "1s",
		},
	}

	err := s.registerService(reg1, false)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulAddress,
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)

	containerIP := s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress

	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami2",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: containerIP,
		Check: &api.AgentServiceCheck{
			CheckID:  "some-ok-check",
			TCP:      containerIP + ":80",
			Name:     "some-ok-check",
			Interval: "1s",
			Timeout:  "1s",
		},
	}

	err = s.registerService(reg2, false)
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami"

	// FIXME Need to wait for up to 10 seconds (for consul discovery or traefik to boot up ?)
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami2"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami2", false)
	c.Assert(err, checker.IsNil)
}
