package tls

import (
	"crypto/tls"
	"fmt"
	"strings"
)

// ClientCA defines traefik CA files for a entryPoint
// and it indicates if they are mandatory or have just to be analyzed if provided
type ClientCA struct {
	Files    []string
	Optional bool
}

// TLS configures TLS for an entry point
type TLS struct {
	MinVersion    string `export:"true"`
	CipherSuites  []string
	Certificates  Certificates
	ClientCAFiles []string // Deprecated
	ClientCA      ClientCA
}

// RootCAs hold the CA we want to have in root
type RootCAs []FileOrContent

// DomainsCertificates allows mapping TLS certificates to a list of domains
type DomainsCertificates map[string]*tls.Certificate

// Configuration allows mapping a TLS certificate to a list of entrypoints
type Configuration struct {
	EntryPoints []string
	Certificate *Certificate
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (dc *DomainsCertificates) add(domain string, cert *tls.Certificate) error {
	dc.Get().(map[string]*tls.Certificate)[domain] = cert
	return nil
}

// Get method allow getting the map stored into the DomainsCertificates
func (dc *DomainsCertificates) Get() interface{} {
	return map[string]*tls.Certificate(*dc)
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (r *RootCAs) String() string {
	sliceOfString := make([]string, len([]FileOrContent(*r)))
	for key, value := range *r {
		sliceOfString[key] = value.String()
	}
	return strings.Join(sliceOfString, ",")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (r *RootCAs) Set(value string) error {
	rootCAs := strings.Split(value, ",")
	if len(rootCAs) == 0 {
		return fmt.Errorf("bad RootCAs format: %s", value)
	}
	for _, rootCA := range rootCAs {
		*r = append(*r, FileOrContent(rootCA))
	}
	return nil
}

// Get return the RootCAs list
func (r *RootCAs) Get() interface{} {
	return RootCAs(*r)
}

// SetValue sets the RootCAs with val
func (r *RootCAs) SetValue(val interface{}) {
	*r = RootCAs(val.(RootCAs))
}

// Type is type of the struct
func (r *RootCAs) Type() string {
	return "rootcas"
}

// SortTLSConfigurationPerEntryPoints converts TLS configuration sorted by Certificates into TLS configuration sorted by EntryPoints
func SortTLSConfigurationPerEntryPoints(configurations []*Configuration, epConfiguration map[string]*DomainsCertificates) error {
	if epConfiguration == nil {
		epConfiguration = make(map[string]*DomainsCertificates)
	}
	for _, conf := range configurations {
		for _, ep := range conf.EntryPoints {
			if err := conf.Certificate.AppendCertificates(epConfiguration, ep); err != nil {
				return err
			}
		}
	}
	return nil
}
