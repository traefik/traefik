package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

func generateDefaultCertificate() (*tls.Certificate, error) {
	randomBytes := make([]byte, 100)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	zBytes := sha256.Sum256(randomBytes)
	z := hex.EncodeToString(zBytes[:sha256.Size])
	domain := fmt.Sprintf("%s.%s.traefik.default", z[:32], z[32:])

	certPEM, keyPEM, err := generateKeyPair(domain, time.Time{})
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}

func generateKeyPair(domain string, expiration time.Time) ([]byte, []byte, error) {
	rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivKey)})

	certPEM, err := generatePemCert(rsaPrivKey, domain, expiration)
	if err != nil {
		return nil, nil, err
	}
	return certPEM, keyPEM, nil
}

func generatePemCert(privKey *rsa.PrivateKey, domain string, expiration time.Time) ([]byte, error) {
	derBytes, err := generateDerCert(privKey, expiration, domain)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}), nil
}

func generateDerCert(privKey *rsa.PrivateKey, expiration time.Time, domain string) ([]byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	if expiration.IsZero() {
		expiration = time.Now().Add(365)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "TRAEFIK DEFAULT CERT",
		},
		NotBefore: time.Now(),
		NotAfter:  expiration,

		KeyUsage:              x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
}

// TLSSNI01ChallengeCert returns a certificate and target domain for the `tls-sni-01` challenge
func TLSSNI01ChallengeCert(keyAuth string) (ChallengeCert, string, error) {
	// generate a new RSA key for the certificates
	var tempPrivKey crypto.PrivateKey
	tempPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return ChallengeCert{}, "", err
	}
	rsaPrivKey := tempPrivKey.(*rsa.PrivateKey)
	rsaPrivPEM := pemEncode(rsaPrivKey)

	zBytes := sha256.Sum256([]byte(keyAuth))
	z := hex.EncodeToString(zBytes[:sha256.Size])
	domain := fmt.Sprintf("%s.%s.acme.invalid", z[:32], z[32:])
	tempCertPEM, err := generatePemCert(rsaPrivKey, domain, time.Time{})
	if err != nil {
		return ChallengeCert{}, "", err
	}

	certificate, err := tls.X509KeyPair(tempCertPEM, rsaPrivPEM)
	if err != nil {
		return ChallengeCert{}, "", err
	}

	return ChallengeCert{Certificate: tempCertPEM, PrivateKey: rsaPrivPEM, certificate: &certificate}, domain, nil
}
func pemEncode(data interface{}) []byte {
	var pemBlock *pem.Block
	switch key := data.(type) {
	case *ecdsa.PrivateKey:
		keyBytes, _ := x509.MarshalECPrivateKey(key)
		pemBlock = &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	case *rsa.PrivateKey:
		pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	case *x509.CertificateRequest:
		pemBlock = &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: key.Raw}
	case []byte:
		pemBlock = &pem.Block{Type: "CERTIFICATE", Bytes: []byte(data.([]byte))}
	}

	return pem.EncodeToMemory(pemBlock)
}
