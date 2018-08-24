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

const xForwardedTLSClientCert = "X-Forwarded-Tls-Client-Cert"
const xForwardedTLSClientCertInfos = "X-Forwarded-Tls-Client-Cert-Infos"

// TLSClientCertificateInfos is a struct for specifying the configuration for the tlsClientHeaders middleware.
type TLSClientCertificateInfos struct {
	NotAfter  bool
	NotBefore bool
	Subject   *TLSCLientCertificateSubjectInfos
	Sans      bool
}

// TLSCLientCertificateSubjectInfos contains the configuration for the certificate subject infos.
type TLSCLientCertificateSubjectInfos struct {
	Country      bool
	Province     bool
	Locality     bool
	Organization bool
	CommonName   bool
	SerialNumber bool
}

// TLSClientHeaders is a middleware that helps setup a few tls infos features.
type TLSClientHeaders struct {
	PEM   bool                       // pass the sanitized pem to the backend in a specific header
	Infos *TLSClientCertificateInfos // pass selected informations from the client certificate
}

func newTLSCLientCertificateSubjectInfos(infos *types.TLSCLientCertificateSubjectInfos) *TLSCLientCertificateSubjectInfos {
	if infos == nil {
		return nil
	}

	return &TLSCLientCertificateSubjectInfos{
		SerialNumber: infos.SerialNumber,
		CommonName:   infos.CommonName,
		Country:      infos.Country,
		Locality:     infos.Locality,
		Organization: infos.Organization,
		Province:     infos.Province,
	}
}

func newTLSClientInfos(infos *types.TLSClientCertificateInfos) *TLSClientCertificateInfos {
	if infos == nil {
		return nil
	}

	return &TLSClientCertificateInfos{
		NotBefore: infos.NotBefore,
		NotAfter:  infos.NotAfter,
		Sans:      infos.Sans,
		Subject:   newTLSCLientCertificateSubjectInfos(infos.Subject),
	}
}

// NewTLSClientHeaders constructs a new TLSClientHeaders instance from supplied frontend header struct.
func NewTLSClientHeaders(frontend *types.Frontend) *TLSClientHeaders {
	if frontend == nil {
		return nil
	}

	var pem bool
	var infos *TLSClientCertificateInfos

	if frontend.PassTLSClientCert != nil {
		conf := frontend.PassTLSClientCert
		pem = conf.PEM
		infos = newTLSClientInfos(conf.Infos)
	}

	return &TLSClientHeaders{
		PEM:   pem,
		Infos: infos,
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

// getSubjectInfos extract the requested informations from the certificate subject
func (s *TLSClientHeaders) getSubjectInfos(cs *pkix.Name) string {
	var subject string

	if s.Infos != nil && s.Infos.Subject != nil {
		options := s.Infos.Subject

		var content []string

		if options.Country && len(cs.Country) > 0 {
			content = append(content, fmt.Sprintf("C=%s", cs.Country[0]))
		}

		if options.Province && len(cs.Province) > 0 {
			content = append(content, fmt.Sprintf("ST=%s", cs.Province[0]))
		}

		if options.Locality && len(cs.Locality) > 0 {
			content = append(content, fmt.Sprintf("L=%s", cs.Locality[0]))
		}

		if options.Organization && len(cs.Organization) > 0 {
			content = append(content, fmt.Sprintf("O=%s", cs.Organization[0]))
		}

		if options.CommonName && len(cs.CommonName) > 0 {
			content = append(content, fmt.Sprintf("CN=%s", cs.CommonName))
		}

		if len(content) > 0 {
			subject = `Subject="` + strings.Join(content, ",") + `"`
		}
	}

	return subject
}

// getXForwardedTLSClientCertInfos Build a string with the wanted client certificates informations
// like Subject="C=%s,ST=%s,L=%s,O=%s,CN=%s",NB=%d,NA=%d,SAN=%s;
func (s *TLSClientHeaders) getXForwardedTLSClientCertInfos(certs []*x509.Certificate) string {
	var headerValues []string

	for _, peerCert := range certs {
		var values []string
		var sans string
		var nb string
		var na string

		subject := s.getSubjectInfos(&peerCert.Subject)
		if len(subject) > 0 {
			values = append(values, subject)
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
			headerContent := s.getXForwardedTLSClientCertInfos(r.TLS.PeerCertificates)
			r.Header.Set(xForwardedTLSClientCertInfos, url.QueryEscape(headerContent))
		} else {
			log.Warn("Try to extract certificate on a request without TLS")
		}
	}
}
