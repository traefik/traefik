package integration

import (
	"fmt"
	"net"
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
	consulClient      *api.Client
	consulAgentClient *api.Client
	consulURL         string
}

func (s *ConsulCatalogSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "consul_catalog")
	s.composeUp(c)

	s.consulURL = "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "consul"), "8500")

	var err error
	s.consulClient, err = api.NewClient(&api.Config{
		Address: s.consulURL,
	})
	c.Check(err, check.IsNil)

	// Wait for consul to elect itself leader
	err = s.waitToElectConsulLeader()
	c.Assert(err, checker.IsNil)

	s.consulAgentClient, err = api.NewClient(&api.Config{
		Address: "http://" + net.JoinHostPort(s.getComposeServiceIP(c, "consul-agent"), "8500"),
	})
	c.Check(err, check.IsNil)
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

func (s *ConsulCatalogSuite) waitForConnectCA() error {
	return try.Do(15*time.Second, func() error {
		caroots, _, err := s.consulClient.Connect().CARoots(nil)

		if err != nil || len(caroots.Roots) == 0 {
			return fmt.Errorf("connect CA not fully initialized. %w", err)
		}

		return nil
	})
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
		Address: s.getComposeServiceIP(c, "whoami1"),
	}
	err := s.registerService(reg1, false)
	c.Assert(err, checker.IsNil)

	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami2",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP(c, "whoami2"),
	}
	err = s.registerService(reg2, false)
	c.Assert(err, checker.IsNil)

	reg3 := &api.AgentServiceRegistration{
		ID:      "whoami3",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP(c, "whoami3"),
	}
	err = s.registerService(reg3, false)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/default_not_exposed.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

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
	containerIP := s.getComposeServiceIP(c, "whoami1")

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
		ConsulAddress: s.consulURL,
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/default_not_exposed.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestSimpleConfiguration(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP(c, "whoami1"),
	}
	err := s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami.consul.localhost"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestSimpleConfigurationWithWatch(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple_watch.toml", tempObjects)
	defer os.Remove(file)

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP(c, "whoami1"),
	}
	err := s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami.consul.localhost"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	whoamiIP := s.getComposeServiceIP(c, "whoami1")
	reg.Check = &api.AgentServiceCheck{
		CheckID:  "some-ok-check",
		TCP:      whoamiIP + ":80",
		Name:     "some-ok-check",
		Interval: "1s",
		Timeout:  "1s",
	}

	err = s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1"))
	c.Assert(err, checker.IsNil)

	reg.Check = &api.AgentServiceCheck{
		CheckID:  "some-failing-check",
		TCP:      ":80",
		Name:     "some-failing-check",
		Interval: "1s",
		Timeout:  "1s",
	}

	err = s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestRegisterServiceWithoutIP(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
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
	defer s.killCmd(cmd)

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
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Port:    80,
		Address: s.getComposeServiceIP(c, "whoami1"),
	}
	err := s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

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
		ConsulAddress: s.consulURL,
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
		Address: s.getComposeServiceIP(c, "whoamitcp"),
	}

	err := s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

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
		ConsulAddress: s.consulURL,
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
		Address: s.getComposeServiceIP(c, "whoami1"),
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
		Address: s.getComposeServiceIP(c, "whoami2"),
	}
	err = s.registerService(reg2, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

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
		ConsulAddress: s.consulURL,
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
		Address: s.getComposeServiceIP(c, "whoami1"),
	}
	err := s.registerService(reg1, false)
	c.Assert(err, checker.IsNil)

	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: s.getComposeServiceIP(c, "whoami2"),
	}
	err = s.registerService(reg2, true)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2"))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/rawdata", nil)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200),
		try.BodyContainsOr(s.getComposeServiceIP(c, "whoami1"), s.getComposeServiceIP(c, "whoami2")))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami", false)
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami", true)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithOneMissingLabels(c *check.C) {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
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
		Address: s.getComposeServiceIP(c, "whoami1"),
	}

	err := s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	// Start traefik
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "my.super.host"

	// TODO Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.Request(req, 1500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithHealthCheck(c *check.C) {
	whoamiIP := s.getComposeServiceIP(c, "whoami1")
	tags := []string{
		"traefik.enable=true",
		"traefik.http.routers.router1.rule=Path(`/whoami`)",
		"traefik.http.routers.router1.service=service1",
		"traefik.http.services.service1.loadBalancer.server.url=http://" + whoamiIP,
	}

	reg1 := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: whoamiIP,
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
		ConsulAddress: s.consulURL,
	}

	file := s.adaptFile(c, "fixtures/consul_catalog/simple.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)

	whoami2IP := s.getComposeServiceIP(c, "whoami2")
	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami2",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: whoami2IP,
		Check: &api.AgentServiceCheck{
			CheckID:  "some-ok-check",
			TCP:      whoami2IP + ":80",
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

	// TODO Need to wait for up to 10 seconds (for consul discovery or traefik to boot up ?)
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami2"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami2", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulConnect(c *check.C) {
	// Wait for consul to fully initialize connect CA
	err := s.waitForConnectCA()
	c.Assert(err, checker.IsNil)

	connectIP := s.getComposeServiceIP(c, "connect")
	reg := &api.AgentServiceRegistration{
		ID:   "uuid-api1",
		Name: "uuid-api",
		Tags: []string{
			"traefik.enable=true",
			"traefik.consulcatalog.connect=true",
			"traefik.http.routers.router1.rule=Path(`/`)",
			"traefik.http.routers.router1.service=service1",
			"traefik.http.services.service1.loadBalancer.server.url=https://" + connectIP,
		},
		Connect: &api.AgentServiceConnect{
			Native: true,
		},
		Port:    443,
		Address: connectIP,
	}
	err = s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	whoamiIP := s.getComposeServiceIP(c, "whoami1")
	regWhoami := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers.router2.rule=Path(`/whoami`)",
			"traefik.http.routers.router2.service=whoami",
		},
		Port:    80,
		Address: whoamiIP,
	}
	err = s.registerService(regWhoami, false)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}
	file := s.adaptFile(c, "fixtures/consul_catalog/connect.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8000/", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("uuid-api1", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulConnect_ByDefault(c *check.C) {
	// Wait for consul to fully initialize connect CA
	err := s.waitForConnectCA()
	c.Assert(err, checker.IsNil)

	connectIP := s.getComposeServiceIP(c, "connect")
	reg := &api.AgentServiceRegistration{
		ID:   "uuid-api1",
		Name: "uuid-api",
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers.router1.rule=Path(`/`)",
			"traefik.http.routers.router1.service=service1",
			"traefik.http.services.service1.loadBalancer.server.url=https://" + connectIP,
		},
		Connect: &api.AgentServiceConnect{
			Native: true,
		},
		Port:    443,
		Address: connectIP,
	}
	err = s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	whoamiIP := s.getComposeServiceIP(c, "whoami1")
	regWhoami := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami1",
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers.router2.rule=Path(`/whoami`)",
			"traefik.http.routers.router2.service=whoami",
		},
		Port:    80,
		Address: whoamiIP,
	}
	err = s.registerService(regWhoami, false)
	c.Assert(err, checker.IsNil)

	whoami2IP := s.getComposeServiceIP(c, "whoami2")
	regWhoami2 := &api.AgentServiceRegistration{
		ID:   "whoami2",
		Name: "whoami2",
		Tags: []string{
			"traefik.enable=true",
			"traefik.consulcatalog.connect=false",
			"traefik.http.routers.router2.rule=Path(`/whoami2`)",
			"traefik.http.routers.router2.service=whoami2",
		},
		Port:    80,
		Address: whoami2IP,
	}
	err = s.registerService(regWhoami2, false)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}
	file := s.adaptFile(c, "fixtures/consul_catalog/connect_by_default.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8000/", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 10*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami2", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("uuid-api1", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami2", false)
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestConsulConnect_NotAware(c *check.C) {
	// Wait for consul to fully initialize connect CA
	err := s.waitForConnectCA()
	c.Assert(err, checker.IsNil)

	connectIP := s.getComposeServiceIP(c, "connect")
	reg := &api.AgentServiceRegistration{
		ID:   "uuid-api1",
		Name: "uuid-api",
		Tags: []string{
			"traefik.enable=true",
			"traefik.consulcatalog.connect=true",
			"traefik.http.routers.router1.rule=Path(`/`)",
			"traefik.http.routers.router1.service=service1",
			"traefik.http.services.service1.loadBalancer.server.url=https://" + connectIP,
		},
		Connect: &api.AgentServiceConnect{
			Native: true,
		},
		Port:    443,
		Address: connectIP,
	}
	err = s.registerService(reg, false)
	c.Assert(err, checker.IsNil)

	whoamiIP := s.getComposeServiceIP(c, "whoami1")
	regWhoami := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers.router2.rule=Path(`/whoami`)",
			"traefik.http.routers.router2.service=whoami",
		},
		Port:    80,
		Address: whoamiIP,
	}
	err = s.registerService(regWhoami, false)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}
	file := s.adaptFile(c, "fixtures/consul_catalog/connect_not_aware.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8000/", 10*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("uuid-api1", false)
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami1", false)
	c.Assert(err, checker.IsNil)
}
