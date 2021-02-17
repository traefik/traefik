package consulcatalog

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/consul/agent/connect"
)

// extractCertURI returns the first URI SAN from the leaf certificate presented
// in the slice. The slice is expected to be the passed from
// tls.Conn.ConnectionState().PeerCertificates and requires that the leaf has at
// least one URI and the first URI is the correct one to use.
func extractCertURI(certs []*x509.Certificate) (*url.URL, error) {
	if len(certs) < 1 {
		return nil, errors.New("no peer certificate presented")
	}

	// Only check the first cert assuming this is the only leaf. It's not clear if
	// services might ever legitimately present multiple leaf certificates or if
	// the slice is just to allow presenting the whole chain of intermediates.
	cert := certs[0]

	// Our certs will only ever have a single URI for now so only check that
	if len(cert.URIs) < 1 {
		return nil, errors.New("peer certificate invalid")
	}

	return cert.URIs[0], nil
}

// verifyServerCertMatchesURI is used on tls connections dialed to a connect
// server to ensure that the certificate it presented has the correct identity.
func verifyServerCertMatchesURI(certs []*x509.Certificate, expected connect.CertURI) error {
	expectedStr := expected.URI().String()

	gotURI, err := extractCertURI(certs)
	if err != nil {
		return errors.New("peer certificate mismatch")
	}

	// Override the hostname since we rely on x509 constraints to limit ability to
	// spoof the trust domain if needed (i.e. because a root is shared with other
	// PKI or Consul clusters). This allows for seamless migrations between trust
	// domains.
	expectURI := expected.URI()
	expectURI.Host = gotURI.Host
	if strings.EqualFold(gotURI.String(), expectURI.String()) {
		return nil
	}

	return fmt.Errorf("peer certificate mismatch got %s, want %s",
		gotURI.String(), expectedStr)
}

// verifyChain performs standard TLS verification without enforcing remote
// hostname matching.
func verifyChain(tlsCfg *tls.Config, rawCerts [][]byte, client bool) (*x509.Certificate, error) {
	// Fetch leaf and intermediates. This is based on code form tls handshake.
	if len(rawCerts) < 1 {
		return nil, errors.New("tls: no certificates from peer")
	}
	certs := make([]*x509.Certificate, len(rawCerts))
	for i, asn1Data := range rawCerts {
		cert, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			return nil, errors.New("tls: failed to parse certificate from peer: " + err.Error())
		}
		certs[i] = cert
	}

	cas := tlsCfg.RootCAs
	if client {
		cas = tlsCfg.ClientCAs
	}

	opts := x509.VerifyOptions{
		Roots:         cas,
		Intermediates: x509.NewCertPool(),
	}
	if !client {
		// Server side only sets KeyUsages in tls. This defaults to ServerAuth in
		// x509 lib. See
		// https://github.com/golang/go/blob/ee7dd810f9ca4e63ecfc1d3044869591783b8b74/src/crypto/x509/verify.go#L866-L868
		opts.KeyUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}

	// All but the first cert are intermediates
	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}
	_, err := certs[0].Verify(opts)
	return certs[0], err
}
