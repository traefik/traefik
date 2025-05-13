package integration

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type ConsulCatalogSuite struct {
	BaseSuite
	consulClient      *api.Client
	consulAgentClient *api.Client
	consulURL         string
}

func TestConsulCatalogSuite(t *testing.T) {
	suite.Run(t, new(ConsulCatalogSuite))
}

func (s *ConsulCatalogSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("consul_catalog")
	s.composeUp()

	s.consulURL = "http://" + net.JoinHostPort(s.getComposeServiceIP("consul"), "8500")

	var err error
	s.consulClient, err = api.NewClient(&api.Config{
		Address: s.consulURL,
	})
	require.NoError(s.T(), err)

	// Wait for consul to elect itself leader
	err = s.waitToElectConsulLeader()
	require.NoError(s.T(), err)

	s.consulAgentClient, err = api.NewClient(&api.Config{
		Address: "http://" + net.JoinHostPort(s.getComposeServiceIP("consul-agent"), "8500"),
	})
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
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

func (s *ConsulCatalogSuite) TestWithNotExposedByDefaultAndDefaultsSettings() {
	reg1 := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP("whoami1"),
	}
	err := s.registerService(reg1, false)
	require.NoError(s.T(), err)

	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami2",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP("whoami2"),
	}
	err = s.registerService(reg2, false)
	require.NoError(s.T(), err)

	reg3 := &api.AgentServiceRegistration{
		ID:      "whoami3",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP("whoami3"),
	}
	err = s.registerService(reg3, false)
	require.NoError(s.T(), err)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}

	file := s.adaptFile("fixtures/consul_catalog/default_not_exposed.toml", tempObjects)

	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "whoami"

	err = try.Request(req, 2*time.Second,
		try.StatusCodeIs(200),
		try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second,
		try.StatusCodeIs(200),
		try.BodyContains(
			fmt.Sprintf(`"http://%s:80":"UP"`, reg1.Address),
			fmt.Sprintf(`"http://%s:80":"UP"`, reg2.Address),
			fmt.Sprintf(`"http://%s:80":"UP"`, reg3.Address),
		))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
	err = s.deregisterService("whoami2", false)
	require.NoError(s.T(), err)
	err = s.deregisterService("whoami3", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestByLabels() {
	containerIP := s.getComposeServiceIP("whoami1")

	reg := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers.router1.rule=Path(`/whoami`)",
		},
		Port:    80,
		Address: containerIP,
	}
	err := s.registerService(reg, false)
	require.NoError(s.T(), err)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}

	file := s.adaptFile("fixtures/consul_catalog/default_not_exposed.toml", tempObjects)

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestSimpleConfiguration() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/simple.toml", tempObjects)

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP("whoami1"),
	}
	err := s.registerService(reg, false)
	require.NoError(s.T(), err)

	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "whoami.consul.localhost"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1"))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestSimpleConfigurationWithWatch() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/simple_watch.toml", tempObjects)

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: s.getComposeServiceIP("whoami1"),
	}
	err := s.registerService(reg, false)
	require.NoError(s.T(), err)

	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "whoami.consul.localhost"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1"))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	whoamiIP := s.getComposeServiceIP("whoami1")
	reg.Check = &api.AgentServiceCheck{
		CheckID:  "some-ok-check",
		TCP:      whoamiIP + ":80",
		Name:     "some-ok-check",
		Interval: "1s",
		Timeout:  "1s",
	}

	err = s.registerService(reg, false)
	require.NoError(s.T(), err)

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1"))
	require.NoError(s.T(), err)

	reg.Check = &api.AgentServiceCheck{
		CheckID:  "some-failing-check",
		TCP:      ":80",
		Name:     "some-failing-check",
		Interval: "1s",
		Timeout:  "1s",
	}

	err = s.registerService(reg, false)
	require.NoError(s.T(), err)

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestRegisterServiceWithoutIP() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/simple.toml", tempObjects)

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Tags:    []string{"traefik.enable=true"},
		Port:    80,
		Address: "",
	}
	err := s.registerService(reg, false)
	require.NoError(s.T(), err)

	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/http/services", nil)
	require.NoError(s.T(), err)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("whoami@consulcatalog", "\"http://127.0.0.1:80\": \"UP\""))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestDefaultConsulService() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/simple.toml", tempObjects)

	reg := &api.AgentServiceRegistration{
		ID:      "whoami1",
		Name:    "whoami",
		Port:    80,
		Address: s.getComposeServiceIP("whoami1"),
	}
	err := s.registerService(reg, false)
	require.NoError(s.T(), err)

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "whoami.consul.localhost"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1"))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithTCPLabels() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/simple.toml", tempObjects)

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
		Address: s.getComposeServiceIP("whoamitcp"),
	}

	err := s.registerService(reg, false)
	require.NoError(s.T(), err)

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`my.super.host`)"))
	require.NoError(s.T(), err)

	who, err := guessWho("127.0.0.1:8000", "my.super.host", true)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), who, "whoamitcp")

	err = s.deregisterService("whoamitcp", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithLabels() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/simple.toml", tempObjects)

	// Start a container with some tags
	reg1 := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.http.Routers.Super.Rule=Host(`my.super.host`)",
		},
		Port:    80,
		Address: s.getComposeServiceIP("whoami1"),
	}

	err := s.registerService(reg1, false)
	require.NoError(s.T(), err)

	// Start another container by replacing a '.' by a '-'
	reg2 := &api.AgentServiceRegistration{
		ID:   "whoami2",
		Name: "whoami",
		Tags: []string{
			"traefik.http.Routers.SuperHost.Rule=Host(`my-super.host`)",
		},
		Port:    80,
		Address: s.getComposeServiceIP("whoami2"),
	}
	err = s.registerService(reg2, false)
	require.NoError(s.T(), err)

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "my-super.host"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1"))
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "my.super.host"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami2"))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami2", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestSameServiceIDOnDifferentConsulAgent() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/default_not_exposed.toml", tempObjects)

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
		Address: s.getComposeServiceIP("whoami1"),
	}
	err := s.registerService(reg1, false)
	require.NoError(s.T(), err)

	reg2 := &api.AgentServiceRegistration{
		ID:      "whoami",
		Name:    "whoami",
		Tags:    tags,
		Port:    80,
		Address: s.getComposeServiceIP("whoami2"),
	}
	err = s.registerService(reg2, true)
	require.NoError(s.T(), err)

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	require.NoError(s.T(), err)
	req.Host = "my.super.host"

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2"))
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/rawdata", nil)
	require.NoError(s.T(), err)

	err = try.Request(req, 2*time.Second, try.StatusCodeIs(200),
		try.BodyContainsOr(s.getComposeServiceIP("whoami1"), s.getComposeServiceIP("whoami2")))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami", false)
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami", true)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithOneMissingLabels() {
	tempObjects := struct {
		ConsulAddress string
		DefaultRule   string
	}{
		ConsulAddress: s.consulURL,
		DefaultRule:   "Host(`{{ normalize .Name }}.consul.localhost`)",
	}

	file := s.adaptFile("fixtures/consul_catalog/simple.toml", tempObjects)

	// Start a container with some tags
	reg := &api.AgentServiceRegistration{
		ID:   "whoami1",
		Name: "whoami",
		Tags: []string{
			"traefik.random.value=my.super.host",
		},
		Port:    80,
		Address: s.getComposeServiceIP("whoami1"),
	}

	err := s.registerService(reg, false)
	require.NoError(s.T(), err)

	// Start traefik
	s.traefikCmd(withConfigFile(file))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/version", nil)
	require.NoError(s.T(), err)
	req.Host = "my.super.host"

	// TODO Need to wait than 500 milliseconds more (for swarm or traefik to boot up ?)
	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.Request(req, 1500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestConsulServiceWithHealthCheck() {
	whoamiIP := s.getComposeServiceIP("whoami1")
	tags := []string{
		"traefik.enable=true",
		"traefik.http.routers.router1.rule=Path(`/whoami`)",
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
	require.NoError(s.T(), err)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}

	file := s.adaptFile("fixtures/consul_catalog/simple.toml", tempObjects)

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)

	whoami2IP := s.getComposeServiceIP("whoami2")
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
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/whoami", nil)
	require.NoError(s.T(), err)
	req.Host = "whoami"

	// TODO Need to wait for up to 10 seconds (for consul discovery or traefik to boot up ?)
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(200), try.BodyContainsOr("Hostname: whoami2"))
	require.NoError(s.T(), err)

	err = s.deregisterService("whoami2", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestConsulConnect() {
	// Wait for consul to fully initialize connect CA
	err := s.waitForConnectCA()
	require.NoError(s.T(), err)

	connectIP := s.getComposeServiceIP("connect")
	reg := &api.AgentServiceRegistration{
		ID:   "uuid-api1",
		Name: "uuid-api",
		Tags: []string{
			"traefik.enable=true",
			"traefik.consulcatalog.connect=true",
			"traefik.http.routers.router1.rule=Path(`/`)",
		},
		Connect: &api.AgentServiceConnect{
			Native: true,
		},
		Port:    443,
		Address: connectIP,
	}
	err = s.registerService(reg, false)
	require.NoError(s.T(), err)

	whoamiIP := s.getComposeServiceIP("whoami1")
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
	require.NoError(s.T(), err)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}
	file := s.adaptFile("fixtures/consul_catalog/connect.toml", tempObjects)

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8000/", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = s.deregisterService("uuid-api1", false)
	require.NoError(s.T(), err)
	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestConsulConnect_ByDefault() {
	// Wait for consul to fully initialize connect CA
	err := s.waitForConnectCA()
	require.NoError(s.T(), err)

	connectIP := s.getComposeServiceIP("connect")
	reg := &api.AgentServiceRegistration{
		ID:   "uuid-api1",
		Name: "uuid-api",
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers.router1.rule=Path(`/`)",
		},
		Connect: &api.AgentServiceConnect{
			Native: true,
		},
		Port:    443,
		Address: connectIP,
	}
	err = s.registerService(reg, false)
	require.NoError(s.T(), err)

	whoamiIP := s.getComposeServiceIP("whoami1")
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
	require.NoError(s.T(), err)

	whoami2IP := s.getComposeServiceIP("whoami2")
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
	require.NoError(s.T(), err)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}
	file := s.adaptFile("fixtures/consul_catalog/connect_by_default.toml", tempObjects)

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8000/", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 10*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami2", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = s.deregisterService("uuid-api1", false)
	require.NoError(s.T(), err)
	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
	err = s.deregisterService("whoami2", false)
	require.NoError(s.T(), err)
}

func (s *ConsulCatalogSuite) TestConsulConnect_NotAware() {
	// Wait for consul to fully initialize connect CA
	err := s.waitForConnectCA()
	require.NoError(s.T(), err)

	connectIP := s.getComposeServiceIP("connect")
	reg := &api.AgentServiceRegistration{
		ID:   "uuid-api1",
		Name: "uuid-api",
		Tags: []string{
			"traefik.enable=true",
			"traefik.consulcatalog.connect=true",
			"traefik.http.routers.router1.rule=Path(`/`)",
		},
		Connect: &api.AgentServiceConnect{
			Native: true,
		},
		Port:    443,
		Address: connectIP,
	}
	err = s.registerService(reg, false)
	require.NoError(s.T(), err)

	whoamiIP := s.getComposeServiceIP("whoami1")
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
	require.NoError(s.T(), err)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.consulURL,
	}
	file := s.adaptFile("fixtures/consul_catalog/connect_not_aware.toml", tempObjects)

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8000/", 10*time.Second, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	err = s.deregisterService("uuid-api1", false)
	require.NoError(s.T(), err)
	err = s.deregisterService("whoami1", false)
	require.NoError(s.T(), err)
}
