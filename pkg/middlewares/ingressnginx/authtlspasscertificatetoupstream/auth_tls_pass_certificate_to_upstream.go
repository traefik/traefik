package authtlspasscertificatetoupstream

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tls"
)

const typeName = "AuthTLSPassCertificateToUpstream"

// Nginx header names.
const (
	sslClientCert      = "Ssl-Client-Cert"
	sslClientVerify    = "Ssl-Client-Verify"
	sslClientSubjectDN = "Ssl-Client-Subject-Dn"
	sslClientIssuerDN  = "Ssl-Client-Issuer-Dn"
)

type authTLSPassCertificateToUpstream struct {
	next           http.Handler
	name           string
	clientAuthType string
	caCertPool     *x509.CertPool
}

func NewAuthTLSPassCertificateToUpstream(ctx context.Context, next http.Handler, config dynamic.AuthTLSPassCertificateToUpstream, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	// caCertPool only needed to do internal validation if VerifyClient is optional_no_ca.
	var caCertPool *x509.CertPool
	if config.ClientAuthType == tls.RequestClientCert && len(config.CAFiles) > 0 {
		caCertPool = x509.NewCertPool()
		for _, ca := range config.CAFiles {
			if !caCertPool.AppendCertsFromPEM([]byte(ca)) {
				return nil, errors.New("failed to parse CA certificate")
			}
		}
	}

	return &authTLSPassCertificateToUpstream{
		next:           next,
		name:           name,
		clientAuthType: config.ClientAuthType,
		caCertPool:     caCertPool,
	}, nil
}

func (p *authTLSPassCertificateToUpstream) GetTracingInformation() (string, string) {
	return p.name, typeName
}

func (p *authTLSPassCertificateToUpstream) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), p.name, typeName)
	ctx := logger.WithContext(req.Context())

	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		logger.Debug().Msg("Tried to extract a certificate on a request without mutual TLS")
		req.Header.Set(sslClientVerify, "NONE")
		p.next.ServeHTTP(rw, req)
		return
	}

	// Nginx only returns the leaf certificate.
	cert := req.TLS.PeerCertificates[0]

	clientVerify := "SUCCESS"
	// Go's RequestClientCert doesn't verify at TLS level, so we have to verify in the middleware to return the correct Ssl-Client-Verify header.
	// For other cases, validation happens during the handshake, so if it reaches this middleware, it means that the certificate is valid.
	if p.clientAuthType == tls.RequestClientCert {
		_, err := cert.Verify(x509.VerifyOptions{Roots: p.caCertPool})
		if err != nil {
			clientVerify = "FAILED:" + err.Error()
		}
	}

	req.Header.Set(sslClientVerify, clientVerify)
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
