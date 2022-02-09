package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type ProxyProtocolSuite struct {
	BaseSuite
	gatewayIP string
	haproxyIP string
	whoamiIP  string
}

func (s *ProxyProtocolSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "proxy-protocol")
	s.composeUp(c)

	s.gatewayIP = s.getContainerIP(c, "traefik")
	s.haproxyIP = s.getComposeServiceIP(c, "haproxy")
	s.whoamiIP = s.getComposeServiceIP(c, "whoami")
}

func (s *ProxyProtocolSuite) TestProxyProtocolTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyIP, WhoamiIP: s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyIP+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.gatewayIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2Trusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyIP, WhoamiIP: s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyIP+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.gatewayIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolNotTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyIP, WhoamiIP: s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyIP+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.haproxyIP))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2NotTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyIP, WhoamiIP: s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyIP+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.haproxyIP))
	c.Assert(err, checker.IsNil)
}
