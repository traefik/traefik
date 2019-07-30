// Package namesilo A Go client library for accessing the Namesilo API.
package namesilo

import (
	"fmt"
	"net/http"
	"net/url"

	querystring "github.com/google/go-querystring/query"
)

const (
	// DefaultAPIEndpoint The default API endpoint.
	DefaultAPIEndpoint = "https://www.namesilo.com/api"

	// SandboxAPIEndpoint The sandbox API endpoint.
	SandboxAPIEndpoint = "https://sandbox.namesilo.com/api"
)

// Response Codes
const (
	SuccessfulAPIOperation = "300"
	SuccessfulRegistration = "301"
	SuccessfulOrder        = "302"
)

// Client the Namesilo client.
type Client struct {
	Endpoint   string
	HTTPClient *http.Client
}

// NewClient Creates a Namesilo client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		Endpoint:   DefaultAPIEndpoint,
		HTTPClient: httpClient,
	}
}

func (c *Client) get(name string, params interface{}) (*http.Response, error) {
	uri, err := url.Parse(fmt.Sprintf("%s/%s", c.Endpoint, name))
	if err != nil {
		return nil, err
	}

	if params != nil {
		v, err := querystring.Values(params)
		if err != nil {
			return nil, err
		}
		uri.RawQuery = v.Encode()
	}

	return c.HTTPClient.Get(uri.String())
}
