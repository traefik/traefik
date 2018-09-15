package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	"github.com/gorilla/websocket"
	checker "github.com/vdemeester/shakers"
)

type RetrySuite struct{ BaseSuite }

func (s *RetrySuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "retry")
	s.composeProject.Start(c)
}

func (s *RetrySuite) TestRetry(c *check.C) {
	whoamiEndpoint := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/retry/simple.toml", struct {
		WhoamiEndpoint string
	}{whoamiEndpoint})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, try.BodyContains("PathPrefix:/"))
	c.Assert(err, checker.IsNil)

	// This simulates a DialTimeout when connecting to the backend server.
	response, err := http.Get("http://127.0.0.1:8000/")
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusOK)
}

func (s *RetrySuite) TestRetryWebsocket(c *check.C) {
	whoamiEndpoint := s.composeProject.Container(c, "whoami").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/retry/simple.toml", struct {
		WhoamiEndpoint string
	}{whoamiEndpoint})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 60*time.Second, try.BodyContains("PathPrefix:/"))
	c.Assert(err, checker.IsNil)

	// This simulates a DialTimeout when connecting to the backend server.
	_, response, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/echo", nil)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusSwitchingProtocols)

	_, response, err = websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/echo", nil)
	c.Assert(err, checker.IsNil)
	c.Assert(response.StatusCode, checker.Equals, http.StatusSwitchingProtocols)
}
