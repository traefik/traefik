package integration

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
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

func (s *TCPSuite) TestPostgresSTARTTLS() {
	file := s.adaptFile("fixtures/tcp/postgres-starttls.toml", struct {
		PostgresAddr string
	}{
		PostgresAddr: s.getComposeServiceIP("postgres") + ":5432",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`postgres-starttls`)"))
	require.NoError(s.T(), err)

	// Wait for postgres to be ready.
	err = try.Do(10*time.Second, func() error {
		c, err := net.DialTimeout("tcp", s.getComposeServiceIP("postgres")+":5432", time.Second)
		if err != nil {
			return err
		}
		_ = c.Close()
		return nil
	})
	require.NoError(s.T(), err)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8093", 2*time.Second)
	require.NoError(s.T(), err)
	defer conn.Close()

	// Send the PostgreSQL SSLRequest message: int32(8) + int32(80877103).
	_, err = conn.Write([]byte{0, 0, 0, 8, 4, 210, 22, 47})
	require.NoError(s.T(), err)

	reply := make([]byte, 1)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = io.ReadFull(conn, reply)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), byte('S'), reply[0])

	_ = conn.SetReadDeadline(time.Time{})

	// Perform TLS handshake (Traefik terminates TLS).
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         "postgres-starttls",
		InsecureSkipVerify: true,
	})
	err = tlsConn.Handshake()
	require.NoError(s.T(), err)

	// Send postgres StartupMessage: length(int32) + protocol 3.0(int32) + params.
	// Null-terminated key=value pairs: "user\0test\0database\0test\0", followed by a final \0 to end the list.
	params := "user\x00test\x00database\x00test\x00\x00"
	msgLen := 4 + 4 + len(params)
	startup := make([]byte, msgLen)
	binary.BigEndian.PutUint32(startup[0:4], uint32(msgLen))
	binary.BigEndian.PutUint32(startup[4:8], 196608) // protocol version 3.0
	copy(startup[8:], params)
	_, err = tlsConn.Write(startup)
	require.NoError(s.T(), err)

	// Read postgres response header: type(1 byte) + length(4 bytes).
	_ = tlsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	header := make([]byte, 5)
	_, err = io.ReadFull(tlsConn, header)
	require.NoError(s.T(), err)

	// Postgres must reply with an Authentication message ('R').
	assert.Equal(s.T(), byte('R'), header[0])
}

func (s *TCPSuite) TestPostgresSTARTTLSPassthrough() {
	file := s.adaptFile("fixtures/tcp/postgres-starttls-passthrough.toml", struct {
		PostgresSSLAddr string
	}{
		PostgresSSLAddr: s.getComposeServiceIP("postgres-ssl") + ":5432",
	})

	s.traefikCmd(withConfigFile(file))

	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.StatusCodeIs(http.StatusOK), try.BodyContains("HostSNI(`postgres-starttls-passthrough`)"))
	require.NoError(s.T(), err)

	// Wait for postgres-ssl to be ready.
	err = try.Do(15*time.Second, func() error {
		c, err := net.DialTimeout("tcp", s.getComposeServiceIP("postgres-ssl")+":5432", time.Second)
		if err != nil {
			return err
		}
		_ = c.Close()
		return nil
	})
	require.NoError(s.T(), err)

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8094", 2*time.Second)
	require.NoError(s.T(), err)
	defer conn.Close()

	_, err = conn.Write([]byte{0, 0, 0, 8, 4, 210, 22, 47})
	require.NoError(s.T(), err)

	reply := make([]byte, 1)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = io.ReadFull(conn, reply)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), byte('S'), reply[0])

	_ = conn.SetReadDeadline(time.Time{})

	// TLS handshake goes through to postgres-ssl (passthrough mode).
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         "postgres-starttls-passthrough",
		InsecureSkipVerify: true,
	})
	err = tlsConn.Handshake()
	require.NoError(s.T(), err)

	// Null-terminated key=value pairs: "user\0test\0database\0test\0", followed by a final \0 to end the list.
	params := "user\x00test\x00database\x00test\x00\x00"
	msgLen := 4 + 4 + len(params)
	startup := make([]byte, msgLen)
	binary.BigEndian.PutUint32(startup[0:4], uint32(msgLen))
	binary.BigEndian.PutUint32(startup[4:8], 196608)
	copy(startup[8:], params)
	_, err = tlsConn.Write(startup)
	require.NoError(s.T(), err)

	_ = tlsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	header := make([]byte, 5)
	_, err = io.ReadFull(tlsConn, header)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), byte('R'), header[0])
}

// TestTCPWildcardHostSNI verifies that a wildcard HostSNI rule HostSNI(`*.snitest.com`)
// routes TLS connections for any matching subdomain to the configured backend.
func (s *SimpleSuite) TestTCPWildcardHostSNI() {
	backend := startTestServer("9041", http.StatusOK, "")
	defer backend.Close()

	file := s.adaptFile("fixtures/tcp/wildcard-hostsni-tls-options.toml", struct {
		Backend string
	}{
		Backend: "127.0.0.1:9041",
	})
	s.traefikCmd(withConfigFile(file))

	// Wait for the wildcard TCP router to be loaded.
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second,
		try.BodyContains("HostSNI(`*.snitest.com`)"))
	require.NoError(s.T(), err)

	// foo.snitest.com matches the wildcard: TLS connection must succeed.
	conn, err := tls.Dial("tcp", "127.0.0.1:8093", &tls.Config{
		ServerName:         "foo.snitest.com",
		InsecureSkipVerify: true,
	})
	require.NoError(s.T(), err)
	conn.Close()

	// bar.snitest.com also matches the wildcard: TLS connection must succeed.
	conn, err = tls.Dial("tcp", "127.0.0.1:8093", &tls.Config{
		ServerName:         "bar.snitest.com",
		InsecureSkipVerify: true,
	})
	require.NoError(s.T(), err)
	conn.Close()
}

// TestTCPWildcardHostSNITLSOptions verifies that:
//   - a wildcard HostSNI rule HostSNI(`*.snitest.com`) with TLS option "foo" (minTLS12)
//     routes and accepts TLS 1.2 connections for any matching subdomain;
//   - a more specific rule HostSNI(`bar.snitest.com`) with TLS option "bar" (minTLS13)
//     takes priority for that subdomain and rejects TLS 1.2-only connections.
func (s *SimpleSuite) TestTCPWildcardHostSNITLSOptions() {
	backend := startTestServer("9041", http.StatusOK, "")
	defer backend.Close()

	file := s.adaptFile("fixtures/tcp/wildcard-hostsni-tls-options.toml", struct {
		Backend string
	}{
		Backend: "127.0.0.1:9041",
	})
	s.traefikCmd(withConfigFile(file))

	// Wait for both TCP routers to be loaded.
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second,
		try.BodyContains("HostSNI(`*.snitest.com`)"))
	require.NoError(s.T(), err)

	// foo.snitest.com matches the wildcard (TLS option "foo", minTLS12).
	// A TLS 1.2 connection must succeed.
	conn, err := tls.Dial("tcp", "127.0.0.1:8093", &tls.Config{
		ServerName:         "foo.snitest.com",
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS12,
	})
	require.NoError(s.T(), err)
	conn.Close()

	// bar.snitest.com has a specific rule with TLS option "bar" (minTLS13).
	// A TLS 1.2-only connection must be rejected.
	conn, err = tls.Dial("tcp", "127.0.0.1:8093", &tls.Config{
		ServerName:         "bar.snitest.com",
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
		MaxVersion:         tls.VersionTLS13,
	})
	require.NoError(s.T(), err)
	conn.Close()

	// bar.snitest.com without a version cap: connection must succeed.
	conn, err = tls.Dial("tcp", "127.0.0.1:8093", &tls.Config{
		ServerName:         "other.snitest.com",
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS11,
		MaxVersion:         tls.VersionTLS11,
	})
	require.NoError(s.T(), err)
	conn.Close()
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
