package acme

import (
	"crypto/tls"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/types"
)

var _ challenge.Provider = (*challengeTLSALPN)(nil)

type challengeTLSALPN struct {
	Store ChallengeStore
}

func (c *challengeTLSALPN) Present(domain, token, keyAuth string) error {
	log.WithoutContext().WithField(log.ProviderName, "acme").
		Debugf("TLS Challenge Present temp certificate for %s", domain)

	certPEMBlock, keyPEMBlock, err := tlsalpn01.ChallengeBlocks(domain, keyAuth)
	if err != nil {
		return err
	}

	cert := &Certificate{Certificate: certPEMBlock, Key: keyPEMBlock, Domain: types.Domain{Main: "TEMP-" + domain}}
	return c.Store.AddTLSChallenge(domain, cert)
}

func (c *challengeTLSALPN) CleanUp(domain, token, keyAuth string) error {
	log.WithoutContext().WithField(log.ProviderName, "acme").
		Debugf("TLS Challenge CleanUp temp certificate for %s", domain)

	return c.Store.RemoveTLSChallenge(domain)
}

// GetTLSALPNCertificate Get the temp certificate for ACME TLS-ALPN-O1 challenge.
func (p *Provider) GetTLSALPNCertificate(domain string) (*tls.Certificate, error) {
	cert, err := p.ChallengeStore.GetTLSChallenge(domain)
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
