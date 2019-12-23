package certificate

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"errors"
)

// PublicKeyAlgorithmType defines which public key algorithm type a certificate has
type PublicKeyAlgorithmType byte

const (
	// RSA indicates an RSA public algorithm type certificate
	RSA PublicKeyAlgorithmType = iota

	// EC indicates an ECDSA or Ed25519 public algorithm type certificate
	EC
)

var (
	certTypeToStringMap = map[PublicKeyAlgorithmType]string{
		RSA: "RSA",
		EC:  "EC",
	}
)

// String return the string representation of this certificate type value
func (value PublicKeyAlgorithmType) String() string {
	return certTypeToStringMap[value]
}

// GetCertificateType determines the public algorithm type of the given certificate
func GetCertificateType(cert *tls.Certificate) (PublicKeyAlgorithmType, error) {
	switch cert.PrivateKey.(type) {
	case *ecdsa.PrivateKey, *ed25519.PrivateKey:
		return EC, nil
	case *rsa.PrivateKey:
		return RSA, nil
	}
	return 0, errors.New("unknown certificate type")
}
