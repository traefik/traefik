package types

import (
	"fmt"
	"strings"
)

// Domain holds a domain name with SANs
type Domain struct {
	Main string
	SANs []string
}

// ToStrArray convert a domain into an array of strings
func (d *Domain) ToStrArray() []string {
	var domains []string
	if len(d.Main) > 0 {
		domains = []string{d.Main}
	}
	return append(domains, d.SANs...)
}

// Set sets a domains from an array of strings
func (d *Domain) Set(domains []string) {
	if len(domains) > 0 {
		d.Main = domains[0]
		d.SANs = domains[1:]
	}
}

// Domains parse []Domain
type Domains []Domain

// Set []Domain
func (ds *Domains) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}

	// get function
	slice := strings.FieldsFunc(str, fargs)
	if len(slice) < 1 {
		return fmt.Errorf("parse error ACME.Domain. Unable to parse %s", str)
	}

	d := Domain{
		Main: slice[0],
	}

	if len(slice) > 1 {
		d.SANs = slice[1:]
	}

	*ds = append(*ds, d)
	return nil
}

// Get []Domain
func (ds *Domains) Get() interface{} { return []Domain(*ds) }

// String returns []Domain in string
func (ds *Domains) String() string { return fmt.Sprintf("%+v", *ds) }

// SetValue sets []Domain into the parser
func (ds *Domains) SetValue(val interface{}) {
	*ds = val.([]Domain)
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
