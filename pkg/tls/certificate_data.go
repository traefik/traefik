package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/types"
)

// CertificateData holds runtime data for runtime TLS certificate handling.
type CertificateData struct {
	config      *Certificate
	ocsp        *OCSP
	Certificate *tls.Certificate
}

// CertificateCollection defines traefik CertificateCollection type.
type CertificateCollection []CertificateData

// AppendCertificate appends a Certificate to a certificates map keyed by store name.
func (c *CertificateData) AppendCertificate(certs map[string]map[string]*CertificateData, storeName string) error {
	certContent, err := c.config.CertFile.Read()
	if err != nil {
		return fmt.Errorf("unable to read CertFile: %w", err)
	}

	keyContent, err := c.config.KeyFile.Read()
	if err != nil {
		return fmt.Errorf("unable to read KeyFile: %w", err)
	}
	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return fmt.Errorf("unable to generate TLS certificate: %w", err)
	}

	parsedCert, _ := x509.ParseCertificate(tlsCert.Certificate[0])

	var SANs []string
	if parsedCert.Subject.CommonName != "" {
		SANs = append(SANs, strings.ToLower(parsedCert.Subject.CommonName))
	}
	if parsedCert.DNSNames != nil {
		for _, dnsName := range parsedCert.DNSNames {
			if dnsName != parsedCert.Subject.CommonName {
				SANs = append(SANs, strings.ToLower(dnsName))
			}
		}
	}
	if parsedCert.IPAddresses != nil {
		for _, ip := range parsedCert.IPAddresses {
			if ip.String() != parsedCert.Subject.CommonName {
				SANs = append(SANs, strings.ToLower(ip.String()))
			}
		}
	}

	// Guarantees the order to produce a unique cert key.
	sort.Strings(SANs)
	certKey := strings.Join(SANs, ",")

	certExists := false
	if certs[storeName] == nil {
		certs[storeName] = make(map[string]*CertificateData)
	} else {
		for domains := range certs[storeName] {
			if domains == certKey {
				certExists = true
				break
			}
		}
	}
	if certExists {
		log.Debug().Msgf("Skipping addition of certificate for domain(s) %q, to TLS Store %s, as it already exists for this store.", certKey, storeName)
	} else {
		log.Debug().Msgf("Adding certificate for domain(s) %s", certKey)

		ocspClient, err := NewOCSP(c.config.OCSP, &tlsCert)
		if err != nil {
			log.Debug().Msgf("Error creating OCSP client for domain(s) %s, skipping", SANs[0])
		}

		certs[storeName][certKey] = &CertificateData{
			Certificate: &tlsCert,
			ocsp:        ocspClient,
			config: &Certificate{
				SANs: SANs,
			},
		}
	}

	return err
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (c *CertificateCollection) String() string {
	if len(*c) == 0 {
		return ""
	}
	var result []string
	for _, certificate := range *c {
		result = append(result, certificate.config.CertFile.String()+","+certificate.config.KeyFile.String())
	}
	return strings.Join(result, ";")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (c *CertificateCollection) Set(value string) error {
	certificates := strings.Split(value, ";")
	for _, certificate := range certificates {
		files := strings.Split(certificate, ",")
		if len(files) != 2 {
			return fmt.Errorf("bad certificates format: %s", value)
		}
		*c = append(*c, CertificateData{
			config: &Certificate{
				CertFile: types.FileOrContent(files[0]),
				KeyFile:  types.FileOrContent(files[1]),
				OCSP: types.OCSPConfig{
					DisableStapling: false,
				},
			},
		})
	}
	return nil
}

// Type is type of the struct.
func (c *CertificateCollection) Type() string {
	return "CertificateCollection"
}
