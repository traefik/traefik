package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func main() {
	// generate CA
	caKey, caCert := generateCA("Test CA")
	saveKeyAndCert("integration/fixtures/ocsp/ca.key", "integration/fixtures/ocsp/ca.crt", caKey, caCert)

	// server certificate
	serverKey, serverCert := generateCert("server.local", caKey, caCert)
	saveKeyAndCert("integration/fixtures/ocsp/server.key", "integration/fixtures/ocsp/server.crt", serverKey, serverCert)

	// default certificate
	defaultKey, defaultCert := generateCert("default.local", caKey, caCert)
	saveKeyAndCert("integration/fixtures/ocsp/default.key", "integration/fixtures/ocsp/default.crt", defaultKey, defaultCert)
}

func generateCA(commonName string) (*ecdsa.PrivateKey, *x509.Certificate) {
	// generate a private key for the CA
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// create a self-signed CA certificate
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10 ans
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
		OCSPServer:            []string{"ocsp.example.com"},
	}

	caCertDER, _ := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	caCert, _ := x509.ParseCertificate(caCertDER)

	return caKey, caCert
}

func generateCert(commonName string, caKey *ecdsa.PrivateKey, caCert *x509.Certificate) (*ecdsa.PrivateKey, *x509.Certificate) {
	// create a private key for the certificate
	certKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// create a certificate signed by the CA
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * 365 * 24 * time.Hour), // 1 an
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		OCSPServer:            []string{"ocsp.example.com"},
	}

	certDER, _ := x509.CreateCertificate(rand.Reader, certTemplate, caCert, &certKey.PublicKey, caKey)
	cert, _ := x509.ParseCertificate(certDER)

	return certKey, cert
}

func saveKeyAndCert(keyFile, certFile string, key *ecdsa.PrivateKey, cert *x509.Certificate) {
	// save the private key
	keyOut, _ := os.Create(keyFile)
	defer keyOut.Close()

	// Marshal the private key to ASN.1 DER format
	privateKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		panic(err)
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKey})
	if err != nil {
		panic(err)
	}

	// save the certificate
	certOut, _ := os.Create(certFile)
	defer certOut.Close()
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		panic(err)
	}
}
