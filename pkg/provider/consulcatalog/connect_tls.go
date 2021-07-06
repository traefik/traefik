package consulcatalog

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
)

// connectCert holds our certificates as a client of the Consul Connect protocol.
type connectCert struct {
	root []string
	leaf keyPair
	// err is used to propagate to the caller (Provide) any error occurring within the certificate watcher goroutines.
	err error
}

func (c *connectCert) getRoot() []traefiktls.FileOrContent {
	var result []traefiktls.FileOrContent
	for _, r := range c.root {
		result = append(result, traefiktls.FileOrContent(r))
	}
	return result
}

func (c *connectCert) getLeaf() traefiktls.Certificate {
	return traefiktls.Certificate{
		CertFile: traefiktls.FileOrContent(c.leaf.cert),
		KeyFile:  traefiktls.FileOrContent(c.leaf.key),
	}
}

func (c *connectCert) isReady() bool {
	return c != nil && len(c.root) > 0 && c.leaf.cert != "" && c.leaf.key != ""
}

func (c *connectCert) equals(other *connectCert) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	if len(c.root) != len(other.root) {
		return false
	}
	for i, v := range c.root {
		if v != other.root[i] {
			return false
		}
	}
	return c.leaf == other.leaf
}

func (c *connectCert) serversTransport(item itemData) *dynamic.ServersTransport {
	return &dynamic.ServersTransport{
		// This ensures that the config changes whenever the verifier function changes
		ServerName: fmt.Sprintf("%s-%s-%s", item.Namespace, item.Datacenter, item.Name),
		// InsecureSkipVerify is needed because Go wants to verify a hostname otherwise
		InsecureSkipVerify: true,
		RootCAs:            c.getRoot(),
		Certificates: traefiktls.Certificates{
			c.getLeaf(),
		},
		CertVerifier: newVerifierData(item),
	}
}

// verifierData implements the CertVerifier interface.
type verifierData struct {
	spiffeIDServiceURI *url.URL
}

func newVerifierData(item itemData) *verifierData {
	spiffeIDService := connect.SpiffeIDService{
		Namespace:  item.Namespace,
		Datacenter: item.Datacenter,
		Service:    item.Name,
	}

	return &verifierData{spiffeIDServiceURI: spiffeIDService.URI()}
}

func (v verifierData) VerifyPeerCertificate(cfg *tls.Config, rawCerts [][]byte, _ [][]*x509.Certificate) error {
	// We should use RootCAs here:
	// https://github.com/hashicorp/consul/blob/cd428060f6547afddd9e0060c07b2a2c862da801/connect/tls.go#L279-L282
	// called via https://github.com/hashicorp/consul/blob/cd428060f6547afddd9e0060c07b2a2c862da801/connect/tls.go#L258
	cert, err := verifyChain(cfg.RootCAs, rawCerts)
	if err != nil {
		return err
	}

	return v.verifyServerCertMatchesURI(cert)
}

// verifyServerCertMatchesURI is used on tls connections dialed to a Connect server
// to ensure that the certificate it presented has the correct identity.
func (v verifierData) verifyServerCertMatchesURI(cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("peer certificate mismatch: no peer certificate presented")
	}

	// Our certs will only ever have a single URI for now so only check that
	if len(cert.URIs) < 1 {
		return errors.New("peer certificate mismatch: peer certificate invalid")
	}

	gotURI := cert.URIs[0]

	// Override the hostname since we rely on x509 constraints to limit ability to spoof the trust domain if needed
	// (i.e. because a root is shared with other PKI or Consul clusters).
	// This allows for seamless migrations between trust domains.

	expectURI := &url.URL{}
	*expectURI = *v.spiffeIDServiceURI
	expectURI.Host = gotURI.Host

	if strings.EqualFold(gotURI.String(), expectURI.String()) {
		return nil
	}

	return fmt.Errorf("peer certificate mismatch got %s, want %s", gotURI, v.spiffeIDServiceURI)
}

// verifyChain performs standard TLS verification without enforcing remote hostname matching.
func verifyChain(rootCAs *x509.CertPool, rawCerts [][]byte) (*x509.Certificate, error) {
	// Fetch leaf and intermediates. This is based on code form tls handshake.
	if len(rawCerts) < 1 {
		return nil, errors.New("tls: no certificates from peer")
	}

	certs := make([]*x509.Certificate, len(rawCerts))
	for i, asn1Data := range rawCerts {
		cert, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			return nil, fmt.Errorf("tls: failed to parse certificate from peer: %w", err)
		}

		certs[i] = cert
	}

	opts := x509.VerifyOptions{
		Roots:         rootCAs,
		Intermediates: x509.NewCertPool(),
	}

	// All but the first cert are intermediates
	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}

	_, err := certs[0].Verify(opts)
	if err != nil {
		return nil, err
	}

	return certs[0], nil
}
