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

// ChallengeCert stores a challenge certificate
type ChallengeCert struct {
	Certificate    []byte
	PrivateKey     []byte
	TLSCertificate *tls.Certificate
}

// RootCAs hold the CA we want to have in root
type RootCAs []FileOrContent

// FileOrContent hold a file path or content
type FileOrContent string

// Certificate holds a SSL cert/key pair
// Certs and Key could be either a file path, or the file content itself
type Certificate struct {
	CertFile FileOrContent
	KeyFile  FileOrContent
}

// Configuration allows mapping a TLS certificate to a list of entrypoints
type Configuration struct {
	EntryPoints []string
	Certificate *Certificate
}

// DomainsCertificates allows mapping TLS certificates to a list of domains
type DomainsCertificates map[string]*tls.Certificate

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (dc *DomainsCertificates) add(domain string, cert *tls.Certificate) error {
	dc.Get().(map[string]*tls.Certificate)[domain] = cert
	return nil
}

// Get method allow getting the map stored into the DomainsCertificates
func (dc *DomainsCertificates) Get() interface{} {
	var domainCerts map[string]*tls.Certificate
	domainCerts = *dc
	return domainCerts
}

// TLS configures TLS for an entry point
type TLS struct {
	MinVersion    string `export:"true"`
	CipherSuites  []string
	Certificates  Certificates
	ClientCAFiles []string
}

// Certificates defines traefik certificates type
// Certs and Keys could be either a file path, or the file content itself
type Certificates []Certificate

func (f FileOrContent) String() string {
	return string(f)
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

// FromUserToServerTLSConfiguration converts TLS configuration sorted by Certificates into TLS configuration sorted by EntryPoints
func FromUserToServerTLSConfiguration(configurations []*Configuration, epConfiguration map[string]*DomainsCertificates) error {
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

//CreateTLSConfig creates a TLS config from Certificate structures
func (c *Certificates) CreateTLSConfig(entryPointName string) (*tls.Config, map[string]*DomainsCertificates, error) {
	config := &tls.Config{}
	domainsCertificates := make(map[string]*DomainsCertificates)
	if c.isEmpty() {
		config.Certificates = make([]tls.Certificate, 0)
		cert, err := GenerateDefaultCertificate()
		if err != nil {
			return nil, nil, err
		}
		config.Certificates = append(config.Certificates, *cert)
	} else {
		for _, certificate := range *c {
			err := certificate.AppendCertificates(domainsCertificates, entryPointName)
			if err != nil {
				return nil, nil, err
			}
			for _, certDom := range domainsCertificates {
				for _, cert := range certDom.Get().(map[string]*tls.Certificate) {
					config.Certificates = append(config.Certificates, *cert)
				}
			}
		}
	}
	return config, domainsCertificates, nil
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
	if key == len(*c) {
		return true
	}
	return false
}

// AppendCertificates appends a Certificate to a certificates map sorted by entrypoints
func (c *Certificate) AppendCertificates(certs map[string]*DomainsCertificates, ep string) error {

	certContent, err := c.CertFile.Read()
	if err != nil {
		return err
	}

	keyContent, err := c.KeyFile.Read()
	if err != nil {
		return err
	}
	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return err
	}

	parsedCert, _ := x509.ParseCertificate(tlsCert.Certificate[0])

	certKey := parsedCert.Subject.CommonName
	if parsedCert.DNSNames != nil {
		sort.Strings(parsedCert.DNSNames)
		certKey += fmt.Sprintf("%s,%s", parsedCert.Subject.CommonName, strings.Join(parsedCert.DNSNames, ","))
	}

	certExists := false
	if certs[ep] == nil {
		certs[ep] = new(DomainsCertificates)
		*certs[ep] = make(map[string]*tls.Certificate)
	} else {
		for domains := range *certs[ep] {
			if domains == certKey {
				certExists = true
				break
			}
		}
	}
	if certExists {
		log.Warnf("Into EntryPoint %s, try to add certificate for domains which already have a certificate (%s). The new certificate will not be append to the EntryPoint.", ep, certKey)
	} else {
		log.Debugf("Add certificate for domains %s", certKey)
		certs[ep].add(certKey, &tlsCert)
	}

	return nil
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
