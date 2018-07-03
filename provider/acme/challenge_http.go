package acme

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/xenolf/lego/acme"
)

var _ acme.ChallengeProviderTimeout = (*challengeHTTP)(nil)

type challengeHTTP struct {
	Store Store
}

// Present presents a challenge to obtain new ACME certificate
func (c *challengeHTTP) Present(domain, token, keyAuth string) error {
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
func (c *challengeHTTP) CleanUp(domain, token, keyAuth string) error {
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
func (c *challengeHTTP) Timeout() (timeout, interval time.Duration) {
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

// AddRoutes add routes on internal router
func (p *Provider) AddRoutes(router *mux.Router) {
	router.Methods(http.MethodGet).
		Path(acme.HTTP01ChallengePath("{token}")).
		Handler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			vars := mux.Vars(req)
			if token, ok := vars["token"]; ok {
				domain, _, err := net.SplitHostPort(req.Host)
				if err != nil {
					log.Debugf("Unable to split host and port: %v. Fallback to request host.", err)
					domain = req.Host
				}

				tokenValue := getTokenValue(token, domain, p.Store)
				if len(tokenValue) > 0 {
					rw.WriteHeader(http.StatusOK)
					_, err = rw.Write(tokenValue)
					if err != nil {
						log.Errorf("Unable to write token : %v", err)
					}
					return
				}
			}
			rw.WriteHeader(http.StatusNotFound)
		}))
}
