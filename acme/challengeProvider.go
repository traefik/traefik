package acme

import (
	"crypto/tls"
	"sync"

	"crypto/x509"
	"github.com/xenolf/lego/acme"
)

var _ acme.ChallengeProvider = (*inMemoryChallengeProvider)(nil)

type inMemoryChallengeProvider struct {
	challengeCerts map[string]*tls.Certificate
	lock           sync.RWMutex
}

func newWrapperChallengeProvider() *inMemoryChallengeProvider {
	return &inMemoryChallengeProvider{
		challengeCerts: map[string]*tls.Certificate{},
	}
}

func (c *inMemoryChallengeProvider) getCertificate(domain string) (cert *tls.Certificate, exists bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if cert, ok := c.challengeCerts[domain]; ok {
		return cert, true
	}
	return nil, false
}

func (c *inMemoryChallengeProvider) Present(domain, token, keyAuth string) error {
	cert, _, err := acme.TLSSNI01ChallengeCert(keyAuth)
	if err != nil {
		return err
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	for i := range cert.Leaf.DNSNames {
		c.challengeCerts[cert.Leaf.DNSNames[i]] = &cert
	}

	return nil
}

func (c *inMemoryChallengeProvider) CleanUp(domain, token, keyAuth string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.challengeCerts, domain)
	return nil
}
