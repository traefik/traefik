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
)

const typeName = "AuthTLSPassCertificateToUpstream"

// Nginx header names.
const (
	sslClientCert      = "Ssl-Client-Cert"
	sslClientVerify    = "Ssl-Client-Verify"
	sslClientSubjectDN = "Ssl-Client-Subject-Dn"
	sslClientIssuerDN  = "Ssl-Client-Issuer-Dn"
)

// managedHeaders holds, in their normalized form, the header names this middleware owns.
var managedHeaders = func() map[string]struct{} {
	names := []string{sslClientCert, sslClientVerify, sslClientSubjectDN, sslClientIssuerDN}

	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[normalizeHeaderName(name)] = struct{}{}
	}

	return set
}()

// deleteManagedHeaders removes from headers every name this middleware owns, whichever spelling it
// was received with. A client has no legitimate reason to send any of them, and they are removed
// before the middleware writes its own values, on the mutual TLS path as well as without it.
//
// The deletion goes through the map directly, and not through Header.Del, which canonicalizes the
// name it is given and would therefore miss the very aliasing spellings this guards against.
func deleteManagedHeaders(headers http.Header) {
	for key := range headers {
		if _, ok := managedHeaders[normalizeHeaderName(key)]; ok {
			delete(headers, key)
		}
	}
}

// normalizeHeaderName upper-cases the letters of name, and replaces every byte that is neither a
// letter nor a digit with a dash.
//
// Go canonicalizes a header name on dashes only, whereas the backends deriving variable names from
// the header names, which are the consumers of the Ssl-Client-* fields, replace every character that
// is neither a letter nor a digit with an underscore. They read Ssl-Client-Verify, Ssl_Client_Verify
// and Ssl.Client.Verify as the same variable, and all of those spellings reach the handlers: the
// fourteen characters building such an alias (!, #, $, %, &, ', *, +, ., ^, _, `, |, ~) are all
// valid in a header name. Comparing this form is what recognizes them as one name.
func normalizeHeaderName(name string) string {
	buf := []byte(name)
	for i, c := range buf {
		switch {
		case c >= 'a' && c <= 'z':
			buf[i] = c - ('a' - 'A')
		case c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		default:
			buf[i] = '-'
		}
	}

	return string(buf)
}

type authTLSPassCertificateToUpstream struct {
	next            http.Handler
	name            string
	caCertPool      *x509.CertPool
	passCertificate bool
}

func NewAuthTLSPassCertificateToUpstream(ctx context.Context, next http.Handler, config dynamic.AuthTLSPassCertificateToUpstream, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	var caCertPool *x509.CertPool
	if len(config.CAFiles) > 0 {
		caCertPool = x509.NewCertPool()
		for _, ca := range config.CAFiles {
			if !caCertPool.AppendCertsFromPEM([]byte(ca)) {
				return nil, errors.New("failed to parse CA certificate")
			}
		}
	}

	return &authTLSPassCertificateToUpstream{
		next:            next,
		name:            name,
		caCertPool:      caCertPool,
		passCertificate: config.PassCertificate,
	}, nil
}

func (p *authTLSPassCertificateToUpstream) GetTracingInformation() (string, string) {
	return p.name, typeName
}

func (p *authTLSPassCertificateToUpstream) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), p.name, typeName)
	ctx := logger.WithContext(req.Context())

	// This middleware owns the Ssl-Client-* headers, so no value a client sent for them may reach the
	// upstream, whether the request carries a client certificate or not.
	deleteManagedHeaders(req.Header)

	// Nginx builds these headers from the $ssl_client_* variables, which are empty outside of a TLS
	// connection, and proxy_set_header omits a header whose value is empty.
	// A plaintext request therefore carries none of them, rather than carrying NONE, which would tell
	// the backend that the client presented no certificate over a connection that was never TLS.
	if req.TLS == nil {
		logger.Debug().Msg("Tried to extract a certificate on a request without TLS")
		p.next.ServeHTTP(rw, req)
		return
	}

	if len(req.TLS.PeerCertificates) == 0 {
		logger.Debug().Msg("Tried to extract a certificate on a request without mutual TLS")
		req.Header.Set(sslClientVerify, "NONE")
		p.next.ServeHTTP(rw, req)
		return
	}

	// Nginx only returns the leaf certificate.
	cert := req.TLS.PeerCertificates[0]

	req.Header.Set(sslClientVerify, p.verify(req.TLS.PeerCertificates))
	req.Header.Set(sslClientSubjectDN, cert.Subject.String())
	req.Header.Set(sslClientIssuerDN, cert.Issuer.String())
	if p.passCertificate {
		req.Header.Set(sslClientCert, extractCertificatePEM(ctx, cert))
	}

	p.next.ServeHTTP(rw, req)
}

// verify returns the Ssl-Client-Verify value for the certificates the client presented.
// The handshake may not have run with the TLS option this middleware was built for (conflicting TLS options fall back to the default one),
// so the result never relies on it, and comes from the CA certificates the middleware was given.
func (p *authTLSPassCertificateToUpstream) verify(certs []*x509.Certificate) string {
	// A nil Roots would verify against the system certificate pool.
	if p.caCertPool == nil {
		return "FAILED:no CA certificate"
	}

	// The options are the ones crypto/tls uses to verify a client certificate during the handshake.
	opts := x509.VerifyOptions{
		Roots:         p.caCertPool,
		Intermediates: x509.NewCertPool(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}

	if _, err := certs[0].Verify(opts); err != nil {
		return "FAILED:" + err.Error()
	}

	return "SUCCESS"
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
