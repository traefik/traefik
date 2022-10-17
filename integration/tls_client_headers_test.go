package integration

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
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
	s.composeUp(c)
}

func (s *TLSClientHeadersSuite) TestTLSClientHeaders(c *check.C) {
	rootCertContent, err := os.ReadFile(rootCertPath)
	c.Assert(err, check.IsNil)
	serverCertContent, err := os.ReadFile(certPemPath)
	c.Assert(err, check.IsNil)
	ServerKeyContent, err := os.ReadFile(certKeyPath)
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
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("PathPrefix(`/foo`)"))
	c.Assert(err, checker.IsNil)

	request, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:8443/foo", nil)
	c.Assert(err, checker.IsNil)

	certificate, err := tls.LoadX509KeyPair(certPemPath, certKeyPath)
	c.Assert(err, checker.IsNil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{certificate},
		},
	}

	err = try.RequestWithTransport(request, 2*time.Second, tr, try.BodyContains("Forwarded-Tls-Client-Cert: MIIDNTCCAh0CFD0QQcHXUJuKwMBYDA+bBExVSP26MA0GCSqGSIb3DQEBCwUAMFYxCzAJBgNVBAYTAkZSMQ8wDQYDVQQIDAZGcmFuY2UxFTATBgNVBAoMDFRyYWVmaWsgTGFiczEQMA4GA1UECwwHdHJhZWZpazENMAsGA1UEAwwEcm9vdDAeFw0yMTAxMDgxNzQ0MjRaFw0zMTAxMDYxNzQ0MjRaMFgxCzAJBgNVBAYTAkZSMQ8wDQYDVQQIDAZGcmFuY2UxFTATBgNVBAoMDFRyYWVmaWsgTGFiczEQMA4GA1UECwwHdHJhZWZpazEPMA0GA1UEAwwGc2VydmVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvYK2z8gLPOfFLgXNWP2460aeJ9vrH47x/lhKLlv4amSDHDx8Cmz/6blOUM8XOfMRW1xx++AgChWN9dx/kf7G2xlA5grZxRvUQ6xj7AvFG9TQUA3muNh2hvm9c3IjaZBNKH27bRKuDIBvZBvXdX4NL/aaFy7w7v7IKxk8j4WkfB23sgyH43g4b7NqKHJugZiedFu5GALmtLbShVOFbjWcre7Wvatdw8dIBmiFJqZQT3UjIuGAgqczIShtLxo4V+XyVkIPmzfPrRV+4zoMFIFOIaj3syyxb4krPBtxhe7nz2cWvvq0wePB2y4YbAAoVY8NYpd5JsMFwZtG6Uk59ygv4QIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQDaPg69wNeFNFisfBJTrscqVCTW+B80gMhpLdxXD+KO0/Wgc5xpB/wLSirNtRQyxAa3+EEcIwJv/wdh8EyjlDLSpFm/8ghntrKhkOfIOPDFE41M5HNfx/Fuh5btKEenOL/XdapqtNUt2ZE4RrsfbL79sPYepa9kDUVi2mCbeH5ollZ0MDU68HpB2YwHbCEuQNk5W3pjYK2NaDkVnxTkfEDM1k+3QydO1lqB5JJmcrs59BEveTqaJ3eeh/0I4OOab6OkTTZ0JNjJp1573oxO+fce/bfGud8xHY5gSN9huU7U6RsgvO7Dhmal/sDNl8XC8oU90hVDVXZdA7ewh4jjaoIv"))
	c.Assert(err, checker.IsNil)
}
