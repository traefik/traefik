package api

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// certificateRepresentation represents a certificate in the API
type certificateRepresentation struct {
	Name                 string    `json:"name"` // certKey (base64 encoded commonName + sorted SANs)
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
	Resolver             string    `json:"resolver,omitempty"`
}

// Interface methods for sort.go compatibility
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

func (c certificateRepresentation) resolver() string {
	return c.Resolver
}

func (c certificateRepresentation) validUntil() time.Time {
	return c.NotAfter
}

// buildCertificateRepresentation builds a certificateRepresentation from an x509 certificate
func buildCertificateRepresentation(cert *x509.Certificate, resolver ...string) certificateRepresentation {
	domains := getDomainsFromCert(cert)
	certKey := buildCertKey(domains)
	keyType, keySize := extractKeyInfo(cert)
	certFingerprint, pubKeyFingerprint := extractFingerprints(cert)
	issuerOrg, issuerCN, issuerCountry := extractIssuerInfo(cert)
	organization, country := extractSubjectInfo(cert)

	return certificateRepresentation{
		Name:                 certKey,
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
		Resolver:             extractResolver(resolver),
	}
}

// extractSANs extracts Subject Alternative Names from a certificate
func extractSANs(cert *x509.Certificate) []string {
	sans := make([]string, 0, len(cert.DNSNames)+len(cert.IPAddresses))
	sans = append(sans, cert.DNSNames...)
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	sort.Strings(sans)
	return sans
}

// extractKeyInfo determines the key type and size from a certificate
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

// extractFingerprints calculates SHA-256 fingerprints for certificate and public key
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

// extractIssuerInfo extracts issuer information from a certificate
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

// extractSubjectInfo extracts subject information from a certificate
func extractSubjectInfo(cert *x509.Certificate) (organization, country string) {
	if len(cert.Subject.Organization) > 0 {
		organization = cert.Subject.Organization[0]
	}
	if len(cert.Subject.Country) > 0 {
		country = cert.Subject.Country[0]
	}
	return organization, country
}

// formatVersion formats the X.509 version for display
func formatVersion(version int) string {
	return fmt.Sprintf("v%d", version)
}

// extractResolver extracts the resolver name from optional parameters
func extractResolver(resolver []string) string {
	if len(resolver) > 0 {
		return resolver[0]
	}
	return ""
}

// getDomainsFromURLEncodedCertKey URL-decodes and then decodes the base64-encoded certKey
func getDomainsFromURLEncodedCertKey(urlEncodedCertKey string) ([]string, error) {
	// URL decode first (certKey may have been URL-encoded when sent in the path)
	certKey, err := url.QueryUnescape(urlEncodedCertKey)
	if err != nil {
		return nil, fmt.Errorf("invalid URL encoding: %w", err)
	}
	
	return getDomainsFromCertKey(certKey)
}

// getDomainsFromCertKey decodes the base64-encoded certKey and returns the list of domains
func getDomainsFromCertKey(certKey string) ([]string, error) {
	// Decode base64 certKey
	decodedBytes, err := base64.URLEncoding.DecodeString(certKey)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate key encoding: %w", err)
	}
	decodedKey := string(decodedBytes)

	// Parse decoded key to get domains (comma-separated commonName + SANs)
	domains := strings.Split(decodedKey, ",")
	if len(domains) == 0 {
		return nil, fmt.Errorf("invalid certificate key: no domains found")
	}

	// Normalize domains
	normalized := make([]string, len(domains))
	for i, domain := range domains {
		normalized[i] = strings.ToLower(strings.TrimSpace(domain))
	}
	sort.Strings(normalized)

	return normalized, nil
}

// buildCertKey creates a base64-encoded certKey from a list of domains
// Deduplicates, lowercases, sorts, and encodes domains
func buildCertKey(domains []string) string {
	// Deduplicate and lowercase using map
	domainsMap := make(map[string]bool)
	for _, domain := range domains {
		domainsMap[strings.ToLower(strings.TrimSpace(domain))] = true
	}
	
	// Convert to sorted slice
	unique := make([]string, 0, len(domainsMap))
	for domain := range domainsMap {
		unique = append(unique, domain)
	}
	sort.Strings(unique)
	
	return base64.URLEncoding.EncodeToString([]byte(strings.Join(unique, ",")))
}

// getDomainsFromCert extracts all domains (CN + SANs) from an x509 certificate
// Returns raw domains without normalization (buildCertKey will normalize)
func getDomainsFromCert(cert *x509.Certificate) []string {
	domains := make([]string, 0, 1+len(cert.DNSNames)+len(cert.IPAddresses))
	
	// Add commonName
	if cert.Subject.CommonName != "" {
		domains = append(domains, cert.Subject.CommonName)
	}
	
	// Add DNS SANs
	domains = append(domains, cert.DNSNames...)
	
	// Add IP SANs
	for _, ip := range cert.IPAddresses {
		domains = append(domains, ip.String())
	}
	
	return domains
}

// getCertificateStatus returns the status of a certificate based on its expiry
func getCertificateStatus(notAfter time.Time) string {
	daysLeft := int(time.Until(notAfter).Hours() / 24)
	if daysLeft < 0 {
		return "disabled"
	}
	if daysLeft < 30 {
		return "warning"
	}
	return "enabled"
}

// sortCertificates sorts certificates based on the given field and direction

