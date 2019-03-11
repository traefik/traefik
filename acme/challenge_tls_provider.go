package acme

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/go-acme/lego/challenge"
	"github.com/go-acme/lego/challenge/tlsalpn01"
)

var _ challenge.ProviderTimeout = (*challengeTLSProvider)(nil)

type challengeTLSProvider struct {
	store cluster.Store
	lock  sync.RWMutex
}

func (c *challengeTLSProvider) getCertificate(domain string) (cert *tls.Certificate, exists bool) {
	log.Debugf("Looking for an existing ACME challenge for %s...", domain)

	if !strings.HasSuffix(domain, ".acme.invalid") {
		return nil, false
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	account := c.store.Get().(*Account)
	if account.ChallengeCerts == nil {
		return nil, false
	}

	account.Init()

	var result *tls.Certificate
	operation := func() error {
		for _, cert := range account.ChallengeCerts {
			for _, dns := range cert.certificate.Leaf.DNSNames {
				if domain == dns {
					result = cert.certificate
					return nil
				}
			}
		}
		return fmt.Errorf("cannot find challenge cert for domain %s", domain)
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("Error getting cert: %v, retrying in %s", err, time)
	}
	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = 60 * time.Second

	err := backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
	if err != nil {
		log.Errorf("Error getting cert: %v", err)
		return nil, false

	}
	return result, true
}

func (c *challengeTLSProvider) Present(domain, token, keyAuth string) error {
	log.Debugf("Challenge Present %s", domain)

	cert, err := tlsALPN01ChallengeCert(domain, keyAuth)
	if err != nil {
		return err
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	transaction, object, err := c.store.Begin()
	if err != nil {
		return err
	}

	account := object.(*Account)
	if account.ChallengeCerts == nil {
		account.ChallengeCerts = map[string]*ChallengeCert{}
	}
	account.ChallengeCerts[domain] = cert

	return transaction.Commit(account)
}

func (c *challengeTLSProvider) CleanUp(domain, token, keyAuth string) error {
	log.Debugf("Challenge CleanUp %s", domain)

	c.lock.Lock()
	defer c.lock.Unlock()

	transaction, object, err := c.store.Begin()
	if err != nil {
		return err
	}

	account := object.(*Account)
	delete(account.ChallengeCerts, domain)

	return transaction.Commit(account)
}

func (c *challengeTLSProvider) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}

func tlsALPN01ChallengeCert(domain, keyAuth string) (*ChallengeCert, error) {
	tempCertPEM, rsaPrivPEM, err := tlsalpn01.ChallengeBlocks(domain, keyAuth)
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(tempCertPEM, rsaPrivPEM)
	if err != nil {
		return nil, err
	}

	return &ChallengeCert{Certificate: tempCertPEM, PrivateKey: rsaPrivPEM, certificate: &certificate}, nil
}
