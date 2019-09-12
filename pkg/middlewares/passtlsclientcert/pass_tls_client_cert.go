package passtlsclientcert

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	xForwardedTLSClientCert     = "X-Forwarded-Tls-Client-Cert"
	xForwardedTLSClientCertInfo = "X-Forwarded-Tls-Client-Cert-Info"
	typeName                    = "PassClientTLSCert"
)

var attributeTypeNames = map[string]string{
	"0.9.2342.19200300.100.1.25": "DC", // Domain component OID - RFC 2247
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

func newDistinguishedNameOptions(info *dynamic.TLSCLientCertificateDNInfo) *DistinguishedNameOptions {
	if info == nil {
		return nil
	}

	return &DistinguishedNameOptions{
		CommonName:          info.CommonName,
		CountryName:         info.Country,
		DomainComponent:     info.DomainComponent,
		LocalityName:        info.Locality,
		OrganizationName:    info.Organization,
		SerialNumber:        info.SerialNumber,
		StateOrProvinceName: info.Province,
	}
}

// passTLSClientCert is a middleware that helps setup a few tls info features.
type passTLSClientCert struct {
	next http.Handler
	name string
	pem  bool                      // pass the sanitized pem to the backend in a specific header
	info *tlsClientCertificateInfo // pass selected information from the client certificate
}

// New constructs a new PassTLSClientCert instance from supplied frontend header struct.
func New(ctx context.Context, next http.Handler, config dynamic.PassTLSClientCert, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	return &passTLSClientCert{
		next: next,
		name: name,
		pem:  config.PEM,
		info: newTLSClientInfo(config.Info),
	}, nil
}

// tlsClientCertificateInfo is a struct for specifying the configuration for the passTLSClientCert middleware.
type tlsClientCertificateInfo struct {
	notAfter  bool
	notBefore bool
	sans      bool
	subject   *DistinguishedNameOptions
	issuer    *DistinguishedNameOptions
}

func newTLSClientInfo(info *dynamic.TLSClientCertificateInfo) *tlsClientCertificateInfo {
	if info == nil {
		return nil
	}

	return &tlsClientCertificateInfo{
		issuer:    newDistinguishedNameOptions(info.Issuer),
		notAfter:  info.NotAfter,
		notBefore: info.NotBefore,
		subject:   newDistinguishedNameOptions(info.Subject),
		sans:      info.Sans,
	}
}

func (p *passTLSClientCert) GetTracingInformation() (string, ext.SpanKindEnum) {
	return p.name, tracing.SpanKindNoneEnum
}

func (p *passTLSClientCert) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := middlewares.GetLoggerCtx(req.Context(), p.name, typeName)

	p.modifyRequestHeaders(ctx, req)
	p.next.ServeHTTP(rw, req)
}

func getDNInfo(ctx context.Context, prefix string, options *DistinguishedNameOptions, cs *pkix.Name) string {
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
		writeParts(ctx, content, cs.Country, "C")
	}

	if options.StateOrProvinceName {
		writeParts(ctx, content, cs.Province, "ST")
	}

	if options.LocalityName {
		writeParts(ctx, content, cs.Locality, "L")
	}

	if options.OrganizationName {
		writeParts(ctx, content, cs.Organization, "O")
	}

	if options.SerialNumber {
		writePart(ctx, content, cs.SerialNumber, "SN")
	}

	if options.CommonName {
		writePart(ctx, content, cs.CommonName, "CN")
	}

	if content.Len() > 0 {
		return prefix + `="` + strings.TrimSuffix(content.String(), ",") + `"`
	}

	return ""
}

func writeParts(ctx context.Context, content io.StringWriter, entries []string, prefix string) {
	for _, entry := range entries {
		writePart(ctx, content, entry, prefix)
	}
}

func writePart(ctx context.Context, content io.StringWriter, entry string, prefix string) {
	if len(entry) > 0 {
		_, err := content.WriteString(fmt.Sprintf("%s=%s,", prefix, entry))
		if err != nil {
			log.FromContext(ctx).Error(err)
		}
	}
}

// getXForwardedTLSClientCertInfo Build a string with the wanted client certificates information
// like Subject="C=%s,ST=%s,L=%s,O=%s,CN=%s",NB=%d,NA=%d,SAN=%s;
func (p *passTLSClientCert) getXForwardedTLSClientCertInfo(ctx context.Context, certs []*x509.Certificate) string {
	var headerValues []string

	for _, peerCert := range certs {
		var values []string
		var sans string
		var nb string
		var na string

		if p.info != nil {
			subject := getDNInfo(ctx, "Subject", p.info.subject, &peerCert.Subject)
			if len(subject) > 0 {
				values = append(values, subject)
			}

			issuer := getDNInfo(ctx, "Issuer", p.info.issuer, &peerCert.Issuer)
			if len(issuer) > 0 {
				values = append(values, issuer)
			}
		}

		ci := p.info
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
func (p *passTLSClientCert) modifyRequestHeaders(ctx context.Context, r *http.Request) {
	logger := log.FromContext(ctx)

	if p.pem {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			r.Header.Set(xForwardedTLSClientCert, getXForwardedTLSClientCert(ctx, r.TLS.PeerCertificates))
		} else {
			logger.Warn("Tried to extract a certificate on a request without mutual TLS")
		}
	}

	if p.info != nil {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			headerContent := p.getXForwardedTLSClientCertInfo(ctx, r.TLS.PeerCertificates)
			r.Header.Set(xForwardedTLSClientCertInfo, url.QueryEscape(headerContent))
		} else {
			logger.Warn("Tried to extract a certificate on a request without mutual TLS")
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
func extractCertificate(ctx context.Context, cert *x509.Certificate) string {
	b := pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
	certPEM := pem.EncodeToMemory(&b)
	if certPEM == nil {
		log.FromContext(ctx).Error("Cannot extract the certificate content")
		return ""
	}
	return sanitize(certPEM)
}

// getXForwardedTLSClientCert Build a string with the client certificates.
func getXForwardedTLSClientCert(ctx context.Context, certs []*x509.Certificate) string {
	var headerValues []string

	for _, peerCert := range certs {
		headerValues = append(headerValues, extractCertificate(ctx, peerCert))
	}

	return strings.Join(headerValues, ",")
}

// getSANs get the Subject Alternate Name values.
func getSANs(cert *x509.Certificate) []string {
	var sans []string
	if cert == nil {
		return sans
	}

	sans = append(sans, cert.DNSNames...)
	sans = append(sans, cert.EmailAddresses...)

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
