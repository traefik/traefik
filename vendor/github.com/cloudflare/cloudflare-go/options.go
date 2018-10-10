package cloudflare

import (
	"net/http"

	"time"

	"golang.org/x/time/rate"
)

// Option is a functional option for configuring the API client.
type Option func(*API) error

// HTTPClient accepts a custom *http.Client for making API calls.
func HTTPClient(client *http.Client) Option {
	return func(api *API) error {
		api.httpClient = client
		return nil
	}
}

// Headers allows you to set custom HTTP headers when making API calls (e.g. for
// satisfying HTTP proxies, or for debugging).
func Headers(headers http.Header) Option {
	return func(api *API) error {
		api.headers = headers
		return nil
	}
}

// Organization allows you to apply account-level changes (Load Balancing, Railguns)
// to an organization instead.
func UsingOrganization(orgID string) Option {
	return func(api *API) error {
		api.organizationID = orgID
		return nil
	}
}

// UsingRateLimit applies a non-default rate limit to client API requests
// If not specified the default of 4rps will be applied
func UsingRateLimit(rps float64) Option {
	return func(api *API) error {
		// because ratelimiter doesnt do any windowing
		// setting burst makes it difficult to enforce a fixed rate
		// so setting it equal to 1 this effectively disables bursting
		// this doesn't check for sensible values, ultimately the api will enforce that the value is ok
		api.rateLimiter = rate.NewLimiter(rate.Limit(rps), 1)
		return nil
	}
}

// UsingRetryPolicy applies a non-default number of retries and min/max retry delays
// This will be used when the client exponentially backs off after errored requests
func UsingRetryPolicy(maxRetries int, minRetryDelaySecs int, maxRetryDelaySecs int) Option {
	// seconds is very granular for a minimum delay - but this is only in case of failure
	return func(api *API) error {
		api.retryPolicy = RetryPolicy{
			MaxRetries:    maxRetries,
			MinRetryDelay: time.Duration(minRetryDelaySecs) * time.Second,
			MaxRetryDelay: time.Duration(maxRetryDelaySecs) * time.Second,
		}
		return nil
	}
}

// UsingLogger can be set if you want to get log output from this API instance
// By default no log output is emitted
func UsingLogger(logger Logger) Option {
	return func(api *API) error {
		api.logger = logger
		return nil
	}
}

// parseOptions parses the supplied options functions and returns a configured
// *API instance.
func (api *API) parseOptions(opts ...Option) error {
	// Range over each options function and apply it to our API type to
	// configure it. Options functions are applied in order, with any
	// conflicting options overriding earlier calls.
	for _, option := range opts {
		err := option(api)
		if err != nil {
			return err
		}
	}

	return nil
}
