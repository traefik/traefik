package servicefabric

import "net/http"

// HTTPClient is an interface for HTTP clients
// to implement. The client only requires
// read-only access to the Service Fabric API,
// thus only the HTTP GET method needs to be implemented
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Transport(transport *http.Transport)
}

// HTTPClientImpl is an implementation of HTTPClient
// that wraps the net/http HTTP client
type httpClientImpl struct {
	client http.Client
}

// NewHTTPClient creates a new HTTPClient instance
func NewHTTPClient(client http.Client) HTTPClient {
	return &httpClientImpl{client: client}
}

// Get is a method that implements a HTTP GET request
func (c *httpClientImpl) Get(url string) (resp *http.Response, err error) {
	return c.client.Get(url)
}

// Transport sets the HTTP client transport property
func (c *httpClientImpl) Transport(transport *http.Transport) {
	c.client.Transport = transport
}
