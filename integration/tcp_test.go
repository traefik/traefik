package integration

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type TCPSuite struct{ BaseSuite }

func (s *TCPSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "tcp")
	s.composeUp(c)
}

func (s *TCPSuite) TestMixed(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/mixed.toml", struct {
		Whoami       string
		WhoamiA      string
		WhoamiB      string
		WhoamiNoCert string
	}{
		Whoami:       "http://" + s.getComposeServiceIP(c, "whoami") + ":80",
		WhoamiA:      s.getComposeServiceIP(c, "whoami-a") + ":8080",
		WhoamiB:      s.getComposeServiceIP(c, "whoami-b") + ":8080",
		WhoamiNoCert: s.getComposeServiceIP(c, "whoami-no-cert") + ":8080",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("Path(`/test`)"))
	c.Assert(err, checker.IsNil)

	// Traefik passes through, termination handled by whoami-a
	out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-a.test")
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-a")

	// Traefik passes through, termination handled by whoami-b
	out, err = guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-b")

	// Termination handled by traefik
	out, err = guessWho("127.0.0.1:8093", "whoami-c.test", true)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-no-cert")

	tr1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:8093/whoami/", nil)
	c.Assert(err, checker.IsNil)
	err = try.RequestWithTransport(req, 10*time.Second, tr1, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)

	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:8093/not-found/", nil)
	c.Assert(err, checker.IsNil)
	err = try.RequestWithTransport(req, 10*time.Second, tr1, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)

	err = try.GetRequest("http://127.0.0.1:8093/test", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8093/not-found", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	c.Assert(err, checker.IsNil)
}

func (s *TCPSuite) TestTLSOptions(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/multi-tls-options.toml", struct {
		WhoamiNoCert string
	}{
		WhoamiNoCert: s.getComposeServiceIP(c, "whoami-no-cert") + ":8080",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`whoami-c.test`)"))
	c.Assert(err, checker.IsNil)

	// Check that we can use a client tls version <= 1.2 with hostSNI 'whoami-c.test'
	out, err := guessWhoTLSMaxVersion("127.0.0.1:8093", "whoami-c.test", true, tls.VersionTLS12)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-no-cert")

	// Check that we can use a client tls version <= 1.3 with hostSNI 'whoami-d.test'
	out, err = guessWhoTLSMaxVersion("127.0.0.1:8093", "whoami-d.test", true, tls.VersionTLS13)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-no-cert")

	// Check that we cannot use a client tls version <= 1.2 with hostSNI 'whoami-d.test'
	_, err = guessWhoTLSMaxVersion("127.0.0.1:8093", "whoami-d.test", true, tls.VersionTLS12)
	c.Assert(err, checker.NotNil)
	c.Assert(err.Error(), checker.Contains, "protocol version not supported")

	// Check that we can't reach a route with an invalid mTLS configuration.
	conn, err := tls.Dial("tcp", "127.0.0.1:8093", &tls.Config{
		ServerName:         "whoami-i.test",
		InsecureSkipVerify: true,
	})
	c.Assert(conn, checker.IsNil)
	c.Assert(err, checker.NotNil)
}

func (s *TCPSuite) TestNonTLSFallback(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/non-tls-fallback.toml", struct {
		WhoamiA      string
		WhoamiB      string
		WhoamiNoCert string
		WhoamiNoTLS  string
	}{
		WhoamiA:      s.getComposeServiceIP(c, "whoami-a") + ":8080",
		WhoamiB:      s.getComposeServiceIP(c, "whoami-b") + ":8080",
		WhoamiNoCert: s.getComposeServiceIP(c, "whoami-no-cert") + ":8080",
		WhoamiNoTLS:  s.getComposeServiceIP(c, "whoami-no-tls") + ":8080",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	c.Assert(err, checker.IsNil)

	// Traefik passes through, termination handled by whoami-a
	out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-a.test")
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-a")

	// Traefik passes through, termination handled by whoami-b
	out, err = guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-b")

	// Termination handled by traefik
	out, err = guessWho("127.0.0.1:8093", "whoami-c.test", true)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-no-cert")

	out, err = guessWho("127.0.0.1:8093", "", false)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-no-tls")
}

func (s *TCPSuite) TestNonTlsTcp(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/non-tls.toml", struct {
		WhoamiNoTLS string
	}{
		WhoamiNoTLS: s.getComposeServiceIP(c, "whoami-no-tls") + ":8080",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	c.Assert(err, checker.IsNil)

	// Traefik will forward every requests on the given port to whoami-no-tls
	out, err := guessWho("127.0.0.1:8093", "", false)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-no-tls")
}

func (s *TCPSuite) TestCatchAllNoTLS(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/catch-all-no-tls.toml", struct {
		WhoamiBannerAddress string
	}{
		WhoamiBannerAddress: s.getComposeServiceIP(c, "whoami-banner") + ":8080",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	c.Assert(err, checker.IsNil)

	// Traefik will forward every requests on the given port to whoami-no-tls
	out, err := welcome("127.0.0.1:8093")
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "Welcome")
}

func (s *TCPSuite) TestCatchAllNoTLSWithHTTPS(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/catch-all-no-tls-with-https.toml", struct {
		WhoamiNoTLSAddress string
		WhoamiURL          string
	}{
		WhoamiNoTLSAddress: s.getComposeServiceIP(c, "whoami-no-tls") + ":8080",
		WhoamiURL:          "http://" + s.getComposeServiceIP(c, "whoami") + ":80",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	c.Assert(err, checker.IsNil)

	req := httptest.NewRequest(http.MethodGet, "https://127.0.0.1:8093/test", nil)
	req.RequestURI = ""

	err = try.RequestWithTransport(req, 500*time.Millisecond, &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TCPSuite) TestMiddlewareWhiteList(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/ip-whitelist.toml", struct {
		WhoamiA string
		WhoamiB string
	}{
		WhoamiA: s.getComposeServiceIP(c, "whoami-a") + ":8080",
		WhoamiB: s.getComposeServiceIP(c, "whoami-b") + ":8080",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`whoami-a.test`)"))
	c.Assert(err, checker.IsNil)

	// Traefik not passes through, ipWhitelist closes connection
	_, err = guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-a.test")
	c.Assert(err, checker.ErrorMatches, "EOF")

	// Traefik passes through, termination handled by whoami-b
	out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "whoami-b")
}

func (s *TCPSuite) TestWRR(c *check.C) {
	file := s.adaptFile(c, "fixtures/tcp/wrr.toml", struct {
		WhoamiB  string
		WhoamiAB string
	}{
		WhoamiB:  s.getComposeServiceIP(c, "whoami-b") + ":8080",
		WhoamiAB: s.getComposeServiceIP(c, "whoami-ab") + ":8080",
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`whoami-b.test`)"))
	c.Assert(err, checker.IsNil)

	call := map[string]int{}
	for i := 0; i < 4; i++ {
		// Traefik passes through, termination handled by whoami-b or whoami-bb
		out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
		c.Assert(err, checker.IsNil)
		switch {
		case strings.Contains(out, "whoami-b"):
			call["whoami-b"]++
		case strings.Contains(out, "whoami-ab"):
			call["whoami-ab"]++
		default:
			call["unknown"]++
		}
		time.Sleep(time.Second)
	}

	c.Assert(call, checker.DeepEquals, map[string]int{"whoami-b": 3, "whoami-ab": 1})
}

func welcome(addr string) (string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return "", err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	out := make([]byte, 2048)
	n, err := conn.Read(out)
	if err != nil {
		return "", err
	}

	return string(out[:n]), nil
}

func guessWho(addr, serverName string, tlsCall bool) (string, error) {
	return guessWhoTLSMaxVersion(addr, serverName, tlsCall, 0)
}

func guessWhoTLSMaxVersion(addr, serverName string, tlsCall bool, tlsMaxVersion uint16) (string, error) {
	var conn net.Conn
	var err error

	if tlsCall {
		conn, err = tls.Dial("tcp", addr, &tls.Config{
			ServerName:         serverName,
			InsecureSkipVerify: true,
			MinVersion:         0,
			MaxVersion:         tlsMaxVersion,
		})
	} else {
		tcpAddr, err2 := net.ResolveTCPAddr("tcp", addr)
		if err2 != nil {
			return "", err2
		}

		conn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return "", err
		}
	}

	if err != nil {
		return "", err
	}
	defer conn.Close()

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

// guessWhoTLSPassthrough guesses service identity and ensures that the
// certificate is valid for the given server name.
func guessWhoTLSPassthrough(addr, serverName string) (string, error) {
	var conn net.Conn
	var err error

	conn, err = tls.Dial("tcp", addr, &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: true,
		MinVersion:         0,
		MaxVersion:         0,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) > 1 {
				return errors.New("tls: more than one certificates from peer")
			}

			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("tls: failed to parse certificate from peer: %w", err)
			}

			if cert.Subject.CommonName == serverName {
				return nil
			}

			if err = cert.VerifyHostname(serverName); err == nil {
				return nil
			}

			return fmt.Errorf("tls: no valid certificate for serverName %s", serverName)
		},
	})

	if err != nil {
		return "", err
	}
	defer conn.Close()

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
