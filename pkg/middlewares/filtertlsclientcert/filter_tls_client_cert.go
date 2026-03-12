package filtertlsclientcert

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "FilterTLSClientCert"

// filterTLSClientCert is a middleware that rejects requests whose client
// certificate Subject or Issuer DN does not match the configured regexes.
type filterTLSClientCert struct {
	next          http.Handler
	name          string
	subjectRegexp *regexp.Regexp
	issuerRegexp  *regexp.Regexp
}

// New creates a new FilterTLSClientCert middleware instance.
func New(ctx context.Context, next http.Handler, config dynamic.FilterTLSClientCert, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	m := &filterTLSClientCert{
		next: next,
		name: name,
	}

	if config.Subject != "" {
		re, err := regexp.Compile(config.Subject)
		if err != nil {
			return nil, fmt.Errorf("invalid subject regex: %w", err)
		}

		m.subjectRegexp = re
	}

	if config.Issuer != "" {
		re, err := regexp.Compile(config.Issuer)
		if err != nil {
			return nil, fmt.Errorf("invalid issuer regex: %w", err)
		}

		m.issuerRegexp = re
	}

	return m, nil
}

func (f *filterTLSClientCert) GetTracingInformation() (string, string) {
	return f.name, typeName
}

func (f *filterTLSClientCert) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), f.name, typeName)

	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		logger.Debug().Msg("Rejecting request: no client certificate presented")
		http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	for _, cert := range req.TLS.PeerCertificates {
		if f.isAllowed(cert) {
			f.next.ServeHTTP(rw, req)
			return
		}
	}

	logger.Debug().Msg("Rejecting request: no client certificate matched the filter")
	http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

// isAllowed returns true if the certificate satisfies all configured regex filters.
func (f *filterTLSClientCert) isAllowed(cert *x509.Certificate) bool {
	if f.subjectRegexp != nil {
		if !f.subjectRegexp.MatchString(buildDN(&cert.Subject)) {
			return false
		}
	}

	if f.issuerRegexp != nil {
		if !f.issuerRegexp.MatchString(buildDN(&cert.Issuer)) {
			return false
		}
	}

	return true
}

// buildDN builds a string representation of the distinguished name using
// short attribute names (e.g. "CN=foo,O=bar,C=CH"), matching the format
// produced by OpenSSL / nginx ssl_client_s_dn.
func buildDN(name *pkix.Name) string {
	var parts []string

	for _, v := range name.Country {
		parts = append(parts, "C="+v)
	}

	for _, v := range name.Province {
		parts = append(parts, "ST="+v)
	}

	for _, v := range name.Locality {
		parts = append(parts, "L="+v)
	}

	for _, v := range name.Organization {
		parts = append(parts, "O="+v)
	}

	for _, v := range name.OrganizationalUnit {
		parts = append(parts, "OU="+v)
	}

	if name.CommonName != "" {
		parts = append(parts, "CN="+name.CommonName)
	}

	return strings.Join(parts, ",")
}
