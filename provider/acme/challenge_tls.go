package acme

import (
	"crypto/tls"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/go-acme/lego/challenge"
	"github.com/go-acme/lego/challenge/tlsalpn01"
)

var _ challenge.Provider = (*challengeTLSALPN)(nil)

type challengeTLSALPN struct {
	Store Store
}

func (c *challengeTLSALPN) Present(domain, token, keyAuth string) error {
	log.Debugf("TLS Challenge Present temp certificate for %s", domain)

	certPEMBlock, keyPEMBlock, err := tlsalpn01.ChallengeBlocks(domain, keyAuth)
	if err != nil {
		return err
	}

	cert := &Certificate{Certificate: certPEMBlock, Key: keyPEMBlock, Domain: types.Domain{Main: "TEMP-" + domain}}
	return c.Store.AddTLSChallenge(domain, cert)
}

func (c *challengeTLSALPN) CleanUp(domain, token, keyAuth string) error {
	log.Debugf("TLS Challenge CleanUp temp certificate for %s", domain)

	return c.Store.RemoveTLSChallenge(domain)
}

// GetTLSALPNCertificate Get the temp certificate for ACME TLS-ALPN-O1 challenge.
func (p *Provider) GetTLSALPNCertificate(domain string) (*tls.Certificate, error) {
	cert, err := p.Store.GetTLSChallenge(domain)
	if err != nil {
		return nil, err
	}

	if cert == nil {
		return nil, nil
	}

	certificate, err := tls.X509KeyPair(cert.Certificate, cert.Key)
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}
