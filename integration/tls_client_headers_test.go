package integration

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

const (
	rootCertPath = "./fixtures/tlsclientheaders/root.pem"
	certPemPath  = "./fixtures/tlsclientheaders/server.pem"
	certKeyPath  = "./fixtures/tlsclientheaders/server.key"
)

type TLSClientHeadersSuite struct{ BaseSuite }

func (s *TLSClientHeadersSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "tlsclientheaders")
	s.composeProject.Start(c)
}

func (s *TLSClientHeadersSuite) TestTLSClientHeaders(c *check.C) {
	rootCertContent, err := ioutil.ReadFile(rootCertPath)
	c.Assert(err, check.IsNil)
	serverCertContent, err := ioutil.ReadFile(certPemPath)
	c.Assert(err, check.IsNil)
	ServerKeyContent, err := ioutil.ReadFile(certKeyPath)
	c.Assert(err, check.IsNil)

	file := s.adaptFile(c, "fixtures/tlsclientheaders/simple.toml", struct {
		RootCertContent   string
		ServerCertContent string
		ServerKeyContent  string
	}{
		RootCertContent:   string(rootCertContent),
		ServerCertContent: string(serverCertContent),
		ServerKeyContent:  string(ServerKeyContent),
	})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err = cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 2*time.Second, try.BodyContains("PathPrefix:/"))
	c.Assert(err, checker.IsNil)

	request, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:8443", nil)
	c.Assert(err, checker.IsNil)

	certificate, err := tls.LoadX509KeyPair(certPemPath, certKeyPath)
	c.Assert(err, checker.IsNil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{certificate},
		},
	}

	err = try.RequestWithTransport(request, 2*time.Second, tr, try.BodyContains("Forwarded-Tls-Client-Cert: MIIDKjCCAhICCQDKAJTeuq3LHjANBgkqhkiG9w0BAQsFADBXMQswCQYDVQQGEwJGUjEPMA0GA1UECAwGRlJBTkNFMREwDwYDVQQHDAhUT1VMT1VTRTETMBEGA1UECgwKY29udGFpbm91czEPMA0GA1UEAwwGc2VydmVyMB4XDTE4MDMyMTEzNDM0MVoXDTIxMDEwODEzNDM0MVowVzELMAkGA1UEBhMCRlIxDzANBgNVBAgMBkZSQU5DRTERMA8GA1UEBwwIVE9VTE9VU0UxEzARBgNVBAoMCmNvbnRhaW5vdXMxDzANBgNVBAMMBnNlcnZlcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKNHrqc7QcRHIc%2F%2FQW3oAcyl9%2BWFLdEtl86f5hTPoV0MpVgxwc98BA%2B0fPb97GOnj05P7QE%2BZerio5kP80ZUBX%2B0LVVilLWKvK47hZ%2FfxHgvtt95sZFT%2B0AHLk%2Bk%2FD86FIMrFuk8d889fFQ0TJz4cdX4wNYwKt%2FiFNNwaWxc%2BwpGAsZBv9cFh5rAdeix9mzMSa82qaYdp0g51JKAE7oEiXnPg8U7V9YXYwGiSvybCMIqAPy8sumIBNqF%2B7kWQaLtGwN8tEw5xaCFQFaiEmFn7M0xg5cC%2Fkg%2Fz%2FRmGtfRmZOIpnafIyw%2F%2FifXi7hxu%2Ba5ETrxOMW0j2xiBpGThGE5ox8CAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAPYDdGyNWp7R9j2oxZEbQS4lb%2B2Ol1r6PFo%2FzmpB6GK3CSNo65a0DtW%2FITeQi97MMgGS1D3wnaFPrwxtp0mEn7HjUuDcufHBqqBsjYC3NEtt%2ByyxNeYddLD%2FGdFXw4d6wNRdRaFCq5N1CPQzF4VTdoSLDxsOq%2FWAHHc2cyZyOprAqm2UXyWXxn4yWZqzDsZ41%2Fv2f3uMNxeqyIEtNZVzTKQBuwWw%2BjlQKGu0T8Ex1f0jaKI1OPtN5dzaIfO8acHcuNdmnE%2BhVsoqe17Dckxsj1ORf8ZcZ4qvULVouGINQBP4fcl5jv6TOm1U%2BZSk01FcHPmiDEMB6Utyy4ZLHPbmKYg%3D%3D"))
	c.Assert(err, checker.IsNil)
}
