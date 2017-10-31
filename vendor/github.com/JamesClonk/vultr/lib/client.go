package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/juju/ratelimit"
)

const (
	// Version of this libary
	Version = "1.13.0"

	// APIVersion of Vultr
	APIVersion = "v1"

	// DefaultEndpoint to be used
	DefaultEndpoint = "https://api.vultr.com/"

	mediaType = "application/json"
)

// retryableStatusCodes are API response status codes that indicate that
// the failed request can be retried without further actions.
var retryableStatusCodes = map[int]struct{}{
	503: {}, // Rate limit hit
	500: {}, // Internal server error. Try again at a later time.
}

// Client represents the Vultr API client
type Client struct {
	// HTTP client for communication with the Vultr API
	client *http.Client

	// User agent for HTTP client
	UserAgent string

	// Endpoint URL for API requests
	Endpoint *url.URL

	// API key for accessing the Vultr API
	APIKey string

	// Max. number of request attempts
	MaxAttempts int

	// Throttling struct
	bucket *ratelimit.Bucket
}

// Options represents optional settings and flags that can be passed to NewClient
type Options struct {
	// HTTP client for communication with the Vultr API
	HTTPClient *http.Client

	// User agent for HTTP client
	UserAgent string

	// Endpoint URL for API requests
	Endpoint string

	// API rate limitation, calls per duration
	RateLimitation time.Duration

	// Max. number of times to retry API calls
	MaxRetries int
}

// NewClient creates new Vultr API client. Options are optional and can be nil.
func NewClient(apiKey string, options *Options) *Client {
	userAgent := "vultr-go/" + Version
	transport := &http.Transport{
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
	client := http.DefaultClient
	client.Transport = transport
	endpoint, _ := url.Parse(DefaultEndpoint)
	rate := 505 * time.Millisecond
	attempts := 1

	if options != nil {
		if options.HTTPClient != nil {
			client = options.HTTPClient
		}
		if options.UserAgent != "" {
			userAgent = options.UserAgent
		}
		if options.Endpoint != "" {
			endpoint, _ = url.Parse(options.Endpoint)
		}
		if options.RateLimitation != 0 {
			rate = options.RateLimitation
		}
		if options.MaxRetries != 0 {
			attempts = options.MaxRetries + 1
		}
	}

	return &Client{
		UserAgent:   userAgent,
		client:      client,
		Endpoint:    endpoint,
		APIKey:      apiKey,
		MaxAttempts: attempts,
		bucket:      ratelimit.NewBucket(rate, 1),
	}
}

func apiPath(path string) string {
	return fmt.Sprintf("/%s/%s", APIVersion, path)
}

func apiKeyPath(path, apiKey string) string {
	if strings.Contains(path, "?") {
		return path + "&api_key=" + apiKey
	}
	return path + "?api_key=" + apiKey
}

func (c *Client) get(path string, data interface{}) error {
	req, err := c.newRequest("GET", apiPath(path), nil)
	if err != nil {
		return err
	}
	return c.do(req, data)
}

func (c *Client) post(path string, values url.Values, data interface{}) error {
	req, err := c.newRequest("POST", apiPath(path), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	return c.do(req, data)
}

func (c *Client) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
	relPath, err := url.Parse(apiKeyPath(path, c.APIKey))
	if err != nil {
		return nil, err
	}

	url := c.Endpoint.ResolveReference(relPath)

	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.UserAgent)
	req.Header.Add("Accept", mediaType)

	if req.Method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, nil
}

func (c *Client) do(req *http.Request, data interface{}) error {
	// Throttle http requests to avoid hitting Vultr's API rate-limit
	c.bucket.Wait(1)

	// Request body gets drained on each read so we
	// need to save it's content for retrying requests
	var err error
	var requestBody []byte
	if req.Body != nil {
		requestBody, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("Error reading request body: %v", err)
		}
		req.Body.Close()
	}

	var apiError error
	for tryCount := 1; tryCount <= c.MaxAttempts; tryCount++ {
		// Restore request body to the original state
		if requestBody != nil {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusOK {
			if data != nil {
				// avoid unmarshalling problem because Vultr API returns
				// empty array instead of empty map when it shouldn't!
				if string(body) == `[]` {
					data = nil
				} else {
					if err := json.Unmarshal(body, data); err != nil {
						return err
					}
				}
			}
			return nil
		}

		apiError = errors.New(string(body))
		if !isCodeRetryable(resp.StatusCode) {
			break
		}

		delay := backoffDuration(tryCount)
		time.Sleep(delay)
	}

	return apiError
}

// backoffDuration returns the duration to wait before retrying the request.
// Duration is an exponential function of the retry count with a jitter of ~0-30%.
func backoffDuration(retryCount int) time.Duration {
	// Upper limit of delay at ~1 minute
	if retryCount > 7 {
		retryCount = 7
	}

	rand.Seed(time.Now().UnixNano())
	delay := (1 << uint(retryCount)) * (rand.Intn(150) + 500)
	return time.Duration(delay) * time.Millisecond
}

// isCodeRetryable returns true if the given status code means that we should retry.
func isCodeRetryable(statusCode int) bool {
	if _, ok := retryableStatusCodes[statusCode]; ok {
		return true
	}

	return false
}
