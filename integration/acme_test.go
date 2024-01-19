package integration

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/provider/acme"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	"github.com/traefik/traefik/v3/pkg/types"
)

// ACME test suites.
type AcmeSuite struct {
	BaseSuite
	pebbleIP      string
	fakeDNSServer *dns.Server
}

func TestAcmeSuite(t *testing.T) {
	suite.Run(t, new(AcmeSuite))
}

type subCases struct {
	host               string
	expectedCommonName string
	expectedAlgorithm  x509.PublicKeyAlgorithm
}

type acmeTestCase struct {
	template            templateModel
	traefikConfFilePath string
	subCases            []subCases
}

type templateModel struct {
	Domain    types.Domain
	Domains   []types.Domain
	PortHTTP  string
	PortHTTPS string
	Acme      map[string]static.CertificateResolver
}

const (
	// Domain to check
	acmeDomain = "traefik.acme.wtf"

	// Wildcard domain to check
	wildcardDomain = "*.acme.wtf"
)

func (s *AcmeSuite) getAcmeURL() string {
	return fmt.Sprintf("https://%s/dir",
		net.JoinHostPort(s.pebbleIP, "14000"))
}

func setupPebbleRootCA() (*http.Transport, error) {
	path, err := filepath.Abs("fixtures/acme/ssl/pebble.minica.pem")
	if err != nil {
		return nil, err
	}

	os.Setenv("LEGO_CA_CERTIFICATES", path)
	os.Setenv("LEGO_CA_SERVER_NAME", "pebble")

	customCAs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(customCAs); !ok {
		return nil, fmt.Errorf("error creating x509 cert pool from %q: %w", path, err)
	}

	return &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: "pebble",
			RootCAs:    certPool,
		},
	}, nil
}

func (s *AcmeSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("pebble")
	s.composeUp()

	// Retrieving the Docker host ip.
	s.fakeDNSServer = startFakeDNSServer(s.hostIP)
	s.pebbleIP = s.getComposeServiceIP("pebble")

	pebbleTransport, err := setupPebbleRootCA()
	require.NoError(s.T(), err)

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

	require.NoError(s.T(), err)
}

func (s *AcmeSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()

	if s.fakeDNSServer != nil {
		err := s.fakeDNSServer.Shutdown()
		if err != nil {
			log.Info().Msg(err.Error())
		}
	}

	s.composeDown()
}

func (s *AcmeSuite) TestHTTP01Domains() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_domains.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Domains: []types.Domain{{
				Main: "traefik.acme.wtf",
			}},
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestHTTP01StoreDomains() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_store_domains.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Domain: types.Domain{
				Main: "traefik.acme.wtf",
			},
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestHTTP01DomainsInSAN() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_domains.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: "acme.wtf",
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Domains: []types.Domain{{
				Main: "acme.wtf",
				SANs: []string{"traefik.acme.wtf"},
			}},
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRule() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestMultipleResolver() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_multiple_resolvers.toml",
		subCases: []subCases{
			{
				host:               acmeDomain,
				expectedCommonName: acmeDomain,
				expectedAlgorithm:  x509.RSA,
			},
			{
				host:               "tchouk.acme.wtf",
				expectedCommonName: "tchouk.acme.wtf",
				expectedAlgorithm:  x509.ECDSA,
			},
		},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
				}},
				"tchouk": {ACME: &acme.Configuration{
					TLSChallenge: &acme.TLSChallenge{},
					KeyType:      "EC256",
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleECDSA() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.ECDSA,
		}},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
					KeyType:       "EC384",
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleInvalidAlgo() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
					KeyType:       "INVALID",
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleDefaultDynamicCertificatesWithWildcard() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: wildcardDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestHTTP01OnHostRuleDynamicCertificatesWithWildcard() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tls_dynamic.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: wildcardDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestTLSALPN01OnHostRuleTCP() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_tcp.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					TLSChallenge: &acme.TLSChallenge{},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestTLSALPN01OnHostRule() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_base.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					TLSChallenge: &acme.TLSChallenge{},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestTLSALPN01Domains() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_domains.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: acmeDomain,
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Domains: []types.Domain{{
				Main: "traefik.acme.wtf",
			}},
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					TLSChallenge: &acme.TLSChallenge{},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

func (s *AcmeSuite) TestTLSALPN01DomainsInSAN() {
	testCase := acmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_domains.toml",
		subCases: []subCases{{
			host:               acmeDomain,
			expectedCommonName: "acme.wtf",
			expectedAlgorithm:  x509.RSA,
		}},
		template: templateModel{
			Domains: []types.Domain{{
				Main: "acme.wtf",
				SANs: []string{"traefik.acme.wtf"},
			}},
			Acme: map[string]static.CertificateResolver{
				"default": {ACME: &acme.Configuration{
					TLSChallenge: &acme.TLSChallenge{},
				}},
			},
		},
	}

	s.retrieveAcmeCertificate(testCase)
}

// Test Let's encrypt down.
func (s *AcmeSuite) TestNoValidLetsEncryptServer() {
	file := s.adaptFile("fixtures/acme/acme_base.toml", templateModel{
		Acme: map[string]static.CertificateResolver{
			"default": {ACME: &acme.Configuration{
				CAServer:      "http://wrongurl:4001/directory",
				HTTPChallenge: &acme.HTTPChallenge{EntryPoint: "web"},
			}},
		},
	})

	s.traefikCmd(withConfigFile(file))

	// Expected traefik works
	err := try.GetRequest("http://127.0.0.1:8080/api/rawdata", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	require.NoError(s.T(), err)
}

// Doing an HTTPS request and test the response certificate.
func (s *AcmeSuite) retrieveAcmeCertificate(testCase acmeTestCase) {
	if len(testCase.template.PortHTTP) == 0 {
		testCase.template.PortHTTP = ":5002"
	}

	if len(testCase.template.PortHTTPS) == 0 {
		testCase.template.PortHTTPS = ":5001"
	}

	for _, value := range testCase.template.Acme {
		if len(value.ACME.CAServer) == 0 {
			value.ACME.CAServer = s.getAcmeURL()
		}
	}

	file := s.adaptFile(testCase.traefikConfFilePath, testCase.template)

	s.traefikCmd(withConfigFile(file))

	// A real file is needed to have the right mode on acme.json file
	defer os.Remove("/tmp/acme.json")

	backend := startTestServer("9010", http.StatusOK, "")
	defer backend.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// wait for traefik (generating acme account take some seconds)
	err := try.Do(60*time.Second, func() error {
		_, errGet := client.Get("https://127.0.0.1:5001")
		return errGet
	})
	require.NoError(s.T(), err)

	for _, sub := range testCase.subCases {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         sub.host,
				},
				// Needed so that each subcase redoes the SSL handshake
				DisableKeepAlives: true,
			},
		}

		req := testhelpers.MustNewRequest(http.MethodGet, "https://127.0.0.1:5001/", nil)
		req.Host = sub.host
		req.Header.Set("Host", sub.host)
		req.Header.Set("Accept", "*/*")

		var resp *http.Response

		// Retry to send a Request which uses the LE generated certificate
		err := try.Do(60*time.Second, func() error {
			resp, err = client.Do(req)
			if err != nil {
				return err
			}

			cn := resp.TLS.PeerCertificates[0].Subject.CommonName
			if cn != sub.expectedCommonName {
				return fmt.Errorf("domain %s found instead of %s", cn, sub.expectedCommonName)
			}

			return nil
		})

		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
		// Check Domain into response certificate
		assert.Equal(s.T(), sub.expectedCommonName, resp.TLS.PeerCertificates[0].Subject.CommonName)
		assert.Equal(s.T(), sub.expectedAlgorithm, resp.TLS.PeerCertificates[0].PublicKeyAlgorithm)
	}
}
