package certificate

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"errors"
)

// CertificateType defines the which public key algorithm type a certificate has
type CertificateType byte

const (
	// RSA indicates an RSA public algorithm type certificate
	RSA CertificateType = iota

	// EC indicates an ECDSA or Ed25519 public algorithm type certificate
	EC
)

var (
	certTypeToStringMap = map[CertificateType]string{
		RSA: "RSA",
		EC:  "EC",
	}
)

// String return the string representation of this certificate type value
func (value CertificateType) String() string {
	return certTypeToStringMap[value]
}

// GetCertificateType determines the public algorithm type of the given certificate
func GetCertificateType(cert *tls.Certificate) (CertificateType, error) {
	switch cert.PrivateKey.(type) {
	case *ecdsa.PrivateKey, *ed25519.PrivateKey:
		return EC, nil
	case *rsa.PrivateKey:
		return RSA, nil
	}
	return 0, errors.New("unknown certificate type")
}
