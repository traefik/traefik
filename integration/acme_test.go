package integration

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/testhelpers"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// ACME test suites (using libcompose)
type AcmeSuite struct {
	BaseSuite
	boulderIP string
}

// Acme tests configuration
type AcmeTestCase struct {
	onDemand            bool
	traefikConfFilePath string
	domainToCheck       string
	algorithm           x509.PublicKeyAlgorithm
}

const (
	// Domain to check
	acmeDomain = "traefik.acme.wtf"

	// Wildcard domain to check
	wildcardDomain = "*.acme.wtf"

	// Traefik default certificate
	traefikDefaultDomain = "TRAEFIK DEFAULT CERT"
)

func (s *AcmeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "boulder")
	s.composeProject.Start(c)

	s.boulderIP = s.composeProject.Container(c, "boulder").NetworkSettings.IPAddress

	// wait for boulder
	err := try.GetRequest("http://"+s.boulderIP+":4001/directory", 120*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *AcmeSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

// Test ACME provider with certificate at start
func (s *AcmeSuite) TestACMEProviderAtStart(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/provideracme/acme.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test ACME provider with certificate at start
func (s *AcmeSuite) TestACMEProviderAtStartInSAN(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/provideracme/acme_insan.toml",
		onDemand:            false,
		domainToCheck:       "acme.wtf",
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test ACME provider with certificate at start
func (s *AcmeSuite) TestACMEProviderOnHost(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/provideracme/acme_onhost.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test ACME provider with certificate at start ECDSA algo
func (s *AcmeSuite) TestACMEProviderOnHostECDSA(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/provideracme/acme_onhost_ecdsa.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain,
		algorithm:           x509.ECDSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test ACME provider with certificate at start invalid algo default RSA
func (s *AcmeSuite) TestACMEProviderOnHostInvalidAlgo(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/provideracme/acme_onhost_invalid_algo.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test ACME provider with certificate at start and no ACME challenge
func (s *AcmeSuite) TestACMEProviderOnHostWithNoACMEChallenge(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/no_challenge_acme.toml",
		onDemand:            false,
		domainToCheck:       traefikDefaultDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test OnDemand option with none provided certificate and challenge HTTP-01
func (s *AcmeSuite) TestOnDemandRetrieveAcmeCertificateHTTP01(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_http01.toml",
		onDemand:            true,
		domainToCheck:       acmeDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test OnHostRule option with none provided certificate and challenge HTTP-01
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificateHTTP01(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_http01.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test OnHostRule option with none provided certificate and challenge HTTP-01 and web path
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificateHTTP01WithPath(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_http01_web.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test OnDemand option with a wildcard provided certificate
func (s *AcmeSuite) TestOnDemandRetrieveAcmeCertificateWithWildcard(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided.toml",
		onDemand:            true,
		domainToCheck:       wildcardDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test onHostRule option with a wildcard provided certificate
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificateWithWildcard(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided.toml",
		onDemand:            false,
		domainToCheck:       wildcardDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test OnDemand option with a wildcard provided certificate
func (s *AcmeSuite) TestOnDemandRetrieveAcmeCertificateWithDynamicWildcard(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided_dynamic.toml",
		onDemand:            true,
		domainToCheck:       wildcardDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test onHostRule option with a wildcard provided certificate
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificateWithDynamicWildcard(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided_dynamic.toml",
		onDemand:            false,
		domainToCheck:       wildcardDomain,
		algorithm:           x509.RSA}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test Let's encrypt down
func (s *AcmeSuite) TestNoValidLetsEncryptServer(c *check.C) {
	cmd, display := s.traefikCmd(withConfigFile("fixtures/acme/wrong_acme.toml"))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	// Expected traefik works
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 10*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

// Doing an HTTPS request and test the response certificate
func (s *AcmeSuite) retrieveAcmeCertificate(c *check.C, testCase AcmeTestCase) {
	file := s.adaptFile(c, testCase.traefikConfFilePath, struct {
		BoulderHost string
		OnDemand    bool
		OnHostRule  bool
	}{
		BoulderHost: s.boulderIP,
		OnDemand:    testCase.onDemand,
		OnHostRule:  !testCase.onDemand,
	})
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

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// wait for traefik (generating acme account take some seconds)
	err = try.Do(90*time.Second, func() error {
		_, errGet := client.Get("https://127.0.0.1:5001")
		return errGet
	})
	c.Assert(err, checker.IsNil)

	tr = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         acmeDomain,
		},
	}
	client = &http.Client{Transport: tr}

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
		if cn != testCase.domainToCheck {
			return fmt.Errorf("domain %s found instead of %s", cn, testCase.domainToCheck)
		}

		return nil
	})

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, http.StatusOK)
	// Check Domain into response certificate
	c.Assert(resp.TLS.PeerCertificates[0].Subject.CommonName, checker.Equals, testCase.domainToCheck)
	c.Assert(resp.TLS.PeerCertificates[0].PublicKeyAlgorithm, checker.Equals, testCase.algorithm)
}
