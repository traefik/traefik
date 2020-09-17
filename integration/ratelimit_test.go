package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type RateLimitSuite struct {
	BaseSuite
	ServerIP string
}

func (s *RateLimitSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "ratelimit")
	s.composeProject.Start(c)

	s.ServerIP = s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress
}

func (s *RateLimitSuite) TestSimpleConfiguration(c *check.C) {
	file := s.adaptFile(c, "fixtures/ratelimit/simple.toml", struct {
		Server1 string
	}{s.ServerIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("ratelimit"))
	c.Assert(err, checker.IsNil)

	start := time.Now()
	count := 0
	for {
		err = try.GetRequest("http://127.0.0.1:8081/", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
		c.Assert(err, checker.IsNil)
		count++
		if count > 100 {
			break
		}
	}
	stop := time.Now()
	elapsed := stop.Sub(start)
	if elapsed < time.Second*99/100 {
		c.Fatalf("requests throughput was too fast wrt to rate limiting: 100 requests in %v", elapsed)
	}
}
