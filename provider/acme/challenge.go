package acme

import (
	"fmt"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	acme "github.com/xenolf/lego/acmev2"
)

func dnsOverrideDelay(delay flaeg.Duration) error {
	if delay == 0 {
		return nil
	}

	if delay > 0 {
		log.Debugf("Delaying %d rather than validating DNS propagation now.", delay)

		acme.PreCheckDNS = func(_, _ string) (bool, error) {
			time.Sleep(time.Duration(delay) * time.Second)
			return true, nil
		}
	} else {
		return fmt.Errorf("delayBeforeCheck: %d cannot be less than 0", delay)
	}
	return nil
}

func getTokenValue(token, domain string, store Store) []byte {
	log.Debugf("Looking for an existing ACME challenge for token %v...", token)
	var result []byte

	operation := func() error {
		var ok bool
		httpChallenges, err := store.GetHTTPChallenges()
		if err != nil {
			return fmt.Errorf("HTTPChallenges not available : %s", err)
		}
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

func presentHTTPChallenge(domain, token, keyAuth string, store Store) error {
	httpChallenges, err := store.GetHTTPChallenges()
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

	return store.SaveHTTPChallenges(httpChallenges)
}

func cleanUpHTTPChallenge(domain, token string, store Store) error {
	httpChallenges, err := store.GetHTTPChallenges()
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
		return store.SaveHTTPChallenges(httpChallenges)
	}
	return nil
}
