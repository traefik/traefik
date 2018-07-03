package vegadns2client

import (
	"net/http"
	"time"
)

// NewVegaDNSClient - helper to instantiate a client
// Input: url string
// Output: VegaDNSClient
func NewVegaDNSClient(url string) VegaDNSClient {
	httpClient := http.Client{Timeout: 15 * time.Second}
	return VegaDNSClient{
		client:  httpClient,
		baseurl: url,
		version: "1.0",
		token:   Token{},
	}
}
