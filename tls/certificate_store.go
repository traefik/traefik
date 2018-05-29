package tls

import (
	"crypto/tls"

	"github.com/containous/traefik/safe"
)

// CertificateStore store for dynamic and static certificates
type CertificateStore struct {
	DynamicCerts *safe.Safe
	StaticCerts  *safe.Safe
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
