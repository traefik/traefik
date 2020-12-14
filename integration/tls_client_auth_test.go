package integration

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

const (
	caRootCertPath = "./fixtures/tlsclientauth/ca/ca.pem"
	caCRLPath      = "./fixtures/tlsclientauth/ca/ca.crl"

	caCRLHTTPPort     = 8081      // as defined in the client certificates
	caCRLHTTPEndpoint = "/ca.crl" // as defined in the client certificates

	// client1 is signed and still valid
	client1CertPath = "./fixtures/tlsclientauth/ca/client1.pem"
	client1KeyPath  = "./fixtures/tlsclientauth/ca/client1.key"
	// client2 is signed but was revoked by the CRL
	client2CertPath = "./fixtures/tlsclientauth/ca/client2.pem"
	client2KeyPath  = "./fixtures/tlsclientauth/ca/client2.key"

	serverCertPath = "./fixtures/tlsclientauth/server/server.pem"
	serverKeyPath  = "./fixtures/tlsclientauth/server/server.key"
)

const (
	requestTimeout = 2 * time.Second
)

var errBadCertificateStr = "bad certificate"

type TLSClientAuthSuite struct{ BaseSuite }

func (s *TLSClientAuthSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "tlsclientauth")
	s.composeProject.Start(c)

	http.HandleFunc(caCRLHTTPEndpoint, func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, caCRLPath)
	})
}

func (s *TLSClientAuthSuite) startCRLServer() func() error {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", caCRLHTTPPort)}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("CRL server failed: %v", err)
		}
	}()
	return func() error {
		return srv.Shutdown(context.Background())
	}
}

func (s *TLSClientAuthSuite) testrun(c *check.C, cfp string) (*http.Request, *http.Transport, *http.Transport, func()) {
	var cleanupFuncs []func() error
	cleanupFunc := func() {
		for _, cf := range cleanupFuncs {
			err := cf()
			c.Assert(err, checker.IsNil)
		}
	}

	caRootCertContent, err := ioutil.ReadFile(caRootCertPath)
	c.Assert(err, check.IsNil)
	client1KeyPair, err := tls.LoadX509KeyPair(client1CertPath, client1KeyPath)
	c.Assert(err, checker.IsNil)
	client2KeyPair, err := tls.LoadX509KeyPair(client2CertPath, client2KeyPath)
	c.Assert(err, checker.IsNil)
	serverCertContent, err := ioutil.ReadFile(serverCertPath)
	c.Assert(err, check.IsNil)
	serverKeyContent, err := ioutil.ReadFile(serverKeyPath)
	c.Assert(err, check.IsNil)

	file := s.adaptFile(c, cfp, struct {
		CARootCertContent string
		ServerCertContent string
		ServerKeyContent  string
	}{
		CARootCertContent: string(caRootCertContent),
		ServerCertContent: string(serverCertContent),
		ServerKeyContent:  string(serverKeyContent),
	})
	cleanupFuncs = append(cleanupFuncs, func() error {
		return os.Remove(file)
	})

	cmd, display := s.traefikCmd(withConfigFile(file))
	cleanupFuncs = append(cleanupFuncs, func() error {
		display(c)
		return nil
	})
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	cleanupFuncs = append(cleanupFuncs, func() error {
		s.killCmd(cmd)
		return nil
	})
	// TODO: Why do we need to wait for the server to spin up?
	time.Sleep(2 * time.Second)

	trClient1 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{client1KeyPair},
		},
	}
	trClient2 := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{client2KeyPair},
		},
	}

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:8443", nil)
	c.Assert(err, checker.IsNil)

	return req, trClient1, trClient2, cleanupFunc
}

func (s *TLSClientAuthSuite) requestShouldSucceed(c *check.C, req *http.Request, tr *http.Transport) {
	err := try.RequestWithTransport(req, requestTimeout, tr, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *TLSClientAuthSuite) requestShouldFail(c *check.C, req *http.Request, tr *http.Transport) {
	err := try.RequestWithTransport(req, requestTimeout, tr)
	c.Assert(err.Error(), checker.Contains, errBadCertificateStr)
}

func (s *TLSClientAuthSuite) TestTLSClientAuthCRL(c *check.C) {
	req, trClient1, trClient2, cleanupFunc := s.testrun(c, "./fixtures/tlsclientauth/crl.toml")
	defer cleanupFunc()

	// Request with client1 should succeed (soft-fail due to missing CRL)
	s.requestShouldSucceed(c, req, trClient1)
	// Request with client2 should succeed (soft-fail due to missing CRL)
	s.requestShouldSucceed(c, req, trClient2)

	// Boot up CRL server
	serverShutdownFunc := s.startCRLServer()

	// Request with client1 should succeed
	s.requestShouldSucceed(c, req, trClient1)
	// Request with client2 should fail
	s.requestShouldFail(c, req, trClient2)

	// Shut down CRL server
	c.Assert(serverShutdownFunc(), checker.IsNil)

	// Request with client1 should succeed (due to cached CRL)
	s.requestShouldSucceed(c, req, trClient1)
	// Request with client2 should fail (due to cached CRL)
	s.requestShouldFail(c, req, trClient2)
}

func (s *TLSClientAuthSuite) TestTLSClientAuthCRLStrict(c *check.C) {
	req, trClient1, trClient2, cleanupFunc := s.testrun(c, "./fixtures/tlsclientauth/crl_revocationCheckStrict.toml")
	defer cleanupFunc()

	// Request with client1 should fail (hard-fail due to missing CRL)
	s.requestShouldFail(c, req, trClient1)
	// Request with client2 should fail (hard-fail due to missing CRL)
	s.requestShouldFail(c, req, trClient2)

	// Boot up CRL server
	serverShutdownFunc := s.startCRLServer()

	// Request with client1 should succeed
	s.requestShouldSucceed(c, req, trClient1)
	// Request with client2 should fail
	s.requestShouldFail(c, req, trClient2)

	// Shut down CRL server
	c.Assert(serverShutdownFunc(), checker.IsNil)

	// Request with client1 should succeed (due to cached CRL)
	s.requestShouldSucceed(c, req, trClient1)
	// Request with client2 should fail (due to cached CRL)
	s.requestShouldFail(c, req, trClient2)
}
