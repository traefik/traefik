package tls

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/crypto/ocsp"
)

// CertificateData holds runtime data for runtime TLS certificate handling.
type CertificateData struct {
	config       *Certificate
	Certificate  *tls.Certificate
	OCSPServer   []string
	OCSPResponse *ocsp.Response
}

// CertificateCollection defines traefik CertificateCollection type.
type CertificateCollection []CertificateData

// AppendCertificate appends a Certificate to a certificates map keyed by store name.
func (c *CertificateData) AppendCertificate(certs map[string]map[string]*CertificateData, storeName string) error {
	certContent, err := c.config.CertFile.Read()
	if err != nil {
		return fmt.Errorf("unable to read CertFile : %w", err)
	}

	keyContent, err := c.config.KeyFile.Read()
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
		certs[storeName][certKey] = &CertificateData{
			Certificate: &tlsCert,
			OCSPServer:  parsedCert.OCSPServer,
			config: &Certificate{
				SANs: SANs,
				OCSP: types.OCSPConfig{
					DisableStapling: c.config.OCSP.DisableStapling,
				},
			},
		}
	}

	return err
}

func getOCSPForCert(certificate *CertificateData, issuedCertificate *x509.Certificate, issuerCertificate *x509.Certificate) ([]byte, *ocsp.Response, error) {
	if len(certificate.OCSPServer) == 0 {
		return nil, nil, errors.New("no OCSP server specified in certificate")
	}

	respURL := certificate.OCSPServer[0]
	ocspReq, err := ocsp.CreateRequest(issuedCertificate, issuerCertificate, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating OCSP request: %w", err)
	}

	reader := bytes.NewReader(ocspReq)
	req, err := http.Post(respURL, "application/ocsp-request", reader)
	if err != nil {
		return nil, nil, fmt.Errorf("making OCSP request: %w", err)
	}
	defer req.Body.Close()

	ocspResBytes, err := io.ReadAll(io.LimitReader(req.Body, 1024*1024))
	if err != nil {
		return nil, nil, fmt.Errorf("reading OCSP response: %w", err)
	}

	ocspRes, err := ocsp.ParseResponse(ocspResBytes, issuerCertificate)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing OCSP response: %w", err)
	}

	return ocspResBytes, ocspRes, nil
}

// StapleOCSP populates the ocsp response of the certificate if needed and not disabled by configuration.
func (c *CertificateData) StapleOCSP() error {
	if c.config.OCSP.DisableStapling {
		return nil
	}

	ocspResponse := c.OCSPResponse
	if ocspResponse != nil && time.Now().Before(ocspResponse.ThisUpdate.Add(ocspResponse.NextUpdate.Sub(ocspResponse.ThisUpdate)/2)) {
		return nil
	}

	leaf, _ := x509.ParseCertificate(c.Certificate.Certificate[0])
	var issuerCertificate *x509.Certificate
	if len(c.Certificate.Certificate) == 1 {
		issuerCertificate = leaf
	} else {
		ic, err := x509.ParseCertificate(c.Certificate.Certificate[1])
		if err != nil {
			return fmt.Errorf("cannot parse issuer certificate for %v: %w", c.config.SANs, err)
		}

		issuerCertificate = ic
	}

	ocspBytes, ocspResp, ocspErr := getOCSPForCert(c, leaf, issuerCertificate)
	if ocspErr != nil {
		return fmt.Errorf("no OCSP stapling for %v: %w", c.config.SANs, ocspErr)
	}

	log.Debug().Msgf("ocsp response: %v", ocspResp)
	if ocspResp.Status == ocsp.Good {
		if ocspResp.NextUpdate.After(leaf.NotAfter) {
			return fmt.Errorf("invalid: OCSP response for %v valid after certificate expiration (%s)", c.config.SANs, leaf.NotAfter.Sub(ocspResp.NextUpdate))
		}

		c.Certificate.OCSPStaple = ocspBytes
		c.OCSPResponse = ocspResp
	}

	return nil
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
