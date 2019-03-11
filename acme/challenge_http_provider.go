package acme

import (
	"fmt"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/go-acme/lego/challenge"
)

var _ challenge.ProviderTimeout = (*challengeHTTPProvider)(nil)

type challengeHTTPProvider struct {
	store cluster.Store
	lock  sync.RWMutex
}

func (c *challengeHTTPProvider) getTokenValue(token, domain string) []byte {
	log.Debugf("Looking for an existing ACME challenge for token %v...", token)
	c.lock.RLock()
	defer c.lock.RUnlock()

	account := c.store.Get().(*Account)
	if account.HTTPChallenge == nil {
		return []byte{}
	}

	var result []byte
	operation := func() error {
		var ok bool
		if result, ok = account.HTTPChallenge[token][domain]; !ok {
			return fmt.Errorf("cannot find challenge for token %v", token)
		}
		return nil
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("Error getting challenge for token retrying in %s", time)
	}

	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = 60 * time.Second
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
	if err != nil {
		log.Errorf("Error getting challenge for token: %v", err)
		return []byte{}
	}
	return result
}

func (c *challengeHTTPProvider) Present(domain, token, keyAuth string) error {
	log.Debugf("Challenge Present %s", domain)
	c.lock.Lock()
	defer c.lock.Unlock()

	transaction, object, err := c.store.Begin()
	if err != nil {
		return err
	}

	account := object.(*Account)
	if account.HTTPChallenge == nil {
		account.HTTPChallenge = map[string]map[string][]byte{}
	}

	if _, ok := account.HTTPChallenge[token]; !ok {
		account.HTTPChallenge[token] = map[string][]byte{}
	}

	account.HTTPChallenge[token][domain] = []byte(keyAuth)

	return transaction.Commit(account)
}

func (c *challengeHTTPProvider) CleanUp(domain, token, keyAuth string) error {
	log.Debugf("Challenge CleanUp %s", domain)
	c.lock.Lock()
	defer c.lock.Unlock()

	transaction, object, err := c.store.Begin()
	if err != nil {
		return err
	}

	account := object.(*Account)
	if _, ok := account.HTTPChallenge[token]; ok {
		if _, domainOk := account.HTTPChallenge[token][domain]; domainOk {
			delete(account.HTTPChallenge[token], domain)
		}
		if len(account.HTTPChallenge[token]) == 0 {
			delete(account.HTTPChallenge, token)
		}
	}

	return transaction.Commit(account)
}

func (c *challengeHTTPProvider) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}
