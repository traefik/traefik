package acme

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/safe"
)

var _ challenge.ProviderTimeout = (*challengeHTTP)(nil)

type challengeHTTP struct {
	Store ChallengeStore
}

// Present presents a challenge to obtain new ACME certificate.
func (c *challengeHTTP) Present(domain, token, keyAuth string) error {
	return c.Store.SetHTTPChallengeToken(token, domain, []byte(keyAuth))
}

// CleanUp cleans the challenges when certificate is obtained.
func (c *challengeHTTP) CleanUp(domain, token, keyAuth string) error {
	return c.Store.RemoveHTTPChallengeToken(token, domain)
}

// Timeout calculates the maximum of time allowed to resolved an ACME challenge.
func (c *challengeHTTP) Timeout() (timeout, interval time.Duration) {
	return 60 * time.Second, 5 * time.Second
}

// CreateHandler creates a HTTP handler to expose the token for the HTTP challenge.
func (p *Provider) CreateHandler(notFoundHandler http.Handler) http.Handler {
	router := mux.NewRouter().SkipClean(true)
	router.NotFoundHandler = notFoundHandler

	router.Methods(http.MethodGet).
		Path(http01.ChallengePath("{token}")).
		Handler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			vars := mux.Vars(req)

			ctx := log.With(context.Background(), log.Str(log.ProviderName, p.ResolverName+".acme"))
			logger := log.FromContext(ctx)

			if token, ok := vars["token"]; ok {
				domain, _, err := net.SplitHostPort(req.Host)
				if err != nil {
					logger.Debugf("Unable to split host and port: %v. Fallback to request host.", err)
					domain = req.Host
				}

				tokenValue := getTokenValue(ctx, token, domain, p.ChallengeStore)
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
		}))

	return router
}

func getTokenValue(ctx context.Context, token, domain string, store ChallengeStore) []byte {
	logger := log.FromContext(ctx)
	logger.Debugf("Retrieving the ACME challenge for token %v...", token)

	var result []byte

	operation := func() error {
		var err error
		result, err = store.GetHTTPChallengeToken(token, domain)
		return err
	}

	notify := func(err error, time time.Duration) {
		logger.Errorf("Error getting challenge for token retrying in %s", time)
	}

	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = 60 * time.Second
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), ebo, notify)
	if err != nil {
		logger.Errorf("Cannot retrieve the ACME challenge for token %v: %v", token, err)
		return []byte{}
	}

	return result
}
