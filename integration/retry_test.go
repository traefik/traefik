package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/gorilla/websocket"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type RetrySuite struct {
	BaseSuite
	whoamiIP string
}

func (s *RetrySuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "retry")
	s.composeUp(c)

	s.whoamiIP = s.getComposeServiceIP(c, "whoami")
}

func (s *RetrySuite) TestRetry(c *check.C) {
	file := s.adaptFile(c, "fixtures/retry/simple.toml", struct{ WhoamiIP string }{s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("PathPrefix(`/`)"))
	c.Assert(err, checker.IsNil)

	start := time.Now()
	// This simulates a DialTimeout when connecting to the backend server.
	response, err := http.Get("http://127.0.0.1:8000/")
	duration, allowed := time.Since(start), time.Millisecond*250
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
	c.Assert(int64(duration), checker.LessThan, int64(allowed))
}

func (s *RetrySuite) TestRetryBackoff(c *check.C) {
	file := s.adaptFile(c, "fixtures/retry/backoff.toml", struct{ WhoamiIP string }{s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("PathPrefix(`/`)"))
	c.Assert(err, checker.IsNil)

	start := time.Now()
	// This simulates a DialTimeout when connecting to the backend server.
	response, err := http.Get("http://127.0.0.1:8000/")
	duration := time.Since(start)
	// test case delays: 500 + 700 + 1000ms with randomization.  It should be safely > 1500ms
	minAllowed := time.Millisecond * 1400

	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
	c.Assert(int64(duration), checker.GreaterThan, int64(minAllowed))
}

func (s *RetrySuite) TestRetryWebsocket(c *check.C) {
	file := s.adaptFile(c, "fixtures/retry/simple.toml", struct{ WhoamiIP string }{s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 60*time.Second, try.BodyContains("PathPrefix(`/`)"))
	c.Assert(err, checker.IsNil)

	// This simulates a DialTimeout when connecting to the backend server.
	_, response, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/echo", nil)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusSwitchingProtocols)

	_, response, err = websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/echo", nil)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusSwitchingProtocols)
}
