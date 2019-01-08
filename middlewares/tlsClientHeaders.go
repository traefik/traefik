package middlewares

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

const (
	xForwardedTLSClientCert      = "X-Forwarded-Tls-Client-Cert"
	xForwardedTLSClientCertInfos = "X-Forwarded-Tls-Client-Cert-Infos"
)

var attributeTypeNames = map[string]string{
	"0.9.2342.19200300.100.1.25": "DC", // Domain component OID - RFC 2247
}

// TLSClientCertificateInfos is a struct for specifying the configuration for the tlsClientHeaders middleware.
type TLSClientCertificateInfos struct {
	Issuer    *DistinguishedNameOptions
	NotAfter  bool
	NotBefore bool
	Sans      bool
	Subject   *DistinguishedNameOptions
}

// DistinguishedNameOptions is a struct for specifying the configuration for the distinguished name info.
type DistinguishedNameOptions struct {
	CommonName          bool
	CountryName         bool
	DomainComponent     bool
	LocalityName        bool
	OrganizationName    bool
	SerialNumber        bool
	StateOrProvinceName bool
}

// TLSClientHeaders is a middleware that helps setup a few tls info features.
type TLSClientHeaders struct {
	Infos *TLSClientCertificateInfos // pass selected informations from the client certificate
	PEM   bool                       // pass the sanitized pem to the backend in a specific header
}

func newDistinguishedNameOptions(infos *types.TLSCLientCertificateDNInfos) *DistinguishedNameOptions {
	if infos == nil {
		return nil
	}

	return &DistinguishedNameOptions{
		CommonName:          infos.CommonName,
		CountryName:         infos.Country,
		DomainComponent:     infos.DomainComponent,
		LocalityName:        infos.Locality,
		OrganizationName:    infos.Organization,
		SerialNumber:        infos.SerialNumber,
		StateOrProvinceName: infos.Province,
	}
}

func newTLSClientInfos(infos *types.TLSClientCertificateInfos) *TLSClientCertificateInfos {
	if infos == nil {
		return nil
	}

	return &TLSClientCertificateInfos{
		Issuer:    newDistinguishedNameOptions(infos.Issuer),
		NotAfter:  infos.NotAfter,
		NotBefore: infos.NotBefore,
		Sans:      infos.Sans,
		Subject:   newDistinguishedNameOptions(infos.Subject),
	}
}

// NewTLSClientHeaders constructs a new TLSClientHeaders instance from supplied frontend header struct.
func NewTLSClientHeaders(frontend *types.Frontend) *TLSClientHeaders {
	if frontend == nil {
		return nil
	}

	var addPEM bool
	var infos *TLSClientCertificateInfos

	if frontend.PassTLSClientCert != nil {
		conf := frontend.PassTLSClientCert
		addPEM = conf.PEM
		infos = newTLSClientInfos(conf.Infos)
	}

	return &TLSClientHeaders{
		Infos: infos,
		PEM:   addPEM,
	}
}

func (s *TLSClientHeaders) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	s.ModifyRequestHeaders(r)
	// If there is a next, call it.
	if next != nil {
		next(w, r)
	}
}

// sanitize As we pass the raw certificates, remove the useless data and make it http request compliant
func sanitize(cert []byte) string {
	s := string(cert)
	r := strings.NewReplacer("-----BEGIN CERTIFICATE-----", "",
		"-----END CERTIFICATE-----", "",
		"\n", "")
	cleaned := r.Replace(s)

	return url.QueryEscape(cleaned)
}

// extractCertificate extract the certificate from the request
func extractCertificate(cert *x509.Certificate) string {
	b := pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
	certPEM := pem.EncodeToMemory(&b)
	if certPEM == nil {
		log.Error("Cannot extract the certificate content")
		return ""
	}
	return sanitize(certPEM)
}

// getXForwardedTLSClientCert Build a string with the client certificates
func getXForwardedTLSClientCert(certs []*x509.Certificate) string {
	var headerValues []string

	for _, peerCert := range certs {
		headerValues = append(headerValues, extractCertificate(peerCert))
	}

	return strings.Join(headerValues, ",")
}

// getSANs get the Subject Alternate Name values
func getSANs(cert *x509.Certificate) []string {
	var sans []string
	if cert == nil {
		return sans
	}

	sans = append(cert.DNSNames, cert.EmailAddresses...)

	var ips []string
	for _, ip := range cert.IPAddresses {
		ips = append(ips, ip.String())
	}
	sans = append(sans, ips...)

	var uris []string
	for _, uri := range cert.URIs {
		uris = append(uris, uri.String())
	}

	return append(sans, uris...)
}

func getDNInfos(prefix string, options *DistinguishedNameOptions, cs *pkix.Name) string {
	if options == nil {
		return ""
	}

	content := &strings.Builder{}

	// Manage non standard attributes
	for _, name := range cs.Names {
		// Domain Component - RFC 2247
		if options.DomainComponent && attributeTypeNames[name.Type.String()] == "DC" {
			content.WriteString(fmt.Sprintf("DC=%s,", name.Value))
		}
	}

	if options.CountryName {
		writeParts(content, cs.Country, "C")
	}

	if options.StateOrProvinceName {
		writeParts(content, cs.Province, "ST")
	}

	if options.LocalityName {
		writeParts(content, cs.Locality, "L")
	}

	if options.OrganizationName {
		writeParts(content, cs.Organization, "O")
	}

	if options.SerialNumber {
		writePart(content, cs.SerialNumber, "SN")
	}

	if options.CommonName {
		writePart(content, cs.CommonName, "CN")
	}

	if content.Len() > 0 {
		return prefix + `="` + strings.TrimSuffix(content.String(), ",") + `"`
	}

	return ""
}

func writeParts(content *strings.Builder, entries []string, prefix string) {
	for _, entry := range entries {
		writePart(content, entry, prefix)
	}
}

func writePart(content *strings.Builder, entry string, prefix string) {
	if len(entry) > 0 {
		content.WriteString(fmt.Sprintf("%s=%s,", prefix, entry))
	}
}

// getXForwardedTLSClientCertInfo Build a string with the wanted client certificates informations
// like Subject="DC=%s,C=%s,ST=%s,L=%s,O=%s,CN=%s",NB=%d,NA=%d,SAN=%s;
func (s *TLSClientHeaders) getXForwardedTLSClientCertInfo(certs []*x509.Certificate) string {
	var headerValues []string

	for _, peerCert := range certs {
		var values []string
		var sans string
		var nb string
		var na string

		if s.Infos != nil {
			subject := getDNInfos("Subject", s.Infos.Subject, &peerCert.Subject)
			if len(subject) > 0 {
				values = append(values, subject)
			}

			issuer := getDNInfos("Issuer", s.Infos.Issuer, &peerCert.Issuer)
			if len(issuer) > 0 {
				values = append(values, issuer)
			}
		}

		ci := s.Infos
		if ci != nil {
			if ci.NotBefore {
				nb = fmt.Sprintf("NB=%d", uint64(peerCert.NotBefore.Unix()))
				values = append(values, nb)
			}
			if ci.NotAfter {
				na = fmt.Sprintf("NA=%d", uint64(peerCert.NotAfter.Unix()))
				values = append(values, na)
			}

			if ci.Sans {
				sans = fmt.Sprintf("SAN=%s", strings.Join(getSANs(peerCert), ","))
				values = append(values, sans)
			}
		}

		value := strings.Join(values, ",")
		headerValues = append(headerValues, value)
	}

	return strings.Join(headerValues, ";")
}

// ModifyRequestHeaders set the wanted headers with the certificates informations
func (s *TLSClientHeaders) ModifyRequestHeaders(r *http.Request) {
	if s.PEM {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			r.Header.Set(xForwardedTLSClientCert, getXForwardedTLSClientCert(r.TLS.PeerCertificates))
		} else {
			log.Warn("Try to extract certificate on a request without TLS")
		}
	}

	if s.Infos != nil {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			headerContent := s.getXForwardedTLSClientCertInfo(r.TLS.PeerCertificates)
			r.Header.Set(xForwardedTLSClientCertInfos, url.QueryEscape(headerContent))
		} else {
			log.Warn("Try to extract certificate on a request without TLS")
		}
	}
}
