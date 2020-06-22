package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/tls/generate"
)

var (
	// MinVersion Map of allowed TLS minimum versions.
	MinVersion = map[string]uint16{
		`VersionTLS10`: tls.VersionTLS10,
		`VersionTLS11`: tls.VersionTLS11,
		`VersionTLS12`: tls.VersionTLS12,
		`VersionTLS13`: tls.VersionTLS13,
	}

	// MaxVersion Map of allowed TLS maximum versions.
	MaxVersion = map[string]uint16{
		`VersionTLS10`: tls.VersionTLS10,
		`VersionTLS11`: tls.VersionTLS11,
		`VersionTLS12`: tls.VersionTLS12,
		`VersionTLS13`: tls.VersionTLS13,
	}

	// CurveIDs is a Map of TLS elliptic curves from crypto/tls
	// Available CurveIDs defined at https://godoc.org/crypto/tls#CurveID,
	// also allowing rfc names defined at https://tools.ietf.org/html/rfc8446#section-4.2.7
	CurveIDs = map[string]tls.CurveID{
		`secp256r1`: tls.CurveP256,
		`CurveP256`: tls.CurveP256,
		`secp384r1`: tls.CurveP384,
		`CurveP384`: tls.CurveP384,
		`secp521r1`: tls.CurveP521,
		`CurveP521`: tls.CurveP521,
		`x25519`:    tls.X25519,
		`X25519`:    tls.X25519,
	}
)

// Certificate holds a SSL cert/key pair
// Certs and Key could be either a file path, or the file content itself.
type Certificate struct {
	CertFile FileOrContent `json:"certFile,omitempty" toml:"certFile,omitempty" yaml:"certFile,omitempty"`
	KeyFile  FileOrContent `json:"keyFile,omitempty" toml:"keyFile,omitempty" yaml:"keyFile,omitempty"`
}

// Certificates defines traefik certificates type
// Certs and Keys could be either a file path, or the file content itself.
type Certificates []Certificate

// FileOrContent hold a file path or content.
type FileOrContent string

func (f FileOrContent) String() string {
	return string(f)
}

// IsPath returns true if the FileOrContent is a file path, otherwise returns false.
func (f FileOrContent) IsPath() bool {
	_, err := os.Stat(f.String())
	return err == nil
}

func (f FileOrContent) Read() ([]byte, error) {
	var content []byte
	if f.IsPath() {
		var err error
		content, err = ioutil.ReadFile(f.String())
		if err != nil {
			return nil, err
		}
	} else {
		content = []byte(f)
	}
	return content, nil
}

// CreateTLSConfig creates a TLS config from Certificate structures.
func (c *Certificates) CreateTLSConfig(entryPointName string) (*tls.Config, error) {
	config := &tls.Config{}
	domainsCertificates := make(map[string]map[string]*tls.Certificate)

	if c.isEmpty() {
		config.Certificates = []tls.Certificate{}

		cert, err := generate.DefaultCertificate()
		if err != nil {
			return nil, err
		}

		config.Certificates = append(config.Certificates, *cert)
	} else {
		for _, certificate := range *c {
			err := certificate.AppendCertificate(domainsCertificates, entryPointName)
			if err != nil {
				log.Errorf("Unable to add a certificate to the entryPoint %q : %v", entryPointName, err)
				continue
			}

			for _, certDom := range domainsCertificates {
				for _, cert := range certDom {
					config.Certificates = append(config.Certificates, *cert)
				}
			}
		}
	}
	return config, nil
}

// isEmpty checks if the certificates list is empty.
func (c *Certificates) isEmpty() bool {
	if len(*c) == 0 {
		return true
	}
	var key int
	for _, cert := range *c {
		if len(cert.CertFile.String()) != 0 && len(cert.KeyFile.String()) != 0 {
			break
		}
		key++
	}
	return key == len(*c)
}

// AppendCertificate appends a Certificate to a certificates map keyed by entrypoint.
func (c *Certificate) AppendCertificate(certs map[string]map[string]*tls.Certificate, ep string) error {
	certContent, err := c.CertFile.Read()
	if err != nil {
		return fmt.Errorf("unable to read CertFile : %w", err)
	}

	keyContent, err := c.KeyFile.Read()
	if err != nil {
		return fmt.Errorf("unable to read KeyFile : %w", err)
	}
	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return fmt.Errorf("unable to generate TLS certificate : %w", err)
	}

	parsedCert, _ := x509.ParseCertificate(tlsCert.Certificate[0])

	var SANs []string
	if parsedCert.Subject.CommonName != "" {
		SANs = append(SANs, strings.ToLower(parsedCert.Subject.CommonName))
	}
	if parsedCert.DNSNames != nil {
		sort.Strings(parsedCert.DNSNames)
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
	certKey := strings.Join(SANs, ",")

	certExists := false
	if certs[ep] == nil {
		certs[ep] = make(map[string]*tls.Certificate)
	} else {
		for domains := range certs[ep] {
			if domains == certKey {
				existingCertX509, err := x509.ParseCertificate(certs[ep][domains].Certificate[0])

				if err != nil {
					// If we can't parse the existing certificate then things have gone
					// really wrong at some point. In this case we continue so that the
					// existing certificate gets replaced by the one we want to append
					log.Errorf("Unable to parse existing certificate for domain(s) %s", certKey)
					break
				}

				if !CompareX509TimeBoundaries(existingCertX509, parsedCert) {
					certExists = true
				}

				break
			}
		}
	}

	if certExists {
		log.Debugf("Skipping addition of certificate for domain(s) %q, to EntryPoint %s, as it already exists for this Entrypoint.", certKey, ep)
	} else if _, found := certs[ep][certKey]; found {
		log.Debugf("Replacing certificate for domain(s) %s", certKey)
		certs[ep][certKey] = &tlsCert
	} else {
		log.Debugf("Adding certificate for domain(s) %s", certKey)
		certs[ep][certKey] = &tlsCert
	}

	return err
}

// GetTruncatedCertificateName truncates the certificate name.
func (c *Certificate) GetTruncatedCertificateName() string {
	certName := c.CertFile.String()

	// Truncate certificate information only if it's a well formed certificate content with more than 50 characters
	if !c.CertFile.IsPath() && strings.HasPrefix(certName, certificateHeader) && len(certName) > len(certificateHeader)+50 {
		certName = strings.TrimPrefix(c.CertFile.String(), certificateHeader)[:50]
	}

	return certName
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (c *Certificates) String() string {
	if len(*c) == 0 {
		return ""
	}
	var result []string
	for _, certificate := range *c {
		result = append(result, certificate.CertFile.String()+","+certificate.KeyFile.String())
	}
	return strings.Join(result, ";")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (c *Certificates) Set(value string) error {
	certificates := strings.Split(value, ";")
	for _, certificate := range certificates {
		files := strings.Split(certificate, ",")
		if len(files) != 2 {
			return fmt.Errorf("bad certificates format: %s", value)
		}
		*c = append(*c, Certificate{
			CertFile: FileOrContent(files[0]),
			KeyFile:  FileOrContent(files[1]),
		})
	}
	return nil
}

// Type is type of the struct.
func (c *Certificates) Type() string {
	return "certificates"
}

// CompareX509TimeBoundaries returns true if target certificate time boundaries are
// "better" than the origin certificate.
//
// origin & target have the same boundaries -> target
// origin & target are both valid -> the one with the greatest NotAfter boundary
// origin & target are both invalid ->
//     both have future boundaries -> the one with the smallest NotBefore boundary
//     both have past boundaries -> the one with the greatest NotAfter boundary
//     only origin has future boundaries -> origin
//     only target has future boundaries -> target
// origin only is valid -> origin
// target only is valid -> target
//
//nolint:gocritic,golint
func CompareX509TimeBoundaries(origin *x509.Certificate, target *x509.Certificate) bool {
	// If boundaries are the same we keep origin
	if origin.NotBefore.Equal(target.NotBefore) && origin.NotAfter.Equal(target.NotAfter) {
		return false
	}

	originIsValid := false
	targetIsValid := false
	now := time.Now()

	if now.After(origin.NotBefore) && now.Before(origin.NotAfter) {
		originIsValid = true
	}

	if now.After(target.NotBefore) && now.Before(target.NotAfter) {
		targetIsValid = true
	}

	if originIsValid && targetIsValid {
		// Both certs are valid, in this case we choose the one with the greatest
		// future time limit boundary (greatest NotAfter)
		if origin.NotAfter.After(target.NotAfter) {
			return false
		} else {
			return true
		}
	} else if !originIsValid && !targetIsValid {
		// Both certs are invalid
		if now.Before(origin.NotBefore) && now.Before(target.NotBefore) {
			// Both certificates have future boundaries, in this case we take the
			// one that is going to be valid the soonest
			if origin.NotBefore.Before(target.NotBefore) {
				return false
			} else {
				return true
			}
		} else if now.After(origin.NotAfter) && now.After(target.NotAfter) {
			// Both certificates have past boundaries, in this case we take the one
			// with the NotAfter boundary the closest to now
			if origin.NotAfter.After(target.NotAfter) {
				return false
			} else {
				return true
			}
		} else {
			// One certificate has past boundaries, one has future boundaries, then
			// in this case we choose the one in the future
			if origin.NotBefore.After(now) {
				return false
			} else {
				return true
			}
		}
	} else if originIsValid && !targetIsValid {
		// Only the origin certificate is valid
		return false
	} else if !originIsValid && targetIsValid {
		// Only the target certificate is valid
		return true
	}

	log.Warnf("CompareX509TimeBoundaries was not able to choose between 2 certificates")
	log.Debugf("origin CN=%s NotBefore=%s NotAfter=%s", origin.Subject.CommonName, origin.NotBefore, origin.NotAfter)
	log.Debugf("target CN=%s NotBefore=%s NotAfter=%s", target.Subject.CommonName, target.NotBefore, target.NotAfter)

	return false
}
