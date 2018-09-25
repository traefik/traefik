package integration

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/provider/label"
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
	s.createConsulClient(config, c)

	// Wait for consul to elect itself leader
	err := s.waitToElectConsulLeader()
	c.Assert(err, checker.IsNil)

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
func (s *ConsulCatalogSuite) createConsulClient(config *api.Config, c *check.C) *api.Client {
	consulClient, err := api.NewClient(config)
	if err != nil {
		c.Fatalf("Error creating consul client. %v", err)
	}
	s.consulClient = consulClient
	return consulClient
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

func (s *ConsulCatalogSuite) registerAgentService(name string, address string, port int, tags []string, withHealthCheck bool) error {
	agent := s.consulClient.Agent()
	var healthCheck *api.AgentServiceCheck
	if withHealthCheck {
		healthCheck = &api.AgentServiceCheck{
			HTTP:     "http://" + address,
			Interval: "10s",
		}
	} else {
		healthCheck = nil
	}
	return agent.ServiceRegister(
		&api.AgentServiceRegistration{
			ID:      address,
			Tags:    tags,
			Name:    name,
			Address: address,
			Port:    port,
			Check:   healthCheck,
		},
	)
}

func (s *ConsulCatalogSuite) registerCheck(name string, address string, port int) error {
	agent := s.consulClient.Agent()
	checkRegistration := &api.AgentCheckRegistration{
		ID:        fmt.Sprintf("%s-%s", name, address),
		Name:      name,
		ServiceID: address,
	}
	checkRegistration.HTTP = fmt.Sprintf("http://%s:%d/health", address, port)
	checkRegistration.Interval = "2s"
	checkRegistration.CheckID = address
	return agent.CheckRegister(checkRegistration)
}

func (s *ConsulCatalogSuite) deregisterAgentService(address string) error {
	agent := s.consulClient.Agent()
	return agent.ServiceDeregister(address)
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

func (s *ConsulCatalogSuite) consulEnableServiceMaintenance(name string) error {
	return s.consulClient.Agent().EnableServiceMaintenance(name, fmt.Sprintf("Maintenance mode for service %s", name))
}

func (s *ConsulCatalogSuite) consulDisableServiceMaintenance(name string) error {
	return s.consulClient.Agent().DisableServiceMaintenance(name)
}

func (s *ConsulCatalogSuite) consulEnableNodeMaintenance() error {
	return s.consulClient.Agent().EnableNodeMaintenance("Maintenance mode for node")
}

func (s *ConsulCatalogSuite) consulDisableNodeMaintenance() error {
	return s.consulClient.Agent().DisableNodeMaintenance()
}

func (s *ConsulCatalogSuite) TestSimpleConfiguration(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.endpoint="+s.consulIP+":8500")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// TODO validate : run on 80
	// Expected a 404 as we did not configure anything
	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestSingleService(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	whoami := s.composeProject.Container(c, "whoami1")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	s.deregisterService("test", whoami.NetworkSettings.IPAddress)
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestExposedByDefaultFalseSingleService(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=false",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestExposedByDefaultFalseSimpleServiceMultipleNode(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=false",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")
	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	whoami2 := s.composeProject.Container(c, "whoami2")
	err = s.registerService("test", whoami2.NetworkSettings.IPAddress, 80, []string{label.TraefikEnable + "=true"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami2.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestExposedByDefaultTrueSimpleServiceMultipleNode(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")
	whoami2 := s.composeProject.Container(c, "whoami2")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{"name=whoami1"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	err = s.registerService("test", whoami2.NetworkSettings.IPAddress, 80, []string{"name=whoami2"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami2.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 60*time.Second,
		try.BodyContains(whoami.NetworkSettings.IPAddress, whoami2.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestRefreshConfigWithMultipleNodeWithoutHealthCheck(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")
	whoami2 := s.composeProject.Container(c, "whoami2")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{"name=whoami1"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	err = s.registerAgentService("test", whoami.NetworkSettings.IPAddress, 80, []string{"name=whoami1"}, true)
	c.Assert(err, checker.IsNil, check.Commentf("Error registering agent service"))
	defer s.deregisterAgentService(whoami.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 60*time.Second,
		try.BodyContains(whoami.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)

	err = s.registerService("test", whoami2.NetworkSettings.IPAddress, 80, []string{"name=whoami2"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 60*time.Second,
		try.BodyContains(whoami.NetworkSettings.IPAddress, whoami2.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)

	s.deregisterService("test", whoami2.NetworkSettings.IPAddress)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 60*time.Second,
		try.BodyContains(whoami.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)

	err = s.registerService("test", whoami2.NetworkSettings.IPAddress, 80, []string{"name=whoami2"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami2.NetworkSettings.IPAddress)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 60*time.Second,
		try.BodyContains(whoami.NetworkSettings.IPAddress, whoami2.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestBasicAuthSimpleService(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{
		label.TraefikFrontendAuthBasic + "=test:$2a$06$O5NksJPAcgrC9MuANkSoE.Xe9DSg7KcLLFYNr1Lj6hPcMmvgwxhme,test2:$2y$10$xP1SZ70QbZ4K2bTGKJOhpujkpcLxQcB3kEPF6XAV19IdcqsZTyDEe",
	})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusUnauthorized), try.HasBody())
	c.Assert(err, checker.IsNil)

	req.SetBasicAuth("test", "test")
	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	req.SetBasicAuth("test2", "test2")
	err = try.Request(req, 5*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestRefreshConfigTagChange(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=false",
		"--consulCatalog.watch=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80,
		[]string{"name=whoami1", label.TraefikEnable + "=false", label.TraefikBackendCircuitBreakerExpression + "=NetworkErrorRatio() > 0.5"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 5*time.Second,
		try.BodyContains(whoami.NetworkSettings.IPAddress))
	c.Assert(err, checker.NotNil)

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80,
		[]string{"name=whoami1", label.TraefikEnable + "=true", label.TraefikBackendCircuitBreakerExpression + "=ResponseCodeRatio(500, 600, 0, 600) > 0.5"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 20*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 60*time.Second,
		try.BodyContains(whoami.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestCircuitBreaker(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--retry",
		"--retry.attempts=1",
		"--forwardingTimeouts.dialTimeout=5s",
		"--forwardingTimeouts.responseHeaderTimeout=10s",
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=false",
		"--consulCatalog.watch=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")
	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80,
		[]string{"name=whoami1", label.TraefikEnable + "=true", label.TraefikBackendCircuitBreakerExpression + "=NetworkErrorRatio() > 0.5"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	whoami2 := s.composeProject.Container(c, "whoami2")
	err = s.registerService("test", whoami2.NetworkSettings.IPAddress, 42,
		[]string{"name=whoami2", label.TraefikEnable + "=true", label.TraefikBackendCircuitBreakerExpression + "=NetworkErrorRatio() > 0.5"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami2.NetworkSettings.IPAddress)

	whoami3 := s.composeProject.Container(c, "whoami3")
	err = s.registerService("test", whoami3.NetworkSettings.IPAddress, 42,
		[]string{"name=whoami3", label.TraefikEnable + "=true", label.TraefikBackendCircuitBreakerExpression + "=NetworkErrorRatio() > 0.5"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))
	defer s.deregisterService("test", whoami3.NetworkSettings.IPAddress)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 20*time.Second, try.StatusCodeIs(http.StatusServiceUnavailable), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestRefreshConfigPortChange(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.exposedByDefault=false",
		"--consulCatalog.watch=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	whoami := s.composeProject.Container(c, "whoami1")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 81, []string{"name=whoami1", "traefik.enable=true"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 20*time.Second, try.StatusCodeIs(http.StatusBadGateway))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 5*time.Second, try.BodyContains(whoami.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{"name=whoami1", label.TraefikEnable + "=true"})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	defer s.deregisterService("test", whoami.NetworkSettings.IPAddress)

	err = try.GetRequest("http://127.0.0.1:8080/api/providers/consul_catalog/backends", 60*time.Second, try.BodyContains(whoami.NetworkSettings.IPAddress))
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 20*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestRetryWithConsulServer(c *check.C) {
	// Scale consul to 0 to be able to start traefik before and test retry
	s.composeProject.Scale(c, "consul", 0)

	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.watch=false",
		"--consulCatalog.exposedByDefault=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	// Request should fail
	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Scale consul to 1
	s.composeProject.Scale(c, "consul", 1)
	s.waitToElectConsulLeader()

	whoami := s.composeProject.Container(c, "whoami1")
	// Register service
	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80, []string{})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	// Provider consul catalog should be present
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.BodyContains("consul_catalog"))
	c.Assert(err, checker.IsNil)

	// Should be ok
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestServiceWithMultipleHealthCheck(c *check.C) {
	// Scale consul to 0 to be able to start traefik before and test retry
	s.composeProject.Scale(c, "consul", 0)

	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.watch=false",
		"--consulCatalog.exposedByDefault=true",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	// Request should fail
	err = try.Request(req, 2*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Scale consul to 1
	s.composeProject.Scale(c, "consul", 1)
	s.waitToElectConsulLeader()

	whoami := s.composeProject.Container(c, "whoami1")
	// Register service
	err = s.registerAgentService("test", whoami.NetworkSettings.IPAddress, 80, []string{"name=whoami1"}, true)
	c.Assert(err, checker.IsNil, check.Commentf("Error registering agent service"))
	defer s.deregisterAgentService(whoami.NetworkSettings.IPAddress)

	// Register one healthcheck
	err = s.registerCheck("test", whoami.NetworkSettings.IPAddress, 80)
	c.Assert(err, checker.IsNil, check.Commentf("Error registering check"))

	// Provider consul catalog should be present
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.BodyContains("consul_catalog"))
	c.Assert(err, checker.IsNil)

	// Should be ok
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Change health value of service to critical
	reqHealth, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:80/health", whoami.NetworkSettings.IPAddress), bytes.NewBuffer([]byte("500")))
	c.Assert(err, checker.IsNil)
	reqHealth.Host = "test.consul.localhost"

	err = try.Request(reqHealth, 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Should be a 404
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Change health value of service to passing
	reqHealth, err = http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:80/health", whoami.NetworkSettings.IPAddress), bytes.NewBuffer([]byte("200")))
	c.Assert(err, checker.IsNil)
	err = try.Request(reqHealth, 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	// Should be a 200
	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestMaintenanceMode(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	whoami := s.composeProject.Container(c, "whoami1")

	err = s.registerAgentService("test", whoami.NetworkSettings.IPAddress, 80, []string{}, false)
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Enable service maintenance mode
	err = s.consulEnableServiceMaintenance(whoami.NetworkSettings.IPAddress)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Disable service maintenance mode
	err = s.consulDisableServiceMaintenance(whoami.NetworkSettings.IPAddress)
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Enable node maintenance mode
	err = s.consulEnableNodeMaintenance()
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusNotFound), try.HasBody())
	c.Assert(err, checker.IsNil)

	// Disable node maintenance mode
	err = s.consulDisableNodeMaintenance()
	c.Assert(err, checker.IsNil)

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestMultipleFrontendRule(c *check.C) {
	cmd, display := s.traefikCmd(
		withConfigFile("fixtures/consul_catalog/simple.toml"),
		"--consulCatalog",
		"--consulCatalog.endpoint="+s.consulIP+":8500",
		"--consulCatalog.domain=consul.localhost")
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Wait for Traefik to turn ready.
	err = try.GetRequest("http://127.0.0.1:8000/", 2*time.Second, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	whoami := s.composeProject.Container(c, "whoami1")

	err = s.registerService("test", whoami.NetworkSettings.IPAddress, 80,
		[]string{
			"traefik.frontends.service1.rule=Host:whoami1.consul.localhost",
			"traefik.frontends.service2.rule=Host:whoami2.consul.localhost",
		})
	c.Assert(err, checker.IsNil, check.Commentf("Error registering service"))

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.consul.localhost"

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami1.consul.localhost"

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "whoami2.consul.localhost"

	err = try.Request(req, 10*time.Second, try.StatusCodeIs(http.StatusOK), try.HasBody())
	c.Assert(err, checker.IsNil)
}
