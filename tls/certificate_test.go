package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificates_CreateTLSConfig(t *testing.T) {
	var cert Certificates
	certificate1, err := generateCertificate("test1.traefik.com")
	require.NoError(t, err)
	require.NotNil(t, certificate1)

	cert = append(cert, *certificate1)

	certificate2, err := generateCertificate("test2.traefik.com")
	require.NoError(t, err)
	require.NotNil(t, certificate2)

	cert = append(cert, *certificate2)

	config, err := cert.CreateTLSConfig("http")
	require.NoError(t, err)

	assert.Len(t, config.NameToCertificate, 2)
	assert.Contains(t, config.NameToCertificate, "test1.traefik.com")
	assert.Contains(t, config.NameToCertificate, "test2.traefik.com")
}

func generateCertificate(domain string) (*Certificate, error) {
	certPEM, keyPEM, err := keyPair(domain, time.Time{})
	if err != nil {
		return nil, err
	}

	return &Certificate{
		CertFile: FileOrContent(certPEM),
		KeyFile:  FileOrContent(keyPEM),
	}, nil
}

// keyPair generates cert and key files
func keyPair(domain string, expiration time.Time) ([]byte, []byte, error) {
	rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivKey)})

	certPEM, err := PemCert(rsaPrivKey, domain, expiration)
	if err != nil {
		return nil, nil, err
	}
	return certPEM, keyPEM, nil
}

// PemCert generates PEM cert file
func PemCert(privKey *rsa.PrivateKey, domain string, expiration time.Time) ([]byte, error) {
	derBytes, err := derCert(privKey, expiration, domain)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}), nil
}

func derCert(privKey *rsa.PrivateKey, expiration time.Time, domain string) ([]byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	if expiration.IsZero() {
		expiration = time.Now().Add(365 * (24 * time.Hour))
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: domain,
		},
		NotBefore: time.Now(),
		NotAfter:  expiration,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
}
