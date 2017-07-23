package integration

import (
	"crypto/tls"
	"fmt"
	"net/http"
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
}

const (
	// Domain to check
	acmeDomain = "traefik.acme.wtf"

	// Wildcard domain to check
	wildcardDomain = "*.acme.wtf"
)

func (s *AcmeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "boulder")
	s.composeProject.Start(c)

	s.boulderIP = s.composeProject.Container(c, "boulder").NetworkSettings.IPAddress

	// wait for boulder
	err := try.GetRequest("http://"+s.boulderIP+":4000/directory", 120*time.Second, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *AcmeSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

// Test OnDemand option with none provided certificate
func (s *AcmeSuite) TestOnDemandRetrieveAcmeCertificate(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme.toml",
		onDemand:            true,
		domainToCheck:       acmeDomain}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test OnHostRule option with none provided certificate
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificate(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test OnDemand option with a wildcard provided certificate
func (s *AcmeSuite) TestOnDemandRetrieveAcmeCertificateWithWildcard(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided.toml",
		onDemand:            true,
		domainToCheck:       wildcardDomain}

	s.retrieveAcmeCertificate(c, testCase)
}

// Test onHostRule option with a wildcard provided certificate
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificateWithWildcard(c *check.C) {
	testCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided.toml",
		onDemand:            false,
		domainToCheck:       wildcardDomain}

	s.retrieveAcmeCertificate(c, testCase)
}

// Doing an HTTPS request and test the response certificate
func (s *AcmeSuite) retrieveAcmeCertificate(c *check.C, testCase AcmeTestCase) {
	file := s.adaptFile(c, testCase.traefikConfFilePath, struct {
		BoulderHost          string
		OnDemand, OnHostRule bool
	}{
		BoulderHost: s.boulderIP,
		OnDemand:    testCase.onDemand,
		OnHostRule:  !testCase.onDemand,
	})

	cmd, output := s.cmdTraefik(withConfigFile(file))
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	backend := startTestServer("9010", http.StatusOK)
	defer backend.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// wait for traefik (generating acme account take some seconds)
	err = try.Do(90*time.Second, func() error {
		_, err := client.Get("https://127.0.0.1:5001")
		return err
	})
	// TODO: waiting a refactor of integration tests
	s.displayTraefikLog(c, output)
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
			return fmt.Errorf("domain %s found in place of %s", cn, testCase.domainToCheck)
		}

		return nil
	})

	c.Assert(err, checker.IsNil)
	c.Assert(resp.StatusCode, checker.Equals, http.StatusOK)
	// Check Domain into response certificate
	c.Assert(resp.TLS.PeerCertificates[0].Subject.CommonName, checker.Equals, testCase.domainToCheck)
}
