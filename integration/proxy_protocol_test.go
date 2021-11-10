package integration

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type ProxyProtocolSuite struct{ BaseSuite }

func (s *ProxyProtocolSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "proxy-protocol")
	err := s.dockerService.Up(context.Background(), s.composeProject, api.UpOptions{})
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolTrusted(c *check.C) {
	gatewayIP := "haproxy"
	haproxyIP := "haproxy"
	whoamiIP := "whoami"

	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyIP+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+gatewayIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2Trusted(c *check.C) {
	gatewayIP := "haproxy"
	haproxyIP := "haproxy"
	whoamiIP := "whoami"

	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyIP+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+gatewayIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolNotTrusted(c *check.C) {
	haproxyIP := "haproxy"
	whoamiIP := "whoami"

	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyIP+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+haproxyIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2NotTrusted(c *check.C) {
	haproxyIP := "haproxy"
	whoamiIP := "whoami"

	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+haproxyIP+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+haproxyIP))
	c.Assert(err, checker.IsNil)
}
