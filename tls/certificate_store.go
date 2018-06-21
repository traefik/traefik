package tls

import (
	"crypto/tls"
	"sort"
	"strings"

	"github.com/containous/traefik/safe"
)

// CertificateStore store for dynamic and static certificates
type CertificateStore struct {
	DynamicCerts       *safe.Safe
	StaticCerts        *safe.Safe
	DefaultCertificate *tls.Certificate
	CertCache          map[string]*tls.Certificate
	SniStrict          bool
}

// GetAllDomains return a slice with all the certificate domain
func (c CertificateStore) GetAllDomains() []string {
	var allCerts []string

	// Get static certificates
	if c.StaticCerts != nil && c.StaticCerts.Get() != nil {
		for domains := range c.StaticCerts.Get().(map[string]*tls.Certificate) {
			allCerts = append(allCerts, domains)
		}
	}

	// Get dynamic certificates
	if c.DynamicCerts != nil && c.DynamicCerts.Get() != nil {
		for domains := range c.DynamicCerts.Get().(map[string]*tls.Certificate) {
			allCerts = append(allCerts, domains)
		}
	}
	return allCerts
}

// GetBestCertificate returns the best match certificate, and caches the response
func (c CertificateStore) GetBestCertificate(domainToCheck string) *tls.Certificate {
	if c.CertCache == nil {
		c.CertCache = map[string]*tls.Certificate{}
	}

	if c.CertCache[domainToCheck] != nil {
		return c.CertCache[domainToCheck]
	}

	matchedCerts := map[string]*tls.Certificate{}
	if c.DynamicCerts.Get() != nil {
		for domains, cert := range c.DynamicCerts.Get().(map[string]*tls.Certificate) {
			for _, certDomain := range strings.Split(domains, ",") {
				if MatchDomain(domainToCheck, certDomain) {
					matchedCerts[domainToCheck] = cert
				}
			}
		}
	}
	if c.StaticCerts.Get() != nil {
		for domains, cert := range c.StaticCerts.Get().(map[string]*tls.Certificate) {
			for _, certDomain := range strings.Split(domains, ",") {
				if MatchDomain(domainToCheck, certDomain) {
					matchedCerts[domainToCheck] = cert
				}
			}
		}
	}
	if len(matchedCerts) > 0 {
		//sort map by keys
		keys := make([]string, 0, len(matchedCerts))
		for k := range matchedCerts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		c.CertCache[domainToCheck] = matchedCerts[keys[len(keys)-1]]
		return c.CertCache[domainToCheck]

	}
	return nil
}

// ContainsCertificates checks if there are any certs in the store
func (c CertificateStore) ContainsCertificates() bool {
	return c.StaticCerts.Get() != nil || c.DynamicCerts.Get() != nil
}

// ResetCache clears the cache in the store
func (c CertificateStore) ResetCache() {
	c.CertCache = map[string]*tls.Certificate{}
}

// MatchDomain return true if a domain match the cert domain
func MatchDomain(domain string, certDomain string) bool {
	if domain == certDomain {
		return true
	}

	for len(certDomain) > 0 && certDomain[len(certDomain)-1] == '.' {
		certDomain = certDomain[:len(certDomain)-1]
	}

	labels := strings.Split(domain, ".")
	for i := range labels {
		labels[i] = "*"
		candidate := strings.Join(labels, ".")
		if certDomain == candidate {
			return true
		}
	}
	return false
}
