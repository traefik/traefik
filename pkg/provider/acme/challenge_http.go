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

	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/logs"
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
	logger := log.Ctx(req.Context()).With().Str(logs.ProviderName, "acme").Logger()

	token, err := getPathParam(req.URL)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to get token")
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if token != "" {
		domain, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			logger.Debug().Err(err).Msg("Unable to split host and port. Fallback to request host.")
			domain = req.Host
		}

		tokenValue := c.getTokenValue(logger.WithContext(req.Context()), token, domain)
		if len(tokenValue) > 0 {
			rw.WriteHeader(http.StatusOK)
			_, err = rw.Write(tokenValue)
			if err != nil {
				logger.Error().Err(err).Msg("Unable to write token")
			}
			return
		}
	}

	rw.WriteHeader(http.StatusNotFound)
}

func (c *ChallengeHTTP) getTokenValue(ctx context.Context, token, domain string) []byte {
	logger := log.Ctx(ctx)
	logger.Debug().Msgf("Retrieving the ACME challenge for %s (token %q)...", domain, token)

	c.lock.RLock()
	defer c.lock.RUnlock()

	if _, ok := c.httpChallenges[token]; !ok {
		logger.Error().Msgf("Cannot retrieve the ACME challenge for %s (token %q)", domain, token)
		return nil
	}

	result, ok := c.httpChallenges[token][domain]
	if !ok {
		logger.Error().Msgf("Cannot retrieve the ACME challenge for %s (token %q)", domain, token)
		return nil
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
