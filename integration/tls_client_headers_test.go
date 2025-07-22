package integration

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/try"
)

const (
	rootCertPath = "./fixtures/tlsclientheaders/root.pem"
	certPemPath  = "./fixtures/tlsclientheaders/server.pem"
	certKeyPath  = "./fixtures/tlsclientheaders/server.key"
)

type TLSClientHeadersSuite struct{ BaseSuite }

func TestTLSClientHeadersSuite(t *testing.T) {
	suite.Run(t, new(TLSClientHeadersSuite))
}

func (s *TLSClientHeadersSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	s.createComposeProject("tlsclientheaders")
	s.composeUp()
}

func (s *TLSClientHeadersSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *TLSClientHeadersSuite) TestTLSClientHeaders() {
	rootCertContent, err := os.ReadFile(rootCertPath)
	assert.NoError(s.T(), err)
	serverCertContent, err := os.ReadFile(certPemPath)
	assert.NoError(s.T(), err)
	ServerKeyContent, err := os.ReadFile(certKeyPath)
	assert.NoError(s.T(), err)

	file := s.adaptFile("fixtures/tlsclientheaders/simple.toml", struct {
		RootCertContent   string
		ServerCertContent string
		ServerKeyContent  string
	}{
		RootCertContent:   string(rootCertContent),
		ServerCertContent: string(serverCertContent),
		ServerKeyContent:  string(ServerKeyContent),
	})

	s.traefikCmd(withConfigFile(file))

	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 2*time.Second, try.BodyContains("PathPrefix(`/foo`)"))
	require.NoError(s.T(), err)

	request, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:8443/foo", nil)
	require.NoError(s.T(), err)

	certificate, err := tls.LoadX509KeyPair(certPemPath, certKeyPath)
	require.NoError(s.T(), err)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{certificate},
		},
	}

	err = try.RequestWithTransport(request, 2*time.Second, tr, try.BodyContains("Forwarded-Tls-Client-Cert: MIIDNTCCAh0CFD0QQcHXUJuKwMBYDA+bBExVSP26MA0GCSqGSIb3DQEBCwUAMFYxCzAJBgNVBAYTAkZSMQ8wDQYDVQQIDAZGcmFuY2UxFTATBgNVBAoMDFRyYWVmaWsgTGFiczEQMA4GA1UECwwHdHJhZWZpazENMAsGA1UEAwwEcm9vdDAeFw0yMTAxMDgxNzQ0MjRaFw0zMTAxMDYxNzQ0MjRaMFgxCzAJBgNVBAYTAkZSMQ8wDQYDVQQIDAZGcmFuY2UxFTATBgNVBAoMDFRyYWVmaWsgTGFiczEQMA4GA1UECwwHdHJhZWZpazEPMA0GA1UEAwwGc2VydmVyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvYK2z8gLPOfFLgXNWP2460aeJ9vrH47x/lhKLlv4amSDHDx8Cmz/6blOUM8XOfMRW1xx++AgChWN9dx/kf7G2xlA5grZxRvUQ6xj7AvFG9TQUA3muNh2hvm9c3IjaZBNKH27bRKuDIBvZBvXdX4NL/aaFy7w7v7IKxk8j4WkfB23sgyH43g4b7NqKHJugZiedFu5GALmtLbShVOFbjWcre7Wvatdw8dIBmiFJqZQT3UjIuGAgqczIShtLxo4V+XyVkIPmzfPrRV+4zoMFIFOIaj3syyxb4krPBtxhe7nz2cWvvq0wePB2y4YbAAoVY8NYpd5JsMFwZtG6Uk59ygv4QIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQDaPg69wNeFNFisfBJTrscqVCTW+B80gMhpLdxXD+KO0/Wgc5xpB/wLSirNtRQyxAa3+EEcIwJv/wdh8EyjlDLSpFm/8ghntrKhkOfIOPDFE41M5HNfx/Fuh5btKEenOL/XdapqtNUt2ZE4RrsfbL79sPYepa9kDUVi2mCbeH5ollZ0MDU68HpB2YwHbCEuQNk5W3pjYK2NaDkVnxTkfEDM1k+3QydO1lqB5JJmcrs59BEveTqaJ3eeh/0I4OOab6OkTTZ0JNjJp1573oxO+fce/bfGud8xHY5gSN9huU7U6RsgvO7Dhmal/sDNl8XC8oU90hVDVXZdA7ewh4jjaoIv"))
	require.NoError(s.T(), err)
}
