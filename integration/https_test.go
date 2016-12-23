package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"time"

	"github.com/go-check/check"

	checker "github.com/vdemeester/shakers"
)

// HTTPSSuite
type HTTPSSuite struct{ BaseSuite }

// TestWithSNIConfigHandshake involves a client sending a SNI hostname of
// "snitest.com", which happens to match the CN of 'snitest.com.crt'. The test
// verifies that traefik presents the correct certificate.
func (s *HTTPSSuite) TestWithSNIConfigHandshake(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/https/https_sni.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		NextProtos:         []string{"h2", "http/1.1"},
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.IsNil, check.Commentf("failed to connect to server"))

	defer conn.Close()
	err = conn.Handshake()
	c.Assert(err, checker.IsNil, check.Commentf("TLS handshake error"))

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("snitest.com")
	c.Assert(err, checker.IsNil, check.Commentf("certificate did not match SNI servername"))

	proto := conn.ConnectionState().NegotiatedProtocol
	c.Assert(proto, checker.Equals, "h2")
}

// TestWithSNIConfigRoute involves a client sending HTTPS requests with
// SNI hostnames of "snitest.org" and "snitest.com". The test verifies
// that traefik routes the requests to the expected backends.
func (s *HTTPSSuite) TestWithSNIConfigRoute(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/https/https_sni.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	backend1 := startTestServer("9010", 204)
	backend2 := startTestServer("9020", 205)
	defer backend1.Close()
	defer backend2.Close()

	time.Sleep(2000 * time.Millisecond)

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

	client := &http.Client{Transport: tr1}
	req, _ := http.NewRequest("GET", "https://127.0.0.1:4443/", nil)
	req.Host = "snitest.com"
	req.Header.Set("Host", "snitest.com")
	req.Header.Set("Accept", "*/*")
	resp, err := client.Do(req)
	c.Assert(err, checker.IsNil)
	// Expected a 204 (from backend1)
	c.Assert(resp.StatusCode, checker.Equals, 204)

	client = &http.Client{Transport: tr2}
	req, _ = http.NewRequest("GET", "https://127.0.0.1:4443/", nil)
	req.Host = "snitest.org"
	req.Header.Set("Host", "snitest.org")
	req.Header.Set("Accept", "*/*")
	resp, err = client.Do(req)
	c.Assert(err, checker.IsNil)
	// Expected a 205 (from backend2)
	c.Assert(resp.StatusCode, checker.Equals, 205)
}

// TestWithClientCertificateAuthentication
// The client has to send a certificate signed by a CA trusted by the server
func (s *HTTPSSuite) TestWithClientCertificateAuthentication(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/https/clientca/https_1ca1config.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	// Connection without client certificate should fail
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.NotNil, check.Commentf("should not be allowed to connect to server"))

	// Connect with client certificate signed by ca1
	cert, err := tls.LoadX509KeyPair("fixtures/https/clientca/client1.crt", "fixtures/https/clientca/client1.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.IsNil, check.Commentf("failed to connect to server"))

	conn.Close()

	// Connect with client signed by ca2 should fail
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client2.crt", "fixtures/https/clientca/client2.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.NotNil, check.Commentf("should not be allowed to connect to server"))

}

// TestWithClientCertificateAuthentication
// Use two CA:s and test that clients with client signed by either of them can connect
func (s *HTTPSSuite) TestWithClientCertificateAuthenticationMultipeCAs(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/https/clientca/https_2ca1config.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	// Connection without client certificate should fail
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.NotNil, check.Commentf("should not be allowed to connect to server"))

	// Connect with client signed by ca1
	cert, err := tls.LoadX509KeyPair("fixtures/https/clientca/client1.crt", "fixtures/https/clientca/client1.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.IsNil, check.Commentf("failed to connect to server"))

	conn.Close()

	// Connect with client signed by ca2
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client2.crt", "fixtures/https/clientca/client2.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.IsNil, check.Commentf("failed to connect to server"))
	conn.Close()

	// Connect with client signed by ca3 should fail
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client3.crt", "fixtures/https/clientca/client3.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.NotNil, check.Commentf("should not be allowed to connect to server"))
}

// TestWithClientCertificateAuthentication
// Use two CA:s in two different files and test that clients with client signed by either of them can connect
func (s *HTTPSSuite) TestWithClientCertificateAuthenticationMultipeCAsMultipleFiles(c *check.C) {
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/https/clientca/https_2ca2config.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(500 * time.Millisecond)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	// Connection without client certificate should fail
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.NotNil, check.Commentf("should not be allowed to connect to server"))

	// Connect with client signed by ca1
	cert, err := tls.LoadX509KeyPair("fixtures/https/clientca/client1.crt", "fixtures/https/clientca/client1.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.IsNil, check.Commentf("failed to connect to server"))

	conn.Close()

	// Connect with client signed by ca2
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client2.crt", "fixtures/https/clientca/client2.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.IsNil, check.Commentf("failed to connect to server"))
	conn.Close()

	// Connect with client signed by ca3 should fail
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "snitest.com",
		Certificates:       []tls.Certificate{},
	}
	cert, err = tls.LoadX509KeyPair("fixtures/https/clientca/client3.crt", "fixtures/https/clientca/client3.key")
	c.Assert(err, checker.IsNil, check.Commentf("unable to load client certificate and key"))
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

	conn, err = tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.NotNil, check.Commentf("should not be allowed to connect to server"))
}

func startTestServer(port string, statusCode int) (ts *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
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
	return
}
