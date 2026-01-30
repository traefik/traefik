package passtlsclientcertnginx

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const nginxTypeName = "PassTLSClientCertNginx"

// Nginx header names.
const (
	sslClientCert      = "ssl-client-cert"
	sslClientVerify    = "ssl-client-verify"
	sslClientSubjectDN = "ssl-client-subject-dn"
	sslClientIssuerDN  = "ssl-client-issuer-dn"
)

type passTLSClientCertNginx struct {
	next         http.Handler
	name         string
	verifyClient string
}

func NewPassTLSClientCertNginx(ctx context.Context, next http.Handler, config dynamic.PassTLSClientCertNginx, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, nginxTypeName).Debug().Msg("Creating middleware")

	return &passTLSClientCertNginx{
		next:         next,
		name:         name,
		verifyClient: config.VerifyClient,
	}, nil
}

func (p *passTLSClientCertNginx) GetTracingInformation() (string, string) {
	return p.name, nginxTypeName
}

func (p *passTLSClientCertNginx) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), p.name, nginxTypeName)
	ctx := logger.WithContext(req.Context())

	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		logger.Debug().Msg("Tried to extract a certificate on a request without mutual TLS")
		req.Header.Set(sslClientVerify, "NONE")
		p.next.ServeHTTP(rw, req)
		return
	}

	// Nginx only returns the leaf certificate.
	cert := req.TLS.PeerCertificates[0]

	// Go limitation, where TLS validation for RequestClientCert will not return VerifiedChains, so that we are not able to know whether the certificate is valid.
	// For other cases, validation happens during the handshake, so if it reaches this middleware, it means that the certificate is valid.
	if p.verifyClient != "optional_no_ca" {
		req.Header.Set(sslClientVerify, "SUCCESS")
	}

	req.Header.Set(sslClientSubjectDN, cert.Subject.String())
	req.Header.Set(sslClientIssuerDN, cert.Issuer.String())
	req.Header.Set(sslClientCert, extractCertificatePEM(ctx, cert))

	p.next.ServeHTTP(rw, req)
}

func extractCertificatePEM(ctx context.Context, cert *x509.Certificate) string {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if certPEM == nil {
		log.Ctx(ctx).Error().Msg("Cannot extract the certificate content")
		return ""
	}
	// To match Nginx format, where spaces are converted into %20.
	return strings.ReplaceAll(url.QueryEscape(string(certPEM)), "+", "%20")
}
