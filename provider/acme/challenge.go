package acme

import (
	"fmt"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/xenolf/lego/acme"
)

func dnsOverrideDelay(delay flaeg.Duration) error {
	if delay == 0 {
		return nil
	}

	if delay > 0 {
		log.Debugf("Delaying %d rather than validating DNS propagation now.", delay)

		acme.PreCheckDNS = func(_, _ string) (bool, error) {
			time.Sleep(time.Duration(delay))
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
		var err error
		result, err = store.GetHTTPChallengeToken(token, domain)
		return err
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
	return store.SetHTTPChallengeToken(token, domain, []byte(keyAuth))
}

func cleanUpHTTPChallenge(domain, token string, store Store) error {
	return store.RemoveHTTPChallengeToken(token, domain)
}
