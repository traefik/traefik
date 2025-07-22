package integration

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

type TCPSuite struct{ BaseSuite }

func TestTCPSuite(t *testing.T) {
	suite.Run(t, new(TCPSuite))
}

func (s *TCPSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("tcp")
	s.composeUp()
}

func (s *TCPSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TCPSuite) TestMixed() {
	file := s.adaptFile("fixtures/tcp/mixed.toml", struct {
		Whoami       string
		WhoamiA      string
		WhoamiB      string
		WhoamiNoCert string
	}{
		Whoami:       "http://" + s.getComposeServiceIP("whoami") + ":80",
		WhoamiA:      s.getComposeServiceIP("whoami-a") + ":8080",
		WhoamiB:      s.getComposeServiceIP("whoami-b") + ":8080",
		WhoamiNoCert: s.getComposeServiceIP("whoami-no-cert") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("Path(`/test`)"))
	require.NoError(s.T(), err)

	// Traefik passes through, termination handled by whoami-a
	out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-a.test")
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-a")

	// Traefik passes through, termination handled by whoami-b
	out, err = guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-b")

	// Termination handled by traefik
	out, err = guessWho("127.0.0.1:8093", "whoami-c.test", true)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-no-cert")

	tr1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:8093/whoami/", nil)
	require.NoError(s.T(), err)
	err = try.RequestWithTransport(req, 10*time.Second, tr1, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:8093/not-found/", nil)
	require.NoError(s.T(), err)
	err = try.RequestWithTransport(req, 10*time.Second, tr1, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8093/test", 500*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
	err = try.GetRequest("http://127.0.0.1:8093/not-found", 500*time.Millisecond, try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

func (s *TCPSuite) TestTLSOptions() {
	file := s.adaptFile("fixtures/tcp/multi-tls-options.toml", struct {
		WhoamiNoCert string
	}{
		WhoamiNoCert: s.getComposeServiceIP("whoami-no-cert") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`whoami-c.test`)"))
	require.NoError(s.T(), err)

	// Check that we can use a client tls version <= 1.2 with hostSNI 'whoami-c.test'
	out, err := guessWhoTLSMaxVersion("127.0.0.1:8093", "whoami-c.test", true, tls.VersionTLS12)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-no-cert")

	// Check that we can use a client tls version <= 1.3 with hostSNI 'whoami-d.test'
	out, err = guessWhoTLSMaxVersion("127.0.0.1:8093", "whoami-d.test", true, tls.VersionTLS13)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-no-cert")

	// Check that we cannot use a client tls version <= 1.2 with hostSNI 'whoami-d.test'
	_, err = guessWhoTLSMaxVersion("127.0.0.1:8093", "whoami-d.test", true, tls.VersionTLS12)
	assert.ErrorContains(s.T(), err, "protocol version not supported")

	// Check that we can't reach a route with an invalid mTLS configuration.
	conn, err := tls.Dial("tcp", "127.0.0.1:8093", &tls.Config{
		ServerName:         "whoami-i.test",
		InsecureSkipVerify: true,
	})
	assert.Nil(s.T(), conn)
	assert.Error(s.T(), err)
}

func (s *TCPSuite) TestNonTLSFallback() {
	file := s.adaptFile("fixtures/tcp/non-tls-fallback.toml", struct {
		WhoamiA      string
		WhoamiB      string
		WhoamiNoCert string
		WhoamiNoTLS  string
	}{
		WhoamiA:      s.getComposeServiceIP("whoami-a") + ":8080",
		WhoamiB:      s.getComposeServiceIP("whoami-b") + ":8080",
		WhoamiNoCert: s.getComposeServiceIP("whoami-no-cert") + ":8080",
		WhoamiNoTLS:  s.getComposeServiceIP("whoami-no-tls") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	require.NoError(s.T(), err)

	// Traefik passes through, termination handled by whoami-a
	out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-a.test")
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-a")

	// Traefik passes through, termination handled by whoami-b
	out, err = guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-b")

	// Termination handled by traefik
	out, err = guessWho("127.0.0.1:8093", "whoami-c.test", true)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-no-cert")

	out, err = guessWho("127.0.0.1:8093", "", false)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-no-tls")
}

func (s *TCPSuite) TestNonTlsTcp() {
	file := s.adaptFile("fixtures/tcp/non-tls.toml", struct {
		WhoamiNoTLS string
	}{
		WhoamiNoTLS: s.getComposeServiceIP("whoami-no-tls") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	require.NoError(s.T(), err)

	// Traefik will forward every requests on the given port to whoami-no-tls
	out, err := guessWho("127.0.0.1:8093", "", false)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-no-tls")
}

func (s *TCPSuite) TestCatchAllNoTLS() {
	file := s.adaptFile("fixtures/tcp/catch-all-no-tls.toml", struct {
		WhoamiBannerAddress string
	}{
		WhoamiBannerAddress: s.getComposeServiceIP("whoami-banner") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	require.NoError(s.T(), err)

	// Traefik will forward every requests on the given port to whoami-no-tls
	out, err := welcome("127.0.0.1:8093")
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "Welcome")
}

func (s *TCPSuite) TestCatchAllNoTLSWithHTTPS() {
	file := s.adaptFile("fixtures/tcp/catch-all-no-tls-with-https.toml", struct {
		WhoamiNoTLSAddress string
		WhoamiURL          string
	}{
		WhoamiNoTLSAddress: s.getComposeServiceIP("whoami-no-tls") + ":8080",
		WhoamiURL:          "http://" + s.getComposeServiceIP("whoami") + ":80",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`*`)"))
	require.NoError(s.T(), err)

	req := httptest.NewRequest(http.MethodGet, "https://127.0.0.1:8093/test", nil)
	req.RequestURI = ""

	err = try.RequestWithTransport(req, 500*time.Millisecond, &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *TCPSuite) TestMiddlewareAllowList() {
	file := s.adaptFile("fixtures/tcp/ip-allowlist.toml", struct {
		WhoamiA string
		WhoamiB string
	}{
		WhoamiA: s.getComposeServiceIP("whoami-a") + ":8080",
		WhoamiB: s.getComposeServiceIP("whoami-b") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`whoami-a.test`)"))
	require.NoError(s.T(), err)

	// Traefik not passes through, ipAllowList closes connection
	_, err = guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-a.test")
	assert.ErrorIs(s.T(), err, io.EOF)

	// Traefik passes through, termination handled by whoami-b
	out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-b")
}

func (s *TCPSuite) TestMiddlewareWhiteList() {
	file := s.adaptFile("fixtures/tcp/ip-whitelist.toml", struct {
		WhoamiA string
		WhoamiB string
	}{
		WhoamiA: s.getComposeServiceIP("whoami-a") + ":8080",
		WhoamiB: s.getComposeServiceIP("whoami-b") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`whoami-a.test`)"))
	require.NoError(s.T(), err)

	// Traefik not passes through, ipWhiteList closes connection
	_, err = guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-a.test")
	assert.ErrorIs(s.T(), err, io.EOF)

	// Traefik passes through, termination handled by whoami-b
	out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
	require.NoError(s.T(), err)
	assert.Contains(s.T(), out, "whoami-b")
}

func (s *TCPSuite) TestWRR() {
	file := s.adaptFile("fixtures/tcp/wrr.toml", struct {
		WhoamiB  string
		WhoamiAB string
	}{
		WhoamiB:  s.getComposeServiceIP("whoami-b") + ":8080",
		WhoamiAB: s.getComposeServiceIP("whoami-ab") + ":8080",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`whoami-b.test`)"))
	require.NoError(s.T(), err)

	call := map[string]int{}
	for range 4 {
		// Traefik passes through, termination handled by whoami-b or whoami-bb
		out, err := guessWhoTLSPassthrough("127.0.0.1:8093", "whoami-b.test")
		require.NoError(s.T(), err)
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

	assert.Equal(s.T(), map[string]int{"whoami-b": 3, "whoami-ab": 1}, call)
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
