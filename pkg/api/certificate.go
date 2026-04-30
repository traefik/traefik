package api

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

const (
	certStatusEnabled = "enabled"
	certStatusWarning = "warning"
	certStatusExpired = "expired"
)

// certificateRepresentation represents a certificate in the API.
type certificateRepresentation struct {
	Name                 string    `json:"name"` // SHA-256 fingerprint of the DER-encoded certificate.
	SANs                 []string  `json:"sans"`
	NotAfter             time.Time `json:"notAfter"`
	NotBefore            time.Time `json:"notBefore"`
	SerialNumber         string    `json:"serialNumber"`
	CommonName           string    `json:"commonName"`
	IssuerOrg            string    `json:"issuerOrg,omitempty"`
	IssuerCN             string    `json:"issuerCN,omitempty"`
	IssuerCountry        string    `json:"issuerCountry,omitempty"`
	Organization         string    `json:"organization,omitempty"`
	Country              string    `json:"country,omitempty"`
	Version              string    `json:"version"`
	KeyType              string    `json:"keyType"`
	KeySize              int       `json:"keySize,omitempty"`
	SignatureAlgorithm   string    `json:"signatureAlgorithm"`
	CertFingerprint      string    `json:"certFingerprint"`
	PublicKeyFingerprint string    `json:"publicKeyFingerprint"`
	Status               string    `json:"status"`
}

// Interface methods for sort.go compatibility.
func (c certificateRepresentation) name() string {
	return c.CommonName
}

func (c certificateRepresentation) status() string {
	return c.Status
}

func (c certificateRepresentation) issuer() string {
	if c.IssuerOrg != "" {
		return c.IssuerOrg
	}
	return c.IssuerCN
}

func (c certificateRepresentation) validUntil() time.Time {
	return c.NotAfter
}

// buildCertificateRepresentation builds a certificateRepresentation from an x509 certificate.
func buildCertificateRepresentation(cert *x509.Certificate) certificateRepresentation {
	keyType, keySize := extractKeyInfo(cert)
	certFingerprint, pubKeyFingerprint := extractFingerprints(cert)
	issuerOrg, issuerCN, issuerCountry := extractIssuerInfo(cert)
	organization, country := extractSubjectInfo(cert)

	return certificateRepresentation{
		Name:                 certFingerprint,
		SANs:                 extractSANs(cert),
		NotAfter:             cert.NotAfter,
		NotBefore:            cert.NotBefore,
		SerialNumber:         cert.SerialNumber.String(),
		CommonName:           cert.Subject.CommonName,
		IssuerOrg:            issuerOrg,
		IssuerCN:             issuerCN,
		IssuerCountry:        issuerCountry,
		Organization:         organization,
		Country:              country,
		Version:              formatVersion(cert.Version),
		KeyType:              keyType,
		KeySize:              keySize,
		SignatureAlgorithm:   cert.SignatureAlgorithm.String(),
		CertFingerprint:      certFingerprint,
		PublicKeyFingerprint: pubKeyFingerprint,
		Status:               getCertificateStatus(cert.NotAfter),
	}
}

// extractSANs extracts Subject Alternative Names from a certificate.
func extractSANs(cert *x509.Certificate) []string {
	sans := make([]string, 0, len(cert.DNSNames)+len(cert.IPAddresses))
	sans = append(sans, cert.DNSNames...)
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	sort.Strings(sans)
	return sans
}

// extractKeyInfo determines the key type and size from a certificate.
func extractKeyInfo(cert *x509.Certificate) (keyType string, keySize int) {
	keyType = "Unknown"
	keySize = 0

	switch pubKey := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		keyType = "RSA"
		keySize = pubKey.N.BitLen()
	case *ecdsa.PublicKey:
		keyType = "ECDSA"
		keySize = pubKey.Curve.Params().BitSize
	}

	return keyType, keySize
}

// extractFingerprints calculates SHA-256 fingerprints for certificate and public key.
func extractFingerprints(cert *x509.Certificate) (certFingerprint, pubKeyFingerprint string) {
	certHash := sha256.Sum256(cert.Raw)
	certFingerprint = hex.EncodeToString(certHash[:])

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err == nil {
		pubKeyHash := sha256.Sum256(pubKeyBytes)
		pubKeyFingerprint = hex.EncodeToString(pubKeyHash[:])
	}

	return certFingerprint, pubKeyFingerprint
}

// extractIssuerInfo extracts issuer information from a certificate.
func extractIssuerInfo(cert *x509.Certificate) (org, cn, country string) {
	if len(cert.Issuer.Organization) > 0 {
		org = cert.Issuer.Organization[0]
	}
	cn = cert.Issuer.CommonName
	if len(cert.Issuer.Country) > 0 {
		country = cert.Issuer.Country[0]
	}
	return org, cn, country
}

// extractSubjectInfo extracts subject information from a certificate.
func extractSubjectInfo(cert *x509.Certificate) (organization, country string) {
	if len(cert.Subject.Organization) > 0 {
		organization = cert.Subject.Organization[0]
	}
	if len(cert.Subject.Country) > 0 {
		country = cert.Subject.Country[0]
	}
	return organization, country
}

// formatVersion formats the X.509 version for display.
func formatVersion(version int) string {
	return fmt.Sprintf("v%d", version)
}

// getCertificateStatus returns the status of a certificate based on its expiry.
func getCertificateStatus(notAfter time.Time) string {
	remaining := time.Until(notAfter)
	if remaining < 0 {
		return certStatusExpired
	}
	// Show warning for certificates with validity less than 30 days left.
	if remaining < 30*24*time.Hour {
		return certStatusWarning
	}
	return certStatusEnabled
}
