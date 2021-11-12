package integration

import (
	"context"
	"net/http"
	"os"
	"time"

	composeapi "github.com/docker/compose/v2/pkg/api"
	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type ProxyProtocolSuite struct{ BaseSuite }

func (s *ProxyProtocolSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "proxy-protocol")
	err := s.dockerService.Up(context.Background(), s.composeProject, composeapi.UpOptions{})
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolTrusted(c *check.C) {
	gatewayHost := s.getServiceIP(c, "traefik")
	haproxyHost := s.getServiceIP(c, "haproxy")
	whoamiHost := s.getServiceIP(c, "whoami")

	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyHost, WhoamiIP: whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyHost+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+gatewayHost))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2Trusted(c *check.C) {
	gatewayHost := s.getServiceIP(c, "traefik")
	haproxyHost := s.getServiceIP(c, "haproxy")
	whoamiHost := s.getServiceIP(c, "whoami")

	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyHost, WhoamiIP: whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyHost+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+gatewayHost))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolNotTrusted(c *check.C) {
	haproxyHost := s.getServiceIP(c, "haproxy")
	whoamiHost := s.getServiceIP(c, "whoami")

	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyHost, WhoamiIP: whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyHost+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+haproxyHost))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2NotTrusted(c *check.C) {
	haproxyHost := s.getServiceIP(c, "haproxy")
	whoamiHost := s.getServiceIP(c, "whoami")

	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyHost, WhoamiIP: whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyHost+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+haproxyHost))
	c.Assert(err, checker.IsNil)
}
