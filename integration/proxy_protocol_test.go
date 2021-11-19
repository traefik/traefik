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
	gatewayHost string
	haproxyHost string
	whoamiHost  string
}

func (s *ProxyProtocolSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "proxy-protocol")
	s.composeUp(c)

	s.gatewayHost = s.getContainerIP(c, "traefik")
	s.haproxyHost = s.getComposeServiceIP(c, "haproxy")
	s.whoamiHost = s.getComposeServiceIP(c, "whoami")
}

func (s *ProxyProtocolSuite) TestProxyProtocolTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyHost, WhoamiIP: s.whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyHost+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.gatewayHost))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2Trusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/with.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyHost, WhoamiIP: s.whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyHost+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.gatewayHost))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolNotTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyHost, WhoamiIP: s.whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyHost+"/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.haproxyHost))
	c.Assert(err, checker.IsNil)
}

func (s *ProxyProtocolSuite) TestProxyProtocolV2NotTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/without.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{HaproxyIP: s.haproxyHost, WhoamiIP: s.whoamiHost})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://"+s.haproxyHost+":81/whoami", 1*time.Second,
		try.StatusCodeIs(http.StatusOK),
		try.BodyContains("X-Forwarded-For: "+s.haproxyHost))
	c.Assert(err, checker.IsNil)
}
