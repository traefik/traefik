package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type ProxyProtocolSuite struct{ BaseSuite }

func (s *ProxyProtocolSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "proxy-protocol")
	s.composeProject.Start(c)
}

func (s *ProxyProtocolSuite) TestProxyProtocolTrusted(c *check.C) {
	gatewayIP := s.composeProject.Container(c, "haproxy").NetworkSettings.Gateway
	haproxyIP := s.composeProject.Container(c, "haproxy").NetworkSettings.IPAddress
	whoamiIP := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://"+haproxyIP+"/whoami", 500*time.Millisecond,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+gatewayIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2Trusted(c *check.C) {
	gatewayIP := s.composeProject.Container(c, "haproxy").NetworkSettings.Gateway
	haproxyIP := s.composeProject.Container(c, "haproxy").NetworkSettings.IPAddress
	whoamiIP := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://"+haproxyIP+":81/whoami", 500*time.Millisecond,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+gatewayIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolNotTrusted(c *check.C) {
	haproxyIP := s.composeProject.Container(c, "haproxy").NetworkSettings.IPAddress
	whoamiIP := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://"+haproxyIP+"/whoami", 500*time.Millisecond,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+haproxyIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2NotTrusted(c *check.C) {
	haproxyIP := s.composeProject.Container(c, "haproxy").NetworkSettings.IPAddress
	whoamiIP := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: haproxyIP, WhoamiIP: whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://"+haproxyIP+":81/whoami", 500*time.Millisecond,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+haproxyIP))
	c.Assert(err, checker.IsNil)
}
