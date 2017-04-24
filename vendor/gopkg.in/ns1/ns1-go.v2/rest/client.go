package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	clientVersion    = "2.0.0"
	defaultEndpoint  = "https://api.nsone.net/v1/"
	defaultUserAgent = "go-ns1/" + clientVersion

	headerAuth          = "X-NSONE-Key"
	headerRateLimit     = "X-Ratelimit-Limit"
	headerRateRemaining = "X-Ratelimit-Remaining"
	headerRatePeriod    = "X-Ratelimit-Period"
)

// Doer is a single method interface that allows a user to extend/augment an http.Client instance.
// Note: http.Client satisfies the Doer interface.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client manages communication with the NS1 Rest API.
type Client struct {
	// httpClient handles all rest api communication,
	// and expects an *http.Client.
	httpClient Doer

	// NS1 rest endpoint, overrides default if given.
	Endpoint *url.URL

	// NS1 api key (value for http request header 'X-NSONE-Key').
	APIKey string

	// NS1 go rest user agent (value for http request header 'User-Agent').
	UserAgent string

	// Func to call after response is returned in Do
	RateLimitFunc func(RateLimit)

	// From the excellent github-go client.
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for communicating with different components of the NS1 API.
	APIKeys       *APIKeysService
	DataFeeds     *DataFeedsService
	DataSources   *DataSourcesService
	Jobs          *JobsService
	Notifications *NotificationsService
	Records       *RecordsService
	Settings      *SettingsService
	Teams         *TeamsService
	Users         *UsersService
	Warnings      *WarningsService
	Zones         *ZonesService
}

// NewClient constructs and returns a reference to an instantiated Client.
func NewClient(httpClient Doer, options ...func(*Client)) *Client {
	endpoint, _ := url.Parse(defaultEndpoint)

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{
		httpClient:    httpClient,
		Endpoint:      endpoint,
		RateLimitFunc: defaultRateLimitFunc,
		UserAgent:     defaultUserAgent,
	}

	c.common.client = c
	c.APIKeys = (*APIKeysService)(&c.common)
	c.DataFeeds = (*DataFeedsService)(&c.common)
	c.DataSources = (*DataSourcesService)(&c.common)
	c.Jobs = (*JobsService)(&c.common)
	c.Notifications = (*NotificationsService)(&c.common)
	c.Records = (*RecordsService)(&c.common)
	c.Settings = (*SettingsService)(&c.common)
	c.Teams = (*TeamsService)(&c.common)
	c.Users = (*UsersService)(&c.common)
	c.Warnings = (*WarningsService)(&c.common)
	c.Zones = (*ZonesService)(&c.common)

	for _, option := range options {
		option(c)
	}
	return c
}

type service struct {
	client *Client
}

// SetHTTPClient sets a Client instances' httpClient.
func SetHTTPClient(httpClient Doer) func(*Client) {
	return func(c *Client) { c.httpClient = httpClient }
}

// SetAPIKey sets a Client instances' APIKey.
func SetAPIKey(key string) func(*Client) {
	return func(c *Client) { c.APIKey = key }
}

// SetEndpoint sets a Client instances' Endpoint.
func SetEndpoint(endpoint string) func(*Client) {
	return func(c *Client) { c.Endpoint, _ = url.Parse(endpoint) }
}

// SetUserAgent sets a Client instances' user agent.
func SetUserAgent(ua string) func(*Client) {
	return func(c *Client) { c.UserAgent = ua }
}

// SetRateLimitFunc sets a Client instances' RateLimitFunc.
func SetRateLimitFunc(ratefunc func(rl RateLimit)) func(*Client) {
	return func(c *Client) { c.RateLimitFunc = ratefunc }
}

// Do satisfies the Doer interface.
func (c Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	rl := parseRate(resp)
	c.RateLimitFunc(rl)

	if v != nil {
		// Try to unmarshal body into given type using streaming decoder.
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, err
		}
	}

	return resp, err
}

// NewRequest constructs and returns a http.Request.
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	uri := c.Endpoint.ResolveReference(rel)

	// Encode body as json
	buf := new(bytes.Buffer)
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, uri.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add(headerAuth, c.APIKey)
	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// Response wraps stdlib http response.
type Response struct {
	*http.Response
}

// Error contains all http responses outside the 2xx range.
type Error struct {
	Resp    *http.Response
	Message string
}

// Satisfy std lib error interface.
func (re *Error) Error() string {
	return fmt.Sprintf("%v %v: %d %v", re.Resp.Request.Method, re.Resp.Request.URL, re.Resp.StatusCode, re.Message)
}

// CheckResponse handles parsing of rest api errors. Returns nil if no error.
func CheckResponse(resp *http.Response) error {
	if c := resp.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	restErr := &Error{Resp: resp}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return restErr
	}

	err = json.Unmarshal(b, restErr)
	if err != nil {
		return err
	}

	return restErr
}

// RateLimitFunc is rate limiting strategy for the Client instance.
type RateLimitFunc func(RateLimit)

// RateLimit stores X-Ratelimit-* headers
type RateLimit struct {
	Limit     int
	Remaining int
	Period    int
}

var defaultRateLimitFunc = func(rl RateLimit) {}

// PercentageLeft returns the ratio of Remaining to Limit as a percentage
func (rl RateLimit) PercentageLeft() int {
	return rl.Remaining * 100 / rl.Limit
}

// WaitTime returns the time.Duration ratio of Period to Limit
func (rl RateLimit) WaitTime() time.Duration {
	return (time.Second * time.Duration(rl.Period)) / time.Duration(rl.Limit)
}

// WaitTimeRemaining returns the time.Duration ratio of Period to Remaining
func (rl RateLimit) WaitTimeRemaining() time.Duration {
	return (time.Second * time.Duration(rl.Period)) / time.Duration(rl.Remaining)
}

// RateLimitStrategySleep sets RateLimitFunc to sleep by WaitTimeRemaining
func (c *Client) RateLimitStrategySleep() {
	c.RateLimitFunc = func(rl RateLimit) {
		remaining := rl.WaitTimeRemaining()
		time.Sleep(remaining)
	}
}

// parseRate parses rate related headers from http response.
func parseRate(resp *http.Response) RateLimit {
	var rl RateLimit

	if limit := resp.Header.Get(headerRateLimit); limit != "" {
		rl.Limit, _ = strconv.Atoi(limit)
	}
	if remaining := resp.Header.Get(headerRateRemaining); remaining != "" {
		rl.Remaining, _ = strconv.Atoi(remaining)
	}
	if period := resp.Header.Get(headerRatePeriod); period != "" {
		rl.Period, _ = strconv.Atoi(period)
	}

	return rl
}

// SetTimeParam sets a url timestamp query param given the parameters name.
func SetTimeParam(key string, t time.Time) func(*url.Values) {
	return func(v *url.Values) { v.Set(key, strconv.Itoa(int(t.Unix()))) }
}

// SetBoolParam sets a url boolean query param given the parameters name.
func SetBoolParam(key string, b bool) func(*url.Values) {
	return func(v *url.Values) { v.Set(key, strconv.FormatBool(b)) }
}

// SetStringParam sets a url string query param given the parameters name.
func SetStringParam(key, val string) func(*url.Values) {
	return func(v *url.Values) { v.Set(key, val) }
}

// SetIntParam sets a url integer query param given the parameters name.
func SetIntParam(key string, val int) func(*url.Values) {
	return func(v *url.Values) { v.Set(key, strconv.Itoa(val)) }
}
