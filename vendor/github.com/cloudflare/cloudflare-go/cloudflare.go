// Package cloudflare implements the Cloudflare v4 API.
package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

const apiURL = "https://api.cloudflare.com/client/v4"

const (
	// AuthKeyEmail specifies that we should authenticate with API key and email address
	AuthKeyEmail = 1 << iota
	// AuthUserService specifies that we should authenticate with a User-Service key
	AuthUserService
	// AuthToken specifies that we should authenticate with an API Token
	AuthToken
)

// API holds the configuration for the current API client. A client should not
// be modified concurrently.
type API struct {
	APIKey            string
	APIEmail          string
	APIUserServiceKey string
	APIToken          string
	BaseURL           string
	OrganizationID    string
	UserAgent         string
	headers           http.Header
	httpClient        *http.Client
	authType          int
	rateLimiter       *rate.Limiter
	retryPolicy       RetryPolicy
	logger            Logger
}

// newClient provides shared logic for New and NewWithUserServiceKey
func newClient(opts ...Option) (*API, error) {
	silentLogger := log.New(ioutil.Discard, "", log.LstdFlags)

	api := &API{
		BaseURL:     apiURL,
		headers:     make(http.Header),
		rateLimiter: rate.NewLimiter(rate.Limit(4), 1), // 4rps equates to default api limit (1200 req/5 min)
		retryPolicy: RetryPolicy{
			MaxRetries:    3,
			MinRetryDelay: time.Duration(1) * time.Second,
			MaxRetryDelay: time.Duration(30) * time.Second,
		},
		logger: silentLogger,
	}

	err := api.parseOptions(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "options parsing failed")
	}

	// Fall back to http.DefaultClient if the package user does not provide
	// their own.
	if api.httpClient == nil {
		api.httpClient = http.DefaultClient
	}

	return api, nil
}

// New creates a new Cloudflare v4 API client.
func New(key, email string, opts ...Option) (*API, error) {
	if key == "" || email == "" {
		return nil, errors.New(errEmptyCredentials)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.APIKey = key
	api.APIEmail = email
	api.authType = AuthKeyEmail

	return api, nil
}

// NewWithAPIToken creates a new Cloudflare v4 API client using API Tokens
func NewWithAPIToken(token string, opts ...Option) (*API, error) {
	if token == "" {
		return nil, errors.New(errEmptyAPIToken)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.APIToken = token
	api.authType = AuthToken

	return api, nil
}

// NewWithUserServiceKey creates a new Cloudflare v4 API client using service key authentication.
func NewWithUserServiceKey(key string, opts ...Option) (*API, error) {
	if key == "" {
		return nil, errors.New(errEmptyCredentials)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.APIUserServiceKey = key
	api.authType = AuthUserService

	return api, nil
}

// SetAuthType sets the authentication method (AuthKeyEmail, AuthToken, or AuthUserService).
func (api *API) SetAuthType(authType int) {
	api.authType = authType
}

// ZoneIDByName retrieves a zone's ID from the name.
func (api *API) ZoneIDByName(zoneName string) (string, error) {
	res, err := api.ListZonesContext(context.TODO(), WithZoneFilter(zoneName))
	if err != nil {
		return "", errors.Wrap(err, "ListZonesContext command failed")
	}

	if len(res.Result) > 1 && api.OrganizationID == "" {
		return "", errors.New("ambiguous zone name used without an account ID")
	}

	for _, zone := range res.Result {
		if api.OrganizationID != "" {
			if zone.Name == zoneName && api.OrganizationID == zone.Account.ID {
				return zone.ID, nil
			}
		} else {
			if zone.Name == zoneName {
				return zone.ID, nil
			}
		}
	}

	return "", errors.New("Zone could not be found")
}

// makeRequest makes a HTTP request and returns the body as a byte slice,
// closing it before returning. params will be serialized to JSON.
func (api *API) makeRequest(method, uri string, params interface{}) ([]byte, error) {
	return api.makeRequestWithAuthType(context.TODO(), method, uri, params, api.authType)
}

func (api *API) makeRequestContext(ctx context.Context, method, uri string, params interface{}) ([]byte, error) {
	return api.makeRequestWithAuthType(ctx, method, uri, params, api.authType)
}

func (api *API) makeRequestWithHeaders(method, uri string, params interface{}, headers http.Header) ([]byte, error) {
	return api.makeRequestWithAuthTypeAndHeaders(context.TODO(), method, uri, params, api.authType, headers)
}

func (api *API) makeRequestWithAuthType(ctx context.Context, method, uri string, params interface{}, authType int) ([]byte, error) {
	return api.makeRequestWithAuthTypeAndHeaders(ctx, method, uri, params, authType, nil)
}

func (api *API) makeRequestWithAuthTypeAndHeaders(ctx context.Context, method, uri string, params interface{}, authType int, headers http.Header) ([]byte, error) {
	// Replace nil with a JSON object if needed
	var jsonBody []byte
	var err error

	if params != nil {
		if paramBytes, ok := params.([]byte); ok {
			jsonBody = paramBytes
		} else {
			jsonBody, err = json.Marshal(params)
			if err != nil {
				return nil, errors.Wrap(err, "error marshalling params to JSON")
			}
		}
	} else {
		jsonBody = nil
	}

	var resp *http.Response
	var respErr error
	var reqBody io.Reader
	var respBody []byte
	for i := 0; i <= api.retryPolicy.MaxRetries; i++ {
		if jsonBody != nil {
			reqBody = bytes.NewReader(jsonBody)
		}
		if i > 0 {
			// expect the backoff introduced here on errored requests to dominate the effect of rate limiting
			// don't need a random component here as the rate limiter should do something similar
			// nb time duration could truncate an arbitrary float. Since our inputs are all ints, we should be ok
			sleepDuration := time.Duration(math.Pow(2, float64(i-1)) * float64(api.retryPolicy.MinRetryDelay))

			if sleepDuration > api.retryPolicy.MaxRetryDelay {
				sleepDuration = api.retryPolicy.MaxRetryDelay
			}
			// useful to do some simple logging here, maybe introduce levels later
			api.logger.Printf("Sleeping %s before retry attempt number %d for request %s %s", sleepDuration.String(), i, method, uri)
			time.Sleep(sleepDuration)

		}
		err = api.rateLimiter.Wait(context.TODO())
		if err != nil {
			return nil, errors.Wrap(err, "Error caused by request rate limiting")
		}
		resp, respErr = api.request(ctx, method, uri, reqBody, authType, headers)

		// retry if the server is rate limiting us or if it failed
		// assumes server operations are rolled back on failure
		if respErr != nil || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			// if we got a valid http response, try to read body so we can reuse the connection
			// see https://golang.org/pkg/net/http/#Client.Do
			if respErr == nil {
				respBody, err = ioutil.ReadAll(resp.Body)
				resp.Body.Close()

				respErr = errors.Wrap(err, "could not read response body")

				api.logger.Printf("Request: %s %s got an error response %d: %s\n", method, uri, resp.StatusCode,
					strings.Replace(strings.Replace(string(respBody), "\n", "", -1), "\t", "", -1))
			} else {
				api.logger.Printf("Error performing request: %s %s : %s \n", method, uri, respErr.Error())
			}
			continue
		} else {
			respBody, err = ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, errors.Wrap(err, "could not read response body")
			}
			break
		}
	}
	if respErr != nil {
		return nil, respErr
	}

	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, errors.Errorf("HTTP status %d: invalid credentials", resp.StatusCode)
	case resp.StatusCode == http.StatusForbidden:
		return nil, errors.Errorf("HTTP status %d: insufficient permissions", resp.StatusCode)
	case resp.StatusCode == http.StatusServiceUnavailable,
		resp.StatusCode == http.StatusBadGateway,
		resp.StatusCode == http.StatusGatewayTimeout,
		resp.StatusCode == 522,
		resp.StatusCode == 523,
		resp.StatusCode == 524:
		return nil, errors.Errorf("HTTP status %d: service failure", resp.StatusCode)
	// This isn't a great solution due to the way the `default` case is
	// a catch all and that the `filters/validate-expr` returns a HTTP 400
	// yet the clients need to use the HTTP body as a JSON string.
	case resp.StatusCode == 400 && strings.HasSuffix(resp.Request.URL.Path, "/filters/validate-expr"):
		return nil, errors.Errorf("%s", respBody)
	default:
		var s string
		if respBody != nil {
			s = string(respBody)
		}
		return nil, errors.Errorf("HTTP status %d: content %q", resp.StatusCode, s)
	}

	return respBody, nil
}

// request makes a HTTP request to the given API endpoint, returning the raw
// *http.Response, or an error if one occurred. The caller is responsible for
// closing the response body.
func (api *API) request(ctx context.Context, method, uri string, reqBody io.Reader, authType int, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, api.BaseURL+uri, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request creation failed")
	}
	req.WithContext(ctx)

	combinedHeaders := make(http.Header)
	copyHeader(combinedHeaders, api.headers)
	copyHeader(combinedHeaders, headers)
	req.Header = combinedHeaders

	if authType&AuthKeyEmail != 0 {
		req.Header.Set("X-Auth-Key", api.APIKey)
		req.Header.Set("X-Auth-Email", api.APIEmail)
	}
	if authType&AuthUserService != 0 {
		req.Header.Set("X-Auth-User-Service-Key", api.APIUserServiceKey)
	}
	if authType&AuthToken != 0 {
		req.Header.Set("Authorization", "Bearer "+api.APIToken)
	}

	if api.UserAgent != "" {
		req.Header.Set("User-Agent", api.UserAgent)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}

	return resp, nil
}

// Returns the base URL to use for API endpoints that exist for both accounts and organizations.
// If an Organization option was used when creating the API instance, returns the org URL.
//
// accountBase is the base URL for endpoints referring to the current user. It exists as a
// parameter because it is not consistent across APIs.
func (api *API) userBaseURL(accountBase string) string {
	if api.OrganizationID != "" {
		return "/accounts/" + api.OrganizationID
	}
	return accountBase
}

// copyHeader copies all headers for `source` and sets them on `target`.
// based on https://godoc.org/github.com/golang/gddo/httputil/header#Copy
func copyHeader(target, source http.Header) {
	for k, vs := range source {
		target[k] = vs
	}
}

// ResponseInfo contains a code and message returned by the API as errors or
// informational messages inside the response.
type ResponseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response is a template.  There will also be a result struct.  There will be a
// unique response type for each response, which will include this type.
type Response struct {
	Success  bool           `json:"success"`
	Errors   []ResponseInfo `json:"errors"`
	Messages []ResponseInfo `json:"messages"`
}

// ResultInfo contains metadata about the Response.
type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	Total      int `json:"total_count"`
}

// RawResponse keeps the result as JSON form
type RawResponse struct {
	Response
	Result json.RawMessage `json:"result"`
}

// Raw makes a HTTP request with user provided params and returns the
// result as untouched JSON.
func (api *API) Raw(method, endpoint string, data interface{}) (json.RawMessage, error) {
	res, err := api.makeRequest(method, endpoint, data)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	var r RawResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// PaginationOptions can be passed to a list request to configure paging
// These values will be defaulted if omitted, and PerPage has min/max limits set by resource
type PaginationOptions struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

// RetryPolicy specifies number of retries and min/max retry delays
// This config is used when the client exponentially backs off after errored requests
type RetryPolicy struct {
	MaxRetries    int
	MinRetryDelay time.Duration
	MaxRetryDelay time.Duration
}

// Logger defines the interface this library needs to use logging
// This is a subset of the methods implemented in the log package
type Logger interface {
	Printf(format string, v ...interface{})
}

// ReqOption is a functional option for configuring API requests
type ReqOption func(opt *reqOption)
type reqOption struct {
	params url.Values
}

// WithZoneFilter applies a filter based on zone name.
func WithZoneFilter(zone string) ReqOption {
	return func(opt *reqOption) {
		opt.params.Set("name", zone)
	}
}

// WithPagination configures the pagination for a response.
func WithPagination(opts PaginationOptions) ReqOption {
	return func(opt *reqOption) {
		opt.params.Set("page", strconv.Itoa(opts.Page))
		opt.params.Set("per_page", strconv.Itoa(opts.PerPage))
	}
}
