package acme

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

func TestGeneratePrivateKey(t *testing.T) {
	key, err := generatePrivateKey(RSA2048)
	if err != nil {
		t.Error("Error generating private key:", err)
	}
	if key == nil {
		t.Error("Expected key to not be nil, but it was")
	}
}

func TestGenerateCSR(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	csr, err := generateCsr(key, "fizz.buzz", nil, true)
	if err != nil {
		t.Error("Error generating CSR:", err)
	}
	if csr == nil || len(csr) == 0 {
		t.Error("Expected CSR with data, but it was nil or length 0")
	}
}

func TestPEMEncode(t *testing.T) {
	buf := bytes.NewBufferString("TestingRSAIsSoMuchFun")

	reader := MockRandReader{b: buf}
	key, err := rsa.GenerateKey(reader, 32)
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	data := pemEncode(key)

	if data == nil {
		t.Fatal("Expected result to not be nil, but it was")
	}
	if len(data) != 127 {
		t.Errorf("Expected PEM encoding to be length 127, but it was %d", len(data))
	}
}

func TestPEMCertExpiration(t *testing.T) {
	privKey, err := generatePrivateKey(RSA2048)
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	expiration := time.Now().Add(365)
	expiration = expiration.Round(time.Second)
	certBytes, err := generateDerCert(privKey.(*rsa.PrivateKey), expiration, "test.com")
	if err != nil {
		t.Fatal("Error generating cert:", err)
	}

	buf := bytes.NewBufferString("TestingRSAIsSoMuchFun")

	// Some random string should return an error.
	if ctime, err := GetPEMCertExpiration(buf.Bytes()); err == nil {
		t.Errorf("Expected getCertExpiration to return an error for garbage string but returned %v", ctime)
	}

	// A DER encoded certificate should return an error.
	if _, err := GetPEMCertExpiration(certBytes); err == nil {
		t.Errorf("Expected getCertExpiration to return an error for DER certificates but returned none.")
	}

	// A PEM encoded certificate should work ok.
	pemCert := pemEncode(derCertificateBytes(certBytes))
	if ctime, err := GetPEMCertExpiration(pemCert); err != nil || !ctime.Equal(expiration.UTC()) {
		t.Errorf("Expected getCertExpiration to return %v but returned %v. Error: %v", expiration, ctime, err)
	}
}

type MockRandReader struct {
	b *bytes.Buffer
}

func (r MockRandReader) Read(p []byte) (int, error) {
	return r.b.Read(p)
}
