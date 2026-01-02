package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// startECHServer starts a TLS server that supports Encrypted Client Hello (ECH).
func startECHServer(bind string, cert tls.Certificate, echKey tls.EncryptedClientHelloKey) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, ECH-enabled TLS server!")
	})

	server := &http.Server{
		Addr:    bind,
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS13,
			EncryptedClientHelloKeys: []tls.EncryptedClientHelloKey{echKey},
		},
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func TestECH(t *testing.T) {
	const commonName = "server.local"

	echKey, err := NewECHKey(commonName)
	if err != nil {
		t.Fatalf("Failed to generate ECH key: %v", err)
	}

	echKeyBytes, err := MarshalECHKey(echKey)
	if err != nil {
		t.Fatalf("Failed to marshal ECH key: %v", err)
	}

	newKey, err := UnmarshalECHKey(echKeyBytes)
	if err != nil {
		t.Fatalf("Failed to unmarshal ECH key: %v", err)
	}

	if !reflect.DeepEqual(*echKey, *newKey) {
		t.Fatal("Parsed ECH key does not match original")
	}

	testCert, err := generateCert(commonName)
	if err != nil {
		t.Fatalf("Failed to load certs: %v", err)
	}

	echConfigList, err := ECHConfigToConfigList(echKey.Config)
	if err != nil {
		t.Fatalf("Failed to convert ECH config to config list: %v", err)
	}

	go startECHServer("localhost:8443", testCert, *echKey)
	time.Sleep(1 * time.Second) // Wait for the server to start
	response, err := RequestWithECH(ECHRequestConfig[[]byte]{
		URL:      "https://localhost:8443/",
		Host:     commonName,
		ECH:      echConfigList,
		Insecure: true,
	})

	if err != nil {
		t.Fatalf("Failed to make ECH request: %v", err)
	}
	if string(response) != "Hello, ECH-enabled TLS server!" {
		t.Fatalf("Unexpected response from ECH server: %s", response)
	}
}

func generateCert(commonName string) (tls.Certificate, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	keyBytes := x509.MarshalPKCS1PrivateKey(rsaKey)
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyBytes,
		},
	)

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * 10 * time.Hour)

	template := x509.Certificate{
		SerialNumber:          big.NewInt(0),
		Subject:               pkix.Name{CommonName: commonName},
		DNSNames:              []string{commonName},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyAgreement | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &rsaKey.PublicKey, rsaKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}
	certPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		},
	)

	return tls.X509KeyPair(certPEM, keyPEM)
}
