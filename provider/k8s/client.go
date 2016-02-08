package k8s

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/containous/traefik/safe"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"net/url"
	"strings"
)

const (
	// APIEndpoint defines the base path for kubernetes API resources.
	APIEndpoint        = "/api/v1"
	defaultService     = "/namespaces/default/services"
	extentionsEndpoint = "/apis/extensions/v1beta1"
	defaultIngress     = "/ingresses"
)

// Client is a client for the Kubernetes master.
type Client struct {
	endpointURL string
	tls         *tls.Config
	token       string
	caCert      []byte
}

// NewClient returns a new Kubernetes client.
// The provided host is an url (scheme://hostname[:port]) of a
// Kubernetes master without any path.
// The provided client is an authorized http.Client used to perform requests to the Kubernetes API master.
func NewClient(baseURL string, caCert []byte, token string) (*Client, error) {
	validURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %v", baseURL, err)
	}
	return &Client{
		endpointURL: strings.TrimSuffix(validURL.String(), "/"),
		token:       token,
		caCert:      caCert,
	}, nil
}

// GetIngresses returns all services in the cluster
func (c *Client) GetIngresses(predicate func(Ingress) bool) ([]Ingress, error) {
	getURL := c.endpointURL + extentionsEndpoint + defaultIngress
	request := gorequest.New().Get(getURL)
	if len(c.token) > 0 {
		request.Header["Authorization"] = "Bearer " + c.token
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(c.caCert)
		c.tls = &tls.Config{RootCAs: pool}
	}
	res, body, errs := request.TLSClientConfig(c.tls).EndBytes()
	if errs != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", getURL, errs)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d GET %q: %q", res.StatusCode, getURL, string(body))
	}

	var ingressList IngressList
	if err := json.Unmarshal(body, &ingressList); err != nil {
		return nil, fmt.Errorf("failed to decode list of ingress resources: %v", err)
	}
	ingresses := ingressList.Items[:0]
	for _, ingress := range ingressList.Items {
		if predicate(ingress) {
			ingresses = append(ingresses, ingress)
		}
	}
	return ingresses, nil
}

// WatchIngresses returns all services in the cluster
func (c *Client) WatchIngresses(predicate func(Ingress) bool, stopCh <-chan bool) (chan interface{}, chan error, error) {
	watchCh := make(chan interface{})
	errCh := make(chan error)

	getURL := c.endpointURL + extentionsEndpoint + defaultIngress + "?watch=true"

	// Make request to Kubernetes API
	request := gorequest.New().Get(getURL)
	if len(c.token) > 0 {
		request.Set("Authorization", "Bearer "+c.token)
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(c.caCert)
		c.tls = &tls.Config{RootCAs: pool}
	}
	req, err := request.TLSClientConfig(c.tls).MakeRequest()
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create request: GET %q : %v", getURL, err)
	}
	request.Client.Transport = request.Transport
	res, err := request.Client.Do(req)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to make request: GET %q: %v", getURL, err)
	}

	shouldStop := safe.New(false)

	go func() {
		select {
		case <-stopCh:
			shouldStop.Set(true)
			res.Body.Close()
			return
		}
	}()

	go func() {
		defer close(watchCh)
		defer close(errCh)
		for {
			var ingressList interface{}
			if err := json.NewDecoder(res.Body).Decode(&ingressList); err != nil {
				if !shouldStop.Get().(bool) {
					errCh <- fmt.Errorf("failed to decode list of ingress resources: %v", err)
				}
				return
			}

			watchCh <- ingressList
		}
	}()
	return watchCh, errCh, nil
}

// GetServices returns all services in the cluster
func (c *Client) GetServices(predicate func(Service) bool) ([]Service, error) {
	getURL := c.endpointURL + APIEndpoint + defaultService

	// Make request to Kubernetes API
	request := gorequest.New().Get(getURL)
	if len(c.token) > 0 {
		request.Header["Authorization"] = "Bearer " + c.token
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(c.caCert)
		c.tls = &tls.Config{RootCAs: pool}
	}
	res, body, errs := request.TLSClientConfig(c.tls).EndBytes()
	if errs != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", getURL, errs)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d GET %q: %q", res.StatusCode, getURL, string(body))
	}

	var serviceList ServiceList
	if err := json.Unmarshal(body, &serviceList); err != nil {
		return nil, fmt.Errorf("failed to decode list of services resources: %v", err)
	}
	services := serviceList.Items[:0]
	for _, service := range serviceList.Items {
		if predicate(service) {
			services = append(services, service)
		}
	}
	return services, nil
}
