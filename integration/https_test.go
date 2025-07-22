package integration

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/net/http2"
)

// HTTPSSuite tests suite.
type HTTPSSuite struct{ BaseSuite }

func TestHTTPSSuite(t *testing.T) {
	suite.Run(t, &HTTPSSuite{})
}

// TestWithSNIConfigHandshake involves a client sending a SNI hostname of
// "snitest.com", which happens to match the CN of 'snitest.com.crt'. The test
// verifies that traefik presents the correct certificate.
func (s *HTTPSSuite) TestWithSNIConfigHandshake() {
	file := s.adaptFile("fixtures/https/https_sni.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.org`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		NextProtos:         []string{"h2", "http/1.1"},
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	defer conn.Close()
	err = conn.Handshake()
	assert.NoError(s.T(), err, "TLS handshake error")

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("snitest.com")
	assert.NoError(s.T(), err, "certificate did not match SNI servername")

	proto := conn.ConnectionState().NegotiatedProtocol
	assert.Equal(s.T(), "h2", proto)
}

// TestWithSNIConfigRoute involves a client sending HTTPS requests with
// SNI hostnames of "snitest.org" and "snitest.com". The test verifies
// that traefik routes the requests to the expected backends.
func (s *HTTPSSuite) TestWithSNIConfigRoute() {
	file := s.adaptFile("fixtures/https/https_sni.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`snitest.org`)"))
	require.NoError(s.T(), err)

	backend1 := startTestServer("9010", http.StatusNoContent, "")
	backend2 := startTestServer("9020", http.StatusResetContent, "")
	defer backend1.Close()
	defer backend2.Close()

	err = try.GetRequest(backend1.URL, 1*time.Second, try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)
	err = try.GetRequest(backend2.URL, 1*time.Second, try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	tr1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "snitest.com",
		},
	}
	tr2 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "snitest.org",
		},
	}

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr1.TLSClientConfig.ServerName
	req.Header.Set("Host", tr1.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 30*time.Second, tr1, try.HasCn(tr1.TLSClientConfig.ServerName), try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr2.TLSClientConfig.ServerName
	req.Header.Set("Host", tr2.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 30*time.Second, tr2, try.HasCn(tr2.TLSClientConfig.ServerName), try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)
}

// TestWithTLSOptions  verifies that traefik routes the requests with the associated tls options.

func (s *HTTPSSuite) TestWithTLSOptions() {
	file := s.adaptFile("fixtures/https/https_tls_options.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`snitest.org`)"))
	require.NoError(s.T(), err)

	backend1 := startTestServer("9010", http.StatusNoContent, "")
	backend2 := startTestServer("9020", http.StatusResetContent, "")
	defer backend1.Close()
	defer backend2.Close()

	err = try.GetRequest(backend1.URL, 1*time.Second, try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)
	err = try.GetRequest(backend2.URL, 1*time.Second, try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	tr1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS12,
			ServerName:         "snitest.com",
		},
	}

	tr2 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS12,
			ServerName:         "snitest.org",
		},
	}

	tr3 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS11,
			ServerName:         "snitest.org",
		},
	}

	// With valid TLS options and request
	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr1.TLSClientConfig.ServerName
	req.Header.Set("Host", tr1.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 30*time.Second, tr1, try.HasCn(tr1.TLSClientConfig.ServerName), try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)

	// With a valid TLS version
	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr2.TLSClientConfig.ServerName
	req.Header.Set("Host", tr2.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 3*time.Second, tr2, try.HasCn(tr2.TLSClientConfig.ServerName), try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	// With a bad TLS version
	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr3.TLSClientConfig.ServerName
	req.Header.Set("Host", tr3.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")
	client := http.Client{
		Transport: tr3,
	}
	_, err = client.Do(req)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "tls: no supported versions satisfy MinVersion and MaxVersion")

	//	with unknown tls option
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("unknown TLS options: unknown@file"))
	require.NoError(s.T(), err)
}

// TestWithConflictingTLSOptions checks that routers with same SNI but different TLS options get fallbacked to the default TLS options.

func (s *HTTPSSuite) TestWithConflictingTLSOptions() {
	file := s.adaptFile("fixtures/https/https_tls_options.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`snitest.net`)"))
	require.NoError(s.T(), err)

	backend1 := startTestServer("9010", http.StatusNoContent, "")
	backend2 := startTestServer("9020", http.StatusResetContent, "")
	defer backend1.Close()
	defer backend2.Close()

	err = try.GetRequest(backend1.URL, 1*time.Second, try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)
	err = try.GetRequest(backend2.URL, 1*time.Second, try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	tr4 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS11,
			ServerName:         "snitest.net",
		},
	}

	trDefault := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS12,
			ServerName:         "snitest.net",
		},
	}

	// With valid TLS options and request
	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = trDefault.TLSClientConfig.ServerName
	req.Header.Set("Host", trDefault.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 30*time.Second, trDefault, try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)

	// With a bad TLS version
	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr4.TLSClientConfig.ServerName
	req.Header.Set("Host", tr4.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")
	client := http.Client{
		Transport: tr4,
	}
	_, err = client.Do(req)
	assert.ErrorContains(s.T(), err, "tls: no supported versions satisfy MinVersion and MaxVersion")

	// with unknown tls option
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains(fmt.Sprintf("found different TLS options for routers on the same host %v, so using the default TLS options instead", tr4.TLSClientConfig.ServerName)))
	require.NoError(s.T(), err)
}

// TestWithSNIStrictNotMatchedRequest involves a client sending a SNI hostname of
// "snitest.org", which does not match the CN of 'snitest.com.crt'. The test
// verifies that traefik closes the connection.

func (s *HTTPSSuite) TestWithSNIStrictNotMatchedRequest() {
	file := s.adaptFile("fixtures/https/https_sni_strict.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.com`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.org",
		NextProtos:         []string{"h2", "http/1.1"},
	}
	// Connection with no matching certificate should fail
	_, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.Error(s.T(), err, "failed to connect to server")
}

// TestWithDefaultCertificate involves a client sending a SNI hostname of
// "snitest.org", which does not match the CN of 'snitest.com.crt'. The test
// verifies that traefik returns the default certificate.

func (s *HTTPSSuite) TestWithDefaultCertificate() {
	file := s.adaptFile("fixtures/https/https_sni_default_cert.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.com`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.org",
		NextProtos:         []string{"h2", "http/1.1"},
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	defer conn.Close()
	err = conn.Handshake()
	assert.NoError(s.T(), err, "TLS handshake error")

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("snitest.com")
	assert.NoError(s.T(), err, "server did not serve correct default certificate")

	proto := cs.NegotiatedProtocol
	assert.Equal(s.T(), "h2", proto)
}

// TestWithDefaultCertificateNoSNI involves a client sending a request with no ServerName
// which does not match the CN of 'snitest.com.crt'. The test
// verifies that traefik returns the default certificate.

func (s *HTTPSSuite) TestWithDefaultCertificateNoSNI() {
	file := s.adaptFile("fixtures/https/https_sni_default_cert.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.com`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	defer conn.Close()
	err = conn.Handshake()
	assert.NoError(s.T(), err, "TLS handshake error")

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("snitest.com")
	assert.NoError(s.T(), err, "server did not serve correct default certificate")

	proto := cs.NegotiatedProtocol
	assert.Equal(s.T(), "h2", proto)
}

// TestWithOverlappingCertificate involves a client sending a SNI hostname of
// "www.snitest.com", which matches the CN of two static certificates:
// 'wildcard.snitest.com.crt', and `www.snitest.com.crt`. The test
// verifies that traefik returns the non-wildcard certificate.

func (s *HTTPSSuite) TestWithOverlappingStaticCertificate() {
	file := s.adaptFile("fixtures/https/https_sni_default_cert.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.com`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "www.snitest.com",
		NextProtos:         []string{"h2", "http/1.1"},
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	defer conn.Close()
	err = conn.Handshake()
	assert.NoError(s.T(), err, "TLS handshake error")

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("www.snitest.com")
	assert.NoError(s.T(), err, "server did not serve correct default certificate")

	proto := cs.NegotiatedProtocol
	assert.Equal(s.T(), "h2", proto)
}

// TestWithOverlappingCertificate involves a client sending a SNI hostname of
// "www.snitest.com", which matches the CN of two dynamic certificates:
// 'wildcard.snitest.com.crt', and `www.snitest.com.crt`. The test
// verifies that traefik returns the non-wildcard certificate.

func (s *HTTPSSuite) TestWithOverlappingDynamicCertificate() {
	file := s.adaptFile("fixtures/https/dynamic_https_sni_default_cert.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.com`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "www.snitest.com",
		NextProtos:         []string{"h2", "http/1.1"},
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	defer conn.Close()
	err = conn.Handshake()
	assert.NoError(s.T(), err, "TLS handshake error")

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("www.snitest.com")
	assert.NoError(s.T(), err, "server did not serve correct default certificate")

	proto := cs.NegotiatedProtocol
	assert.Equal(s.T(), "h2", proto)
}

// TestWithClientCertificateAuthentication
// The client can send a certificate signed by a CA trusted by the server but it's optional.

func (s *HTTPSSuite) TestWithClientCertificateAuthentication() {
	file := s.adaptFile("fixtures/https/clientca/https_1ca1config.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.org`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	// Connection without client certificate should fail
	_, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "should be allowed to connect to server")

	// Connect with client certificate signed by ca1
	cert, err := tls.LoadX509KeyPair("fixtures/https/clientca/client1.crt", "fixtures/https/clientca/client1.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	conn.Close()

	// Connect with client certificate not signed by ca1
	cert, err = tls.LoadX509KeyPair("fixtures/https/snitest.org.cert", "fixtures/https/snitest.org.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	conn.Close()

	// Connect with client signed by ca2 should fail
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client2.crt", "fixtures/https/clientca/client2.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	_, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "should be allowed to connect to server")
}

// TestWithClientCertificateAuthentication
// Use two CA:s and test that clients with client signed by either of them can connect.

func (s *HTTPSSuite) TestWithClientCertificateAuthenticationMultipleCAs() {
	server1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) { _, _ = rw.Write([]byte("server1")) }))
	server2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) { _, _ = rw.Write([]byte("server2")) }))
	defer func() {
		server1.Close()
		server2.Close()
	}()

	file := s.adaptFile("fixtures/https/clientca/https_2ca1config.toml", struct {
		Server1 string
		Server2 string
	}{
		Server1: server1.URL,
		Server2: server2.URL,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`snitest.org`)"))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443", nil)
	require.NoError(s.T(), err)
	req.Host = "snitest.com"

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}

	client := http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   1 * time.Second,
	}

	// Connection without client certificate should fail
	_, err = client.Do(req)
	assert.Error(s.T(), err)

	cert, err := tls.LoadX509KeyPair("fixtures/https/clientca/client1.crt", "fixtures/https/clientca/client1.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	// Connect with client signed by ca1
	_, err = client.Do(req)
	require.NoError(s.T(), err)

	// Connect with client signed by ca2
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}

	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client2.crt", "fixtures/https/clientca/client2.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	client = http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   1 * time.Second,
	}

	// Connect with client signed by ca1
	_, err = client.Do(req)
	require.NoError(s.T(), err)

	// Connect with client signed by ca3 should fail
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}

	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client3.crt", "fixtures/https/clientca/client3.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	client = http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   1 * time.Second,
	}

	// Connect with client signed by ca1
	_, err = client.Do(req)
	assert.Error(s.T(), err)
}

// TestWithClientCertificateAuthentication
// Use two CA:s in two different files and test that clients with client signed by either of them can connect.

func (s *HTTPSSuite) TestWithClientCertificateAuthenticationMultipleCAsMultipleFiles() {
	server1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) { _, _ = rw.Write([]byte("server1")) }))
	server2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) { _, _ = rw.Write([]byte("server2")) }))
	defer func() {
		server1.Close()
		server2.Close()
	}()

	file := s.adaptFile("fixtures/https/clientca/https_2ca2config.toml", struct {
		Server1 string
		Server2 string
	}{
		Server1: server1.URL,
		Server2: server2.URL,
	})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`snitest.org`)"))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443", nil)
	require.NoError(s.T(), err)
	req.Host = "snitest.com"

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}

	client := http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   1 * time.Second,
	}

	// Connection without client certificate should fail
	_, err = client.Do(req)
	assert.Error(s.T(), err)

	// Connect with client signed by ca1
	cert, err := tls.LoadX509KeyPair("fixtures/https/clientca/client1.crt", "fixtures/https/clientca/client1.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	_, err = client.Do(req)
	require.NoError(s.T(), err)

	// Connect with client signed by ca2
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}

	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client2.crt", "fixtures/https/clientca/client2.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	client = http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   1 * time.Second,
	}

	_, err = client.Do(req)
	require.NoError(s.T(), err)

	// Connect with client signed by ca3 should fail
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}

	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client3.crt", "fixtures/https/clientca/client3.key")
	assert.NoError(s.T(), err, "unable to load client certificate and key")
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	client = http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   1 * time.Second,
	}

	_, err = client.Do(req)
	assert.Error(s.T(), err)
}

func (s *HTTPSSuite) TestWithRootCAsContentForHTTPSOnBackend() {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	file := s.adaptFile("fixtures/https/rootcas/https.toml", struct{ BackendHost string }{backend.URL})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains(backend.URL))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8081/ping", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func (s *HTTPSSuite) TestWithRootCAsFileForHTTPSOnBackend() {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	file := s.adaptFile("fixtures/https/rootcas/https_with_file.toml", struct{ BackendHost string }{backend.URL})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains(backend.URL))
	require.NoError(s.T(), err)

	err = try.GetRequest("http://127.0.0.1:8081/ping", 1*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

func startTestServer(port string, statusCode int, textContent string) (ts *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if textContent != "" {
			_, _ = w.Write([]byte(textContent))
		}
	})
	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		panic(err)
	}

	ts = &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	ts.Start()
	return ts
}

// TestWithSNIDynamicConfigRouteWithNoChange involves a client sending HTTPS requests with
// SNI hostnames of "snitest.org" and "snitest.com". The test verifies
// that traefik routes the requests to the expected backends thanks to given certificate if possible
// otherwise thanks to the default one.
func (s *HTTPSSuite) TestWithSNIDynamicConfigRouteWithNoChange() {
	dynamicConfFileName := s.adaptFile("fixtures/https/dynamic_https.toml", struct{}{})
	confFileName := s.adaptFile("fixtures/https/dynamic_https_sni.toml", struct {
		DynamicConfFileName string
	}{
		DynamicConfFileName: dynamicConfFileName,
	})
	s.traefikCmd(withConfigFile(confFileName))

	tr1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "snitest.org",
		},
	}

	tr2 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "snitest.com",
		},
	}

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`"+tr1.TLSClientConfig.ServerName+"`)"))
	require.NoError(s.T(), err)

	backend1 := startTestServer("9010", http.StatusNoContent, "")
	backend2 := startTestServer("9020", http.StatusResetContent, "")
	defer backend1.Close()
	defer backend2.Close()

	err = try.GetRequest(backend1.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)
	err = try.GetRequest(backend2.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr1.TLSClientConfig.ServerName
	req.Header.Set("Host", tr1.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	// snitest.org certificate must be used yet && Expected a 204 (from backend1)
	err = try.RequestWithTransport(req, 30*time.Second, tr1, try.HasCn(tr1.TLSClientConfig.ServerName), try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr2.TLSClientConfig.ServerName
	req.Header.Set("Host", tr2.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	// snitest.com certificate does not exist, default certificate has to be used && Expected a 205 (from backend2)
	err = try.RequestWithTransport(req, 30*time.Second, tr2, try.HasCn("TRAEFIK DEFAULT CERT"), try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)
}

// TestWithSNIDynamicConfigRouteWithChange involves a client sending HTTPS requests with
// SNI hostnames of "snitest.org" and "snitest.com". The test verifies
// that traefik updates its configuration when the HTTPS configuration is modified and
// it routes the requests to the expected backends thanks to given certificate if possible
// otherwise thanks to the default one.

func (s *HTTPSSuite) TestWithSNIDynamicConfigRouteWithChange() {
	dynamicConfFileName := s.adaptFile("fixtures/https/dynamic_https.toml", struct{}{})
	confFileName := s.adaptFile("fixtures/https/dynamic_https_sni.toml", struct {
		DynamicConfFileName string
	}{
		DynamicConfFileName: dynamicConfFileName,
	})
	s.traefikCmd(withConfigFile(confFileName))

	tr1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "snitest.com",
		},
	}

	tr2 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "snitest.org",
		},
	}

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`"+tr2.TLSClientConfig.ServerName+"`)"))
	require.NoError(s.T(), err)

	backend1 := startTestServer("9010", http.StatusNoContent, "")
	backend2 := startTestServer("9020", http.StatusResetContent, "")
	defer backend1.Close()
	defer backend2.Close()

	err = try.GetRequest(backend1.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusNoContent))
	require.NoError(s.T(), err)
	err = try.GetRequest(backend2.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	// Change certificates configuration file content
	s.modifyCertificateConfFileContent(tr1.TLSClientConfig.ServerName, dynamicConfFileName)

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr1.TLSClientConfig.ServerName
	req.Header.Set("Host", tr1.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 30*time.Second, tr1, try.HasCn(tr1.TLSClientConfig.ServerName), try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)

	req, err = http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr2.TLSClientConfig.ServerName
	req.Header.Set("Host", tr2.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 30*time.Second, tr2, try.HasCn("TRAEFIK DEFAULT CERT"), try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

// TestWithSNIDynamicConfigRouteWithTlsConfigurationDeletion involves a client sending HTTPS requests with
// SNI hostnames of "snitest.org" and "snitest.com". The test verifies
// that traefik updates its configuration when the HTTPS configuration is modified, even if it totally deleted, and
// it routes the requests to the expected backends thanks to given certificate if possible
// otherwise thanks to the default one.

func (s *HTTPSSuite) TestWithSNIDynamicConfigRouteWithTlsConfigurationDeletion() {
	dynamicConfFileName := s.adaptFile("fixtures/https/dynamic_https.toml", struct{}{})
	confFileName := s.adaptFile("fixtures/https/dynamic_https_sni.toml", struct {
		DynamicConfFileName string
	}{
		DynamicConfFileName: dynamicConfFileName,
	})
	s.traefikCmd(withConfigFile(confFileName))

	tr2 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "snitest.org",
		},
	}

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`"+tr2.TLSClientConfig.ServerName+"`)"))
	require.NoError(s.T(), err)

	backend2 := startTestServer("9020", http.StatusResetContent, "")

	defer backend2.Close()

	err = try.GetRequest(backend2.URL, 500*time.Millisecond, try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443/", nil)
	require.NoError(s.T(), err)
	req.Host = tr2.TLSClientConfig.ServerName
	req.Header.Set("Host", tr2.TLSClientConfig.ServerName)
	req.Header.Set("Accept", "*/*")

	err = try.RequestWithTransport(req, 30*time.Second, tr2, try.HasCn(tr2.TLSClientConfig.ServerName), try.StatusCodeIs(http.StatusResetContent))
	require.NoError(s.T(), err)

	// Change certificates configuration file content
	s.modifyCertificateConfFileContent("", dynamicConfFileName)

	err = try.RequestWithTransport(req, 30*time.Second, tr2, try.HasCn("TRAEFIK DEFAULT CERT"), try.StatusCodeIs(http.StatusNotFound))
	require.NoError(s.T(), err)
}

// modifyCertificateConfFileContent replaces the content of a HTTPS configuration file.
func (s *HTTPSSuite) modifyCertificateConfFileContent(certFileName, confFileName string) {
	file, err := os.OpenFile("./"+confFileName, os.O_WRONLY, os.ModeExclusive)
	require.NoError(s.T(), err)
	defer func() {
		file.Close()
	}()
	err = file.Truncate(0)
	require.NoError(s.T(), err)

	// If certificate file is not provided, just truncate the configuration file
	if len(certFileName) > 0 {
		tlsConf := dynamic.Configuration{
			TLS: &dynamic.TLSConfiguration{
				Certificates: []*traefiktls.CertAndStores{
					{
						Certificate: traefiktls.Certificate{
							CertFile: types.FileOrContent("fixtures/https/" + certFileName + ".cert"),
							KeyFile:  types.FileOrContent("fixtures/https/" + certFileName + ".key"),
						},
					},
				},
			},
		}

		var confBuffer bytes.Buffer
		err := toml.NewEncoder(&confBuffer).Encode(tlsConf)
		require.NoError(s.T(), err)

		_, err = file.Write(confBuffer.Bytes())
		require.NoError(s.T(), err)
	}
}

func (s *HTTPSSuite) TestEntryPointHttpsRedirectAndPathModification() {
	file := s.adaptFile("fixtures/https/https_redirect.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 5*time.Second, try.BodyContains("Host(`example.com`)"))
	require.NoError(s.T(), err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	testCases := []struct {
		desc  string
		hosts []string
		path  string
	}{
		{
			desc:  "Stripped URL redirect",
			hosts: []string{"example.com", "foo.com", "bar.com"},
			path:  "/api",
		},
		{
			desc:  "Stripped URL with trailing slash redirect",
			hosts: []string{"example.com", "example2.com", "foo.com", "foo2.com", "bar.com", "bar2.com"},
			path:  "/api/",
		},
		{
			desc:  "Stripped URL with path redirect",
			hosts: []string{"example.com", "example2.com", "foo.com", "foo2.com", "bar.com", "bar2.com"},
			path:  "/api/bacon",
		},
		{
			desc:  "Stripped URL with path and trailing slash redirect",
			hosts: []string{"example.com", "example2.com", "foo.com", "foo2.com", "bar.com", "bar2.com"},
			path:  "/api/bacon/",
		},
		{
			desc:  "Root Path with redirect",
			hosts: []string{"test.com", "test2.com", "pow.com", "pow2.com"},
			path:  "/",
		},
		{
			desc:  "Path modify with redirect",
			hosts: []string{"test.com", "test2.com", "pow.com", "pow2.com"},
			path:  "/wtf",
		},
		{
			desc:  "Path modify with trailing slash redirect",
			hosts: []string{"test.com", "test2.com", "pow.com", "pow2.com"},
			path:  "/wtf/",
		},
		{
			desc:  "Path modify with matching path segment redirect",
			hosts: []string{"test.com", "test2.com", "pow.com", "pow2.com"},
			path:  "/wtf/foo",
		},
	}

	for _, test := range testCases {
		sourceURL := fmt.Sprintf("http://127.0.0.1:8888%s", test.path)
		for _, host := range test.hosts {
			req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
			require.NoError(s.T(), err)
			req.Host = host

			resp, err := client.Do(req)
			require.NoError(s.T(), err)
			resp.Body.Close()

			location := resp.Header.Get("Location")
			expected := "https://" + net.JoinHostPort(host, "8443") + test.path

			assert.Equal(s.T(), expected, location)
		}
	}
}

// TestWithSNIDynamicCaseInsensitive involves a client sending a SNI hostname of
// "bar.www.snitest.com", which matches the DNS SAN of '*.WWW.SNITEST.COM'. The test
// verifies that traefik presents the correct certificate.
func (s *HTTPSSuite) TestWithSNIDynamicCaseInsensitive() {
	file := s.adaptFile("fixtures/https/https_sni_case_insensitive_dynamic.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("HostRegexp(`[a-z1-9-]+\\\\.www\\\\.snitest\\\\.com`)"))
	require.NoError(s.T(), err)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "bar.www.snitest.com",
		NextProtos:         []string{"h2", "http/1.1"},
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	assert.NoError(s.T(), err, "failed to connect to server")

	defer conn.Close()
	err = conn.Handshake()
	assert.NoError(s.T(), err, "TLS handshake error")

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("*.WWW.SNITEST.COM")
	assert.NoError(s.T(), err, "certificate did not match SNI servername")

	proto := conn.ConnectionState().NegotiatedProtocol
	assert.Equal(s.T(), "h2", proto)
}

// TestWithDomainFronting verify the domain fronting behavior
func (s *HTTPSSuite) TestWithDomainFronting() {
	backend := startTestServer("9010", http.StatusOK, "server1")
	defer backend.Close()
	backend2 := startTestServer("9020", http.StatusOK, "server2")
	defer backend2.Close()
	backend3 := startTestServer("9030", http.StatusOK, "server3")
	defer backend3.Close()

	file := s.adaptFile("fixtures/https/https_domain_fronting.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`site1.www.snitest.com`)"))
	require.NoError(s.T(), err)

	testCases := []struct {
		desc               string
		hostHeader         string
		serverName         string
		expectedError      bool
		expectedContent    string
		expectedStatusCode int
	}{
		{
			desc:               "SimpleCase",
			hostHeader:         "site1.www.snitest.com",
			serverName:         "site1.www.snitest.com",
			expectedContent:    "server1",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "Simple case with port in the Host Header",
			hostHeader:         "site3.www.snitest.com:4443",
			serverName:         "site3.www.snitest.com",
			expectedContent:    "server3",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "Spaces after the host header",
			hostHeader:         "site3.www.snitest.com ",
			serverName:         "site3.www.snitest.com",
			expectedError:      true,
			expectedContent:    "server3",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "Spaces after the servername",
			hostHeader:         "site3.www.snitest.com",
			serverName:         "site3.www.snitest.com ",
			expectedContent:    "server3",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "Spaces after the servername and host header",
			hostHeader:         "site3.www.snitest.com ",
			serverName:         "site3.www.snitest.com ",
			expectedError:      true,
			expectedContent:    "server3",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "Domain Fronting with same tlsOptions should follow header",
			hostHeader:         "site1.www.snitest.com",
			serverName:         "site2.www.snitest.com",
			expectedContent:    "server1",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "Domain Fronting with same tlsOptions should follow header (2)",
			hostHeader:         "site2.www.snitest.com",
			serverName:         "site1.www.snitest.com",
			expectedContent:    "server2",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "Domain Fronting with different tlsOptions should produce a 421",
			hostHeader:         "site2.www.snitest.com",
			serverName:         "site3.www.snitest.com",
			expectedContent:    "",
			expectedStatusCode: http.StatusMisdirectedRequest,
		},
		{
			desc:               "Domain Fronting with different tlsOptions should produce a 421 (2)",
			hostHeader:         "site3.www.snitest.com",
			serverName:         "site1.www.snitest.com",
			expectedContent:    "",
			expectedStatusCode: http.StatusMisdirectedRequest,
		},
		{
			desc:               "Case insensitive",
			hostHeader:         "sIte1.www.snitest.com",
			serverName:         "sitE1.www.snitest.com",
			expectedContent:    "server1",
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, test := range testCases {
		req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:4443", nil)
		require.NoError(s.T(), err)
		req.Host = test.hostHeader

		err = try.RequestWithTransport(req, 500*time.Millisecond, &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: test.serverName}}, try.StatusCodeIs(test.expectedStatusCode), try.BodyContains(test.expectedContent))
		if test.expectedError {
			assert.Error(s.T(), err)
		} else {
			require.NoError(s.T(), err)
		}
	}
}

// TestWithInvalidTLSOption verifies the behavior when using an invalid tlsOption configuration.
func (s *HTTPSSuite) TestWithInvalidTLSOption() {
	backend := startTestServer("9010", http.StatusOK, "server1")
	defer backend.Close()

	file := s.adaptFile("fixtures/https/https_invalid_tls_options.toml", struct{}{})
	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 500*time.Millisecond, try.BodyContains("Host(`snitest.com`)"))
	require.NoError(s.T(), err)

	testCases := []struct {
		desc       string
		serverName string
	}{
		{
			desc:       "With invalid TLS Options specified",
			serverName: "snitest.com",
		},
		{
			desc:       "With invalid Default TLS Options",
			serverName: "snitest.org",
		},
		{
			desc: "With TLS Options without servername (fallback to default)",
		},
	}

	for _, test := range testCases {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		if test.serverName != "" {
			tlsConfig.ServerName = test.serverName
		}

		conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
		assert.Error(s.T(), err, "connected to server successfully")
		assert.Nil(s.T(), conn)
	}
}

func (s *SimpleSuite) TestMaxConcurrentStream() {
	file := s.adaptFile("fixtures/https/max_concurrent_stream.toml", struct{}{})

	s.traefikCmd(withConfigFile(file), "--log.level=DEBUG", "--accesslog")

	// Wait for traefik.
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", time.Second, try.BodyContains("api@internal"))
	require.NoError(s.T(), err)

	// Add client self-signed cert.
	roots := x509.NewCertPool()
	certContent, err := os.ReadFile("./resources/tls/local.cert")
	require.NoError(s.T(), err)

	roots.AppendCertsFromPEM(certContent)

	// Open a connection to inspect SettingsFrame.
	conn, err := tls.Dial("tcp", "127.0.0.1:8000", &tls.Config{
		RootCAs:    roots,
		NextProtos: []string{"h2"},
	})
	require.NoError(s.T(), err)

	framer := http2.NewFramer(nil, conn)
	frame, err := framer.ReadFrame()
	require.NoError(s.T(), err)

	fr, ok := frame.(*http2.SettingsFrame)
	require.True(s.T(), ok)

	maxConcurrentStream, ok := fr.Value(http2.SettingMaxConcurrentStreams)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), uint32(42), maxConcurrentStream)
}
