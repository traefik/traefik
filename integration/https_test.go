package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"time"

	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
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
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:4443", tlsConfig)
	c.Assert(err, checker.IsNil, check.Commentf("failed to connect to server"))

	defer conn.Close()
	err = conn.Handshake()
	c.Assert(err, checker.IsNil, check.Commentf("TLS handshake error"))

	cs := conn.ConnectionState()
	err = cs.PeerCertificates[0].VerifyHostname("snitest.com")
	c.Assert(err, checker.IsNil, check.Commentf("certificate did not match SNI servername"))
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
