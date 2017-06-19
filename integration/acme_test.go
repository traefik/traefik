package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/containous/traefik/integration/utils"
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

// Domain to check
const acmeDomain = "traefik.acme.wtf"

// Wildcard domain to chekc
const wildcardDomain = "*.acme.wtf"

func (s *AcmeSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "boulder")
	s.composeProject.Start(c)

	s.boulderIP = s.composeProject.Container(c, "boulder").NetworkSettings.IPAddress

	// wait for boulder
	err := utils.Try(120*time.Second, func() error {
		resp, err := http.Get("http://" + s.boulderIP + ":4000/directory")
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New("Expected http 200 from boulder")
		}
		return nil
	})

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
	aTestCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme.toml",
		onDemand:            true,
		domainToCheck:       acmeDomain}
	s.retrieveAcmeCertificate(c, aTestCase)
}

// Test OnHostRule option with none provided certificate
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificate(c *check.C) {
	aTestCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme.toml",
		onDemand:            false,
		domainToCheck:       acmeDomain}
	s.retrieveAcmeCertificate(c, aTestCase)
}

// Test OnDemand option with a wildcard provided certificate
func (s *AcmeSuite) TestOnDemandRetrieveAcmeCertificateWithWildcard(c *check.C) {
	aTestCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided.toml",
		onDemand:            true,
		domainToCheck:       wildcardDomain}
	s.retrieveAcmeCertificate(c, aTestCase)
}

// Test onHostRule option with a wildcard provided certificate
func (s *AcmeSuite) TestOnHostRuleRetrieveAcmeCertificateWithWildcard(c *check.C) {
	aTestCase := AcmeTestCase{
		traefikConfFilePath: "fixtures/acme/acme_provided.toml",
		onDemand:            false,
		domainToCheck:       wildcardDomain}
	s.retrieveAcmeCertificate(c, aTestCase)
}

// Doing an HTTPS request and test the response certificate
func (s *AcmeSuite) retrieveAcmeCertificate(c *check.C, a AcmeTestCase) {
	file := s.adaptFile(c, a.traefikConfFilePath, struct {
		BoulderHost          string
		OnDemand, OnHostRule bool
	}{s.boulderIP, a.onDemand, !a.onDemand})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	backend := startTestServer("9010", 200)
	defer backend.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// wait for traefik (generating acme account take some seconds)
	err = utils.Try(30*time.Second, func() error {
		_, err := client.Get("https://127.0.0.1:5001")
		if err != nil {
			return err
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	tr = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         acmeDomain,
		},
	}
	client = &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", "https://127.0.0.1:5001/", nil)
	req.Host = acmeDomain
	req.Header.Set("Host", acmeDomain)
	req.Header.Set("Accept", "*/*")

	var resp *http.Response
	// Retry to send a Request which uses the LE generated certificate
	err = utils.Try(60*time.Second, func() error {
		resp, err = client.Do(req)
		// /!\ If connection is not closed, SSLHandshake will only be done during the first trial /!\
		req.Close = true
		if err != nil {
			return err
		} else if resp.TLS.PeerCertificates[0].Subject.CommonName != a.domainToCheck {
			return errors.New("Domain " + resp.TLS.PeerCertificates[0].Subject.CommonName + " found in place of " + a.domainToCheck)
		}
		return nil
	})
	c.Assert(err, checker.IsNil)
	// Check Domain into response certificate
	c.Assert(resp.TLS.PeerCertificates[0].Subject.CommonName, checker.Equals, a.domainToCheck)
	// Expected a 200
	c.Assert(resp.StatusCode, checker.Equals, 200)

}
