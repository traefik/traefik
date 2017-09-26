package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
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

// DynamicConfigurations contains array of DynamicConfiguration deserialized from TOML file
type DynamicConfigurations struct {
	TLS []*DynamicConfiguration
}

// Certificate holds a SSL cert/key pair
// Certs and Key could be either a file path, or the file content itself
type Certificate struct {
	CertFile FileOrContent
	KeyFile  FileOrContent
}

// DynamicConfiguration allows mapping a TLS certificate to a list of entrypoints
type DynamicConfiguration struct {
	EntryPoints []string
	Certificate *Certificate
}

// TLS configures TLS for an entry point
type TLS struct {
	MinVersion    string `export:"true"`
	CipherSuites  []string
	Certificates  Certificates
	ClientCAFiles []string
}

// EntrypointsCertificates allows sorting certificates by entrypoint
type EntrypointsCertificates map[string][]*Certificate

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

// Get return the EntryPoints map
func (r *RootCAs) Get() interface{} {
	return RootCAs(*r)
}

// SetValue sets the EntryPoints map with val
func (r *RootCAs) SetValue(val interface{}) {
	*r = RootCAs(val.(RootCAs))
}

// Type is type of the struct
func (r *RootCAs) Type() string {
	return "rootcas"
}

// ConvertTLSDynamicsToTLSConfiguration converts TLS configuration sorted by Certificates into TLS configuration sorted by EntryPoints
func (t *DynamicConfigurations) ConvertTLSDynamicsToTLSConfiguration() *EntrypointsCertificates {
	tlsConfiguration := make(EntrypointsCertificates)
	for _, tlsDynamic := range t.TLS {
		for _, ep := range tlsDynamic.EntryPoints {
			tlsConfiguration[ep] = append(tlsConfiguration[ep], tlsDynamic.Certificate)
		}
	}
	return &tlsConfiguration
}

// Diff return a EntrypointsCertificates which contains for all entrypoints defined in c, all the certificates which are missing in the otherConfiguration
func (c *EntrypointsCertificates) Diff(otherConfiguration *EntrypointsCertificates) map[string]Certificates {
	diffConf := make(map[string]Certificates)
	if otherConfiguration != nil {
		otherMap := make(EntrypointsCertificates)
		for key, value := range *otherConfiguration {
			otherMap[key] = value
		}
		for ep, certsToCheck := range *c {
			otherCerts, exists := otherMap[ep]
			if exists {
				otherCertsMap := make(map[Certificate]struct{}, len(otherCerts))
				for _, certif := range otherCerts {
					otherCertsMap[*certif] = struct{}{}
				}
				for _, certToCheck := range certsToCheck {
					_, exists := otherCertsMap[*certToCheck]
					if !exists {
						diffConf[ep] = append(diffConf[ep], *certToCheck)
					}
				}
			} else {
				for _, newCert := range certsToCheck {
					diffConf[ep] = append(diffConf[ep], *newCert)
				}
			}
		}
	} else {
		for ep, newCerts := range *c {
			for _, newCert := range newCerts {
				diffConf[ep] = append(diffConf[ep], *newCert)
			}
		}
	}
	return diffConf
}

//CreateTLSConfig creates a TLS config from Certificate structures
func (c *Certificates) CreateTLSConfig() (*tls.Config, map[string]*tls.Certificate, error) {
	config := &tls.Config{}
	certsMap := make(map[string]*tls.Certificate)
	if len(*c) == 0 {
		config.Certificates = make([]tls.Certificate, 0)
		cert, err := GenerateDefaultCertificate()
		if err != nil {
			return nil, nil, err
		}
		config.Certificates = append(config.Certificates, *cert)
	} else {
		err := c.AppendCertificates(certsMap)
		if err != nil {
			return nil, nil, err
		}
		for _, cert := range certsMap {
			config.Certificates = append(config.Certificates, *cert)
		}
	}
	return config, certsMap, nil
}

// AppendCertificates appends a Certificate to a certificates map sorted by entrypoints
func (c *Certificates) AppendCertificates(certs map[string]*tls.Certificate) error {
	certsSlice := []Certificate(*c)
	for _, cert := range certsSlice {
		certContent, err := cert.CertFile.Read()
		if err != nil {
			return err
		}

		keyContent, err := cert.KeyFile.Read()
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
		_, exists := certs[certKey]
		if exists {
			log.Warnf("Try to add certificate for domains which already have a certificate (%s). The new certificate will not be appened to the EntryPoint.", certKey)
		} else {
			log.Debugf("Add certificate for domains %s", certKey)
			certs[certKey] = &tlsCert
		}
	}
	return nil
}

// DeleteCertificates deletes a Certificate from a certificates map sorted by entrypoints
func (c *Certificates) DeleteCertificates(certs map[string]*tls.Certificate) error {

	for _, cert := range *c {
		certContent, err := cert.CertFile.Read()
		if err != nil {
			return err
		}

		keyContent, err := cert.KeyFile.Read()
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

		storedCert, exists := certs[certKey]
		if exists && reflect.DeepEqual(*storedCert, tlsCert) {
			log.Debugf("Delete certificate for domains %s", certKey)
			delete(certs, certKey)
		} else if exists {
			log.Warnf("Try to delete certificate for domains which already have another one (%s). This certificate will not be deleted from the EntryPoint.", certKey)
		}
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
