package passtlsclientcert

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
)

const (
	xForwardedTLSClientCert      = "X-Forwarded-Tls-Client-Cert"
	xForwardedTLSClientCertInfos = "X-Forwarded-Tls-Client-Cert-infos"
	typeName                     = "PassClientTLSCert"
)

// passTLSClientCert is a middleware that helps setup a few tls info features.
type passTLSClientCert struct {
	next  http.Handler
	name  string
	pem   bool                       // pass the sanitized pem to the backend in a specific header
	infos *tlsClientCertificateInfos // pass selected information from the client certificate
}

// New constructs a new PassTLSClientCert instance from supplied frontend header struct.
func New(ctx context.Context, next http.Handler, config config.PassTLSClientCert, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	return &passTLSClientCert{
		next:  next,
		name:  name,
		pem:   config.PEM,
		infos: newTLSClientInfos(config.Infos),
	}, nil
}

// tlsClientCertificateInfos is a struct for specifying the configuration for the passTLSClientCert middleware.
type tlsClientCertificateInfos struct {
	notAfter  bool
	notBefore bool
	subject   *tlsCLientCertificateSubjectInfos
	sans      bool
}

func newTLSClientInfos(infos *config.TLSClientCertificateInfos) *tlsClientCertificateInfos {
	if infos == nil {
		return nil
	}

	return &tlsClientCertificateInfos{
		notBefore: infos.NotBefore,
		notAfter:  infos.NotAfter,
		sans:      infos.Sans,
		subject:   newTLSCLientCertificateSubjectInfos(infos.Subject),
	}
}

// tlsCLientCertificateSubjectInfos contains the configuration for the certificate subject infos.
type tlsCLientCertificateSubjectInfos struct {
	country      bool
	province     bool
	locality     bool
	Organization bool
	commonName   bool
	serialNumber bool
}

func newTLSCLientCertificateSubjectInfos(infos *config.TLSCLientCertificateSubjectInfos) *tlsCLientCertificateSubjectInfos {
	if infos == nil {
		return nil
	}

	return &tlsCLientCertificateSubjectInfos{
		serialNumber: infos.SerialNumber,
		commonName:   infos.CommonName,
		country:      infos.Country,
		locality:     infos.Locality,
		Organization: infos.Organization,
		province:     infos.Province,
	}
}

func (p *passTLSClientCert) GetTracingInformation() (string, ext.SpanKindEnum) {
	return p.name, tracing.SpanKindNoneEnum
}

func (p *passTLSClientCert) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), p.name, typeName)
	p.modifyRequestHeaders(logger, req)
	p.next.ServeHTTP(rw, req)
}

// getSubjectInfos extract the requested information from the certificate subject.
func (p *passTLSClientCert) getSubjectInfos(cs *pkix.Name) string {
	var subject string

	if p.infos != nil && p.infos.subject != nil {
		options := p.infos.subject

		var content []string

		if options.country && len(cs.Country) > 0 {
			content = append(content, fmt.Sprintf("C=%s", cs.Country[0]))
		}

		if options.province && len(cs.Province) > 0 {
			content = append(content, fmt.Sprintf("ST=%s", cs.Province[0]))
		}

		if options.locality && len(cs.Locality) > 0 {
			content = append(content, fmt.Sprintf("L=%s", cs.Locality[0]))
		}

		if options.Organization && len(cs.Organization) > 0 {
			content = append(content, fmt.Sprintf("O=%s", cs.Organization[0]))
		}

		if options.commonName && len(cs.CommonName) > 0 {
			content = append(content, fmt.Sprintf("CN=%s", cs.CommonName))
		}

		if len(content) > 0 {
			subject = `Subject="` + strings.Join(content, ",") + `"`
		}
	}

	return subject
}

// getXForwardedTLSClientCertInfos Build a string with the wanted client certificates information
// like Subject="C=%s,ST=%s,L=%s,O=%s,CN=%s",NB=%d,NA=%d,SAN=%s;
func (p *passTLSClientCert) getXForwardedTLSClientCertInfos(certs []*x509.Certificate) string {
	var headerValues []string

	for _, peerCert := range certs {
		var values []string
		var sans string
		var nb string
		var na string

		subject := p.getSubjectInfos(&peerCert.Subject)
		if len(subject) > 0 {
			values = append(values, subject)
		}

		ci := p.infos
		if ci != nil {
			if ci.notBefore {
				nb = fmt.Sprintf("NB=%d", uint64(peerCert.NotBefore.Unix()))
				values = append(values, nb)
			}
			if ci.notAfter {
				na = fmt.Sprintf("NA=%d", uint64(peerCert.NotAfter.Unix()))
				values = append(values, na)
			}

			if ci.sans {
				sans = fmt.Sprintf("SAN=%s", strings.Join(getSANs(peerCert), ","))
				values = append(values, sans)
			}
		}

		value := strings.Join(values, ",")
		headerValues = append(headerValues, value)
	}

	return strings.Join(headerValues, ";")
}

// modifyRequestHeaders set the wanted headers with the certificates information.
func (p *passTLSClientCert) modifyRequestHeaders(logger logrus.FieldLogger, r *http.Request) {
	if p.pem {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			r.Header.Set(xForwardedTLSClientCert, getXForwardedTLSClientCert(logger, r.TLS.PeerCertificates))
		} else {
			logger.Warn("Try to extract certificate on a request without TLS")
		}
	}

	if p.infos != nil {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			headerContent := p.getXForwardedTLSClientCertInfos(r.TLS.PeerCertificates)
			r.Header.Set(xForwardedTLSClientCertInfos, url.QueryEscape(headerContent))
		} else {
			logger.Warn("Try to extract certificate on a request without TLS")
		}
	}
}

// sanitize As we pass the raw certificates, remove the useless data and make it http request compliant.
func sanitize(cert []byte) string {
	s := string(cert)
	r := strings.NewReplacer("-----BEGIN CERTIFICATE-----", "",
		"-----END CERTIFICATE-----", "",
		"\n", "")
	cleaned := r.Replace(s)

	return url.QueryEscape(cleaned)
}

// extractCertificate extract the certificate from the request.
func extractCertificate(logger logrus.FieldLogger, cert *x509.Certificate) string {
	b := pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
	certPEM := pem.EncodeToMemory(&b)
	if certPEM == nil {
		logger.Error("Cannot extract the certificate content")
		return ""
	}
	return sanitize(certPEM)
}

// getXForwardedTLSClientCert Build a string with the client certificates.
func getXForwardedTLSClientCert(logger logrus.FieldLogger, certs []*x509.Certificate) string {
	var headerValues []string

	for _, peerCert := range certs {
		headerValues = append(headerValues, extractCertificate(logger, peerCert))
	}

	return strings.Join(headerValues, ",")
}

// getSANs get the Subject Alternate Name values.
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
