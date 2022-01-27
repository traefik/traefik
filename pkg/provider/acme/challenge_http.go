package acme

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/safe"
)

// ChallengeHTTP HTTP challenge provider implements challenge.Provider.
type ChallengeHTTP struct {
	httpChallenges map[string]map[string][]byte
	lock           sync.RWMutex
}

// NewChallengeHTTP creates a new ChallengeHTTP.
func NewChallengeHTTP() *ChallengeHTTP {
	return &ChallengeHTTP{
		httpChallenges: make(map[string]map[string][]byte),
	}
}

// Present presents a challenge to obtain new ACME certificate.
func (c *ChallengeHTTP) Present(domain, token, keyAuth string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.httpChallenges[token]; !ok {
		c.httpChallenges[token] = map[string][]byte{}
	}

	c.httpChallenges[token][domain] = []byte(keyAuth)

	return nil
}

// CleanUp cleans the challenges when certificate is obtained.
func (c *ChallengeHTTP) CleanUp(domain, token, _ string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.httpChallenges == nil && len(c.httpChallenges) == 0 {
		return nil
	}

	if _, ok := c.httpChallenges[token]; ok {
		delete(c.httpChallenges[token], domain)

		if len(c.httpChallenges[token]) == 0 {
			delete(c.httpChallenges, token)
		}
	}

	return nil
}

// Timeout calculates the maximum of time allowed to resolved an ACME challenge.
func (c *ChallengeHTTP) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}

func (c *ChallengeHTTP) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := log.With(req.Context(), log.Str(log.ProviderName, "acme"))
	logger := log.FromContext(ctx)

	token, err := getPathParam(req.URL)
	if err != nil {
		logger.Errorf("Unable to get token: %v.", err)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if token != "" {
		domain, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			logger.Debugf("Unable to split host and port: %v. Fallback to request host.", err)
			domain = req.Host
		}

		tokenValue := c.getTokenValue(ctx, token, domain)
		if len(tokenValue) > 0 {
			rw.WriteHeader(http.StatusOK)
			_, err = rw.Write(tokenValue)
			if err != nil {
				logger.Errorf("Unable to write token: %v", err)
			}
			return
		}
	}

	rw.WriteHeader(http.StatusNotFound)
}

func (c *ChallengeHTTP) getTokenValue(ctx context.Context, token, domain string) []byte {
	logger := log.FromContext(ctx)
	logger.Debugf("Retrieving the ACME challenge for %s (token %q)...", domain, token)

	var result []byte

	operation := func() error {
		c.lock.RLock()
		defer c.lock.RUnlock()

		if _, ok := c.httpChallenges[token]; !ok {
			return fmt.Errorf("cannot find challenge for token %q (%s)", token, domain)
		}

		var ok bool
		result, ok = c.httpChallenges[token][domain]
		if !ok {
			return fmt.Errorf("cannot find challenge for %s (token %q)", domain, token)
		}

		return nil
	}

	notify := func(err error, time time.Duration) {
		logger.Errorf("Error getting challenge for token retrying in %s", time)
	}

	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = 60 * time.Second
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
	if err != nil {
		logger.Errorf("Cannot retrieve the ACME challenge for %s (token %q): %v", domain, token, err)
		return []byte{}
	}

	return result
}

func getPathParam(uri *url.URL) (string, error) {
	exp := regexp.MustCompile(fmt.Sprintf(`^%s([^/]+)/?$`, http01.ChallengePath("")))
	parts := exp.FindStringSubmatch(uri.Path)

	if len(parts) != 2 {
		return "", errors.New("missing token")
	}

	return parts[1], nil
}
