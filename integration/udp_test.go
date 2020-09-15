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
	s.composeProject.Start(c)
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
	whoamiAIP := s.composeProject.Container(c, "whoami-a").NetworkSettings.IPAddress
	whoamiBIP := s.composeProject.Container(c, "whoami-b").NetworkSettings.IPAddress
	whoamiCIP := s.composeProject.Container(c, "whoami-c").NetworkSettings.IPAddress
	whoamiDIP := s.composeProject.Container(c, "whoami-d").NetworkSettings.IPAddress

	file := s.adaptFile(c, "fixtures/udp/wrr.toml", struct {
		WhoamiAIP string
		WhoamiBIP string
		WhoamiCIP string
		WhoamiDIP string
	}{
		WhoamiAIP: whoamiAIP,
		WhoamiBIP: whoamiBIP,
		WhoamiCIP: whoamiCIP,
		WhoamiDIP: whoamiDIP,
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("whoami-a"))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8093/who", 5*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	stop := make(chan struct{})
	go func() {
		call := map[string]int{}
		for i := 0; i < 4; i++ {
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
		c.Assert(call, checker.DeepEquals, map[string]int{"whoami-a": 2, "whoami-b": 1, "whoami-c": 1})
		close(stop)
	}()

	select {
	case <-stop:
	case <-time.Tick(5 * time.Second):
		c.Error("Timeout")
	}
}
