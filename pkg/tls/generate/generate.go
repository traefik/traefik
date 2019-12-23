package generate

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/containous/traefik/v2/pkg/tls/certificate"
)

// DefaultDomain Traefik domain for the default certificate
const DefaultDomain = "TRAEFIK DEFAULT CERT"

// DefaultCertificate generates random TLS certificates
func DefaultCertificate(certType certificate.CertificateType) (*tls.Certificate, error) {
	randomBytes := make([]byte, 100)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	zBytes := sha256.Sum256(randomBytes)
	z := hex.EncodeToString(zBytes[:sha256.Size])
	domain := fmt.Sprintf("%s.%s.traefik.default", z[:32], z[32:])

	certPEM, keyPEM, err := KeyPair(domain, time.Time{}, certType)
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}

// KeyPair generates cert and key files
func KeyPair(domain string, expiration time.Time, certType certificate.CertificateType) ([]byte, []byte, error) {
	var keyPEM []byte
	var privKey crypto.PrivateKey

	switch certType {
	case certificate.EC:
		ecdsaPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		ecdsaBytes, err := x509.MarshalECPrivateKey(ecdsaPrivKey)
		if err != nil {
			return nil, nil, err
		}
		keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: ecdsaBytes})
		privKey = ecdsaPrivKey
	case certificate.RSA:
		rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, nil, err
		}
		keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivKey)})
		privKey = rsaPrivKey
	}

	certPEM, err := PemCert(privKey, domain, expiration)
	if err != nil {
		return nil, nil, err
	}
	return certPEM, keyPEM, nil
}

// PemCert generates PEM cert file
func PemCert(privKey crypto.PrivateKey, domain string, expiration time.Time) ([]byte, error) {
	derBytes, err := derCert(privKey, expiration, domain)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}), nil
}

func derCert(privKey crypto.PrivateKey, expiration time.Time, domain string) ([]byte, error) {
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
			CommonName: DefaultDomain,
		},
		NotBefore: time.Now(),
		NotAfter:  expiration,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	var pubKey crypto.PublicKey
	switch v := privKey.(type) {
	case *ecdsa.PrivateKey:
		pubKey = &v.PublicKey
	case *ed25519.PrivateKey:
		edPubKey := v.Public()
		pubKey = &edPubKey
	case *rsa.PrivateKey:
		pubKey = &v.PublicKey
	default:
		return nil, errors.New("Unknown public key algorithm")
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, pubKey, privKey)
}
