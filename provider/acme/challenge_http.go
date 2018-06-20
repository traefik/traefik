package acme

import (
	"fmt"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/xenolf/lego/acme"
)

var _ acme.ChallengeProviderTimeout = (*ChallengeHTTP)(nil)

type ChallengeHTTP struct {
	Store Store
}

// Present presents a challenge to obtain new ACME certificate
func (c *ChallengeHTTP) Present(domain, token, keyAuth string) error {
	httpChallenges, err := c.Store.GetHTTPChallenges()
	if err != nil {
		return fmt.Errorf("unable to get HTTPChallenges : %s", err)
	}

	if httpChallenges == nil {
		httpChallenges = map[string]map[string][]byte{}
	}

	if _, ok := httpChallenges[token]; !ok {
		httpChallenges[token] = map[string][]byte{}
	}

	httpChallenges[token][domain] = []byte(keyAuth)

	return c.Store.SaveHTTPChallenges(httpChallenges)
}

// CleanUp cleans the challenges when certificate is obtained
func (c *ChallengeHTTP) CleanUp(domain, token, keyAuth string) error {
	httpChallenges, err := c.Store.GetHTTPChallenges()
	if err != nil {
		return fmt.Errorf("unable to get HTTPChallenges : %s", err)
	}

	log.Debugf("Challenge CleanUp for domain %s", domain)

	if _, ok := httpChallenges[token]; ok {
		if _, domainOk := httpChallenges[token][domain]; domainOk {
			delete(httpChallenges[token], domain)
		}
		if len(httpChallenges[token]) == 0 {
			delete(httpChallenges, token)
		}
		return c.Store.SaveHTTPChallenges(httpChallenges)
	}
	return nil
}

// Timeout calculates the maximum of time allowed to resolved an ACME challenge
func (c *ChallengeHTTP) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}

func getTokenValue(token, domain string, store Store) []byte {
	log.Debugf("Looking for an existing ACME challenge for token %v...", token)
	var result []byte

	operation := func() error {
		httpChallenges, err := store.GetHTTPChallenges()
		if err != nil {
			return fmt.Errorf("HTTPChallenges not available : %s", err)
		}

		var ok bool
		if result, ok = httpChallenges[token][domain]; !ok {
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
