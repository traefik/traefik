package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/traefik/traefik/v2/pkg/log"
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
	KeyFile  FileOrContent `json:"keyFile,omitempty" toml:"keyFile,omitempty" yaml:"keyFile,omitempty" loggable:"false"`
}

// Certificates defines traefik certificates type
// Certs and Keys could be either a file path, or the file content itself.
type Certificates []Certificate

// GetCertificates retrieves the certificates as slice of tls.Certificate.
func (c Certificates) GetCertificates() []tls.Certificate {
	var certs []tls.Certificate

	for _, certificate := range c {
		cert, err := certificate.GetCertificate()
		if err != nil {
			log.WithoutContext().Debugf("Error while getting certificate: %v", err)
			continue
		}

		certs = append(certs, cert)
	}

	return certs
}

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
		content, err = os.ReadFile(f.String())
		if err != nil {
			return nil, err
		}
	} else {
		content = []byte(f)
	}
	return content, nil
}

// AppendCertificate appends a Certificate to a certificates map keyed by store name.
func (c *Certificate) AppendCertificate(certs map[string]map[string]*tls.Certificate, storeName string) error {
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
		certs[storeName] = make(map[string]*tls.Certificate)
	} else {
		for domains := range certs[storeName] {
			if domains == certKey {
				certExists = true
				break
			}
		}
	}
	if certExists {
		log.Debugf("Skipping addition of certificate for domain(s) %q, to TLS Store %s, as it already exists for this store.", certKey, storeName)
	} else {
		log.Debugf("Adding certificate for domain(s) %s", certKey)
		certs[storeName][certKey] = &tlsCert
	}

	return err
}

// GetCertificate retrieves Certificate as tls.Certificate.
func (c *Certificate) GetCertificate() (tls.Certificate, error) {
	certContent, err := c.CertFile.Read()
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to read CertFile : %w", err)
	}

	keyContent, err := c.KeyFile.Read()
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to read KeyFile : %w", err)
	}

	cert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("unable to generate TLS certificate : %w", err)
	}

	return cert, nil
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

// VerifyPeerCertificate verifies the chain certificates and their URI.
func VerifyPeerCertificate(uri string, cfg *tls.Config, rawCerts [][]byte) error {
	// TODO: Refactor to avoid useless verifyChain (ex: when insecureskipverify is false)
	cert, err := verifyChain(cfg.RootCAs, rawCerts)
	if err != nil {
		return err
	}

	if len(uri) > 0 {
		return verifyServerCertMatchesURI(uri, cert)
	}

	return nil
}

// verifyServerCertMatchesURI is used on tls connections dialed to a server
// to ensure that the certificate it presented has the correct URI.
func verifyServerCertMatchesURI(uri string, cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("peer certificate mismatch: no peer certificate presented")
	}

	// Our certs will only ever have a single URI for now so only check that
	if len(cert.URIs) < 1 {
		return errors.New("peer certificate mismatch: peer certificate invalid")
	}

	gotURI := cert.URIs[0]

	// Override the hostname since we rely on x509 constraints to limit ability to spoof the trust domain if needed
	// (i.e. because a root is shared with other PKI or Consul clusters).
	// This allows for seamless migrations between trust domains.

	expectURI := &url.URL{}
	id, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("%q is not a valid URI", uri)
	}
	*expectURI = *id
	expectURI.Host = gotURI.Host

	if strings.EqualFold(gotURI.String(), expectURI.String()) {
		return nil
	}

	return fmt.Errorf("peer certificate mismatch got %s, want %s", gotURI, uri)
}

// verifyChain performs standard TLS verification without enforcing remote hostname matching.
func verifyChain(rootCAs *x509.CertPool, rawCerts [][]byte) (*x509.Certificate, error) {
	// Fetch leaf and intermediates. This is based on code form tls handshake.
	if len(rawCerts) < 1 {
		return nil, errors.New("tls: no certificates from peer")
	}

	certs := make([]*x509.Certificate, len(rawCerts))
	for i, asn1Data := range rawCerts {
		cert, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			return nil, fmt.Errorf("tls: failed to parse certificate from peer: %w", err)
		}

		certs[i] = cert
	}

	opts := x509.VerifyOptions{
		Roots:         rootCAs,
		Intermediates: x509.NewCertPool(),
	}

	// All but the first cert are intermediates
	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}

	_, err := certs[0].Verify(opts)
	if err != nil {
		return nil, err
	}

	return certs[0], nil
}
