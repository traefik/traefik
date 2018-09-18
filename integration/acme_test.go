package integration

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/go-check/check"
	"github.com/miekg/dns"
	checker "github.com/vdemeester/shakers"
)

// ACME test suites (using libcompose)
type AcmeSuite struct {
	BaseSuite
	pebbleIP      string
	fakeDNSServer *dns.Server
}

type acmeTestCase struct {
	template            templateModel
	traefikConfFilePath string
	expectedCommonName  string
	expectedAlgorithm   x509.PublicKeyAlgorithm
}

type templateModel struct {
	PortHTTP  string
	PortHTTPS string
	Acme      acme.Configuration
}

const (
	// Domain to check
	acmeDomain = "traefik.acme.wtf"

	// Wildcard domain to check
	wildcardDomain = "*.acme.wtf"
)

func (s *AcmeSuite) getAcmeURL() string {
	return fmt.Sprintf("https://%s:14000/dir", s.pebbleIP)
}

func setupPebbleRootCA() (*http.Transport, error) {
	path, err := filepath.Abs("fixtures/acme/ssl/pebble.minica.pem")
	if err != nil {
		return nil, err
	}

	os.Setenv("LEGO_CA_CERTIFICATES", path)
	os.Setenv("LEGO_CA_SERVER_NAME", "pebble")

	customCAs, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(customCAs); !ok {
		return nil, fmt.Errorf("error creating x509 cert pool from %q: %v", path, err)
	}

	return &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: "pebble",
			RootCAs:    certPool,
		},
	}, nil
}

func (s *AcmeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "pebble")
	s.composeProject.Start(c)

	s.fakeDNSServer = startFakeDNSServer()

	s.pebbleIP = s.composeProject.Container(c, "pebble").NetworkSettings.IPAddress

	pebbleTransport, err := setupPebbleRootCA()
	if err != nil {
		c.Fatal(err)
	}

	// wait for pebble
	req := testhelpers.MustNewRequest(http.MethodGet, s.getAcmeURL(), nil)

	client := &http.Client{
		Transport: pebbleTransport,
	}

	err = try.Do(5*time.Second, func() error {
		resp, errGet := client.Do(req)
		if errGet != nil {
			return errGet
		}
		return try.StatusCodeIs(http.StatusOK)(resp)
	})
	c.Assert(err, checker.IsNil)
}

func (s *AcmeSuite) TearDownSuite(c *check.C) {
	err := s.fakeDNSServer.Shutdown()
	if err != nil {
		c.Log(err)
	}

	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *AcmeSuite) TestHTTP01DomainsAtStart(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				Domains: types.Domains{types.Domain{
					Main: "traefik.acme.wtf",
				}},
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01DomainsInSANAtStart(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				Domains: types.Domains{types.Domain{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf"},
				}},
			},
		},
		expectedCommonName: "acme.wtf",
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRule(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnHostRule:    true,
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleECDSA(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnHostRule:    true,
				KeyType:       "EC384",
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.ECDSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleInvalidAlgo(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnHostRule:    true,
				KeyType:       "INVALID",
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleWithPath(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_http01_web_path.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnHostRule:    true,
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleStaticCertificatesWithWildcard(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnHostRule:    true,
			},
		},
		expectedCommonName: wildcardDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleDynamicCertificatesWithWildcard(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls_dynamic.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnHostRule:    true,
			},
		},
		expectedCommonName: wildcardDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnDemand(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnDemand:      true,
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnDemandStaticCertificatesWithWildcard(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnDemand:      true,
			},
		},
		expectedCommonName: wildcardDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnDemandStaticCertificatesWithWildcardMultipleEntrypoints(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls_multiple_entrypoints.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnDemand:      true,
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestHTTP01OnDemandDynamicCertificatesWithWildcard(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls_dynamic.toml",
		template: templateModel{
			Acme: acme.Configuration{
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
				OnDemand:      true,
			},
		},
		expectedCommonName: wildcardDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestTLSALPN01OnHostRule(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				TLSChallenge: &acme.TLSChallenge{},
				OnHostRule:   true,
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestTLSALPN01OnDemand(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				TLSChallenge: &acme.TLSChallenge{},
				OnDemand:     true,
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestTLSALPN01DomainsAtStart(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				TLSChallenge: &acme.TLSChallenge{},
				Domains: types.Domains{types.Domain{
					Main: "traefik.acme.wtf",
				}},
			},
		},
		expectedCommonName: acmeDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestTLSALPN01DomainsInSANAtStart(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		template: templateModel{
			Acme: acme.Configuration{
				TLSChallenge: &acme.TLSChallenge{},
				Domains: types.Domains{types.Domain{
					Main: "acme.wtf",
					SANs: []string{"traefik.acme.wtf"},
				}},
			},
		},
		expectedCommonName: "acme.wtf",
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

func (s *AcmeSuite) TestTLSALPN01DomainsWithProvidedWildcardDomainAtStart(c *check.C) {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls.toml",
		template: templateModel{
			Acme: acme.Configuration{
				TLSChallenge: &acme.TLSChallenge{},
				Domains: types.Domains{types.Domain{
					Main: acmeDomain,
				}},
			},
		},
		expectedCommonName: wildcardDomain,
		expectedAlgorithm:  x509.RSA,
	}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test Let's encrypt down
func (s *AcmeSuite) TestNoValidLetsEncryptServer(c *check.C) {
	file := s.adaptFile(c, "fixtures/acme/acme_base.toml", templateModel{
		Acme: acme.Configuration{
			CAServer:      "http://wrongurl:4001/directory",
			HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "http"},
			OnHostRule:    true,
		},
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected traefik works
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

// Doing an HTTPS request and test the response certificate
func (s *AcmeSuite) retrieveAcmeCertificate(c *check.C, testCase acmeTestCase) {
	if len(testCase.template.PortHTTP) == 0 {
		testCase.template.PortHTTP = ":5002"
	}

	if len(testCase.template.PortHTTPS) == 0 {
		testCase.template.PortHTTPS = ":5001"
	}

	if len(testCase.template.Acme.CAServer) == 0 {
		testCase.template.Acme.CAServer = s.getAcmeURL()
	}

	file := s.adaptFile(c, testCase.traefikConfFilePath, testCase.template)
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()
	// A real file is needed to have the right mode on acme.json file
	defer os.Remove("/tmp/acme.json")

	backend := startTestServer("9010", http.StatusOK)
	defer backend.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// wait for traefik (generating acme account take some seconds)
	err = try.Do(90*time.Second, func() error {
		_, errGet := client.Get("https://127.0.0.1:5001")
		return errGet
	})
	c.Assert(err, checker.IsNil)

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         acmeDomain,
			},
		},
	}

	req := testhelpers.MustNewRequest(http.MethodGet, "https://127.0.0.1:5001/", nil)
	req.Host = acmeDomain
	req.Header.Set("Host", acmeDomain)
	req.Header.Set("Accept", "*/*")

	var resp *http.Response

	// Retry to send a Request which uses the LE generated certificate
	err = try.Do(60*time.Second, func() error {
		resp, err = client.Do(req)

		// /!\ If connection is not closed, SSLHandshake will only be done during the first trial /!\
		req.Close = true

		if err != nil {
			return err
		}

		cn := resp.TLS.PeerCertificates[0].Subject.CommonName
		if cn != testCase.expectedCommonName {
			return fmt.Errorf("domain %s found instead of %s", cn, testCase.expectedCommonName)
		}

		return nil
	})

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, http.StatusOK)
	// Check Domain into response certificate
	c.Assert(resp.TLS.PeerCertificates[0].Subject.CommonName, checker.Equals, testCase.expectedCommonName)
	c.Assert(resp.TLS.PeerCertificates[0].PublicKeyAlgorithm, checker.Equals, testCase.expectedAlgorithm)
}
