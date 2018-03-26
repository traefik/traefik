package egoscale

import (
	"crypto/tls"
	"net/http"
)

// NewClient creates a CloudStack API client
func NewClient(endpoint string, apiKey string, apiSecret string) *Client {
	cs := &Client{
		client: &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		},
		endpoint:  endpoint,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
	return cs
}
