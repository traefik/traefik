package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/tls/generate"
)

var (
	// MinVersion Map of allowed TLS minimum versions
	MinVersion = map[string]uint16{
		`VersionTLS10`: tls.VersionTLS10,
		`VersionTLS11`: tls.VersionTLS11,
		`VersionTLS12`: tls.VersionTLS12,
	}

	// CipherSuites Map of TLS CipherSuites from crypto/tls
	// Available CipherSuites defined at https://golang.org/pkg/crypto/tls/#pkg-constants
	CipherSuites = map[string]uint16{
		`TLS_RSA_WITH_RC4_128_SHA`:                tls.TLS_RSA_WITH_RC4_128_SHA,
		`TLS_RSA_WITH_3DES_EDE_CBC_SHA`:           tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		`TLS_RSA_WITH_AES_128_CBC_SHA`:            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		`TLS_RSA_WITH_AES_256_CBC_SHA`:            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		`TLS_RSA_WITH_AES_128_CBC_SHA256`:         tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		`TLS_RSA_WITH_AES_128_GCM_SHA256`:         tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		`TLS_RSA_WITH_AES_256_GCM_SHA384`:         tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		`TLS_ECDHE_ECDSA_WITH_RC4_128_SHA`:        tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		`TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA`:    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		`TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA`:    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		`TLS_ECDHE_RSA_WITH_RC4_128_SHA`:          tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		`TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA`:     tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		`TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA`:      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		`TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA`:      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		`TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256`: tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		`TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256`:   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		`TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`:   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		`TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256`: tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		`TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`:   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		`TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384`: tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		`TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305`:    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		`TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305`:  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}
)

// Certificate holds a SSL cert/key pair
// Certs and Key could be either a file path, or the file content itself
type Certificate struct {
	CertFile FileOrContent
	KeyFile  FileOrContent
}

// Certificates defines traefik certificates type
// Certs and Keys could be either a file path, or the file content itself
type Certificates []Certificate

// FileOrContent hold a file path or content
type FileOrContent string

func (f FileOrContent) String() string {
	return string(f)
}

// IsPath returns true if the FileOrContent is a file path, otherwise returns false
func (f FileOrContent) IsPath() bool {
	_, err := os.Stat(f.String())
	return err == nil
}

func (f FileOrContent) Read() ([]byte, error) {
	var content []byte
	if _, err := os.Stat(f.String()); err == nil {
		content, err = ioutil.ReadFile(f.String())
		if err != nil {
			return nil, err
		}
	} else {
		content = []byte(f)
	}
	return content, nil
}

// CreateTLSConfig creates a TLS config from Certificate structures
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

// isEmpty checks if the certificates list is empty
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
		return fmt.Errorf("unable to read CertFile : %v", err)
	}

	keyContent, err := c.KeyFile.Read()
	if err != nil {
		return fmt.Errorf("unable to read KeyFile : %v", err)
	}
	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return fmt.Errorf("unable to generate TLS certificate : %v", err)
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
				certExists = true
				break
			}
		}
	}
	if certExists {
		log.Warnf("Skipping addition of certificate for domain(s) %q, to EntryPoint %s, as it already exists for this Entrypoint.", certKey, ep)
	} else {
		log.Debugf("Adding certificate for domain(s) %s", certKey)
		certs[ep][certKey] = &tlsCert
	}

	return err
}

func (c *Certificate) getTruncatedCertificateName() string {
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

// Type is type of the struct
func (c *Certificates) Type() string {
	return "certificates"
}
