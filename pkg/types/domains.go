package types

import (
	"strings"
)

// +k8s:deepcopy-gen=true

// Domain holds a domain name with SANs.
type Domain struct {
	// Main defines the main domain name.
	Main string `description:"Default subject name." json:"main,omitempty" toml:"main,omitempty" yaml:"main,omitempty"`
	// SANs defines the subject alternative domain names.
	SANs []string `description:"Subject alternative names." json:"sans,omitempty" toml:"sans,omitempty" yaml:"sans,omitempty"`
}

// ToStrArray convert a domain into an array of strings.
func (d *Domain) ToStrArray() []string {
	var domains []string
	if len(d.Main) > 0 {
		domains = []string{d.Main}
	}
	return append(domains, d.SANs...)
}

// Set sets a domains from an array of strings.
func (d *Domain) Set(domains []string) {
	if len(domains) > 0 {
		d.Main = domains[0]
		d.SANs = domains[1:]
	}
}

// MatchDomain returns true if a domain match the cert domain.
func MatchDomain(domain, certDomain string) bool {
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

// CanonicalDomain returns a lower case domain with trim space.
func CanonicalDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}
