package acme

import (
	"crypto/tls"
	"strings"
	"sync"

	"fmt"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/xenolf/lego/acme"
	"time"
)

var _ acme.ChallengeProviderTimeout = (*challengeProvider)(nil)

type challengeProvider struct {
	store cluster.Store
	lock  sync.RWMutex
}

func (c *challengeProvider) getCertificate(domain string) (cert *tls.Certificate, exists bool) {
	log.Debugf("Challenge GetCertificate %s", domain)
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
		return fmt.Errorf("Cannot find challenge cert for domain %s", domain)
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

func (c *challengeProvider) Present(domain, token, keyAuth string) error {
	log.Debugf("Challenge Present %s", domain)
	cert, _, err := TLSSNI01ChallengeCert(keyAuth)
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
	account.ChallengeCerts[domain] = &cert
	return transaction.Commit(account)
}

func (c *challengeProvider) CleanUp(domain, token, keyAuth string) error {
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

func (c *challengeProvider) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}
