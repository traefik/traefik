package integration

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type UDPSuite struct{ BaseSuite }

func TestUDPSuite(t *testing.T) {
	suite.Run(t, new(UDPSuite))
}

func (s *UDPSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("udp")
	s.composeUp()
}

func (s *UDPSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
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

func (s *UDPSuite) TestWRR() {
	file := s.adaptFile("fixtures/udp/wrr.toml", struct {
		WhoamiAIP string
		WhoamiBIP string
		WhoamiCIP string
		WhoamiDIP string
	}{
		WhoamiAIP: s.getComposeServiceIP("whoami-a"),
		WhoamiBIP: s.getComposeServiceIP("whoami-b"),
		WhoamiCIP: s.getComposeServiceIP("whoami-c"),
		WhoamiDIP: s.getComposeServiceIP("whoami-d"),
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("whoami-a"))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8093/who", 5*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	stop := make(chan struct{})
	go func() {
		call := map[string]int{}
		for range 8 {
			out, err := guessWhoUDP("127.0.0.1:8093")
			require.NoError(s.T(), err)
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
		assert.Equal(s.T(), map[string]int{"whoami-a": 3, "whoami-b": 2, "whoami-c": 3}, call)
		close(stop)
	}()

	select {
	case <-stop:
	case <-time.Tick(5 * time.Second):
		log.Info().Msg("Timeout")
	}
}
