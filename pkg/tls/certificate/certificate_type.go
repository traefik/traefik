package certificate

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"errors"
)

type CertificateType byte

const (
	RSA CertificateType = iota
	EC
)

var (
	certTypeToStringMap = map[CertificateType]string{
		RSA: "RSA",
		EC:  "EC",
	}
)

func (value CertificateType) String() string {
	return certTypeToStringMap[value]
}

func GetCertificateType(cert *tls.Certificate) (CertificateType, error) {
	switch cert.PrivateKey.(type) {
	case *ecdsa.PrivateKey, *ed25519.PrivateKey:
		return EC, nil
	case *rsa.PrivateKey:
		return RSA, nil
	}
	return 0, errors.New("unknown certificate type")
}
