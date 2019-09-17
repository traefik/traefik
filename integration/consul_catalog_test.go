package integration

import (
	"bytes"
	"fmt"
	"github.com/containous/traefik/v2/integration/try"
	"github.com/docker/docker/integration-cli/checker"
	"github.com/go-check/check"
	"net/http"
	"os"
	"strings"
	"time"
)

type ConsulCatalogSuite struct {
	BaseSuite
	ConsulAddr string
}

func (s *ConsulCatalogSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "consul-catalog")
	s.composeProject.Start(c)
	s.ConsulAddr = "http://" + s.composeProject.Container(c, "consul").NetworkSettings.IPAddress + ":8500"
}

func (s *ConsulCatalogSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *ConsulCatalogSuite) registerService(id, name, address, port string, tags []string) error {
	tagsStr := ""
	if len(tags) > 0 {
		tagsStr = "\"" + strings.Join(tags, "\",\"") + "\""
	}

	body := fmt.Sprintf(`{"ID":"%s","Name":"%s","Address":"%s","Port":%s,"Tags":[%s]}`, id, name, address, port, tagsStr)

	req, err := http.NewRequest("PUT", s.ConsulAddr+"/v1/agent/service/register", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}

	return try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
}

func (s *ConsulCatalogSuite) deregisterService(id string) error {
	req, err := http.NewRequest("PUT", s.ConsulAddr+"/v1/agent/service/deregister/"+id, nil)
	if err != nil {
		return err
	}

	return try.Request(req, 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
}

func (s *ConsulCatalogSuite) TestWithNotExposedByDefaultAndDefaultsSettings(c *check.C) {
	var err error

	err = s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", []string{"traefik.enable=true"})
	c.Assert(err, checker.IsNil)
	err = s.registerService("whoami2", "whoami", s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress, "80", []string{"traefik.enable=true"})
	c.Assert(err, checker.IsNil)
	err = s.registerService("whoami3", "whoami", s.composeProject.Container(c, "whoami3").NetworkSettings.IPAddress, "80", []string{"traefik.enable=true"})
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.ConsulAddr,
	}

	file := s.adaptFile(c, "fixtures/consul-catalog/default_not_exposed.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(time.Second * 2)

	err = try.GetRequest("http://127.0.0.1:8000/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1")
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami2")
	c.Assert(err, checker.IsNil)
	err = s.deregisterService("whoami3")
	c.Assert(err, checker.IsNil)
}

func (s *ConsulCatalogSuite) TestByLabels(c *check.C) {
	var err error

	labels := []string{
		"traefik.enable=true",
		"traefik.http.routers.router1.rule=Path(`/whoami`)",
		"traefik.http.routers.router1.service=service1",
		"traefik.http.services.service1.loadBalancer.server.url=http://" + s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress,
	}

	err = s.registerService("whoami1", "whoami", s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress, "80", labels)
	c.Assert(err, checker.IsNil)

	tempObjects := struct {
		ConsulAddress string
	}{
		ConsulAddress: s.ConsulAddr,
	}

	file := s.adaptFile(c, "fixtures/consul-catalog/default_not_exposed.toml", tempObjects)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(time.Second * 2)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK), try.BodyContainsOr("Hostname: whoami1", "Hostname: whoami2", "Hostname: whoami3"))
	c.Assert(err, checker.IsNil)

	err = s.deregisterService("whoami1")
	c.Assert(err, checker.IsNil)
}
