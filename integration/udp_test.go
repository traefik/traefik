package integration

import (
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type UDPSuite struct{ BaseSuite }

func (s *UDPSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "udp")
	s.composeUp(c)
}

func guessWhoUDP(addr string) (string, error) {
	var conn net.Conn
	var err error

	udpAddr, err2 := net.ResolveUDPAddr("udp", addr)
	if err2 != nil {
		return "", err2
	}

	conn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return "", err
	}

	_, err = conn.Write([]byte("WHO"))
	if err != nil {
		return "", err
	}

	out := make([]byte, 2048)
	n, err := conn.Read(out)
	if err != nil {
		return "", err
	}
	return string(out[:n]), nil
}

func (s *UDPSuite) TestWRR(c *check.C) {
	file := s.adaptFile(c, "fixtures/udp/wrr.toml", struct {
		WhoamiAIP string
		WhoamiBIP string
		WhoamiCIP string
		WhoamiDIP string
	}{
		WhoamiAIP: s.getComposeServiceIP(c, "whoami-a"),
		WhoamiBIP: s.getComposeServiceIP(c, "whoami-b"),
		WhoamiCIP: s.getComposeServiceIP(c, "whoami-c"),
		WhoamiDIP: s.getComposeServiceIP(c, "whoami-d"),
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("whoami-a"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8093/who", 5*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	stop := make(chan struct{})
	go func() {
		call := map[string]int{}
		for i := 0; i < 8; i++ {
			out, err := guessWhoUDP("127.0.0.1:8093")
			c.Assert(err, checker.IsNil)
			switch {
			case strings.Contains(out, "whoami-a"):
				call["whoami-a"]++
			case strings.Contains(out, "whoami-b"):
				call["whoami-b"]++
			case strings.Contains(out, "whoami-c"):
				call["whoami-c"]++
			default:
				call["unknown"]++
			}
		}
		c.Assert(call, checker.DeepEquals, map[string]int{"whoami-a": 3, "whoami-b": 2, "whoami-c": 3})
		close(stop)
	}()

	select {
	case <-stop:
	case <-time.Tick(5 * time.Second):
		c.Error("Timeout")
	}
}
